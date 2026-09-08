package mapping_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/internal/mapping"
)

func TestCanonicalPreservesTechnicalChanges(t *testing.T) {
	c := qt.New(t)
	before := canonicalMarkdown(t, "The client **may retry** `Upload` 3 times.")
	c.Assert(before, qt.Equals, canonicalMarkdown(t, "The client may retry `Upload`\r\n3 times."))
	for _, changed := range []string{
		"The client may not retry `Upload` 3 times.",
		"The client may retry `Download` 3 times.",
		"The client may retry `upload` 3 times.",
		"The client may retry `Upload` 4 times.",
		"The client may retry Upload 3 times.",
	} {
		c.Assert(before, qt.Not(qt.Equals), canonicalMarkdown(t, changed), qt.Commentf("%s", changed))
	}
}

func canonicalMarkdown(t *testing.T, text string) string {
	t.Helper()
	c := qt.New(t)
	source := []byte(text)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: source}, extract.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	canonical, err := mapping.Canonical(doc.Blocks[0].MappedText, doc.Source)
	c.Assert(err, qt.IsNil)
	return canonical
}
