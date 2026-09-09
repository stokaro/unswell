package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownEmptyDataRows(t *testing.T) {
	cases := []struct{ name, body string }{
		{"first", "| | |\n| Client | Output |\n"},
		{"middle", "| Input | Ready |\n| | |\n| Client | Output |\n"},
		{"last", "| Client | Output |\n| | |\n"},
		{"adjacent pipes", "|||\n| Client | Output |\n"},
		{"tabs", "|\t|\t|\n| Client | Output |\n"},
		{"repeated", "| | |\n|||\n|   |   |\n| Client | Output |\n"},
		{"short row", "||\n| Client | Output |\n"},
		{"wide row", "|||||\n| Client | Output |\n"},
		{"one cell", "Ordinary prose.\n| Client | Output |\n"},
		{"escaped pipe", "\\|\n| Client | Output |\n"},
		{"no final newline", "| Input | Ready |\n| | |\n| Client | Output |"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, ending := range []string{"\n", "\r\n"} {
				source := strings.ReplaceAll("| Name | Result |\n| --- | --- |\n"+tc.body, "\n", ending)
				doc := parseMarkdownRows(t, source, extract.Options{})
				for _, block := range doc.Blocks {
					qt.New(t).Assert(block.Kind, qt.Equals, "table-cell", qt.Commentf("%q", block.Text))
				}
				assertTableOutput(t, doc, source)
			}
		})
	}
}

func TestMarkdownEmptyRowsInContainers(t *testing.T) {
	table := "| Name | Result |\n| --- | --- |\n| Input | Ready |\n|||\n| | |\n| Client | Output |\n"
	cases := []struct{ name, before, prefix string }{
		{"quote", "", "> "},
		{"nested quote", "", "> > "},
		{"list", "- Item.\n\n", "  "},
		{"bom and front matter", "\ufeff---\nignored: value\n---\n\n", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			source := tc.before + tc.prefix + strings.ReplaceAll(strings.TrimSuffix(table, "\n"), "\n", "\n"+tc.prefix) + "\n"
			source = strings.ReplaceAll(source, "\n", "\r\n")
			doc := parseMarkdownRows(t, source, extract.Options{IncludeQuotes: true})
			assertTableOutput(t, doc, source)
		})
	}
}

func TestMarkdownEmptyRowsPreserveBlockBoundaries(t *testing.T) {
	cases := []struct{ name, boundary string }{
		{"blank", "\n"},
		{"heading", "# Heading\n"},
		{"thematic break", "***\n"},
		{"fenced code", "```\n||\n```\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			source := "| Name | Result |\n| --- | --- |\n|||\n| Client | Output |\n" + tc.boundary + "| After | Table |\n"
			doc := parseMarkdownRows(t, source, extract.Options{})
			assertTableOutput(t, doc, source)
			found := false
			for _, block := range doc.Blocks {
				if strings.Contains(block.Text, "After") {
					c.Assert(block.Kind, qt.Equals, "paragraph")
					c.Assert(block.Text, qt.Contains, "| After | Table |")
					found = true
				}
			}
			c.Assert(found, qt.IsTrue)
		})
	}
}

func TestMarkdownEmptyRowsPreserveProtectedSpans(t *testing.T) {
	cases := []struct {
		reason, protected string
		trailingBlank     int
	}{
		{"code", "```text\n|||\nHidden words.\n```\n", 0},
		{"code", "    |||\n    Hidden words.\n", 1},
		{"quote", "> |||\n> Hidden words.\n", 0},
		{"html", "<div>\n|||\nHidden words.\n</div>\n", 1},
	}
	for _, tc := range cases {
		t.Run(tc.reason, func(t *testing.T) {
			c := qt.New(t)
			prefix := "| Name | Result |\n| --- | --- |\n|||\n| Client | Output |\n\n"
			source := prefix + tc.protected + "\nVisible prose.\n"
			doc := parseMarkdownRows(t, source, extract.Options{})
			assertTableOutput(t, doc, source)
			c.Assert(doc.Excluded, qt.DeepEquals, []document.Exclusion{
				{Span: document.Span{Start: len(prefix), End: len(prefix) + len(tc.protected) + tc.trailingBlank}, Reason: tc.reason},
			})
			c.Assert(doc.Blocks[len(doc.Blocks)-1].Text, qt.Equals, "Visible prose.")
		})
	}
}

func TestMarkdownLonePipePreservesFollowingParagraph(t *testing.T) {
	c := qt.New(t)
	source := "| Name | Result |\n| --- | --- |\n| Client | Output |\n|\n| After | Table |\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "table.md", Format: document.Markdown,
		Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 5)
	c.Assert(doc.Blocks[4].Kind, qt.Equals, "paragraph")
	c.Assert(doc.Blocks[4].Text, qt.Equals, "| | After | Table |")
}

func TestMarkdownEmptyRowsPreserveContextSelection(t *testing.T) {
	c := qt.New(t)
	source := "Visible prose.\n\n| Name | Result |\n| --- | --- |\n| Input | Ready |\n|||\n| Client | Output |\n"
	doc := parseMarkdownRows(t, source, extract.Options{Policy: extract.Policy{Contexts: []string{"paragraph"}}})
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Visible prose.")
	doc = parseMarkdownRows(t, source, extract.Options{Policy: extract.Policy{Contexts: []string{"table-cell"}}})
	c.Assert(doc.Blocks, qt.HasLen, 6)
	assertTableOutput(t, doc, source)
}

func parseMarkdownRows(t *testing.T, source string, options extract.Options) *document.Document {
	t.Helper()
	c := qt.New(t)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "table.md", Format: document.Markdown,
		Bytes: []byte(source)}, options)
	c.Assert(err, qt.IsNil)
	c.Assert(string(doc.Source), qt.Equals, source)
	for _, excluded := range doc.Excluded {
		c.Assert(excluded.Span.Valid(len(source)), qt.IsTrue)
	}
	return &doc
}

func assertTableOutput(t *testing.T, doc *document.Document, source string) {
	t.Helper()
	c := qt.New(t)
	found := 0
	for _, block := range doc.Blocks {
		if block.Text == "Output" {
			found++
			start := strings.Index(source, "Output")
			c.Assert(block.Kind, qt.Equals, "table-cell")
			c.Assert(block.Spans(0, len(block.Text)), qt.DeepEquals,
				[]document.Span{{Start: start, End: start + len("Output")}})
		}
	}
	c.Assert(found, qt.Equals, 1)
}
