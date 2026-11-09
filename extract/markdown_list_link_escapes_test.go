package extract_test

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownListLinksWithEscapedPunctuation(t *testing.T) {
	for _, test := range []struct {
		name, first, second, ending string
	}{
		{"unordered star", "- [Cursor\\*10](https://example.com/cursor)", "- [Clock](https://example.com/clock) has a description.", "\n"},
		{"ordered bracket", "1. [Café\\]10](https://example.com/cafe)", "2. [Next](https://example.com/next) has a description.", "\n"},
		{"nested underscore CRLF", "  - [Item\\_10](https://example.com/item)",
			"  - [Next](https://example.com/next) has a description.", "\r\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			source := test.first + test.ending + test.second + test.ending
			doc, err := extract.Parse(t.Context(), document.Source{Name: "links.md",
				Format: document.Markdown, Bytes: []byte(source)}, extract.Options{IncludeStructure: true})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Source, qt.DeepEquals, []byte(source))
			c.Assert(doc.Hash, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(source))))
			c.Assert(doc.Blocks, qt.HasLen, 2)
			for i, block := range doc.Blocks {
				c.Assert(block.Kind, qt.Equals, "list-item")
				c.Assert(block.List, qt.IsNotNil)
				c.Assert(block.List.Item, qt.Equals, i+1)
				c.Assert(block.List.Items, qt.Equals, 2)
				c.Assert(block.Span.Valid(len(source)), qt.IsTrue)
			}
			label := strings.ReplaceAll(strings.Split(test.first, "](")[0][strings.Index(test.first, "[")+1:], "\\", "")
			c.Assert(doc.Blocks[0].Text, qt.Contains, label)
			labelStart := strings.Index(doc.Blocks[0].Text, label)
			c.Assert(document.Bounds(doc.Blocks[0].Spans(labelStart, labelStart+len(label))), qt.DeepEquals,
				document.Span{Start: strings.Index(source, "[") + 1, End: strings.Index(source, "](")})
			c.Assert(doc.Blocks[1].Text, qt.Contains, "has a description")
		})
	}
}
