package extract_test

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMarkdownNestedLists(t *testing.T) {
	for _, depth := range []int{4, 5, 8, 32} {
		for _, indentation := range []int{2, 4, 0} {
			for _, tail := range []string{"", "\nAfter the list.\n"} {
				name := fmt.Sprintf("depth=%d/indent=%d/tail=%t", depth, indentation, tail != "")
				t.Run(name, func(t *testing.T) {
					checkNestedMarkdown(t, depth, indentation, tail)
				})
			}
		}
	}
}

func checkNestedMarkdown(t *testing.T, depth, indentation int, tail string) {
	t.Helper()
	c := qt.New(t)
	body := nestedMarkdown(depth, indentation) + tail
	if tail == "" {
		body = strings.TrimSuffix(body, "\n")
	}
	text := "\ufeff" + strings.ReplaceAll(body, "\n", "\r\n")
	source := document.Source{Name: "nested.md", Format: document.Markdown, Bytes: []byte(text)}
	doc, err := extract.Parse(t.Context(), source, extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	c.Assert(string(source.Bytes), qt.Equals, text)
	c.Assert(string(doc.Source), qt.Equals, text)
	wantBlocks := depth
	if tail != "" {
		wantBlocks++
	}
	c.Assert(doc.Blocks, qt.HasLen, wantBlocks)
	for i, block := range doc.Blocks[:depth] {
		c.Assert(block.Kind, qt.Equals, "list-item")
		c.Assert(block.List, qt.IsNotNil)
		c.Assert(block.List.Depth, qt.Equals, i+1)
		c.Assert(block.List.Item, qt.Equals, 1)
		c.Assert(block.List.Items, qt.Equals, 1)
		label := fmt.Sprintf("Level %d", i+1)
		c.Assert(block.Text, qt.Contains, label+": The client checks café & scopes.")
		c.Assert(block.Text, qt.Not(qt.Contains), "protected phrase")
		start := strings.Index(text, label)
		client := start + strings.Index(text[start:], "client")
		analysis := strings.Index(block.Text, "client")
		c.Assert(block.Spans(analysis, analysis+len("client")), qt.DeepEquals,
			[]document.Span{{Start: client, End: client + len("client")}})
		entity := start + strings.Index(text[start:], "&amp;")
		analysis = strings.Index(block.Text, "&")
		c.Assert(block.Spans(analysis, analysis+1), qt.DeepEquals,
			[]document.Span{{Start: entity, End: entity + len("&amp;")}})
	}
	if tail != "" {
		last := doc.Blocks[depth]
		c.Assert(strings.TrimSpace(last.Text), qt.Equals, "After the list.")
		c.Assert(last.Kind, qt.Equals, "paragraph")
		c.Assert(last.List, qt.IsNil)
	}
	assertMappedDocument(t, doc, len(text))
}

func TestMarkdownNestedListLimits(t *testing.T) {
	for _, row := range []struct {
		name, source, want string
		options            extract.Options
	}{
		{"blocks", nestedMarkdown(5, 2), "source exceeds 3 prose blocks", extract.Options{MaxBlocks: 3}},
		{"nesting", nestedMarkdown(80, 2), "syntax nesting exceeds 128", extract.Options{}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := extract.Parse(t.Context(),
				document.Source{Name: "nested.md", Format: document.Markdown, Bytes: []byte(row.source)}, row.options)
			c.Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

func TestMarkdownDeepListPreservesCode(t *testing.T) {
	c := qt.New(t)
	const code = "                ```text\n                Hidden code words.\n                ```\n"
	text := nestedMarkdown(8, 2) + "\n" + code + "\n              * Last nested item.\n\nAfter the list.\n"
	doc := parseStructure(t, document.Markdown, text)
	c.Assert(doc.Blocks, qt.HasLen, 10)
	c.Assert(strings.TrimSpace(doc.Blocks[8].Text), qt.Equals, "Last nested item.")
	c.Assert(doc.Blocks[8].List.Depth, qt.Equals, 8)
	c.Assert(strings.TrimSpace(doc.Blocks[9].Text), qt.Equals, "After the list.")
	for _, block := range doc.Blocks {
		c.Assert(block.Text, qt.Not(qt.Contains), "Hidden code words.")
	}
	start := strings.Index(text, "```text")
	end := strings.Index(text, "```\n") + len("```\n")
	var codeRanges []document.Span
	for _, excluded := range doc.Excluded {
		if excluded.Reason == "code" {
			codeRanges = append(codeRanges, excluded.Span)
		}
	}
	c.Assert(codeRanges, qt.DeepEquals, []document.Span{{Start: start, End: end}})
}

func nestedMarkdown(depth, indentation int) string {
	var text strings.Builder
	spaces := 0
	for i := range depth {
		fmt.Fprintf(&text, "%s* Level %d: The **client** checks café &amp; scopes. `protected phrase`\n", strings.Repeat(" ", spaces), i+1)
		step := indentation
		if step == 0 {
			step = 2 + i%3
		}
		spaces += step
	}
	return text.String()
}
