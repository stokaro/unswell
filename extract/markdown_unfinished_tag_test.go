package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/extract"
)

// An unfinished tag has one reading, whether the inline grammar accepts the
// block or rejects it. The second tag of each pair is the spelling that the
// grammar rejects. With a quoted attribute value the tag is HTML source and
// is excluded; without one the bracket is literal and the text is prose.
func TestMarkdownReadsAnUnfinishedTagTheSameWayInEveryBlock(t *testing.T) {
	for _, tc := range []markdownCase{
		{
			name:     "quoted value in a block that the grammar accepts",
			source:   "<ins data=\"x\"\n",
			blocks:   []string{},
			excluded: []string{"inline-protected: <ins data=\"x\""},
		},
		{
			name:     "quoted value in a block that the grammar rejects",
			source:   "<ins data-ad-format=\"x\"\n",
			blocks:   []string{},
			excluded: []string{"inline-protected: <ins data-ad-format=\"x\""},
		},
		{
			name:     "tag over two lines",
			source:   "<ins\n  data=\"fluid\"\n",
			blocks:   []string{},
			excluded: []string{"inline-protected: <ins\n  data=\"fluid\""},
		},
		{
			name:     "tag inside a sentence",
			source:   "See <a href=\"https://example.com\" for details.\n",
			blocks:   []string{"paragraph: See \x00 for details."},
			excluded: []string{"inline-protected: <a href=\"https://example.com\""},
		},
		{
			name:     "single quotes in a list item",
			source:   "- Item <img src='a.png' alt='A picture'\n- Next item.\n",
			blocks:   []string{"list-item: Item \x00", "list-item: Next item."},
			excluded: []string{"inline-protected: <img src='a.png' alt='A picture'"},
		},
		{
			// The tag ends at its last quoted value. The word after it can be prose.
			name:     "quoted value between an unquoted value and a bare word",
			source:   "<ins data=x title=\"y\" async\n",
			blocks:   []string{"paragraph: \x00 async"},
			excluded: []string{"inline-protected: <ins data=x title=\"y\""},
		},
		{
			name:   "unquoted value in a block that the grammar accepts",
			source: "<ins data=x\n",
			blocks: []string{"paragraph: <ins data=x"},
		},
		{
			name:   "unquoted value in a block that the grammar rejects",
			source: "<ins data-ad-format=fluid\n",
			blocks: []string{"paragraph: <ins data-ad-format=fluid"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
		})
	}
}

// Text that only looks like the start of a tag keeps its reading exactly: a
// comparison with an assignment after it, a bracket with a later ">", a
// bracket inside a code span, and a tag without an attribute value.
func TestMarkdownUnfinishedTagRuleLeavesOtherBracketsAlone(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		text         []string
		excluded     []string
	}{
		{"comparison before an assignment", "When n <m and k=2 then stop.\n", []string{"When n <m and k=2 then stop."}, []string{}},
		{"tag without an attribute", "Use <br and go on.\n", []string{"Use <br and go on."}, []string{}},
		{
			"bracket before a later closing bracket", "See <a href=\"x\" and then b > c.\n",
			[]string{"See <a href=\"x\" and then b > c."}, []string{},
		},
		{"finished tag", "See <a href=\"x\">this</a> page.\n", []string{"See  \x00 this \x00  page."}, []string{
			"inline-protected: <a href=\"x\">", "inline-protected: </a>",
		}},
		{"bracket in an image", "See ![<a b=\"c\"](x.png) here.\n", []string{"See  \x00  here."}, []string{
			"inline-protected: ![<a b=\"c\"](x.png)",
		}},
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
