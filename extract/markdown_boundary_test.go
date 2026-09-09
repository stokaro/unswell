package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

const boundaryTable = "| Name | Result |\n| --- | --- |\n| Client | Output |\n"
const boundaryProse = "| After | It is **important** to note that café requests may fail. |\n"

func TestMarkdownLonePipeContainersAndCoordinates(t *testing.T) {
	for _, test := range []struct{ name, before, prefix, kind string }{
		{"plain", "", "", "paragraph"},
		{"quote", "", "> ", "paragraph"},
		{"nested quote", "", "> > ", "paragraph"},
		{"list", "- Item.\n\n", "  ", "list-item"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, boundary := range []string{"|\n", " | \n", "\t|\t\n"} {
				for _, ending := range []string{"\n", "\r\n"} {
					source := test.before + test.prefix + strings.ReplaceAll(
						strings.TrimSuffix(boundaryTable+boundary+boundaryProse, "\n"), "\n", "\n"+test.prefix) + "\n"
					source = strings.ReplaceAll(source, "\n", ending)
					doc := parseMarkdownRows(t, source, extract.Options{IncludeQuotes: true, IncludeStructure: true})
					assertTableOutput(t, doc, source)
					assertBoundaryProse(t, doc, source, test.kind)
				}
			}
		})
	}
}

func assertBoundaryProse(t *testing.T, doc *document.Document, source, kind string) {
	t.Helper()
	c := qt.New(t)
	tables, paragraphs := 0, 0
	for _, block := range doc.Blocks {
		if block.Kind == "table-cell" {
			tables++
		}
		if !strings.Contains(block.Text, "After") {
			continue
		}
		paragraphs++
		c.Assert(block.Kind, qt.Equals, kind)
		c.Assert(block.Text, qt.Contains, "It is important to note that café requests may fail.")
		start := strings.Index(block.Text, "café")
		original := strings.Index(source, "café")
		c.Assert(document.Bounds(block.Spans(start, start+len("café"))), qt.DeepEquals,
			document.Span{Start: original, End: original + len("café")})
	}
	c.Assert(tables, qt.Equals, 4)
	c.Assert(paragraphs, qt.Equals, 1)
}

func TestMarkdownLonePipePreservesSelection(t *testing.T) {
	c := qt.New(t)
	source := boundaryTable + "|\n" + boundaryProse
	doc := parseMarkdownRows(t, source, extract.Options{Policy: extract.Policy{Contexts: []string{"table-cell"}}})
	c.Assert(doc.Blocks, qt.HasLen, 4)
	doc = parseMarkdownRows(t, source, extract.Options{Policy: extract.Policy{Contexts: []string{"paragraph"}}})
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Contains, "After")
	quoted := "> " + strings.ReplaceAll(strings.TrimSuffix(source, "\n"), "\n", "\n> ") + "\n"
	doc = parseMarkdownRows(t, quoted, extract.Options{})
	c.Assert(doc.Blocks, qt.HasLen, 0)
	c.Assert(doc.Excluded, qt.HasLen, 1)
	c.Assert(doc.Excluded[0].Reason, qt.Equals, "quote")
	c.Assert(doc.Excluded[0].Span, qt.DeepEquals, document.Span{Start: 0, End: len(quoted)})
}

func TestMarkdownLonePipePreservesProtectedCodeAndBlankLines(t *testing.T) {
	c := qt.New(t)
	source := boundaryTable + "|\nIt is `important` to note that café requests may fail.\n\n" +
		"```text\n|\nHidden prose.\n```\n\nVisible prose.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(doc.Blocks, qt.HasLen, 6)
	c.Assert(doc.Blocks[4].Kind, qt.Equals, "paragraph")
	c.Assert(doc.Blocks[4].Text, qt.Contains, "It is  \x00  to note that café requests may fail.")
	c.Assert(doc.Blocks[5].Text, qt.Equals, "Visible prose.")
	var protected []string
	for _, excluded := range doc.Excluded {
		if excluded.Reason == "code" || excluded.Reason == "inline-protected" {
			protected = append(protected, source[excluded.Span.Start:excluded.Span.End])
		}
	}
	c.Assert(protected, qt.DeepEquals, []string{"`important`", "```text\n|\nHidden prose.\n```\n"})
}

func TestMarkdownLonePipeAtEOF(t *testing.T) {
	c := qt.New(t)
	doc := parseMarkdownRows(t, boundaryTable+"|", extract.Options{})
	c.Assert(doc.Blocks, qt.HasLen, 5)
	c.Assert(doc.Blocks[4].Text, qt.Equals, "|")
	c.Assert(doc.Blocks[4].Kind, qt.Equals, "paragraph")
}

func TestMarkdownBoundaryAdapterPreservesSharedGrammar(t *testing.T) {
	c := qt.New(t)
	source := []byte(boundaryTable + "|\n" + boundaryProse)
	lang := grammars.MarkdownLanguage()
	before, err := ts.NewParser(lang).ParseStrict(source)
	c.Assert(err, qt.IsNil)
	defer before.Release()
	doc, err := extract.Parse(t.Context(), document.Source{Name: "table.md", Format: document.Markdown, Bytes: source}, extract.Options{})
	c.Assert(err, qt.IsNil)
	assertBoundaryProse(t, &doc, string(source), "paragraph")
	after, err := ts.NewParser(lang).ParseStrict(source)
	c.Assert(err, qt.IsNil)
	defer after.Release()
	c.Assert(after.RootNode().SExpr(lang), qt.Equals, before.RootNode().SExpr(lang))
}
