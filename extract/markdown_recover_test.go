package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func unparsedText(doc *document.Document) []string {
	var texts []string
	for _, exclusion := range doc.Excluded {
		if exclusion.Reason == extract.UnparsedReason {
			texts = append(texts, string(doc.Source[exclusion.Span.Start:exclusion.Span.End]))
		}
	}
	return texts
}

func recoveredText(doc *document.Document) string {
	texts := make([]string, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		texts = append(texts, block.Text)
	}
	return strings.Join(texts, "\n")
}

// One block that the inline grammar rejects is excluded with a reason. The
// paragraphs around it are still extracted with their original text.
func TestMarkdownExcludesABlockTheInlineGrammarRejects(t *testing.T) {
	c := qt.New(t)
	source := "The client retries after a failure.\n\n" +
		"| Target rule           | `<rule-name>`                             |\n" +
		"| Diagnostics           | `<total target-rule diagnostics>`         |\n\n" +
		"The server closes idle connections.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	unparsed := unparsedText(doc)
	c.Assert(unparsed, qt.HasLen, 1)
	c.Assert(unparsed[0], qt.Contains, "`<total target-rule diagnostics>`")
	text := recoveredText(doc)
	c.Assert(text, qt.Contains, "The client retries after a failure.")
	c.Assert(text, qt.Contains, "The server closes idle connections.")
	c.Assert(text, qt.Not(qt.Contains), "Target rule")
}

// When the block grammar rejects the document, the run of lines at the defect
// is excluded and the rest of the document keeps its blocks and coordinates.
func TestMarkdownExcludesARunTheBlockGrammarRejects(t *testing.T) {
	c := qt.New(t)
	source := "Intro paragraph one.\n\n" +
		"3. Access your environment from a selection of IDEs.\n\n" +
		"    <img width=\"864\" alt=\"image\" src=\"https://example.com/a.png\">\n\n" +
		"    <img width=\"1024\" alt=\"image\" src=\"https://example.com/b.png\">\n\n" +
		"Closing paragraph here.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	unparsed := unparsedText(doc)
	c.Assert(len(unparsed) > 0, qt.IsTrue)
	c.Assert(strings.Join(unparsed, ""), qt.Contains, "<img")
	text := recoveredText(doc)
	c.Assert(text, qt.Contains, "Intro paragraph one.")
	c.Assert(text, qt.Contains, "Closing paragraph here.")
	for _, block := range doc.Blocks {
		if block.Text == "Closing paragraph here." {
			c.Assert(string(doc.Source[block.Span.Start:block.Span.End]), qt.Equals, "Closing paragraph here.")
		}
	}
}

// A table row that the grammar builds outside a table is excluded with its run
// of lines; the prose after it is still read.
func TestMarkdownExcludesATableTheGrammarBuildsWrongly(t *testing.T) {
	c := qt.New(t)
	source := "Options follow.\n\n" +
		"| Short form | Long form | Explanation |\n" +
		"| :----------: | :---------: | :-----------: |\n" +
		"||--logging| Enable logging. Output will be saved to the log file.|\n\n" +
		"The client retries after a failure.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(strings.Join(unparsedText(doc), ""), qt.Contains, "--logging")
	c.Assert(recoveredText(doc), qt.Contains, "The client retries after a failure.")
}

// A document without a defect has no unparsed exclusion.
func TestMarkdownRecoveryLeavesValidDocumentsAlone(t *testing.T) {
	c := qt.New(t)
	doc := parseMarkdownRows(t, "The client retries.\n\n- One item.\n- Another item.\n", extract.Options{})
	c.Assert(unparsedText(doc), qt.HasLen, 0)
}
