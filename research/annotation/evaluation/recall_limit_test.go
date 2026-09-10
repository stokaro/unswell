package evaluation_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/evaluation"
)

func close(c *qt.C, got *float64, want float64) {
	c.Helper()
	c.Assert(got, qt.IsNotNil)
	c.Assert(math.Abs(*got-want) < 1e-12, qt.IsTrue, qt.Commentf("got %v want %v", *got, want))
}

// A false-positive limit is what a CI user actually sets. The table answers
// how much recall each limit buys and at which threshold, from the same saved
// responses the other metrics use.
func TestRecallAtFalsePositiveLimitsSweepsThresholds(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.8), observed("c", "one", 1, 0.6)}
	for i := range 9 {
		rows = append(rows, observed("n"+string(rune('0'+i)), "one", 0, 0.1))
	}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	table := result.Micro.RecallAtFalsePositiveLimits
	c.Assert(table, qt.HasLen, 4)
	// Ten negatives: one false positive is a rate of 0.10, allowed only by the last limit.
	for _, point := range table[:3] {
		close(c, point.Recall, 0.5)
		close(c, point.FalsePositiveRate, 0)
		close(c, point.Threshold, 0.9)
	}
	c.Assert(table[3].FalsePositiveLimit, qt.Equals, 0.10)
	close(c, table[3].Recall, 1)
	close(c, table[3].FalsePositiveRate, 0.1)
	close(c, table[3].Threshold, 0.6)
	c.Assert(result.Groups[0].Metrics.RecallAtFalsePositiveLimits, qt.HasLen, 0)
}

// Precision on the evaluation set reflects that set's base rate. The table
// restates it at fixed prevalences so a false-alert rate reads correctly in a
// repository where most prose is fine.
func TestPrevalenceSensitivityAppliesBayesRule(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{observed("a", "one", 1, 0.9), observed("b", "one", 1, 0.7), observed("c", "one", 0, 0.6)}
	for i := range 9 {
		rows = append(rows, observed("n"+string(rune('0'+i)), "one", 0, 0.2))
	}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	// Recall 1.0, false-positive rate 0.1 at the fixed threshold.
	close(c, result.Micro.Recall, 1)
	close(c, result.Micro.FPR, 0.1)
	table := result.Micro.PrevalenceSensitivity
	c.Assert(table, qt.HasLen, 5)
	c.Assert(table[0].Prevalence, qt.Equals, 0.01)
	close(c, table[0].Precision, 0.01/(0.01+0.99*0.1))
	c.Assert(table[4].Prevalence, qt.Equals, 0.5)
	close(c, table[4].Precision, 0.5/(0.5+0.5*0.1))
}

// Without negatives there is no false-positive rate, so neither table can
// claim a value. Absent is not zero risk.
func TestLimitTablesStayUndefinedWithoutNegatives(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{observed("a", "one", 1, 0.9), observed("b", "one", 1, 0.4)}
	result, err := evaluation.Summarize(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Micro.RecallAtFalsePositiveLimits, qt.HasLen, 4)
	for _, point := range result.Micro.RecallAtFalsePositiveLimits {
		c.Assert(point.Recall, qt.IsNil)
	}
	c.Assert(result.Micro.PrevalenceSensitivity, qt.HasLen, 5)
	for _, point := range result.Micro.PrevalenceSensitivity {
		c.Assert(point.Precision, qt.IsNil)
	}
}
