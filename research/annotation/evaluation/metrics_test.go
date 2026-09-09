package evaluation_test

import (
	"context"
	"math"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func observed(id, group string, label int, response float64) evaluation.Observation {
	positive := response >= 0.5
	return evaluation.Observation{UnitID: id, GroupID: group, Label: label, Response: &response, Positive: &positive}
}

func TestMetricsSeparateCoverageAndCoveredErrors(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.7),
		observed("c", "two", 1, 0.2), observed("d", "two", 0, 0.1), {UnitID: "e", GroupID: "two", Label: 1}}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	m := result.Micro
	c.Assert(m.Counts, qt.DeepEquals, evaluation.Counts{Eligible: 5, Covered: 4, Groups: 2, TP: 1, FP: 1, TN: 1, FN: 1,
		AbstainedPositive: 1})
	for _, row := range []struct {
		name string
		got  *float64
		want float64
	}{
		{"coverage", m.Coverage, 0.8}, {"precision", m.Precision, 0.5}, {"recall", m.Recall, 0.5},
		{"FPR", m.FPR, 0.5}, {"full recall", m.FullRecall, 1.0 / 3}, {"Brier", m.Brier, 0.2875},
		{"constant Brier", m.ConstantBrier, 0.25}, {"ECE", m.ECE, 0.425},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(row.got, qt.IsNotNil)
			c.Assert(math.Abs(*row.got-row.want) < 1e-12, qt.IsTrue)
		})
	}
	c.Assert(result.Groups, qt.HasLen, 2)
	c.Assert(result.Groups[0].Metrics.Counts.Covered, qt.Equals, 2)
	c.Assert(result.Groups[1].Metrics.Counts.AbstainedPositive, qt.Equals, 1)
	slices.Reverse(rows)
	again, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
}

func TestReliabilityBoundariesAndUndefinedMetrics(t *testing.T) {
	c := qt.New(t)
	result, err := evaluation.Summarize(t.Context(), []evaluation.Observation{
		observed("a", "group", 0, 0), observed("b", "group", 0, 0.1), observed("c", "group", 1, 1)}, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Micro.Bins[0].Count, qt.Equals, 1)
	c.Assert(result.Micro.Bins[1].Count, qt.Equals, 1)
	c.Assert(result.Micro.Bins[9].Count, qt.Equals, 1)
	c.Assert(result.Micro.Bins[2].MeanResponse, qt.IsNil)
	c.Assert(result.Micro.Bins[2].PositiveRate, qt.IsNil)
	empty, err := evaluation.Summarize(t.Context(), nil, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(empty.Micro.Coverage, qt.IsNil)
	c.Assert(empty.Micro.Brier, qt.IsNil)
	c.Assert(empty.Micro.ECE, qt.IsNil)
	c.Assert(empty.Micro.FPR, qt.IsNil)
	abstained, err := evaluation.Summarize(t.Context(), []evaluation.Observation{{UnitID: "a", GroupID: "g", Label: 1}}, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(abstained.Micro.Recall, qt.IsNil)
	c.Assert(*abstained.Micro.FullRecall, qt.Equals, float64(0))
	c.Assert(abstained.Micro.Brier, qt.IsNil)
}

func TestMetricInputValidationAndCancellation(t *testing.T) {
	for _, row := range []struct {
		name string
		rows []evaluation.Observation
	}{
		{"duplicate", []evaluation.Observation{observed("a", "g", 1, 0.1), observed("a", "g", 0, 0.5)}},
		{"label", []evaluation.Observation{observed("a", "g", 2, 0.1)}},
		{"group", []evaluation.Observation{observed("a", "", 1, 0.1)}},
		{"NaN", []evaluation.Observation{observed("a", "g", 1, math.NaN())}},
		{"range", []evaluation.Observation{observed("a", "g", 1, 1.01)}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := evaluation.Summarize(t.Context(), row.rows, 0.5)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, evaluation.Summary{})
		})
	}
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := evaluation.Summarize(ctx, nil, 0.5)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
