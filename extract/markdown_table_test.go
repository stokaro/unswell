package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownEmptyTableCells(t *testing.T) {
	cases := []struct {
		name, source string
		want         []string
	}{
		{"header", "| | Result |\n| --- | --- |\n| Input | Output |\n", []string{"Result", "Input", "Output"}},
		{"body", "| Name | Result |\n| --- | --- |\n| Input | |\n", []string{"Name", "Result", "Input"}},
		{"whitespace", "| \t | Result |\n| --- | --- |\n| Input | \t |\n", []string{"Result", "Input"}},
		{"all empty", "| | |\n| --- | --- |\n| | |\n", []string{}},
		{"inline code", "| | Result |\n| --- | --- |\n| `Certainly!` | Output |\n", []string{"Result", "Output"}},
		{"unicode", "| | Résumé |\n| --- | --- |\n| 😀 | |\n", []string{"Résumé", "😀"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{
				Name: "table.md", Format: document.Markdown, Bytes: []byte(tc.source),
			}, extract.Options{})
			c.Assert(err, qt.IsNil)
			got := []string{}
			for index, block := range doc.Blocks {
				got = append(got, block.Text)
				c.Assert(block.ID, qt.Equals, index)
				c.Assert(block.Kind, qt.Equals, "table-cell")
				start := strings.Index(tc.source, block.Text)
				c.Assert(start >= 0, qt.IsTrue)
				c.Assert(block.Spans(0, len(block.Text)), qt.DeepEquals,
					[]document.Span{{Start: start, End: start + len(block.Text)}})
			}
			c.Assert(got, qt.DeepEquals, tc.want)
		})
	}
}

func TestEmptyTableCellsPreserveContextSelection(t *testing.T) {
	source := "Read the manual.\n\n| | Result |\n| --- | --- |\n| Input | |\n"
	cases := []struct {
		context string
		want    []string
	}{
		{"paragraph", []string{"Read the manual."}},
		{"table-cell", []string{"Result", "Input"}},
	}
	for _, tc := range cases {
		t.Run(tc.context, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{
				Name: "table.md", Format: document.Markdown, Bytes: []byte(source),
			}, extract.Options{Policy: extract.Policy{Contexts: []string{tc.context}}})
			c.Assert(err, qt.IsNil)
			got := []string{}
			for _, block := range doc.Blocks {
				got = append(got, block.Text)
				c.Assert(block.Kind, qt.Equals, tc.context)
			}
			c.Assert(got, qt.DeepEquals, tc.want)
			for _, excluded := range doc.Excluded {
				c.Assert(excluded.Span.Valid(len(source)), qt.IsTrue)
			}
		})
	}
}
