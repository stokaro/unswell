package glyph_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/glyph"
)

func TestPictographLeavesStatusMarksAlone(t *testing.T) {
	c := qt.New(t)
	for _, r := range "\U0001F680⚡⭐\U0001F1FA" {
		c.Assert(glyph.Pictograph(r), qt.IsTrue, qt.Commentf("%q", r))
	}
	for _, r := range "✅❌✓☐a—→" {
		c.Assert(glyph.Pictograph(r), qt.IsFalse, qt.Commentf("%q", r))
	}
}
