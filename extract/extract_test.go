package extract_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestProseExtraction(t *testing.T) {
	cases := []struct {
		name   string
		format document.Format
		source string
		want   []string
	}{
		{
			"plain",
			document.Plain,
			"First line\r\ncontinued.\r\n\r\nAnother paragraph.",
			[]string{"First line \ncontinued.", "Another paragraph."},
		},
		{"emphasis", document.Markdown, "It is **important** to note that…", []string{"It is important to note that…"}},
		{"entities", document.Markdown, "Let&rsquo;s **dive** into &lt;details&gt;.", []string{"Let’s dive into <details>."}},
		{"code boundary", document.Markdown, "It is `important` to note that.", []string{"It is  \x00  to note that."}},
		{"link", document.Markdown, "Read the [manual](https://example.com/path).", []string{"Read the manual."}},
		{
			"exclude",
			document.Markdown,
			"---\ntitle: ignored\n---\n\n> Quoted text.\n\n```go\ncode()\n```\n\nVisible text.",
			[]string{"Visible text."},
		},
		{
			"table",
			document.Markdown,
			"| Name | Detail |\n| --- | --- |\n| Client | Opens connections. |",
			[]string{"Name", "Detail", "Client", "Opens connections."},
		},
		{
			"comments",
			document.Go,
			"// Package sample provides examples.\npackage sample\n//go:generate command\n// Client opens connections.\ntype Client struct{}\n",
			[]string{" Package sample provides examples. ", " Client opens connections. "},
		},
		{
			"cgo",
			document.Go,
			"package sample\n/* This belongs to C. */\nimport \"C\"\n// A normal comment.\n",
			[]string{" A normal comment. "},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(
				t.Context(),
				document.Source{Name: "sample", Format: tc.format, Bytes: []byte(tc.source)},
				extract.Options{},
			)
			c.Assert(err, qt.IsNil)
			got := make([]string, 0)
			for _, block := range doc.Blocks {
				got = append(got, block.Text)
				c.Assert(block.Map, qt.HasLen, len(block.Text))
				for _, span := range block.Map {
					c.Assert(span.Valid(len(tc.source)), qt.IsTrue)
				}
			}
			c.Assert(got, qt.DeepEquals, tc.want)
		})
	}
}

func TestSegmentMapping(t *testing.T) {
	c := qt.New(t)
	source := "😀 It is **important** to note that.\r\n"
	doc, err := extract.Parse(
		t.Context(),
		document.Source{Name: "a.md", Format: document.Markdown, Bytes: []byte(source)},
		extract.Options{},
	)
	c.Assert(err, qt.IsNil)
	block := doc.Blocks[0]
	start := strings.Index(block.Text, "important")
	spans := block.Spans(start, start+len("important"))
	c.Assert(spans, qt.HasLen, 1)
	c.Assert(source[spans[0].Start:spans[0].End], qt.Equals, "important")
	position, err := document.Locate([]byte(source), spans[0].Start)
	c.Assert(err, qt.IsNil)
	c.Assert(position, qt.Equals, document.Position{Line: 1, Column: 11})
	full := block.Spans(0, len(block.Text))
	c.Assert(len(full) > 1, qt.IsTrue)
}

func TestInputFailures(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := extract.Parse(ctx, document.Source{Name: "a.txt", Format: document.Plain}, extract.Options{})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = extract.Parse(t.Context(), document.Source{Name: "a.html", Format: "html", Bytes: []byte("prose")}, extract.Options{})
	c.Assert(err, qt.ErrorMatches, `unsupported input format "html"`)
}

func FuzzSourceMap(f *testing.F) {
	formats := document.Formats()
	for index, format := range formats {
		seed := "Read the manual."
		if format == document.Markdown {
			seed = "It is **important** to note that 😀 &amp; \\*."
		}
		f.Add(uint8(index), seed)
		if format == document.Markdown {
			f.Add(uint8(index), "| Name | Result |\n| --- | --- |\n|||\n| | |\n| 😀 | Output &amp; `code` |\n")
			f.Add(uint8(index), "<!-- unswell-disable-next-block rule.one -- Required contract wording. -->\n\nRead the manual.")
			f.Add(uint8(index), "| Name | Result |\n| --- | --- |\n|||\n\n"+
				"<!-- unswell-disable-next-block rule.one -- Required contract wording. -->\n\nRead the manual.")
		}
		if format == document.Go {
			f.Add(uint8(index), "package sample\n// unswell-disable-next-block rule.one -- Required contract wording.\n// Read the manual.\n")
			f.Add(uint8(index), "package sample\n// Read the manual.\nconst data = \"\\xff\"\nconst text = \"caf\\xc3\\xa9\"\n")
		}
	}
	f.Fuzz(func(t *testing.T, format uint8, text string) {
		if len(text) > 4096 {
			t.Skip()
		}
		doc, err := extract.Parse(
			t.Context(),
			document.Source{Name: "fuzz-input", Format: formats[int(format)%len(formats)], Bytes: []byte(text)},
			extract.Options{},
		)
		if err != nil {
			return
		}
		assertMappedDocument(t, doc, len(text))
	})
}

func assertMappedDocument(t *testing.T, doc document.Document, size int) {
	t.Helper()
	c := qt.New(t)
	for _, directive := range doc.Directives {
		c.Assert(directive.Span.Valid(size), qt.IsTrue)
	}
	for _, excluded := range doc.Excluded {
		c.Assert(excluded.Span.Valid(size), qt.IsTrue)
	}
	for _, block := range doc.Blocks {
		c.Assert(block.Text, qt.HasLen, len(block.Map))
		for _, span := range block.Map {
			c.Assert(span.Valid(size), qt.IsTrue)
		}
	}
}
