package patterns_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/patterns"
)

func baselineFixture() patterns.BaselineTable {
	return patterns.BaselineTable{
		Measure: patterns.MeasureClosed3, Role: "comment", Sentences: 100, Words: 10000,
		Terms: []patterns.BaselineTerm{
			{Key: "rather than a", Count: 10, PerThousandWords: 1},
			{Key: "in the same", Count: 100, PerThousandWords: 10},
		},
	}
}

// A comparison names what a tree uses more than a baseline does. It needs both
// sides measured, and it must not divide by a rate the baseline never showed.
func TestDeviationNeedsBothSidesAndNamesTheLift(t *testing.T) {
	c := qt.New(t)
	counter, err := patterns.NewCounter(patterns.MeasureClosed3)
	c.Assert(err, qt.IsNil)

	_, err = counter.Compare(baselineFixture(), patterns.DeviationOptions{})
	c.Assert(err, qt.ErrorMatches, ".*needs words on both sides.*")

	for range 6 {
		c.Assert(counter.Add(c.Context(),
			"The plan is refused rather than a rewrite of the whole schema."), qt.IsNil)
	}
	result, err := counter.Compare(baselineFixture(),
		patterns.DeviationOptions{MinCount: 5, MinLift: 2, Top: 10})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Version, qt.Equals, patterns.DeviationVersion)
	c.Assert(result.BaselineRole, qt.Equals, "comment")
	c.Assert(result.TreeWords > 0, qt.IsTrue)

	var found patterns.DeviationTerm
	for _, term := range result.Terms {
		if term.Key == "rather than a" {
			found = term
		}
	}
	c.Assert(found.Key, qt.Equals, "rather than a")
	c.Assert(found.Status, qt.Equals, patterns.DeviationMeasured)
	c.Assert(found.BaselineRate, qt.Equals, float64(1))
	c.Assert(found.Count, qt.Equals, 6)
	c.Assert(found.Lift > 2, qt.IsTrue, qt.Commentf("lift %v", found.Lift))
}

// A key the baseline never reached is not a key with a rate of zero. The two
// are different claims, and the status has to say which one this is.
func TestDeviationMarksAKeyTheBaselineNeverReached(t *testing.T) {
	c := qt.New(t)
	counter, err := patterns.NewCounter(patterns.MeasureClosed3)
	c.Assert(err, qt.IsNil)
	for range 6 {
		c.Assert(counter.Add(c.Context(), "The row is dropped instead of a rewrite of it."), qt.IsNil)
	}
	result, err := counter.Compare(baselineFixture(),
		patterns.DeviationOptions{MinCount: 5, MinLift: 2, Top: 10})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Terms, qt.Not(qt.HasLen), 0)
	for _, term := range result.Terms {
		if term.Status == patterns.DeviationAbsent {
			c.Assert(term.BaselineRate, qt.Equals, float64(0))
			c.Assert(term.BaselineCount, qt.Equals, 0)
			return
		}
	}
	c.Fatal("expected a construction the baseline never reached")
}

func TestDeviationRefusesAMismatchedBaselineAndMeasure(t *testing.T) {
	c := qt.New(t)
	_, err := patterns.NewCounter("invented")
	c.Assert(err, qt.ErrorMatches, `unknown measure "invented"`)

	counter, err := patterns.NewCounter(patterns.MeasureClosed3)
	c.Assert(err, qt.IsNil)
	c.Assert(counter.Add(c.Context(), "The plan is refused rather than a rewrite."), qt.IsNil)
	wrong := baselineFixture()
	wrong.Measure = patterns.MeasureWord3
	_, err = counter.Compare(wrong, patterns.DeviationOptions{})
	c.Assert(err, qt.ErrorMatches, `baseline measures "word-3gram", not "closed-3gram"`)
}
