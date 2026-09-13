package extract_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

// A link carries prose only through its visible label. The inline grammar
// also reads a bracket pair in ordinary prose as a reference link, so a Java
// type such as long[][] reaches the same code with no label at all. A lone
// bracket pair stays ordinary text.
func TestMarkdownLinksWithoutAVisibleLabel(t *testing.T) {
	for _, test := range []struct {
		name, source, keep string
		protected          bool
	}{
		{"array type", "Invoke assertThat(long[][]) on the array.\n", "Invoke assertThat(long", true},
		{"empty inline link", "See [](https://example.com) for the schedule.\n", "See", true},
		{"empty reference link", "See [][guide] for the schedule.\n", "See", true},
		{"lone bracket pair", "The value [] holds nothing.\n", "The value [] holds nothing.", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "note.md",
				Format: document.Markdown, Bytes: []byte(test.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(len(doc.Blocks) > 0, qt.IsTrue)
			boundary := qt.Contains
			if !test.protected {
				boundary = qt.Not(qt.Contains)
			}
			c.Assert(doc.Blocks[0].Text, qt.Contains, test.keep)
			c.Assert(doc.Blocks[0].Text, boundary, "\x00")
			for _, excluded := range doc.Excluded {
				c.Assert(excluded.Span.Valid(len(test.source)), qt.IsTrue)
			}
		})
	}
}
