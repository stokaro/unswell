package evaluation_test

import (
	"context"
	"encoding/json"
	"math"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func paired(id, group string, label int, a, b float64) evaluation.PairedObservation {
	return evaluation.PairedObservation{Candidate: observed(id, group, label, a), Comparator: observed(id, group, label, b)}
}

func comparisonRows() []evaluation.PairedObservation {
	rows := []evaluation.PairedObservation{paired("a", "one", 1, 0.9, 0.1), paired("b", "one", 1, 0.9, 0.1),
		paired("c", "one", 0, 0.1, 0.9), paired("d", "two", 1, 0.1, 0.9), paired("e", "two", 0, 0.1, 0.1),
		paired("f", "three", 1, 0.1, 0.9), paired("g", "three", 0, 0.1, 0.1), paired("h", "three", 0, 0.1, 0.1)}
	rows[5].Candidate.Response, rows[5].Candidate.Positive = nil, nil
	rows[6].Comparator.Response, rows[6].Comparator.Positive = nil, nil
	rows[7].Candidate.Response, rows[7].Candidate.Positive = nil, nil
	rows[7].Comparator.Response, rows[7].Comparator.Positive = nil, nil
	return rows
}

func comparisonMetric(t *testing.T, scope evaluation.PairedScope, id string) evaluation.MetricComparison {
	t.Helper()
	c := qt.New(t)
	i := slices.IndexFunc(scope.Metrics, func(m evaluation.MetricComparison) bool { return m.ID == id })
	c.Assert(i >= 0, qt.IsTrue)
	return scope.Metrics[i]
}

func near(t *testing.T, value *float64, want float64) {
	t.Helper()
	c := qt.New(t)
	c.Assert(value, qt.IsNotNil)
	c.Assert(math.Abs(*value-want) < 1e-12, qt.IsTrue, qt.Commentf("got %.15f, want %.15f", *value, want))
}

func TestPairedSupportAndGroupWeighting(t *testing.T) {
	c := qt.New(t)
	result, err := evaluation.Compare(t.Context(), comparisonRows(), [2]float64{0.5, 0.5})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Replicates, qt.Equals, 10000)
	c.Assert(result.Seed, qt.Equals, 17)
	c.Assert(result.CandidateOnly, qt.Equals, 1)
	c.Assert(result.ComparatorOnly, qt.Equals, 1)
	c.Assert(result.Neither, qt.Equals, 1)
	c.Assert(result.CommonCovered.Candidate.Micro.Counts.Eligible, qt.Equals, 5)
	common := comparisonMetric(t, result.CommonCovered, "recall")
	near(t, common.Micro.Candidate.Value, 2.0/3)
	near(t, common.Micro.Comparator.Value, 1.0/3)
	near(t, common.Micro.Difference.Value, 1.0/3)
	near(t, common.GroupMacro.Difference.Value, 0)
	near(t, common.Micro.Difference.Interval.Lower, -1)
	near(t, common.Micro.Difference.Interval.Upper, 1)
	full := comparisonMetric(t, result.FullFlow, "full_flow_recall")
	near(t, full.Micro.Candidate.Value, 0.5)
	near(t, full.Micro.Comparator.Value, 0.5)
	near(t, full.Micro.Difference.Value, 0)
	near(t, full.GroupMacro.Difference.Value, -1.0/3)
	fpr := comparisonMetric(t, result.CommonCovered, "false_positive_rate")
	near(t, fpr.Micro.Candidate.Value, 0)
	c.Assert(fpr.Micro.Candidate.Interval.Upper, qt.IsNil)
	c.Assert(fpr.Micro.Candidate.Interval.Reason, qt.Equals, "zero_observed_false_flags_require_separate_risk_bound")
	near(t, result.CommonCovered.CandidateZeroFlags.Upper, 1-math.Sqrt(0.05))
	c.Assert(result.CommonCovered.ComparatorZeroFlags.Upper, qt.IsNil)
	_, err = json.Marshal(result)
	c.Assert(err, qt.IsNil)
}

func TestPairedResamplingPreservesPairsAndOrdering(t *testing.T) {
	c := qt.New(t)
	rows := comparisonRows()
	for i := range rows {
		rows[i].Comparator = rows[i].Candidate
	}
	result, err := evaluation.Compare(t.Context(), rows, [2]float64{0.5, 0.5})
	c.Assert(err, qt.IsNil)
	for _, metric := range result.CommonCovered.Metrics {
		near(t, metric.Micro.Difference.Value, 0)
		if metric.ID == "precision" {
			c.Assert(metric.Micro.Difference.Groups, qt.Equals, 1)
			c.Assert(metric.Micro.Difference.Interval.Upper, qt.IsNil)
		} else {
			near(t, metric.Micro.Difference.Interval.Lower, 0)
			near(t, metric.Micro.Difference.Interval.Upper, 0)
		}
	}
	slices.Reverse(rows)
	again, err := evaluation.Compare(t.Context(), rows, [2]float64{0.5, 0.5})
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
}

func TestPairedUndefinedClassesAndMissingReplicates(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.PairedObservation{paired("a", "positive", 1, 0.9, 0.1), paired("b", "negative", 0, 0.1, 0.9)}
	result, err := evaluation.Compare(t.Context(), rows, [2]float64{0.5, 0.5})
	c.Assert(err, qt.IsNil)
	recall := comparisonMetric(t, result.CommonCovered, "recall")
	c.Assert(recall.Micro.Candidate.Groups, qt.Equals, 1)
	c.Assert(recall.Micro.Candidate.Interval.Upper, qt.IsNil)
	c.Assert(recall.Micro.Candidate.Interval.Reason, qt.Equals, "fewer_than_two_defined_groups")
	c.Assert(recall.Micro.Candidate.Interval.Invalid > 2000, qt.IsTrue)
	c.Assert(recall.Micro.Candidate.Interval.Invalid < 3000, qt.IsTrue)
	c.Assert(recall.Micro.Candidate.Interval.Valid+recall.Micro.Candidate.Interval.Invalid, qt.Equals, 10000)
	empty, err := evaluation.Compare(t.Context(), nil, [2]float64{0.5, 0.5})
	c.Assert(err, qt.IsNil)
	c.Assert(empty.FullFlow.Candidate.Micro.Coverage, qt.IsNil)
	c.Assert(comparisonMetric(t, empty.FullFlow, "recall").Micro.Difference.Value, qt.IsNil)
	c.Assert(empty.FullFlow.CandidateZeroFlags.Upper, qt.IsNil)
}

func TestPairedInputErrors(t *testing.T) {
	for _, change := range []struct {
		name   string
		mutate func([]evaluation.PairedObservation)
	}{
		{"unit", func(r []evaluation.PairedObservation) { r[0].Comparator.UnitID = "different" }},
		{"group", func(r []evaluation.PairedObservation) { r[0].Comparator.GroupID = "different" }},
		{"label", func(r []evaluation.PairedObservation) { r[0].Comparator.Label = 0 }},
		{"duplicate", func(r []evaluation.PairedObservation) { r[1] = r[0] }},
		{"NaN", func(r []evaluation.PairedObservation) { value := math.NaN(); r[0].Candidate.Response = &value }},
		{"invented decision", func(r []evaluation.PairedObservation) { r[0].Candidate.Response = nil }},
	} {
		t.Run(change.name, func(t *testing.T) {
			c := qt.New(t)
			rows := comparisonRows()
			change.mutate(rows)
			result, err := evaluation.Compare(t.Context(), rows, [2]float64{0.5, 0.5})
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, evaluation.PairedSummary{})
		})
	}
	c := qt.New(t)
	_, err := evaluation.Compare(t.Context(), make([]evaluation.PairedObservation, 10001), [2]float64{0.5, 0.5})
	c.Assert(err, qt.ErrorMatches, ".*10000 targets")
	_, err = evaluation.Compare(t.Context(), nil, [2]float64{math.Inf(1), 0.5})
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := evaluation.Compare(ctx, nil, [2]float64{0.5, 0.5})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, evaluation.PairedSummary{})
}
