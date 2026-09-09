package corpus

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestRuleExtractionEqualityPreservesContextSets(t *testing.T) {
	c := qt.New(t)
	base := extract.Policy{Contexts: extract.DefaultContexts()}
	initialized := base
	initialized.Languages = map[document.Format]extract.LanguagePolicy{}
	initialized.Exceptions = []extract.Exception{}
	c.Assert(sameRuleExtraction(base, initialized), qt.IsTrue)
	c.Assert(sameRuleExtraction(extract.Policy{}, extract.Policy{Contexts: []string{}}), qt.IsFalse)
	initialized.Contexts = []string{}
	c.Assert(sameRuleExtraction(base, initialized), qt.IsFalse)
	initialized = base
	initialized.Languages = map[document.Format]extract.LanguagePolicy{document.Go: {Contexts: []string{}}}
	c.Assert(sameRuleExtraction(base, initialized), qt.IsFalse)
}
