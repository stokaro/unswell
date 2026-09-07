package extract_test

import (
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestDirectivesUseSourceCommentGrammars(t *testing.T) {
	formats := []document.Format{document.Go, document.JavaScript, document.TypeScript, document.TSX, document.Python,
		document.Rust, document.Java, document.C, document.CPP, document.Bash, document.Shell, document.Zsh, document.Fish,
		document.PowerShell, document.CSharp, document.YAML}
	for _, format := range formats {
		t.Run(string(format), func(t *testing.T) {
			c := qt.New(t)
			prefix := "//"
			if slices.Contains([]document.Format{document.Python, document.Bash, document.Shell, document.Zsh,
				document.Fish, document.PowerShell, document.YAML}, format) {
				prefix = "#"
			}
			text := prefix + " unswell-disable-next-block policy.banned-phrases -- Required contract wording.\n" +
				prefix + " The robust client starts.\n"
			if format == document.Go {
				text = "package sample\n" + text
			}
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: format, Bytes: []byte(text)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Directives, qt.HasLen, 1)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			c.Assert(doc.Blocks[0].Text, qt.Not(qt.Contains), "unswell-")
			c.Assert(doc.Blocks[0].Text, qt.Contains, "The robust client starts.")
			c.Assert(string(doc.Source), qt.Equals, text)
			span := doc.Directives[0].Span
			c.Assert(text[span.Start:span.End], qt.Contains, prefix+" unswell-disable-next-block")
		})
	}
}

func TestDirectiveExamplesAndStringsRemainData(t *testing.T) {
	command := "unswell-disable-next-block policy.banned-phrases -- Required contract wording."
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("```html\n<!-- " + command + " -->\n```\n\nThe client starts.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("> <!-- " + command + " -->\n\nThe client starts.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The example is `<!-- " + command + " -->`.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("<script>\nconst example = '<!-- " + command +
			" -->';\n</script>\n\nThe client starts.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("<div title='<!-- " + command + " -->'>Example</div>\n\nThe client starts.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("<!-- Ordinary comment. --><script>const example = '<!-- " + command +
			" -->';</script>\n\nThe client starts.")},
		{Name: "sample.go", Format: document.Go, Bytes: []byte("package sample\nvar example = \"// " + command + "\"\n")},
		{Name: "sample.go", Format: document.Go, Bytes: []byte("package sample\n//     " + command + "\n// The client starts.\n")},
		{Name: "settings.yaml", Format: document.YAML, Bytes: []byte("example: '# " + command + "'\n")},
		{Name: "guide.txt", Format: document.Plain, Bytes: []byte(command)},
	} {
		c := qt.New(t)
		doc, err := extract.Parse(t.Context(), source, extract.Options{})
		c.Assert(err, qt.IsNil, qt.Commentf("%s", source.Bytes))
		c.Assert(doc.Directives, qt.HasLen, 0, qt.Commentf("%s", source.Bytes))
	}
}

func TestDirectiveNormalizationRejectsUnrelatedSyntaxErrors(t *testing.T) {
	command := "unswell-disable-next-block policy.banned-phrases -- Required contract wording."
	for _, source := range []document.Source{
		{Name: "sample.ps1", Format: document.PowerShell, Bytes: []byte("# " + command + "\n$value = (\n")},
		{Name: "sample.ps1", Format: document.PowerShell, Bytes: []byte("# " + command + "\n<# An unfinished comment.\n")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("A sentence. <!-- ordinary -- comment --> Another sentence.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("A sentence. <!-- " + command +
			" --> The client starts. <!-- ordinary -- comment --> Another sentence.")},
	} {
		c := qt.New(t)
		_, err := extract.Parse(t.Context(), source, extract.Options{})
		c.Assert(err, qt.IsNotNil, qt.Commentf("%s", source.Bytes))
	}
}

func TestPowerShellCommentOnlySourceMapping(t *testing.T) {
	c := qt.New(t)
	text := "\ufeff# unswell-disable-next-block policy.banned-phrases -- Required contract wording.\r\n# The client starts.\r\n"
	doc, err := extract.Parse(t.Context(), document.Source{
		Name: "sample.ps1", Format: document.PowerShell, Bytes: []byte(text),
	}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(string(doc.Source), qt.Equals, text)
	c.Assert(doc.Directives, qt.HasLen, 1)
	c.Assert(doc.Directives[0].Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(doc.Blocks, qt.HasLen, 1)
	for _, span := range doc.Blocks[0].Map {
		c.Assert(span.Valid(len(text)), qt.IsTrue)
	}
}

func TestDirectiveRangesSurviveInlineMarkdownAndCRLF(t *testing.T) {
	c := qt.New(t)
	comment := "<!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. -->"
	text := "\ufeffA clean sentence. " + comment + " The **robust** client starts.\r\n"
	doc, err := extract.Parse(t.Context(), document.Source{
		Name: "guide.md", Format: document.Markdown, Bytes: []byte(text),
	}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Directives, qt.HasLen, 1)
	start := strings.Index(text, comment)
	c.Assert(doc.Directives[0].Span, qt.Equals, document.Span{Start: start, End: start + len(comment)})
	for _, block := range doc.Blocks {
		c.Assert(block.Text, qt.Not(qt.Contains), "Required contract wording")
	}
}

func TestMarkdownDirectiveAfterLeadingBOM(t *testing.T) {
	c := qt.New(t)
	comment := "<!-- unswell-disable-next-block policy.banned-phrases -- Required contract wording. -->"
	text := "\ufeff" + comment + "\r\n\r\nThe client starts.\r\n"
	doc, err := extract.Parse(t.Context(), document.Source{
		Name: "guide.md", Format: document.Markdown, Bytes: []byte(text),
	}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Directives, qt.HasLen, 1)
	c.Assert(doc.Directives[0].Span, qt.Equals, document.Span{Start: len("\ufeff"), End: len("\ufeff") + len(comment)})
	c.Assert(string(doc.Source), qt.Equals, text)
}

func TestMarkdownDirectiveRangesAfterEmptyTableRows(t *testing.T) {
	for _, placement := range []string{"block", "inline"} {
		t.Run(placement, func(t *testing.T) {
			c := qt.New(t)
			comment := "<!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. -->"
			text := "\ufeff| Name | Result |\r\n| --- | --- |\r\n|||\r\n| | |\r\n| 😀 | Output &amp; `code` |\r\n\r\n"
			if placement == "inline" {
				text += "A clean sentence. " + comment + " The **robust** client starts.\r\n"
			} else {
				text += comment + "\r\n\r\nThe **robust** client starts.\r\n"
			}
			doc, err := extract.Parse(t.Context(), document.Source{
				Name: "guide.md", Format: document.Markdown, Bytes: []byte(text),
			}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(string(doc.Source), qt.Equals, text)
			c.Assert(doc.Directives, qt.HasLen, 1)
			start := strings.Index(text, comment)
			c.Assert(doc.Directives[0].Span, qt.Equals, document.Span{Start: start, End: start + len(comment)})
			assertMappedDocument(t, doc, len(text))
		})
	}
}
