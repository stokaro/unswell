package extract_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestMDXComponents(t *testing.T) {
	for _, newline := range []string{"\n", "\r\n"} {
		t.Run(strings.ReplaceAll(newline, "\n", "LF"), func(t *testing.T) {
			c := qt.New(t)
			source := "\ufeff---\ntitle: Hidden front matter\n---\n\n" +
				"import { LinkCard, CardGrid } from '@astrojs/starlight/components';\n\n" +
				"export const links = {\n  title: 'Hidden export',\n\n  url: '/guide'\n};\n\n" +
				"# Guide\n\n<CardGrid>\n\nThe café client retries.\n\n" +
				"<LinkCard title=\"Hidden attribute\" href={links.url} />\n\n" +
				"<a href={`${base}schema.dot`} download>Download the schema.</a>\n\n" +
				"</CardGrid>\n\nHello {frontmatter.title}, try **again**.\n"
			source = strings.ReplaceAll(source, "\n", newline)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX,
				Bytes: []byte(source)},
				extract.Options{IncludeStructure: true})
			c.Assert(err, qt.IsNil)
			c.Assert(string(doc.Source), qt.Equals, source)
			text := mdxText(doc)
			for _, expected := range []string{"Guide", "The café client retries.", "Download the schema.", "Hello", "try again."} {
				c.Assert(text, qt.Contains, expected)
			}
			for _, absent := range []string{"Hidden", "import", "export", "frontmatter", "schema.dot", "CardGrid", "LinkCard"} {
				c.Assert(text, qt.Not(qt.Contains), absent)
			}
			assertMDXExclusion(t, doc, "<a href={`${base}schema.dot`} download>", "mdx-tag")
			assertMDXExclusion(t, doc, "{frontmatter.title}", "mdx-expression")
			assertMDXMapping(t, doc, source)
		})
	}
}

func TestMDXGrammarBoundaries(t *testing.T) {
	for name, region := range map[string]string{
		"object":        "{({value: {nested: '}'}}).value}",
		"template":      "{`value: ${JSON.stringify({value: '}'})}`}",
		"regex":         "{/}/.test('}') ? 'one' : 'two'}",
		"comment":       "{/* closing } and <fake> stay in the comment */}",
		"empty":         "{}",
		"function":      "{(() => {\n const x = {a: '}'};\n return <span>{x.a}</span>;\n})()}",
		"jsx attribute": "<Widget icon={<Icon title={'>'} />} options={{value: '}'}} />",
		"fragment":      "<><Widget /></>",
		"member":        "<UI.Card title=\"a > b\"></UI.Card>",
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			source := "Before " + region + " after."
			doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX,
				Bytes: []byte(source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(mdxText(doc), qt.Contains, "Before")
			c.Assert(mdxText(doc), qt.Contains, "after.")
			c.Assert(mdxText(doc), qt.Contains, "\x00")
			c.Assert(strings.ContainsAny(mdxText(doc), "{}<>"), qt.IsFalse)
		})
	}
}

func TestMDXMarkdownProtection(t *testing.T) {
	c := qt.New(t)
	source := "<Panel>\n\n```mdx\n{unclosed\n<Invalid\nimport broken\n```\n\n" +
		"The `a { < b` token. A \\{literal\\} and \\<literal.\n\n</Panel>\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX,
		Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(mdxText(doc), qt.Not(qt.Contains), "unclosed")
	c.Assert(mdxText(doc), qt.Not(qt.Contains), "a { < b")
	c.Assert(mdxText(doc), qt.Contains, "{literal}")
	c.Assert(mdxText(doc), qt.Contains, "<literal")
}

func TestMDXMalformed(t *testing.T) {
	for _, source := range []string{"Before {broken", "<Panel>", "</Panel>", "<A></B>",
		"<a href={broken>text</a>", "{let x = 1}", "import { broken\n\nThe client retries.",
		"Before " + strings.Repeat("{", 200) + "1"} {
		t.Run(source[:min(25, len(source))], func(t *testing.T) {
			_, err := extract.Parse(t.Context(), document.Source{Name: "broken.mdx", Format: document.MDX,
				Bytes: []byte(source)}, extract.Options{})
			c := qt.New(t)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func assertMDXExclusion(t *testing.T, doc document.Document, text, reason string) {
	t.Helper()
	c := qt.New(t)
	start := bytes.Index(doc.Source, []byte(text))
	c.Assert(start >= 0, qt.IsTrue)
	c.Assert(doc.Excluded, qt.Contains, document.Exclusion{
		Span: document.Span{Start: start, End: start + len(text)}, Reason: reason})
}

func mdxText(doc document.Document) string {
	var result strings.Builder
	for _, block := range doc.Blocks {
		result.WriteString(block.Text)
		result.WriteByte('\n')
	}
	return result.String()
}

func TestMDXIndentedProse(t *testing.T) {
	c := qt.New(t)
	source := "<main>\n  <article>\n    # Guide\n\n    The café client retries.\n\n" +
		"    - First item.\n      - Nested item.\n\n  </article>\n</main>\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: []byte(source)},
		extract.Options{IncludeStructure: true})
	c.Assert(err, qt.IsNil)
	c.Assert(mdxText(doc), qt.Contains, "The café client retries.")
	c.Assert(mdxText(doc), qt.Contains, "Nested item.")
	c.Assert(doc.Blocks[0].Kind, qt.Equals, "heading")
	c.Assert(doc.Blocks[1].Kind, qt.Equals, "paragraph")
	c.Assert(doc.Blocks[2].Kind, qt.Equals, "list-item")
	c.Assert(doc.Blocks[3].Kind, qt.Equals, "list-item")
	c.Assert(doc.Blocks[0].Text, qt.Equals, "Guide")
	c.Assert(source[doc.Blocks[1].Map[0].Start:doc.Blocks[1].Map[0].End], qt.Equals, "T")
}

func assertMDXMapping(t *testing.T, doc document.Document, source string) {
	t.Helper()
	c := qt.New(t)
	for _, block := range doc.Blocks {
		c.Assert(block.Map, qt.HasLen, len(block.Text))
		for offset, span := range block.Map {
			c.Assert(span.Start >= 0 && span.End <= len(source), qt.IsTrue)
			if block.Text[offset] != 0 && strings.ContainsAny(source[span.Start:span.End], "é") {
				c.Assert(source[span.Start:span.End], qt.Equals, "é")
			}
		}
	}
}

func TestMDXExportInParagraph(t *testing.T) {
	c := qt.New(t)
	source := "Collisions can exist in one\nexport and not in another."
	doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: []byte(source)}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(mdxText(doc), qt.Contains, "export and not in another.")
}

func TestMDXLimitsAndCancellation(t *testing.T) {
	for name, row := range map[string]struct{ source, message string }{
		"nesting":      {strings.Repeat("<Panel>", 129), "JSX nesting exceeds 128"},
		"parse budget": {"{'" + strings.Repeat("a", 12000) + strings.Repeat("}", 64) + "'}", "parse-input budget"},
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			_, err := extract.Parse(t.Context(), document.Source{Name: "limit.mdx", Format: document.MDX, Bytes: []byte(row.source)},
				extract.Options{})
			c.Assert(err, qt.IsNotNil)
			c.Assert(err.Error(), qt.Contains, row.message)
		})
	}
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := extract.Parse(ctx, document.Source{Name: "canceled.mdx", Format: document.MDX, Bytes: []byte("<Panel>Text.</Panel>")},
		extract.Options{})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestMDXDetection(t *testing.T) {
	for _, name := range []string{"guide.mdx", "nested/guide.MDX"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			format, ok := extract.Detect(name, nil)
			c.Assert(ok, qt.IsTrue)
			c.Assert(format, qt.Equals, document.MDX)
		})
	}
}

func TestMDXContainerIndentation(t *testing.T) {
	for name, source := range map[string]string{
		"quote": "> <Panel>\n>\n>     The client retries.\n>\n> </Panel>\n",
		"list":  "- <Panel>\n\n      The client retries.\n\n  </Panel>\n",
	} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: []byte(source)},
				extract.Options{IncludeQuotes: true, IncludeStructure: true})
			c.Assert(err, qt.IsNil)
			c.Assert(mdxText(doc), qt.Contains, "The client retries.")
		})
	}
}
