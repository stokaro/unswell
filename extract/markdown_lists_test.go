package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestGrammarListMetadata(t *testing.T) {
	for _, row := range []struct {
		name, text             string
		items                  int
		ordered, task, complex bool
	}{
		{"flat", "- First label\n- Second label", 2, false, false, false},
		{"ordered", "1. First label\n2. Second label", 2, true, false, false},
		{"parenthesis ordered", "1) First label\n2) Second label", 2, true, false, false},
		{"tasks", "- [ ] First label\n- [x] Second label", 2, false, true, false},
		{"nested", "- First label\n  - Inner label\n- Second label", 2, false, false, true},
		{"multiple paragraphs", "- First label\n\n  Another paragraph.\n\n- Second label", 2, false, false, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc := parseStructure(t, document.Markdown, row.text)
			list := doc.Blocks[0].List
			c.Assert(list, qt.IsNotNil)
			c.Assert(*list, qt.Equals, document.ListContext{ID: 1, Item: 1, Items: row.items, Depth: 1,
				Ordered: row.ordered, Task: row.task, Complex: row.complex})
			source := document.Source{Name: "sample.md", Format: document.Markdown, Bytes: doc.Source}
			plain, err := extract.Parse(t.Context(), source, extract.Options{})
			c.Assert(err, qt.IsNil)
			for _, block := range plain.Blocks {
				c.Assert(block.List, qt.IsNil)
			}
		})
	}
}

func TestNestedListIdentity(t *testing.T) {
	c := qt.New(t)
	doc := parseStructure(t, document.Markdown, "- Outer one\n  - Inner one\n  - Inner two\n- Outer two\n\nText.\n\n- Separate list")
	want := []document.ListContext{
		{ID: 1, Item: 1, Items: 2, Depth: 1, Complex: true},
		{ID: 2, Item: 1, Items: 2, Depth: 2},
		{ID: 2, Item: 2, Items: 2, Depth: 2},
		{ID: 1, Item: 2, Items: 2, Depth: 1, Complex: true},
		{ID: 3, Item: 1, Items: 1, Depth: 1},
	}
	var got []document.ListContext
	for _, block := range doc.Blocks {
		if block.List != nil {
			got = append(got, *block.List)
		}
	}
	c.Assert(got, qt.DeepEquals, want)
}
