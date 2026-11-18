package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func parseMDXRecovery(t *testing.T, source string) (document.Document, error) {
	t.Helper()
	return extract.Parse(t.Context(), document.Source{Name: "guide.mdx", Format: document.MDX, Bytes: []byte(source)},
		extract.Options{})
}

// MDX looks for indented prose in a tree of the blocks before it extracts
// them. A run that the block grammar rejects is repaired or excluded there as
// in Markdown, so it no longer fails the file.
func TestMDXRecoversBlocksLikeMarkdown(t *testing.T) {
	tabs := "<Tabs>\n\n    Indented prose here.\n\n</Tabs>\n\n"
	for _, tc := range []struct {
		name, source string
		blocks       []string
		unparsed     []string
	}{
		{
			name:   "list item that starts with a character reference in a link label",
			source: "Intro paragraph.\n\n- [Install &amp; use](#install)\n- [Other](#other)\n\nClosing paragraph.\n",
			blocks: []string{"Intro paragraph.", "Install & use", "Other", "Closing paragraph."},
		},
		{
			name:   "the same list after indented prose",
			source: tabs + "- [Install &amp; use](#install)\n- [Other](#other)\n",
			blocks: []string{"Indented prose here.", "Install & use", "Other"},
		},
		{
			// The label does not end on its line, so no repair applies and
			// the run is excluded.
			name:     "run that stays rejected",
			source:   "Intro paragraph.\n\n* [Install &amp;\n  use](#install)\n* Next item.\n\nClosing paragraph.\n",
			blocks:   []string{"Intro paragraph.", "Closing paragraph."},
			unparsed: []string{"* [Install &amp;\n  use](#install)\n* Next item.\n"},
		},
		{
			name:     "run that stays rejected after indented prose",
			source:   tabs + "* [Install &amp;\n  use](#install)\n* Next item.\n\nClosing paragraph.\n",
			blocks:   []string{"Indented prose here.", "Closing paragraph."},
			unparsed: []string{"* [Install &amp;\n  use](#install)\n* Next item.\n"},
		},
		{
			// The table has no empty cell, so its run is excluded. The span
			// was reported four bytes early, by the indentation removed above.
			name:     "wrongly built table after indented prose",
			source:   tabs + "Text before.\n\n| A | B |\n| --- | --- |\n|x|\n|-\n",
			blocks:   []string{"Indented prose here.", "Text before."},
			unparsed: []string{"| A | B |\n| --- | --- |\n|x|\n|-\n"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			doc, err := parseMDXRecovery(t, tc.source)
			c.Assert(err, qt.IsNil)
			c.Assert(blockTexts(doc), qt.DeepEquals, tc.blocks)
			c.Assert(unparsedText(&doc), qt.DeepEquals, tc.unparsed)
			assertSourceMap(c, doc, tc.source)
			assertVerbatimBlocks(c, &doc, tc.source)
		})
	}
}

// Recovery covers what the Markdown grammar rejects. What MDX itself refuses
// is still an error of the file.
func TestMDXRecoveryStillRefusesDamagedMDX(t *testing.T) {
	for _, tc := range []struct{ name, source, want string }{
		{
			"unclosed element", "Intro paragraph.\n\n- First item.\n\n  <p>\n\n  More of the item.\n",
			`parse mdx: unclosed JSX element "p"`,
		},
		{
			"unclosed expression", "Intro paragraph.\n\n{unclosed\n\nClosing paragraph.\n",
			`parse mdx at byte 18: invalid or unterminated MDX expression.*`,
		},
		{
			"damaged tag", "Intro paragraph.\n\n<Tabs value=>\n\nClosing paragraph.\n",
			`parse mdx at byte 18: invalid or unterminated MDX tag.*`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parseMDXRecovery(t, tc.source)
			qt.New(t).Assert(err, qt.ErrorMatches, tc.want)
		})
	}
}
