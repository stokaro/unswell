package extract_test

import (
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// markdownCase is one document with the blocks and exclusions it must give.
// Blocks are written "kind: text" with white space folded; an exclusion is
// written "reason: text" and matches when an exclusion with that reason holds
// the text.
type markdownCase struct {
	name     string
	source   string
	blocks   []string
	excluded []string
	unparsed []string
}

// checkMarkdownCase parses the case with and without structure and compares
// the whole block list, so a block that should not be there fails the case.
func checkMarkdownCase(t *testing.T, tc markdownCase) {
	t.Helper()
	for _, options := range []extract.Options{{}, {IncludeStructure: true}} {
		c := qt.New(t)
		doc := parseMarkdownRows(t, tc.source, options)
		got := []string{}
		for _, block := range doc.Blocks {
			got = append(got, block.Kind+": "+fields(block.Text))
		}
		c.Assert(got, qt.DeepEquals, tc.blocks)
		c.Assert(unparsedText(doc), qt.DeepEquals, tc.unparsed)
		for _, want := range tc.excluded {
			reason, text, _ := strings.Cut(want, ": ")
			held := strings.Join(excludedTexts(*doc, tc.source, reason), "\x00")
			c.Assert(held, qt.Contains, text, qt.Commentf("exclusion %q", want))
		}
		assertSourceMap(c, *doc, tc.source)
		assertVerbatimBlocks(c, doc, tc.source)
	}
}

// assertVerbatimBlocks checks the mapping of each block whose text stands in
// the source unchanged: its spans must cover exactly those source bytes.
func assertVerbatimBlocks(c *qt.C, doc *document.Document, source string) {
	c.Helper()
	for _, block := range doc.Blocks {
		if strings.Count(source, block.Text) != 1 {
			continue
		}
		start := strings.Index(source, block.Text)
		c.Assert(document.Bounds(block.Spans(0, len(block.Text))), qt.DeepEquals,
			document.Span{Start: start, End: start + len(block.Text)}, qt.Commentf("block %q", block.Text))
	}
}

// Each input is a block that the grammars rejected. It now parses whole: no
// error, no unparsed exclusion, HTML and code excluded, and the prose read.
func TestMarkdownRepairsRejectedInputs(t *testing.T) {
	for _, tc := range slices.Concat(
		rejectedHTMLCases(), rejectedCodeSpanCases(), rejectedTagCases(), rejectedTableAndListCases(),
	) {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
		})
	}
}

// rejectedHTMLCases holds an HTML block in a list item or a quote that a blank
// line ends while more of the container follows.
func rejectedHTMLCases() []markdownCase {
	return []markdownCase{
		{
			name: "fence and HTML in a list item",
			source: "1. Install FFmpeg, LaTeX in terminal using homebrew.\n    ```sh\n    ```\n" +
				"      <summary>💡 An alternative to heavyweight MacTeX bundle.</summary>\n\n   \n",
			blocks:   []string{"list-item: Install FFmpeg, LaTeX in terminal using homebrew."},
			excluded: []string{"html: <summary>💡 An alternative to heavyweight MacTeX bundle.</summary>", "code: ```sh"},
		},
		{
			name:     "HTML at a list marker before a nested list",
			source:   "1. <details><summary>Click if you have an ARM device</summary><p>\n\n    - If you have a ARM 32 bit v6 architecture\n",
			blocks:   []string{"list-item: If you have a ARM 32 bit v6 architecture"},
			excluded: []string{"html: <details><summary>Click if you have an ARM device</summary><p>"},
		},
		{
			name: "two images in a list item",
			source: "3. Access your environment from a selection of IDEs.\n\n" +
				"    <img width=\"864\" alt=\"image\" src=\"https://example.com/a.png\">\n\n" +
				"    <img width=\"1024\" alt=\"image\" src=\"https://example.com/b.png\">\n",
			blocks: []string{"list-item: Access your environment from a selection of IDEs."},
			excluded: []string{
				"html: <img width=\"864\" alt=\"image\" src=\"https://example.com/a.png\">",
				"html: <img width=\"1024\" alt=\"image\" src=\"https://example.com/b.png\">",
			},
		},
		{
			name: "closing tag in a list item",
			source: "*   Includes a built-in console (`--devconsole`) to test your scripts.\n    </p>\n\n" +
				"    [Watch a short demo](demo.mp4) of connect, send, receive.\n",
			blocks: []string{
				"list-item: Includes a built-in console ( \x00 ) to test your scripts.",
				"list-item: Watch a short demo of connect, send, receive.",
			},
			excluded: []string{"html: </p>", "inline-protected: `--devconsole`"},
		},
		{
			name:     "HTML block that holds the end of a processing instruction",
			source:   "- First item.\n\n  <div data-q=\"a?>\"\n  class=\"b\">\n\n  More of the item.\n",
			blocks:   []string{"list-item: First item.", "list-item: More of the item."},
			excluded: []string{"html: <div data-q=\"a?>\"\n  class=\"b\">"},
		},
		{
			name:     "HTML element whose name starts like a preformatted one",
			source:   "- First item.\n\n  <preview>\n\n  More of the item.\n",
			blocks:   []string{"list-item: First item.", "list-item: More of the item."},
			excluded: []string{"html: <preview>"},
		},
		{
			name:     "HTML in a quote",
			source:   "Before.\n\n> Quoted text.\n>\n> <div>x</div>\n>\n> More quoted text.\n\nAfter.\n",
			blocks:   []string{"paragraph: Before.", "paragraph: After."},
			excluded: []string{"quote: <div>x</div>"},
		},
	}
}

// rejectedCodeSpanCases holds a code span with angle brackets.
func rejectedCodeSpanCases() []markdownCase {
	return []markdownCase{
		{
			name: "code spans with angle brackets in rows without a delimiter row",
			source: "| Target rule           | `<rule-name>`                             |\n" +
				"| Diagnostics           | `<total target-rule diagnostics>`         |\n",
			blocks:   []string{"paragraph: | Target rule | \x00 | | Diagnostics | \x00 |"},
			excluded: []string{"inline-protected: `<rule-name>`", "inline-protected: `<total target-rule diagnostics>`"},
		},
		{
			name: "code spans with angle brackets in table cells",
			source: "| Field | Value |\n| --- | --- |\n" +
				"| Target rule           | `<rule-name>`                             |\n" +
				"| Diagnostics           | `<total target-rule diagnostics>`         |\n",
			blocks:   []string{"table-cell: Field", "table-cell: Value", "table-cell: Target rule", "table-cell: Diagnostics"},
			excluded: []string{"inline-protected: `<rule-name>`", "inline-protected: `<total target-rule diagnostics>`"},
		},
		{
			name: "code span with angle brackets in a list item",
			source: "3. **Blocked CDN.** If the host is unreachable, the automatic headless-shell\n" +
				"   download fails. Pass `--browser-executable=<path-to-local chrome-headless-shell>`.\n",
			blocks: []string{
				"list-item: Blocked CDN. If the host is unreachable, the automatic headless-shell download fails. Pass \x00 .",
			},
			excluded: []string{"inline-protected: `--browser-executable=<path-to-local chrome-headless-shell>`"},
		},
		{
			name:     "code span after an escaped backtick",
			source:   "Use \\` then `<total target-rule diagnostics>` here, <ins data-ad-format\n",
			blocks:   []string{"paragraph: Use ` then \x00 here, <ins data-ad-format"},
			excluded: []string{"inline-protected: `<total target-rule diagnostics>`"},
		},
		{
			name:     "code span with two backticks around single ones",
			source:   "Use ``<total `target-rule` diagnostics>`` and `<a b c>` here, <ins data-ad-format\n",
			blocks:   []string{"paragraph: Use \x00 and \x00 here, <ins data-ad-format"},
			excluded: []string{"inline-protected: ``<total `target-rule` diagnostics>``", "inline-protected: `<a b c>`"},
		},
		{
			// The first backtick is inside a tag, so it opens no code span.
			name:   "code span after a tag with a backtick in an attribute",
			source: "See <a title=\"`\" href=\"x\"> and <b>bold</b> then ` and ``<total target-rule diagnostics>`` here.\n",
			blocks: []string{"paragraph: See \x00 and \x00 bold \x00 then ` and \x00 here."},
			excluded: []string{
				"inline-protected: <a title=\"`\" href=\"x\">", "inline-protected: <b>", "inline-protected: </b>",
				"inline-protected: ``<total target-rule diagnostics>``",
			},
		},
	}
}

// rejectedTagCases holds an HTML tag that never closes. With a quoted
// attribute value it is HTML source and stays out of the prose. Without one
// the bracket is literal text, as CommonMark reads it.
func rejectedTagCases() []markdownCase {
	return []markdownCase{
		{
			name:     "unfinished tag after a fence",
			source:   "```python\n```\n<ins\n  data-ad-format=\"fluid\"\n",
			blocks:   []string{},
			excluded: []string{"code: ```python", "inline-protected: <ins\n  data-ad-format=\"fluid\""},
		},
		{
			// A line that starts with ">" is a quote, so the paragraph ends
			// before the bracket that would close the tag.
			name: "unfinished tag whose closing bracket starts the next line",
			source: "<ins\n  class=\"adsbygoogle\"\n  style=\"display:block\"\n  data-ad-format=\"fluid\"\n" +
				"  data-ad-slot=\"5678\"\n></ins>\n\nText after.\n",
			blocks: []string{"paragraph: Text after."},
			excluded: []string{
				"inline-protected: <ins\n  class=\"adsbygoogle\"\n  style=\"display:block\"\n  data-ad-format=\"fluid\"\n  data-ad-slot=\"5678\"",
				"quote: ></ins>",
			},
		},
		{
			name:   "unfinished tag with only an unquoted value",
			source: "<ins data-ad-format=fluid\n",
			blocks: []string{"paragraph: <ins data-ad-format=fluid"},
		},
		{
			name:     "unfinished tag with spaces around an equals sign",
			source:   "Text <ins data-ad-format='fluid' data-ad-slot = \"5\"\n",
			blocks:   []string{"paragraph: Text \x00"},
			excluded: []string{"inline-protected: <ins data-ad-format='fluid' data-ad-slot = \"5\""},
		},
		{
			name:     "unfinished tag in a list item",
			source:   "- Item text.\n  <ins\n  data-ad-format=\"fluid\"\n  data-ad-slot=\"5\"\n- Next item.\n",
			blocks:   []string{"list-item: Item text. \x00", "list-item: Next item."},
			excluded: []string{"inline-protected: <ins\n  data-ad-format=\"fluid\"\n  data-ad-slot=\"5\""},
		},
		{
			// Words after the last attribute value can be prose and stay.
			name:     "two unfinished tags with words between and after",
			source:   "First <ins data-ad-format=\"a\" and second <b data-x-y=\"c\" end\n",
			blocks:   []string{"paragraph: First \x00 and second \x00 end"},
			excluded: []string{"inline-protected: <ins data-ad-format=\"a\"", "inline-protected: <b data-x-y=\"c\""},
		},
		{
			// An attribute needs white space before it, so the tag ends at
			// the first value.
			name:     "unfinished tag before text that looks like an attribute",
			source:   "<ins data-ad-format=\"a\"b=\"c\" end\n",
			blocks:   []string{"paragraph: \x00 b=\"c\" end"},
			excluded: []string{"inline-protected: <ins data-ad-format=\"a\""},
		},
		{
			// A tag name starts with a letter, so the first bracket is text.
			name:     "bracket that starts no tag before an unfinished tag",
			source:   "a <- b data-ad-format=\"x\" and <ins data-ad-slot=\"1\"\n",
			blocks:   []string{"paragraph: a <- b data-ad-format=\"x\" and \x00"},
			excluded: []string{"inline-protected: <ins data-ad-slot=\"1\""},
		},
		{
			name:   "unfinished tags without an attribute value",
			source: "First <ins data-ad-format and second <b data-x-y\n",
			blocks: []string{"paragraph: First <ins data-ad-format and second <b data-x-y"},
		},
		{
			name:   "unfinished tag whose only value has no closing quote",
			source: "<ins data-ad-format=\"fluid\n",
			blocks: []string{"paragraph: <ins data-ad-format=\"fluid"},
		},
	}
}

// rejectedTableAndListCases holds a table row with a zero-width cell and a
// list item that starts with a link label holding a character reference.
func rejectedTableAndListCases() []markdownCase {
	return []markdownCase{
		{
			name: "empty first cell in a table row",
			source: "| Short form | Long form | Explanation |\n| :----------: | :---------: | :-----------: |\n" +
				"||--logging| Enable logging. Output will be saved to the log file.|\n",
			blocks: []string{
				"table-cell: Short form", "table-cell: Long form", "table-cell: Explanation",
				"table-cell: --logging", "table-cell: Enable logging. Output will be saved to the log file.",
			},
		},
		{
			name:   "empty cells in a header, inside a row, and at a row end",
			source: "| A || C |\n| --- | --- | --- |\n| one || three |\n| four | five ||\n|||six|\n",
			blocks: []string{
				"table-cell: A", "table-cell: C", "table-cell: one", "table-cell: three",
				"table-cell: four", "table-cell: five", "table-cell: six",
			},
		},
		{
			name: "entity in the link label of a nested list item",
			source: " * [About](#about)\n    * [Installation &amp; Usage](#installation--usage)\n" +
				" * [About this repository](#about-this-repository)\n",
			blocks: []string{"list-item: About", "list-item: Installation & Usage", "list-item: About this repository"},
		},
		{
			name:     "entity in the link label of a quoted list item",
			source:   "> * [Named &amp; more](#a)\n> * [Plain](#b)\n",
			blocks:   []string{},
			excluded: []string{"quote: > * [Named &amp; more](#a)"},
		},
		{
			name:     "entities in the link labels of a flat list",
			source:   "* [Named &amp; more](#a)\n* [Decimal &#38; more](#b)\n* [Hex &#x26; more][ref]\n\n[ref]: /url\n",
			blocks:   []string{"list-item: Named & more", "list-item: Decimal & more", "list-item: Hex & more"},
			excluded: []string{"link-definition: [ref]: /url"},
		},
	}
}

// With quotes included, the prose of a repaired quote is read.
func TestMarkdownRepairsInAQuote(t *testing.T) {
	for _, tc := range []struct{ name, source, html string }{
		{"HTML", "> Quoted text.\n>\n> <div>\n> Hidden words.\n> </div>\n>\n> More quoted text.\n", "<div>\n> Hidden words.\n> </div>"},
		{"entity in a link label", "> * [Quoted text.](#a) &amp; more\n> * [More &amp; text.](#b)\n>\n> More quoted text.\n", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			doc := parseMarkdownRows(t, tc.source, extract.Options{IncludeQuotes: true})
			c.Assert(unparsedText(doc), qt.HasLen, 0)
			texts := blockTexts(*doc)
			c.Assert(texts[0], qt.Contains, "Quoted text.")
			c.Assert(texts[len(texts)-1], qt.Equals, "More quoted text.")
			c.Assert(strings.Join(excludedTexts(*doc, tc.source, "html"), ""), qt.Contains, tc.html)
			assertSourceMap(c, *doc, tc.source)
			assertVerbatimBlocks(c, doc, tc.source)
		})
	}
}

// Input that looks like a repaired one and that the grammars accept gives the
// blocks and exclusions it gave before the repairs existed.
func TestMarkdownRepairsLeaveAcceptedInputAlone(t *testing.T) {
	for _, tc := range []markdownCase{
		{
			name:     "HTML that ends a list item",
			source:   "- First item.\n\n  <div>x</div>\n\nAfter the list.\n",
			blocks:   []string{"list-item: First item.", "paragraph: After the list."},
			excluded: []string{"html: <div>x</div>\n\n"},
		},
		{
			name:     "HTML comment in a list item",
			source:   "- First item.\n\n  <!-- note -->\n\n  More of the item.\n",
			blocks:   []string{"list-item: First item.", "list-item: More of the item."},
			excluded: []string{"html: <!-- note -->\n"},
		},
		{
			name:     "preformatted HTML in a list item",
			source:   "- First item.\n\n  <pre>\n  hidden words\n  </pre>\n\n  More of the item.\n",
			blocks:   []string{"list-item: First item.", "list-item: More of the item."},
			excluded: []string{"html: <pre>\n  hidden words\n  </pre>\n"},
		},
		{
			name:     "code span with one tag-like word",
			source:   "Pass `<a b>` here.\n",
			blocks:   []string{"paragraph: Pass \x00 here."},
			excluded: []string{"inline-protected: `<a b>`"},
		},
		{
			name:     "tags next to a code span",
			source:   "Use `x` and <b>bold</b> text.\n",
			blocks:   []string{"paragraph: Use \x00 and \x00 bold \x00 text."},
			excluded: []string{"inline-protected: `x`", "inline-protected: <b>", "inline-protected: </b>"},
		},
		{
			name:     "finished tag over two lines",
			source:   "```python\n```\n<ins\n  data-ad-format=\"fluid\">\n",
			blocks:   []string{},
			excluded: []string{"code: ```python", "inline-protected: <ins\n  data-ad-format=\"fluid\">"},
		},
		{
			name:   "table row with fewer cells",
			source: "| A | B | C |\n| --- | --- | --- |\n| one | two |\n",
			blocks: []string{"table-cell: A", "table-cell: B", "table-cell: C", "table-cell: one", "table-cell: two"},
		},
		{
			name:   "table row with more cells",
			source: "| A | B | C |\n| --- | --- | --- |\n| one | two | three | four |\n",
			blocks: []string{
				"table-cell: A", "table-cell: B", "table-cell: C",
				"table-cell: one", "table-cell: two", "table-cell: three", "table-cell: four",
			},
		},
		{
			name:   "blank first cell in a table row",
			source: "| Short form | Long form | Explanation |\n| --- | --- | --- |\n| |--logging| Enable logging.|\n",
			blocks: []string{
				"table-cell: Short form", "table-cell: Long form", "table-cell: Explanation",
				"table-cell: --logging", "table-cell: Enable logging.",
			},
		},
		{
			name:     "adjacent pipes outside a table",
			source:   "Run `a || b` first, or a || b.\n\n```sh\nmake || exit 1\n```\n",
			blocks:   []string{"paragraph: Run \x00 first, or a || b."},
			excluded: []string{"inline-protected: `a || b`", "code: ```sh\nmake || exit 1\n```\n"},
		},
		{
			name:   "entity after the start of a list item",
			source: "* See [Named &amp; more](#a)\n* [Plain](#b)\n",
			blocks: []string{"list-item: See Named & more", "list-item: Plain"},
		},
		{
			name: "one-space bullet with a four-space child",
			source: " * [About](#about)\n    * [Installation and Usage](#installation--usage)\n" +
				" * [About this repository](#about-this-repository)\n",
			blocks: []string{"list-item: About", "list-item: Installation and Usage", "list-item: About this repository"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
		})
	}
}

// Damage that no repair covers keeps its behavior: an unclosed fence is code
// to the end of the file, and a block that still fails is excluded as unparsed.
func TestMarkdownRepairsLeaveOtherDamageAsItWas(t *testing.T) {
	for _, tc := range []markdownCase{
		{
			name:     "unclosed fence",
			source:   "Text before.\n\n```sh\nmake build\n",
			blocks:   []string{"paragraph: Text before."},
			excluded: []string{"code: ```sh\nmake build\n"},
		},
		{
			name:     "HTML block too short to rewrite",
			source:   "- First item.\n\n  <p>\n\n  More of the item.\n\nAfter the list.\n",
			blocks:   []string{"list-item: First item.", "paragraph: After the list."},
			excluded: []string{"html: <p>"},
			unparsed: []string{"  More of the item.\n"},
		},
		{
			name:     "HTML block whose last line holds one byte",
			source:   "- First item.\n\n  <div\n  >\n\n  More of the item.\n",
			blocks:   []string{"list-item: First item."},
			excluded: []string{"html: <div\n  >"},
			unparsed: []string{"  More of the item.\n"},
		},
		{
			name:     "unfinished tag before a later closing bracket",
			source:   "Use `<total target-rule diagnostics>` and <ins data-ad-format 5 > b\n",
			blocks:   []string{},
			unparsed: []string{"Use `<total target-rule diagnostics>` and <ins data-ad-format 5 > b"},
		},
		{
			// The first backtick is inside the comment, so the pair that the
			// repair finds is no code span and the repair is refused.
			name:     "backticks that start inside a comment",
			source:   "See <!-- > ` --> and <b>bold</b> then ` and ``<total target-rule diagnostics>`` here.\n",
			blocks:   []string{},
			unparsed: []string{"See <!-- > ` --> and <b>bold</b> then ` and ``<total target-rule diagnostics>`` here."},
		},
		{
			name:     "table without an empty cell that the grammar builds wrongly",
			source:   "Text before.\n\n| A | B |\n| --- | --- |\n|x|\n|-\n",
			blocks:   []string{"paragraph: Text before."},
			unparsed: []string{"| A | B |\n| --- | --- |\n|x|\n|-\n"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
		})
	}
}

// A run that recovery blanked keeps its source span when the cell repair puts
// a marker byte before it. The marker goes between the pipes of the first
// paragraph, which is harmless there, and the broken table comes last.
func TestMarkdownCellRepairKeepsUnparsedSpans(t *testing.T) {
	c := qt.New(t)
	source := "Run a || b first.\n\n- First item.\n\n  <p>\n\n  More of the item.\n\n| A | B |\n| --- | --- |\n|| two |\n"
	doc := parseMarkdownRows(t, source, extract.Options{})
	c.Assert(unparsedText(doc), qt.DeepEquals, []string{"  More of the item.\n"})
	c.Assert(blockTexts(*doc), qt.DeepEquals, []string{"Run a || b first.", "First item.", "A", "B", "two"})
	assertVerbatimBlocks(c, doc, source)
}

// In MDX, indented prose is read with its indentation removed for the grammar.
// The cell repair keeps that mapping, so every span still points at the source.
func TestMDXCellRepairKeepsIndentedProse(t *testing.T) {
	c := qt.New(t)
	source := "<Tabs>\n\n    Indented prose here.\n\n</Tabs>\n\n| A | B |\n| --- | --- |\n|| two |\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: []byte(source)},
		extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(unparsedText(&doc), qt.HasLen, 0)
	c.Assert(blockTexts(doc), qt.DeepEquals, []string{"Indented prose here.", "A", "B", "two"})
	assertSourceMap(c, doc, source)
	assertVerbatimBlocks(c, &doc, source)
}
