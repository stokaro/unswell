package extract

// White-box tests: Check which brackets the tag rule of an accepted block
// takes, and that a refused trial returns the first tree and leaves the buffer
// alone. The public result is the same whether the rule was not tried or was
// refused, so only the candidates and the tree itself show the difference.

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestUnfinishedTagsNeedAQuotedValueOutsideCode(t *testing.T) {
	for _, row := range []struct {
		name, source string
		want         []string
	}{
		{"quoted value", "See <a href=\"x\" for details.", []string{"<a href=\"x\""}},
		{"two tags", "First <ins data=\"a\" and <b c='d' end", []string{"<ins data=\"a\"", "<b c='d'"}},
		{"unquoted value", "<ins data=x", nil},
		{"no attribute", "Use <br and go on.", nil},
		{"comparison", "When n <m and k=2 then stop.", nil},
		{"later closing bracket", "See <a href=\"x\" and then b > c.", nil},
		{"inside a code span", "Use `a <b c=\"d\"` here.", nil},
		{"value without its closing quote", "<ins data=\"x", nil},
	} {
		t.Run(row.name, func(t *testing.T) {
			var got []string
			for _, span := range unfinishedTags([]byte(row.source)) {
				got = append(got, row.source[span.Start:span.End])
			}
			qt.New(t).Assert(got, qt.DeepEquals, row.want)
		})
	}
}

func TestUnfinishedTagRuleReplacesOnlyAVerifiedTree(t *testing.T) {
	for _, row := range []struct {
		name, source string
		replaced     bool
	}{
		{"tag in a sentence", "See <a href=\"x\" for details.", true},
		{"no candidate", "When n <m and k=2 then stop.", false},
		{"tag text in a link title", "See [the page](url \"<a b='c'\") here.", false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source := []byte(row.source)
			original, err := parseSyntaxTree(t.Context(), source, "markdown_inline")
			c.Assert(err, qt.IsNil)
			c.Assert(original.tree.RootNode().HasErrorOrMissing(), qt.IsFalse)

			restored := protectUnfinishedTags(t.Context(), original, source)
			defer restored.tree.Release()
			c.Assert(restored.tree != original.tree, qt.Equals, row.replaced)
			c.Assert(bytes.Equal(source, []byte(row.source)), qt.Equals, !row.replaced)
			if row.replaced {
				c.Assert(verifyInlineBoundaries(t.Context(), restored, map[document.Span]string{{Start: 4, End: 15}: "html_tag"}), qt.IsNil)
			}
		})
	}
}
