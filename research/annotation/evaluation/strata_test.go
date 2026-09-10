package evaluation_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func stratumFor(strata []evaluation.Stratum, dimension, value string) *evaluation.Stratum {
	for i := range strata {
		if strata[i].Dimension == dimension && strata[i].Value == value {
			return &strata[i]
		}
	}
	return nil
}

// A breakdown is only trustworthy if every unit lands in exactly one stratum
// per dimension and the sample size travels with the numbers.
func TestStratifyBreaksDownByRecordedAttributes(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{
		observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.2),
		observed("c", "two", 1, 0.4), {UnitID: "d", GroupID: "two", Label: 0},
	}
	attributes := map[string]evaluation.Attributes{
		"a": {Words: 12, Role: "paragraph", Language: "markdown", ProseLanguage: "en", Origin: "human"},
		"b": {Words: 30, Role: "paragraph", Language: "go", ProseLanguage: "en", Origin: "human"},
		"c": {Words: 250, Role: "comment", Language: "go", ProseLanguage: "en"},
		"d": {Words: 200, Role: "comment", Language: "go", ProseLanguage: "en", Origin: "machine"},
	}
	strata, err := evaluation.Stratify(t.Context(), rows, attributes, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(strata, qt.HasLen, 11)
	short := stratumFor(strata, "words", "1-19")
	c.Assert(short, qt.IsNotNil)
	c.Assert(short.Metrics.Counts.Eligible, qt.Equals, 1)
	c.Assert(short.Metrics.Counts.TP, qt.Equals, 1)
	long := stratumFor(strata, "words", "200+")
	c.Assert(long, qt.IsNotNil)
	c.Assert(long.Metrics.Counts.Eligible, qt.Equals, 2)
	c.Assert(long.Metrics.Counts.Covered, qt.Equals, 1)
	c.Assert(long.Metrics.Counts.AbstainedNegative, qt.Equals, 1)
	goStratum := stratumFor(strata, "language", "go")
	c.Assert(goStratum, qt.IsNotNil)
	c.Assert(goStratum.Metrics.Counts.Eligible, qt.Equals, 3)
	c.Assert(goStratum.Metrics.Counts.Groups, qt.Equals, 2)
	c.Assert(stratumFor(strata, "origin", "unknown").Metrics.Counts.Eligible, qt.Equals, 1)
	// Strata carry no full-flow tables; those describe the whole set only.
	c.Assert(goStratum.Metrics.RiskCoverage, qt.HasLen, 0)
	c.Assert(goStratum.Metrics.PrevalenceSensitivity, qt.HasLen, 0)
}

// Dropping a unit without attributes would shrink a stratum without saying so.
func TestStratifyRefusesUnitsWithoutAttributes(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{observed("a", "one", 1, 0.9)}
	_, err := evaluation.Stratify(t.Context(), rows, map[string]evaluation.Attributes{}, 0.5)
	c.Assert(err, qt.ErrorMatches, "unit a has no corpus attributes")
	_, err = evaluation.Stratify(t.Context(), rows, map[string]evaluation.Attributes{"a": {}}, 2)
	c.Assert(err, qt.ErrorMatches, "constant baseline must be finite.*")
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = evaluation.Stratify(canceled, rows, map[string]evaluation.Attributes{"a": {}}, 0.5)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

// Corpus metadata cannot turn the record into one row per unit.
func TestStratifyBoundsTheNumberOfStrata(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{}
	attributes := map[string]evaluation.Attributes{}
	for i := range 65 {
		id := "u" + string(rune('A'+i%26)) + string(rune('a'+i/26))
		rows = append(rows, observed(id, "one", 1, 0.9))
		attributes[id] = evaluation.Attributes{Role: id}
	}
	_, err := evaluation.Stratify(t.Context(), rows, attributes, 0.5)
	c.Assert(err, qt.ErrorMatches, "dimension role exceeds 64 strata")
}
