package evaluation_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

// A selective model is only useful if its confident answers are its good ones.
// The curve reports the error rate among the answers a confidence threshold
// would accept, so a reader can see where accuracy starts to fall.
func TestRiskCoverageOrdersDecisionsByConfidence(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{
		observed("a", "one", 1, 0.95), // confident and right
		observed("b", "one", 0, 0.90), // wrong: predicts positive on a negative
		observed("c", "two", 0, 0.20), // confident and right
		observed("d", "two", 0, 0.60), // least confident, and wrong
	}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	curve := result.Micro.RiskCoverage
	c.Assert(curve, qt.HasLen, 4)
	c.Assert(curve[0], qt.Equals, evaluation.RiskPoint{Coverage: 0.25, Accepted: 1, Errors: 0,
		Risk: 0, MinimumConfidence: 0.95})
	c.Assert(curve[1].Accepted, qt.Equals, 2)
	c.Assert(curve[1].Errors, qt.Equals, 1)
	c.Assert(curve[1].Risk, qt.Equals, 0.5)
	c.Assert(curve[3], qt.Equals, evaluation.RiskPoint{Coverage: 1, Accepted: 4, Errors: 2,
		Risk: 0.5, MinimumConfidence: 0.6})
	c.Assert(result.Groups[0].Metrics.RiskCoverage, qt.HasLen, 0)
}

// No threshold separates two equally confident decisions, so the curve cannot
// claim an operating point between them.
func TestRiskCoverageSkipsUnreachableThresholds(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{
		observed("a", "one", 1, 0.8), observed("b", "one", 0, 0.8), observed("c", "one", 0, 0.1),
	}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	curve := result.Micro.RiskCoverage
	// Without the rule this would report three points, one of them unreachable.
	c.Assert(curve, qt.HasLen, 2)
	c.Assert(curve[0].Accepted, qt.Equals, 1)
	c.Assert(curve[0].MinimumConfidence, qt.Equals, 0.9)
	c.Assert(curve[1].Accepted, qt.Equals, 3)
	c.Assert(curve[1].MinimumConfidence, qt.Equals, 0.8)
}

// Abstention lowers coverage. A model that answers half the targets cannot
// present itself as covering all of them, however good its answers are.
func TestRiskCoverageCountsAbstentionsAgainstCoverage(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{
		observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.2),
		{UnitID: "c", GroupID: "one", Label: 1}, {UnitID: "d", GroupID: "one", Label: 0},
	}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	curve := result.Micro.RiskCoverage
	c.Assert(curve, qt.HasLen, 2)
	c.Assert(curve[1].Coverage, qt.Equals, 0.5)
	c.Assert(curve[1].Risk, qt.Equals, 0.0)
}

// Without a covered decision there is no operating point to report, and an
// empty curve is not a zero-risk claim.
func TestRiskCoverageIsAbsentWithoutCoveredDecisions(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{{UnitID: "a", GroupID: "one", Label: 1}}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Micro.RiskCoverage, qt.HasLen, 0)
	c.Assert(result.Micro.Coverage, qt.IsNotNil)
	c.Assert(*result.Micro.Coverage, qt.Equals, 0.0)
}
