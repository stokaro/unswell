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

func TestMarkdownHeadingsAtEOF(t *testing.T) {
	cases := []struct{ name, source string }{
		{"simple", "# The client retries"},
		{"closing marker", "### The client retries ###"},
		{"sixth level", "###### The client retries"},
		{"unicode and protected code", "\ufeff## Café **retries** with `MaxAttempts`"},
		{"quote", "> ## The client retries"},
		{"list", "- Item.\n\n  ## The client retries"},
		{"front matter", "\ufeff---\r\nowner: client\r\n---\r\n\r\n## The client retries"},
		{"empty table rows", "| Name | Result |\n| --- | --- |\n|||\n| | |\n\n## The client retries"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			options := extract.Options{IncludeStructure: true, IncludeQuotes: true,
				Policy: extract.Policy{Contexts: []string{"heading"}}}
			reference := parseMarkdownEOF(t, test.source+"\n", options)
			c.Assert(reference.Blocks, qt.HasLen, 1)
			for _, ending := range []string{"", "\n", "\r\n"} {
				doc := parseMarkdownEOF(t, test.source+ending, options)
				c.Assert(doc.Blocks, qt.DeepEquals, reference.Blocks)
				block := doc.Blocks[0]
				c.Assert(block.Kind, qt.Equals, "heading")
				c.Assert(len(block.Context) > 0, qt.IsTrue)
				at := strings.Index(block.Text, "retries")
				start := strings.LastIndex(test.source, "retries")
				c.Assert(block.Spans(at, at+len("retries")), qt.DeepEquals,
					[]document.Span{{Start: start, End: start + len("retries")}})
			}
		})
	}
}

func TestMarkdownEOFProtectedSpans(t *testing.T) {
	for _, test := range []struct{ reason, source string }{
		{"code", "```text\nHidden words.\n```"},
		{"code", "    Hidden words."},
		{"html", "<div>\nHidden words.\n</div>"},
		{"quote", "> Hidden words."},
		{"link-definition", "[hidden]: https://example.com"},
	} {
		t.Run(test.reason, func(t *testing.T) {
			c := qt.New(t)
			doc := parseMarkdownEOF(t, test.source, extract.Options{})
			c.Assert(doc.Blocks, qt.HasLen, 0)
			c.Assert(doc.Excluded, qt.DeepEquals, []document.Exclusion{
				{Span: document.Span{Start: 0, End: len(test.source)}, Reason: test.reason},
			})
		})
	}
}

func TestLeadingMarkdownQuoteSelection(t *testing.T) {
	c := qt.New(t)
	quote := "> The client retries.\n"
	source := quote + "\nVisible prose."
	excluded := parseMarkdownEOF(t, source, extract.Options{})
	c.Assert(excluded.Blocks, qt.HasLen, 1)
	c.Assert(excluded.Blocks[0].Text, qt.Equals, "Visible prose.")
	c.Assert(excluded.Excluded, qt.DeepEquals, []document.Exclusion{
		{Span: document.Span{Start: 0, End: len(quote)}, Reason: "quote"},
	})
	included := parseMarkdownEOF(t, source, extract.Options{IncludeQuotes: true})
	c.Assert(included.Blocks, qt.HasLen, 2)
	c.Assert(included.Blocks[0].Text, qt.Equals, "The client retries.")
	c.Assert(included.Excluded, qt.HasLen, 0)
	c.Assert(included.Blocks[0].Spans(0, len(included.Blocks[0].Text)), qt.DeepEquals,
		[]document.Span{{Start: 2, End: len(quote) - 1}})
}

func TestMarkdownFramingDoesNotCountAsProse(t *testing.T) {
	for _, source := range []string{"", " \n\t", "# The client retries"} {
		t.Run(source, func(t *testing.T) {
			c := qt.New(t)
			doc := parseMarkdownEOF(t, source, extract.Options{MaxBlocks: 1, IncludeStructure: true})
			if strings.HasPrefix(source, "#") {
				c.Assert(doc.Blocks, qt.HasLen, 1)
				c.Assert(doc.Blocks[0].Text, qt.Equals, "The client retries")
				c.Assert(doc.Blocks[0].ID, qt.Equals, 0)
			} else {
				c.Assert(doc.Blocks, qt.HasLen, 0)
			}
		})
	}
}

func parseMarkdownEOF(t *testing.T, source string, options extract.Options) document.Document {
	t.Helper()
	c := qt.New(t)
	input := []byte(source)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "eof.md", Format: document.Markdown, Bytes: input}, options)
	c.Assert(err, qt.IsNil)
	c.Assert(string(input), qt.Equals, source)
	c.Assert(string(doc.Source), qt.Equals, source)
	c.Assert(doc.Hash, qt.Equals, fmt.Sprintf("%x", sha256.Sum256(input)))
	for _, block := range doc.Blocks {
		c.Assert(block.Span.Valid(len(input)), qt.IsTrue)
		for _, span := range block.Map {
			c.Assert(span.Valid(len(input)), qt.IsTrue)
		}
	}
	for _, exclusion := range doc.Excluded {
		c.Assert(exclusion.Span.Valid(len(input)), qt.IsTrue)
	}
	return doc
}
