package evaluation

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

// PairedObservation retains both responses to the same target and judgment.
type PairedObservation struct {
	Candidate  Observation
	Comparator Observation
}

// Interval describes a numerical group bootstrap, not qualification evidence.
// Bounds are absent when source support is insufficient. Invalid replicates are
// counted separately and never filled with zero.
type Interval struct {
	Status  string   `json:"status"`
	Reason  string   `json:"reason"`
	Lower   *float64 `json:"lower"`
	Upper   *float64 `json:"upper"`
	Valid   int      `json:"valid_replicates"`
	Invalid int      `json:"invalid_replicates"`
}

// Estimate records the point estimate and number of groups defining that metric.
type Estimate struct {
	Value    *float64 `json:"value"`
	Groups   int      `json:"defined_groups"`
	Interval Interval `json:"interval"`
}

// Contrast reports candidate minus comparator. A macro difference averages
// within-group differences only where both rates are defined; it need not equal
// the difference of independently supported group means.
type Contrast struct {
	Candidate  Estimate `json:"candidate"`
	Comparator Estimate `json:"comparator"`
	Difference Estimate `json:"difference"`
}

// MetricComparison separates unit-weighted metrics from equal-group averages.
type MetricComparison struct {
	ID         string   `json:"id"`
	Micro      Contrast `json:"micro"`
	GroupMacro Contrast `json:"group_macro"`
}

// ZeroFlagBound is the one-sided exact 95% zero-event binomial upper bound
// for groups with eligible negatives and any false flag among those negatives.
// It requires independent Bernoulli sampling of the declared group population;
// it never certifies the differently weighted full-corpus micro FPR.
type ZeroFlagBound struct {
	Groups     int      `json:"groups_with_negatives"`
	Flagged    int      `json:"groups_with_false_flags"`
	Upper      *float64 `json:"upper_95"`
	Status     string   `json:"status"`
	Reason     string   `json:"reason"`
	Population string   `json:"population"`
}

// PairedScope retains per-group results and uncertainty for one target scope.
type PairedScope struct {
	Candidate           Summary            `json:"candidate"`
	Comparator          Summary            `json:"comparator"`
	Metrics             []MetricComparison `json:"metrics"`
	CandidateZeroFlags  ZeroFlagBound      `json:"candidate_zero_flags"`
	ComparatorZeroFlags ZeroFlagBound      `json:"comparator_zero_flags"`
}

// PairedSummary reports common support alongside the full eligible target flow.
// Resampling preserves groups, pairs, labels, and abstentions. Metadata declares
// grouping; numerical calculations cannot establish source independence.
type PairedSummary struct {
	Method         string      `json:"method"`
	Replicates     int         `json:"replicates"`
	Seed           int         `json:"seed"`
	FullFlow       PairedScope `json:"full_flow"`
	CommonCovered  PairedScope `json:"common_covered"`
	CandidateOnly  int         `json:"candidate_only_covered"`
	ComparatorOnly int         `json:"comparator_only_covered"`
	Neither        int         `json:"neither_covered"`
}

// Compare calculates paired metrics using 10,000 group bootstrap replicates
// with PCG(17,0) and interpolated empirical percentile intervals. Constants are
// each model's frozen training prevalence. At most 10,000 targets are accepted.
func Compare(ctx context.Context, rows []PairedObservation, constants [2]float64) (PairedSummary, error) {
	if len(rows) > 10000 {
		return PairedSummary{}, fmt.Errorf("comparison exceeds 10000 targets")
	}
	rows = slices.Clone(rows)
	slices.SortFunc(rows, func(a, b PairedObservation) int { return strings.Compare(a.Candidate.UnitID, b.Candidate.UnitID) })
	if err := validatePairs(ctx, rows); err != nil {
		return PairedSummary{}, err
	}
	result := PairedSummary{Method: "paired-source-group-pcg17-percentile-v1", Replicates: bootstrapReplicates, Seed: 17}
	common := commonRows(rows, &result)
	var err error
	result.FullFlow, err = compareScope(ctx, rows, constants)
	if err != nil {
		return PairedSummary{}, err
	}
	result.CommonCovered, err = compareScope(ctx, common, constants)
	if err != nil {
		return PairedSummary{}, err
	}
	return result, nil
}

func validatePairs(ctx context.Context, rows []PairedObservation) error {
	seenA, seenB := make(map[string]bool), make(map[string]bool)
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		a, b := row.Candidate, row.Comparator
		if a.UnitID != b.UnitID || a.GroupID != b.GroupID || a.Label != b.Label {
			return fmt.Errorf("paired targets must have identical unit, group, and label")
		}
		if err := validateObservation(a, seenA); err != nil {
			return err
		}
		if err := validateObservation(b, seenB); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func commonRows(rows []PairedObservation, result *PairedSummary) []PairedObservation {
	common := []PairedObservation{}
	for _, row := range rows {
		switch {
		case row.Candidate.Response != nil && row.Comparator.Response != nil:
			common = append(common, row)
		case row.Candidate.Response != nil:
			result.CandidateOnly++
		case row.Comparator.Response != nil:
			result.ComparatorOnly++
		default:
			result.Neither++
		}
	}
	return common
}

func compareScope(ctx context.Context, rows []PairedObservation, constants [2]float64) (PairedScope, error) {
	a, b := make([]Observation, len(rows)), make([]Observation, len(rows))
	for i, row := range rows {
		a[i], b[i] = row.Candidate, row.Comparator
	}
	var result PairedScope
	var err error
	result.Candidate, err = Summarize(ctx, a, constants[0])
	if err != nil {
		return PairedScope{}, err
	}
	result.Comparator, err = Summarize(ctx, b, constants[1])
	if err != nil {
		return PairedScope{}, err
	}
	groups := pairGroups(rows, constants)
	result.Metrics, err = bootstrap(ctx, groups)
	if err != nil {
		return PairedScope{}, err
	}
	result.CandidateZeroFlags = zeroFlagBound(result.Candidate)
	result.ComparatorZeroFlags = zeroFlagBound(result.Comparator)
	return result, nil
}
