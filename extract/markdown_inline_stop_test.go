package extract_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// On some unfinished tags the inline parser stops before the end of the block
// and returns no tree. That is one damaged block: it is repaired when a repair
// applies and excluded as unparsed otherwise, and the rest of the file is read.
func TestMarkdownReadsAroundABlockThatStopsTheInlineParser(t *testing.T) {
	for _, tc := range []markdownCase{
		{
			// No ">" follows, so the bracket is literal text.
			name:   "attribute with two equals signs",
			source: "Intro paragraph.\n\n<ins data-ad-format==\n\nClosing paragraph.\n",
			blocks: []string{"paragraph: Intro paragraph.", "paragraph: <ins data-ad-format==", "paragraph: Closing paragraph."},
		},
		{
			name:   "attribute value that starts another tag",
			source: "- First item.\n- <a-b c-d=<\n- Last item.\n",
			blocks: []string{"list-item: First item.", "list-item: <a-b c-d=<", "list-item: Last item."},
		},
		{
			// The later ">" keeps the bracket repair away, so the block stays damaged.
			name:     "two unfinished tags before a closing bracket",
			source:   "Intro paragraph.\n\n<ins data-ad-format=\"fluid\"<ins data-ad-format=\"fluid\" 5 > b\n\nClosing paragraph.\n",
			blocks:   []string{"paragraph: Intro paragraph.", "paragraph: Closing paragraph."},
			unparsed: []string{"<ins data-ad-format=\"fluid\"<ins data-ad-format=\"fluid\" 5 > b"},
		},
		{
			name:     "empty attribute value before a closing bracket",
			source:   "Intro paragraph.\n\n<ins data-ad-format=>\n\nClosing paragraph.\n",
			blocks:   []string{"paragraph: Intro paragraph.", "paragraph: Closing paragraph."},
			unparsed: []string{"<ins data-ad-format=>"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			checkMarkdownCase(t, tc)
		})
	}
}

// A parse that a canceled context stops is still an error of the file.
func TestMarkdownCancellationIsNotARejectedBlock(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := extract.Parse(ctx, document.Source{
		Name: "guide.md", Format: document.Markdown, Bytes: []byte("Intro paragraph.\n\n<ins data-ad-format==\n"),
	}, extract.Options{})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
