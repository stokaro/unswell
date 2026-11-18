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

// One block that the inline grammar rejects is excluded with a reason. The
// paragraphs around it are still extracted with their original text. The tag
// is unfinished, and the later ">" keeps the bracket repair away from it.
func TestMarkdownExcludesABlockTheInlineGrammarRejects(t *testing.T) {
	c := qt.New(t)
	source := "The client retries after a failure.\n\n" +
		"<ins data-ad-format 5 > b\n\n" +
		"The server closes idle connections.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(unparsedText(doc), qt.DeepEquals, []string{"<ins data-ad-format 5 > b"})
	c.Assert(blockTexts(*doc), qt.DeepEquals, []string{"The client retries after a failure.", "The server closes idle connections."})
}

// When the block grammar rejects the document, the run of lines at the defect
// is excluded and the rest of the document keeps its blocks and coordinates.
// A lone "<p>" is too short for the HTML repair.
func TestMarkdownExcludesARunTheBlockGrammarRejects(t *testing.T) {
	c := qt.New(t)
	source := "Intro paragraph one.\n\n" +
		"3. Access your environment from a selection of IDEs.\n\n" +
		"    <p>\n\n" +
		"    More text in the item.\n\n" +
		"Closing paragraph here.\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(strings.Join(unparsedText(doc), ""), qt.Contains, "More text in the item.")
	c.Assert(excludedTexts(*doc, source, "html"), qt.HasLen, 1)
	c.Assert(blockTexts(*doc), qt.DeepEquals, []string{
		"Intro paragraph one.", "Access your environment from a selection of IDEs.", "Closing paragraph here.",
	})
	for _, block := range doc.Blocks {
		c.Assert(string(doc.Source[block.Span.Start:block.Span.End]), qt.Equals, block.Text)
	}
}

// A table row that the grammar builds outside a table is excluded with its run
// of lines; the prose before it is still read. The rows have no empty cell, so
// the cell repair does not apply.
func TestMarkdownExcludesATableTheGrammarBuildsWrongly(t *testing.T) {
	c := qt.New(t)
	source := "The client retries after a failure.\n\n" +
		"| Short form | Long form |\n" +
		"| --- | --- |\n" +
		"|x|\n" +
		"|-\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(unparsedText(doc), qt.DeepEquals, []string{"| Short form | Long form |\n| --- | --- |\n|x|\n|-\n"})
	c.Assert(blockTexts(*doc), qt.DeepEquals, []string{"The client retries after a failure."})
}

// A document without a defect has no unparsed exclusion.
func TestMarkdownRecoveryLeavesValidDocumentsAlone(t *testing.T) {
	c := qt.New(t)
	doc := parseMarkdownRows(t, "The client retries.\n\n- One item.\n- Another item.\n", extract.Options{})
	c.Assert(unparsedText(doc), qt.HasLen, 0)
}
