package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownEscapedLinkInAdjacentListItems(t *testing.T) {
	c := qt.New(t)
	source := "- [Cursor\\*10](https://example.com/Cursor-10.html)\n" +
		"- [avmplus](https://example.com/avmplus) AVM2 source code (flash player AVM1(?)AVM2 source code)\n"
	doc, err := extract.Parse(t.Context(), document.Source{
		Name: "README.md", Format: document.Markdown, Bytes: []byte(source),
	}, extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 2)
	c.Assert(doc.Excluded, qt.HasLen, 0)
	c.Assert(doc.Blocks[0].Kind, qt.Equals, "list-item")
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Cursor*10")
	c.Assert(doc.Blocks[0].Span, qt.DeepEquals, document.Span{Start: 3, End: 13})
	c.Assert(doc.Blocks[1].Kind, qt.Equals, "list-item")
	c.Assert(doc.Blocks[1].Text, qt.Equals, "avmplus AVM2 source code (flash player AVM1(?)AVM2 source code)")
	wordStart := strings.Index(source, "avmplus")
	c.Assert(document.Bounds(doc.Blocks[1].Spans(0, len("avmplus"))), qt.DeepEquals,
		document.Span{Start: wordStart, End: wordStart + len("avmplus")})
}
