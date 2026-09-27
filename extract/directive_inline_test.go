package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestInlineDirectiveTrialKeepsExamplesAndLiteralMarkup(t *testing.T) {
	const comment = "<!-- unswell-disable-next-sentence policy.banned-phrases -- Required wording. -->"
	for _, example := range []string{
		"`" + comment + "`",
		"\\" + comment,
		"<span title='" + comment + "'>a label</span>",
		"<span title=\"" + comment + "\">a label</span>",
		"<span title=\"" + comment + " an unfinished attribute",
		"![" + comment + "](image.png)",
	} {
		t.Run(example, func(t *testing.T) {
			c := qt.New(t)
			text := "\ufeffExamples: " + example + ".\r\n\r\nA clean sentence. " + comment + " The robust client starts.\r\n"
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
			doc, err := extract.Parse(t.Context(), source, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Directives, qt.HasLen, 1)
			start := strings.LastIndex(text, comment)
			c.Assert(doc.Directives[0].Span, qt.Equals, document.Span{Start: start, End: start + len(comment)})
			c.Assert(string(doc.Source), qt.Equals, text)
			assertMappedDocument(t, doc, len(text))
		})
	}
}

func TestInlineNonDirectiveDoubleHyphensRemainLiteral(t *testing.T) {
	c := qt.New(t)
	text := "A sentence. <!-- ordinary -- comment --> Another sentence."
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	doc, err := extract.Parse(t.Context(), source, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Directives, qt.HasLen, 0)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Text, qt.Equals, text)
	assertMappedDocument(t, doc, len(text))
}

func TestInlineDirectiveCandidateLimit(t *testing.T) {
	c := qt.New(t)
	text := strings.Repeat("A sentence. <!-- unswell-disable-next-sentence policy.banned-phrases -- Required wording. --> ", 1001)
	_, err := extract.Parse(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}, extract.Options{})
	c.Assert(err, qt.ErrorMatches, ".*inline directive candidates exceed 1000.*")
}
