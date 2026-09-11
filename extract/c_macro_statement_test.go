package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// A function-like macro used as a loop header has no semicolon, which the C
// and C++ grammars cannot know; the extractor recognizes exactly that shape
// and keeps every comment and string at its original offset.
func TestMacroStatementsExtractTheirProse(t *testing.T) {
	source := "/* Walks every pair. */\n" +
		"void walk(json_t *root) {\n" +
		"    json_object_foreach(root, key,\n        value) {\n" +
		"        /* Prints one pair per line. */\n" +
		"        printf(\"Key: %s\\n\", key);\n" +
		"    }\n" +
		"    list_for_each(pos, head) { count++; } // Counts the rest.\n" +
		"}\n"
	for _, format := range []document.Format{document.C, document.CPP} {
		t.Run(string(format), func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "walk.c", Format: format, Bytes: []byte(source)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			texts := []string{}
			for _, block := range doc.Blocks {
				texts = append(texts, block.Text)
				c.Assert(block.Map, qt.HasLen, len(block.Text))
				for _, span := range block.Map {
					c.Assert(span.Valid(len(source)), qt.IsTrue)
				}
			}
			c.Assert(texts, qt.DeepEquals, []string{" Walks every pair.  ", " Prints one pair per line.  ", "Key: %s\n",
				" Counts the rest. "})
		})
	}
}

func TestMacroStatementsWithProseOrOtherErrorsStayInvalid(t *testing.T) {
	for _, row := range []struct{ name, source string }{
		{"string inside the macro call", "void f(void) {\n    each(\"Item %d\", i) { g(i); }\n}\n"},
		{"comment inside the macro call", "void f(void) {\n    each(i, j /* every item */) { g(i); }\n}\n"},
		{"macro statement beside another error", "void f(void) {\n    each(i) { g(i); }\n    int = 3;\n}\n"},
		{"unbalanced body", "void f(void) {\n    each(i) { g(i);\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := extract.Parse(t.Context(), document.Source{Name: "f.c", Format: document.C, Bytes: []byte(row.source)},
				extract.Options{})
			c.Assert(err, qt.ErrorMatches, `parse c: c grammar returned an incomplete or invalid syntax tree`)
		})
	}
}
