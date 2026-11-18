package extract_test

import (
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// exclusionList writes every exclusion of a document as "reason: text".
func exclusionList(doc *document.Document) []string {
	list := []string{}
	for _, excluded := range doc.Excluded {
		list = append(list, excluded.Reason+": "+string(doc.Source[excluded.Span.Start:excluded.Span.End]))
	}
	return list
}

// The inline grammar accepts each block but drops a code span that holds a
// "<": no error, the words of the span in the prose, and sometimes a span made
// of the text between two real spans. Each span is now protected.
func TestMarkdownProtectsCodeSpansWithAngleBrackets(t *testing.T) {
	for _, tc := range []markdownCase{
		{
			name:     "span with words in angle brackets",
			source:   "Use `<total target-rule diagnostics>` here.\n",
			blocks:   []string{"paragraph: Use \x00 here."},
			excluded: []string{"inline-protected: `<total target-rule diagnostics>`"},
		},
		{
			name:     "span that ends with words in angle brackets",
			source:   "Pass `--browser-executable=<path-to-local chrome-headless-shell>`.\n",
			blocks:   []string{"paragraph: Pass \x00 ."},
			excluded: []string{"inline-protected: `--browser-executable=<path-to-local chrome-headless-shell>`"},
		},
		{
			name:     "two spans whose inner backticks the grammar paired",
			source:   "Use `<total target-rule diagnostics>` and `<other rule name>` here.\n",
			blocks:   []string{"paragraph: Use \x00 and \x00 here."},
			excluded: []string{"inline-protected: `<total target-rule diagnostics>`", "inline-protected: `<other rule name>`"},
		},
		{
			name:   "span between spans that the grammar reads",
			source: "Pass `<a b>` and `<total target-rule diagnostics>` and `plain` here.\n",
			blocks: []string{"paragraph: Pass \x00 and \x00 and \x00 here."},
			excluded: []string{
				"inline-protected: `<a b>`", "inline-protected: `<total target-rule diagnostics>`", "inline-protected: `plain`",
			},
		},
		{
			name:     "span in a list item",
			source:   "- Use `<total target-rule diagnostics>` here.\n- Next item.\n",
			blocks:   []string{"list-item: Use \x00 here.", "list-item: Next item."},
			excluded: []string{"inline-protected: `<total target-rule diagnostics>`"},
		},
		{
			// The grammar nests one level for each element before the span
			// that it drops, so this tree was too deep to read at all.
			name:     "span after many other spans",
			source:   strings.Repeat("`key` and ", 70) + "`STRUCT<a INT64>` last.\n",
			blocks:   []string{"paragraph: " + strings.Repeat("\x00 and ", 70) + "\x00 last."},
			excluded: []string{"inline-protected: `STRUCT<a INT64>`", "inline-protected: `key`"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
			doc := parseMarkdownRows(t, tc.source, extract.Options{})
			for _, excluded := range exclusionList(doc) {
				qt.New(t).Assert(excluded, qt.Matches, "inline-protected: `[^`]*`", qt.Commentf("only whole code spans are excluded"))
			}
		})
	}
}

// A block whose code spans the grammar reads keeps its result exactly: the
// same text and the same exclusions as before the repair existed. The last
// rows hold backticks that are no code span, where the repair is tried and
// refused.
func TestMarkdownCodeSpanRepairKeepsWhatTheGrammarReads(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		text         []string
		excluded     []string
	}{
		{
			"span with two words in angle brackets", "Pass `<a b>` here.\n",
			[]string{"Pass  \x00  here."}, []string{"inline-protected: `<a b>`"},
		},
		{
			"generic type", "A `Map<String>` holds text.\n",
			[]string{"A  \x00  holds text."}, []string{"inline-protected: `Map<String>`"},
		},
		{
			"comparison", "Use `a < b` in the filter.\n",
			[]string{"Use  \x00  in the filter."}, []string{"inline-protected: `a < b`"},
		},
		{
			"opening tag with attributes", "Use `<div class=\"a b\">` here.\n",
			[]string{"Use  \x00  here."}, []string{"inline-protected: `<div class=\"a b\">`"},
		},
		{
			"two backticks around single ones", "Use ``<total `target-rule` diagnostics>`` here.\n",
			[]string{"Use  \x00  here."}, []string{"inline-protected: ``<total `target-rule` diagnostics>``"},
		},
		{
			"span after an autolink", "See <https://example.com/a> and `<rule name here>` now.\n",
			[]string{"See  \x00  and  \x00  now."},
			[]string{"inline-protected: <https://example.com/a>", "inline-protected: `<rule name here>`"},
		},
		{
			"tags next to a span", "Use `x` and <b>bold</b> text.\n",
			[]string{"Use  \x00  and  \x00 bold \x00  text."},
			[]string{"inline-protected: `x`", "inline-protected: <b>", "inline-protected: </b>"},
		},
		{
			"many spans, the last one read", strings.Repeat("`key` and ", 70) + "`<a b>` last.\n",
			[]string{strings.Repeat(" \x00  and ", 70) + " \x00  last."},
			append(slices.Repeat([]string{"inline-protected: `key`"}, 70), "inline-protected: `<a b>`"),
		},
		{
			"backticks in a comment", "Text <!-- a > `<b c>` --> more.\n",
			[]string{"Text  \x00  more."}, []string{"inline-protected: <!-- a > `<b c>` -->"},
		},
		{
			"backticks in a link title", "See [the page](url \"`<a b>`\") here.\n",
			[]string{"See the page here."}, []string{},
		},
		{
			"backticks in an image", "See ![`<a b c>`](x.png) here.\n",
			[]string{"See  \x00  here."}, []string{"inline-protected: ![`<a b c>`](x.png)"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			doc := parseMarkdownRows(t, tc.source, extract.Options{})
			text := []string{}
			for _, block := range doc.Blocks {
				text = append(text, block.Text)
			}
			c.Assert(text, qt.DeepEquals, tc.text)
			c.Assert(exclusionList(doc), qt.DeepEquals, tc.excluded)
			assertSourceMap(c, *doc, tc.source)
		})
	}
}
