// Package evaluation summarizes frozen research predictions and independent labels.
// Numerical metrics do not qualify editorial models or establish data provenance.
package evaluation

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
)

// Observation is one eligible binary editorial judgment and an optional response.
// Nil Response means abstention; Positive must be nil in that case as well.
type Observation struct {
	UnitID   string
	GroupID  string
	Label    int
	Response *float64
	Positive *bool
}

// Counts separates covered classification from eligible but unscored targets.
type Counts struct {
	Eligible          int `json:"eligible"`
	Covered           int `json:"covered"`
	Groups            int `json:"groups"`
	TP                int `json:"true_positive"`
	FP                int `json:"false_positive"`
	TN                int `json:"true_negative"`
	FN                int `json:"false_negative"`
	AbstainedPositive int `json:"abstained_positive"`
	AbstainedNegative int `json:"abstained_negative"`
}

// Bin uses fixed equal-width response bins, with 1 included in the final bin.
// Empty bins have no estimated response or positive rate.
type Bin struct {
	Lower        float64  `json:"lower"`
	Upper        float64  `json:"upper"`
	Count        int      `json:"count"`
	MeanResponse *float64 `json:"mean_response"`
	PositiveRate *float64 `json:"positive_rate"`
}

// RiskPoint is one selective-prediction operating point: the error rate among
// the most confident decisions a confidence threshold can accept. Coverage
// counts eligible targets, so abstention lowers it and a model cannot reach full
// coverage by refusing to answer.
type RiskPoint struct {
	Coverage          float64 `json:"coverage"`
	Accepted          int     `json:"accepted"`
	Errors            int     `json:"errors"`
	Risk              float64 `json:"risk"`
	MinimumConfidence float64 `json:"minimum_confidence"`
}

// RecallPoint is the highest recall any response threshold reaches while keeping
// the false-positive rate at or below a stated limit. Nil recall means no
// threshold satisfies the limit, or the set has no positives or no negatives.
type RecallPoint struct {
	FalsePositiveLimit float64  `json:"false_positive_limit"`
	Recall             *float64 `json:"recall"`
	FalsePositiveRate  *float64 `json:"false_positive_rate"`
	Threshold          *float64 `json:"threshold"`
}

// PrevalencePoint is the precision the covered recall and false-positive rate
// imply at an assumed prevalence. A reader can then judge false alerts under
// their own base rate instead of the evaluation set's.
type PrevalencePoint struct {
	Prevalence float64  `json:"prevalence"`
	Precision  *float64 `json:"precision"`
}

// Metrics describes eligible-flow coverage and quality on its covered subset.
// Undefined denominators produce null, never a perfect score or zero risk.
type Metrics struct {
	Counts        Counts   `json:"counts"`
	Coverage      *float64 `json:"coverage"`
	Precision     *float64 `json:"precision"`
	Recall        *float64 `json:"recall"`
	FPR           *float64 `json:"false_positive_rate"`
	FullRecall    *float64 `json:"full_flow_recall"`
	Brier         *float64 `json:"brier"`
	ConstantBrier *float64 `json:"training_constant_brier"`
	ECE           *float64 `json:"ece"`
	Bins          []Bin    `json:"reliability"`
	// RiskCoverage is reported for the full eligible flow only. Per-group curves
	// would multiply the record without adding a decision anyone makes per group.
	RiskCoverage []RiskPoint `json:"risk_coverage,omitempty"`
	// The two tables below are also full-flow only, for the same reason.
	RecallAtFalsePositiveLimits []RecallPoint     `json:"recall_at_false_positive_limits,omitempty"`
	PrevalenceSensitivity       []PrevalencePoint `json:"prevalence_sensitivity,omitempty"`
}

// GroupMetrics retains per-component denominators for later paired comparisons.
type GroupMetrics struct {
	GroupID string  `json:"group_id"`
	Metrics Metrics `json:"metrics"`
}

// Summary contains the full eligible flow and every independent source group.
// Confidence intervals and scientific qualification are separate procedures.
type Summary struct {
	Micro  Metrics        `json:"micro"`
	Groups []GroupMetrics `json:"groups"`
}

// Summarize computes binary-event Brier, ten-bin ECE, and classification metrics.
// Constant is frozen training prevalence, not prevalence estimated from this set.
func Summarize(ctx context.Context, rows []Observation, constant float64) (Summary, error) {
	if !responseValue(constant) {
		return Summary{}, fmt.Errorf("constant baseline must be finite and within [0,1]")
	}
	if len(rows) > 10000 {
		return Summary{}, fmt.Errorf("evaluation exceeds 10000 targets")
	}
	rows = slices.Clone(rows)
	slices.SortFunc(rows, func(a, b Observation) int { return strings.Compare(a.UnitID, b.UnitID) })
	all := newAccumulator()
	groups := make(map[string]*accumulator)
	seen := make(map[string]bool)
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return Summary{}, err
		}
		if err := validateObservation(row, seen); err != nil {
			return Summary{}, err
		}
		group := groups[row.GroupID]
		if group == nil {
			group = newAccumulator()
			groups[row.GroupID] = group
		}
		group.add(row, constant)
		all.add(row, constant)
	}
	result := Summary{Micro: all.finish(len(groups), true), Groups: []GroupMetrics{}}
	for id, group := range groups {
		result.Groups = append(result.Groups, GroupMetrics{id, group.finish(1, false)})
	}
	slices.SortFunc(result.Groups, func(a, b GroupMetrics) int { return strings.Compare(a.GroupID, b.GroupID) })
	if err := ctx.Err(); err != nil {
		return Summary{}, err
	}
	return result, nil
}

func validateObservation(row Observation, seen map[string]bool) error {
	if row.UnitID == "" || row.GroupID == "" || seen[row.UnitID] || (row.Label != 0 && row.Label != 1) {
		return fmt.Errorf("evaluation requires unique unit IDs, groups, and binary labels")
	}
	seen[row.UnitID] = true
	if row.Response == nil {
		if row.Positive != nil {
			return fmt.Errorf("abstention must not invent a threshold decision")
		}
		return nil
	}
	if !responseValue(*row.Response) || row.Positive == nil {
		return fmt.Errorf("available response requires a finite value within [0,1] and a threshold decision")
	}
	return nil
}

func responseValue(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1
}
