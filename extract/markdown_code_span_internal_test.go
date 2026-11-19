package extract

// White-box tests: Check that the code span repair of an accepted block is not
// tried when the tree has every span, and that a refused trial returns the
// first tree and leaves the buffer alone. The public result is the same in
// both cases, so only the tree itself shows that a reading was not replaced.

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCodeSpanRepairReplacesOnlyATreeThatLostASpan(t *testing.T) {
	for _, row := range []struct {
		name, source string
		lost         int
		replaced     bool
	}{
		{"span that the grammar reads", "Pass `<a b>` here.", 0, false},
		{"span without a bracket", "Pass `a b` here.", 0, false},
		{"bracket outside a span", "Pass `a` and <b>bold</b> here.", 0, false},
		{"span that the grammar drops", "Use `<total target-rule diagnostics>` here.", 1, true},
		{"backticks in a comment", "Text <!-- a > `<b c>` --> more.", 1, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source := []byte(row.source)
			original, err := parseSyntaxTree(t.Context(), source, "markdown_inline")
			c.Assert(err, qt.IsNil)
			c.Assert(original.tree.RootNode().HasErrorOrMissing(), qt.IsFalse)
			c.Assert(lostCodeSpans(t.Context(), original, source), qt.HasLen, row.lost)

			restored := restoreCodeSpans(t.Context(), original, source)
			defer restored.tree.Release()
			c.Assert(restored.tree != original.tree, qt.Equals, row.replaced)
			c.Assert(bytes.Equal(source, []byte(row.source)), qt.Equals, !row.replaced)
		})
	}
}
