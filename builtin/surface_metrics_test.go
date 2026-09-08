package builtin

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestProseMeasurementsUseWordsAndUnicodeCharacters(t *testing.T) {
	c := qt.New(t)
	block := document.Block{Sentences: []document.Sentence{{Tokens: []document.Token{
		{Text: "café", Normal: "café", Word: true, Tag: "NN"},
		{Text: "café", Normal: "café", Word: true, Tag: "NN"},
		{Text: "12", Normal: "12", Word: true, Tag: "CD"},
		{Text: "long_identifier", Normal: "long_identifier", Word: true, Protected: true},
		{Text: "."},
	}}}}
	stats, err := measureProse(newEditorialMatcher(context.Background(), rule.View{MaxCandidates: 100}), block)
	c.Assert(err, qt.IsNil)
	c.Assert(stats.words, qt.Equals, 3)
	c.Assert(stats.characters, qt.Equals, 10)
	c.Assert(stats.nouns, qt.Equals, 2)
	c.Assert(stats.frequencies, qt.DeepEquals, map[string]int{"café": 2, "12": 1})
	mean, deviation, shortest, longest := lengthStatistics([]int{2, 4, 6})
	c.Assert(mean, qt.Equals, float64(4))
	c.Assert(deviation, qt.Equals, math.Sqrt(float64(8)/3))
	c.Assert(shortest, qt.Equals, 2)
	c.Assert(longest, qt.Equals, 6)
}

func TestEmptyProseHasNoStatisticalEvidence(t *testing.T) {
	c := qt.New(t)
	stats := proseMeasurements{}
	c.Assert(stats.metrics(), qt.HasLen, 0)
	mean, deviation, shortest, longest := lengthStatistics(nil)
	c.Assert(mean, qt.Equals, float64(0))
	c.Assert(deviation, qt.Equals, float64(0))
	c.Assert(shortest+longest, qt.Equals, 0)
}
