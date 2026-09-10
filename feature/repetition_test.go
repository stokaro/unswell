package feature_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
)

// Repetition ratios describe one unit and nothing beyond it. Technical prose
// repeats terms for good reasons, so these values judge no quality.
func TestRepetitionRatiosDescribeOneUnit(t *testing.T) {
	for _, row := range []struct {
		name                          string
		sentences                     []string
		peak, pair, opener, duplicate float64
	}{
		{"distinct prose", []string{"the cache retries once", "workers drain their backlog now"}, 1.0 / 9, 0, 0, 0},
		{"repeated sentence", []string{"the cache retries once", "the cache retries once"}, 2.0 / 8, 1, 1, 1},
		{"repeated opener only", []string{"the cache retries once", "the workers drain backlogs now"}, 2.0 / 9, 0, 1, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, err := feature.Measure(t.Context(), testBlock(row.sentences...), testIdentity(), testLimits())
			c.Assert(err, qt.IsNil)
			c.Assert(numeric(t, m, "peak-word-frequency-ratio"), qt.Equals, row.peak)
			c.Assert(numeric(t, m, "repeated-bigram-ratio"), qt.Equals, row.pair)
			c.Assert(numeric(t, m, "sentence-opener-repeat-ratio"), qt.Equals, row.opener)
			c.Assert(numeric(t, m, "duplicate-sentence-ratio"), qt.Equals, row.duplicate)
		})
	}
}

func TestRepetitionRatiosAbstainWithoutPairsOrProse(t *testing.T) {
	c := qt.New(t)
	m, err := feature.Measure(t.Context(), testBlock("alpha", "beta"), testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	pair, err := m.Value("repeated-bigram-ratio")
	c.Assert(err, qt.IsNil)
	c.Assert(pair.Number, qt.IsNil)
	c.Assert(pair.Reason, qt.Equals, "no_eligible_pair")
	c.Assert(numeric(t, m, "duplicate-sentence-ratio"), qt.Equals, float64(0))

	empty, err := feature.Measure(t.Context(), testBlock(""), testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	for _, id := range []string{"peak-word-frequency-ratio", "repeated-bigram-ratio",
		"sentence-opener-repeat-ratio", "duplicate-sentence-ratio"} {
		value, err := empty.Value(id)
		c.Assert(err, qt.IsNil)
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, "no_prose_words")
	}
}

// The repetition family joins the existing contract rather than replacing it,
// so every earlier descriptor keeps its definition and position.
func TestRepetitionJoinsTheExistingContract(t *testing.T) {
	c := qt.New(t)
	catalog := feature.Catalog()
	c.Assert(catalog, qt.HasLen, 18)
	c.Assert(catalog[0].ID, qt.Equals, "prose-words")
	c.Assert(catalog[13].ID, qt.Equals, "automated-readability-index")
	families := map[string]int{}
	for _, descriptor := range catalog {
		families[descriptor.Family]++
	}
	c.Assert(families["repetition"], qt.Equals, 4)
	c.Assert(feature.Contract, qt.Equals, "unswell-block-features-v1")
	unit, err := feature.UnitCatalog("sentence")
	c.Assert(err, qt.IsNil)
	c.Assert(unit, qt.HasLen, 18)
	c.Assert(unit[14].Family, qt.Equals, "repetition")
	c.Assert(unit[14].Scope, qt.Equals, "sentence")
}
