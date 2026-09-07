package rule_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestTermContainmentPreservesIndividualBoundaries(t *testing.T) {
	c := qt.New(t)
	ranges := []rule.TokenRange{{BlockID: 1, SentenceID: 2, Start: 4, End: 7}, {BlockID: 1, SentenceID: 2, Start: 2, End: 5}}
	matches, err := rule.NewTermMatches(ranges)
	c.Assert(err, qt.IsNil)
	ranges[0].End = 100
	view := rule.View{TermExemptions: matches}
	sentence := document.Sentence{BlockID: 1, ID: 2, Tokens: make([]document.Token, 10)}
	for _, bounds := range [][2]int{{2, 5}, {3, 4}, {4, 7}, {6, 7}} {
		c.Assert(view.Exempts(sentence, bounds[0], bounds[1]), qt.IsTrue)
	}
	for _, bounds := range [][2]int{{2, 7}, {1, 3}, {6, 8}, {-1, 1}, {4, 4}, {4, 11}} {
		c.Assert(view.Exempts(sentence, bounds[0], bounds[1]), qt.IsFalse)
	}
	sentence.ID++
	c.Assert(view.Exempts(sentence, 2, 5), qt.IsFalse)
	c.Assert((rule.View{}).Exempts(sentence, 2, 5), qt.IsFalse)
	_, err = rule.NewTermMatches([]rule.TokenRange{{Start: 1, End: 0}})
	c.Assert(err, qt.IsNotNil)
}
