package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// A block terminator belongs to the comment that opened it. Removing it from a
// line comment that merely ends with the same bytes cut two bytes that were
// never a delimiter, and for a two-byte comment it produced an inverted span.
func TestCommentTerminatorsBelongToTheirOpener(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		source string
		text   string
	}{
		{"hash comment keeps a stray terminator", document.Python, "# Visible comment text. #>\n",
			" Visible comment text. #> "},
		{"shell comment keeps a stray terminator", document.Bash, "# Visible comment text. #>\n",
			" Visible comment text. #> "},
		{"line comment keeps a block terminator", document.Go, "package a\n// Visible comment text. */\n",
			" Visible comment text. */ "},
		{"block comment loses its terminator", document.C, "/* Visible comment text. */\n",
			" Visible comment text.  "},
		{"here-comment loses its terminator", document.PowerShell, "<# Visible comment text. #>\n",
			" Visible comment text.  "},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Equals, row.text)
			c.Assert(doc.Blocks[0].Map, qt.HasLen, len(doc.Blocks[0].Text))
		})
	}
}

// The two-byte comment that fuzzing found: the opener and the terminator share
// their bytes, so narrowing has to stop at an empty span instead of an inverted
// one.
func TestTwoByteCommentDoesNotInvertItsSpan(t *testing.T) {
	c := qt.New(t)
	doc, err := extract.Parse(t.Context(),
		document.Source{Name: "sample.py", Format: document.Python, Bytes: []byte("#>\n")}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "> ")
	doc, err = extract.Parse(t.Context(),
		document.Source{Name: "sample.c", Format: document.C, Bytes: []byte("/**/\n")}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 0)
}
