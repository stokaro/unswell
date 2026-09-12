package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

const docComment = "/**\n" +
	" * Formats the given date according to the pattern.\n" +
	" * The second sentence of the description.\n" +
	" * @param {Date} date - the original date\n" +
	" * @returns {String} the formatted date string\n" +
	" * @throws {RangeError} the date must be valid\n" +
	" *\n" +
	" * @example\n" +
	" * const result = format(new Date(2014, 1, 11), 'MM/dd/yyyy')\n" +
	" * //=> '02/11/2014'\n" +
	" */\n"

// In a documentation comment the description before the first tag is prose.
// Tag lines carry types, names and short phrases; an example carries code.
// Both are excluded line by line with their own reasons and original spans.
func TestDocCommentTagsAndExamplesAreExcluded(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		code   string
	}{
		{"JavaScript", document.JavaScript, "function format(date, pattern) { return \"Keep the queue running.\"; }\n"},
		{"TypeScript", document.TypeScript, "function format(date: Date, pattern: string): string { return \"Keep the queue running.\"; }\n"},
		{"TSX", document.TSX, "function format(date: Date, pattern: string): string { return \"Keep the queue running.\"; }\n"},
		{"Java", document.Java, "class Sample { String format(String date) { return \"Keep the queue running.\"; } }\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source := docComment + row.code
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 2)
			c.Assert(doc.Blocks[0].Kind, qt.Equals, "comment")
			c.Assert(strings.Join(strings.Fields(doc.Blocks[0].Text), " "), qt.Equals,
				"Formats the given date according to the pattern. The second sentence of the description.")
			c.Assert(doc.Blocks[1].Text, qt.Equals, "Keep the queue running.")
			assertSourceMap(c, doc, source)
			var lines []string
			for _, excluded := range doc.Excluded {
				c.Assert(excluded.Span.Valid(len(source)), qt.IsTrue)
				lines = append(lines, excluded.Reason+": "+strings.TrimSpace(source[excluded.Span.Start:excluded.Span.End]))
			}
			c.Assert(lines, qt.DeepEquals, []string{
				"doc-tag: @param {Date} date - the original date",
				"doc-tag: @returns {String} the formatted date string",
				"doc-tag: @throws {RangeError} the date must be valid",
				"doc-example: @example",
				"doc-example: const result = format(new Date(2014, 1, 11), 'MM/dd/yyyy')",
				"doc-example: //=> '02/11/2014'",
			})
		})
	}
}

// Only a line that starts with "@name" inside a "/**" comment of a format with
// documentation tags opens a tag section. A mention elsewhere stays prose.
func TestAtSignsOutsideDocTagsStayProse(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
	}{
		{"line comment", document.JavaScript, "// @ops owns this queue and answers its pages.\n"},
		{"plain block comment", document.JavaScript, "/*\n * @ops owns this queue and answers its pages.\n */\n"},
		{"tag after prose on the line", document.JavaScript, "/** Mail @ops when the queue is stuck. */\n"},
		{"at sign without a name", document.TypeScript, "/**\n * @ ops owns this queue and answers its pages.\n */\n"},
		{"go block comment", document.Go, "package sample\n\n/**\n * @ops owns this queue and answers its pages.\n */\n"},
		{"python comment", document.Python, "# @ops owns this queue and answers its pages.\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Contains, "@")
			c.Assert(doc.Blocks[0].Text, qt.Contains, "ops")
			c.Assert(doc.Excluded, qt.HasLen, 0)
		})
	}
}

// A tag section runs to the next tag or the end of the comment. A description
// that follows a tag on later lines belongs to that tag, and a second comment
// starts over.
func TestDocTagSectionsEndWithTheComment(t *testing.T) {
	c := qt.New(t)
	source := "/**\n * Returns the queue length.\n * @returns {number} the number of\n * waiting jobs\n */\n" +
		"/** Keep the queue running. */\nconst length = 0;\n"
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "sample.js", Format: document.JavaScript, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(blockTexts(doc), qt.DeepEquals, []string{"Returns the queue length.", "Keep the queue running."})
	c.Assert(doc.Excluded, qt.HasLen, 2)
	for _, excluded := range doc.Excluded {
		c.Assert(excluded.Reason, qt.Equals, "doc-tag")
	}
	c.Assert(strings.TrimSpace(source[doc.Excluded[1].Span.Start:doc.Excluded[1].Span.End]), qt.Equals, "waiting jobs")
}
