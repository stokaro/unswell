package extract

// White-box tests: Feed an orphaned parser token directly to the private block
// visitor to verify its incomplete-tree guard. The public extractor applies the
// boundary adapter first, so it cannot inject the uncorrected grammar's tree.

import (
	"testing"

	qt "github.com/frankban/quicktest"
	ts "github.com/stokaro/gotreesitter"
	"github.com/stokaro/gotreesitter/grammars"

	"github.com/stokaro/unswell/document"
)

func TestMarkdownRejectsOrphanedGrammarToken(t *testing.T) {
	c := qt.New(t)
	lang := grammars.MarkdownLanguage()
	tree, err := ts.NewParser(lang).ParseStrict([]byte("| Name | Result |\n| --- | --- |\n| Client | Output |\n|\n| After | Table |\n"))
	c.Assert(err, qt.IsNil)
	defer tree.Release()
	var orphan *ts.Node
	err = walkSyntax(t.Context(), tree.RootNode(), 0, func(node *ts.Node) (bool, error) {
		if node.Type(lang) == "|" && !markdownTableDelimiter(node.Parent(), lang) {
			orphan = node
		}
		return false, nil
	})
	c.Assert(err, qt.IsNil)
	c.Assert(orphan, qt.IsNotNil)
	reader := markdownReader{ctx: t.Context(), doc: &document.Document{}, syntax: syntaxTree{tree: tree, lang: lang}}
	skip, err := reader.block(orphan)
	c.Assert(skip, qt.IsTrue)
	c.Assert(err, qt.ErrorMatches, "markdown grammar returned an orphaned table delimiter")
}
