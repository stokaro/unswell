package evaluation

import (
	"math"
	"slices"
)

const metricCount = 9

var pairedMetricIDs = [metricCount]string{"coverage", "precision", "recall", "false_positive_rate",
	"full_flow_recall", "full_flow_false_flag_rate", "brier", "training_constant_brier", "ece"}

// Sufficient statistics keep resampling proportional to source groups rather
// than expanding duplicated targets. ECE uses signed residuals in fixed bins.
type pairedCounts struct {
	c                    Counts
	brier, constantBrier float64
	residuals            [10]float64
}

func (p *pairedCounts) add(row Observation, constant float64) {
	a := newAccumulator()
	a.add(row, constant)
	p.merge(pairedCounts{c: a.counts, brier: a.brier, constantBrier: a.constantBrier})
	if row.Response != nil {
		p.residuals[min(int(*row.Response*10), 9)] += *row.Response - float64(row.Label)
	}
}

func (p *pairedCounts) merge(other pairedCounts) {
	p.c.Eligible += other.c.Eligible
	p.c.Covered += other.c.Covered
	p.c.TP += other.c.TP
	p.c.FP += other.c.FP
	p.c.TN += other.c.TN
	p.c.FN += other.c.FN
	p.c.AbstainedPositive += other.c.AbstainedPositive
	p.c.AbstainedNegative += other.c.AbstainedNegative
	p.brier += other.brier
	p.constantBrier += other.constantBrier
	for i, value := range other.residuals {
		p.residuals[i] += value
	}
}

func (p pairedCounts) values() [metricCount]float64 {
	c := p.c
	rates := classificationValues(c)
	ece := 0.0
	for _, residual := range p.residuals {
		ece += math.Abs(residual)
	}
	return [metricCount]float64{rates[0], rates[1], rates[2], rates[3], rates[4], rates[5],
		quotient(p.brier, c.Covered), quotient(p.constantBrier, c.Covered), quotient(ece, c.Covered)}
}

func classificationValues(c Counts) [6]float64 {
	return [6]float64{quotient(float64(c.Covered), c.Eligible), quotient(float64(c.TP), c.TP+c.FP),
		quotient(float64(c.TP), c.TP+c.FN), quotient(float64(c.FP), c.FP+c.TN),
		quotient(float64(c.TP), c.TP+c.FN+c.AbstainedPositive), quotient(float64(c.FP), c.FP+c.TN+c.AbstainedNegative)}
}

func quotient(value float64, count int) float64 {
	if count == 0 {
		return math.NaN()
	}
	return value / float64(count)
}

type pairedGroup struct {
	counts [2]pairedCounts
	values [3][metricCount]float64
}

func pairGroups(rows []PairedObservation, constants [2]float64) []pairedGroup {
	byID := make(map[string]*pairedGroup)
	ids := []string{}
	for _, row := range rows {
		id := row.Candidate.GroupID
		if byID[id] == nil {
			byID[id] = &pairedGroup{}
			ids = append(ids, id)
		}
		byID[id].counts[0].add(row.Candidate, constants[0])
		byID[id].counts[1].add(row.Comparator, constants[1])
	}
	slices.Sort(ids)
	groups := make([]pairedGroup, 0, len(ids))
	for _, id := range ids {
		group := *byID[id]
		group.values = contrastValues(group.counts)
		groups = append(groups, group)
	}
	return groups
}

func contrastValues(counts [2]pairedCounts) [3][metricCount]float64 {
	values := [3][metricCount]float64{counts[0].values(), counts[1].values()}
	for i := range metricCount {
		values[2][i] = values[0][i] - values[1][i]
	}
	return values
}

type pairedAggregate struct {
	counts [2]pairedCounts
	sums   [3][metricCount]float64
	groups [3][metricCount]int
}

func (a *pairedAggregate) add(group pairedGroup) {
	for i := range a.counts {
		a.counts[i].merge(group.counts[i])
	}
	for side, values := range group.values {
		for metric, value := range values {
			if !math.IsNaN(value) {
				a.sums[side][metric] += value
				a.groups[side][metric]++
			}
		}
	}
}

func (a pairedAggregate) values() [2][3][metricCount]float64 {
	values := [2][3][metricCount]float64{contrastValues(a.counts)}
	for side, sums := range a.sums {
		for metric, sum := range sums {
			values[1][side][metric] = quotient(sum, a.groups[side][metric])
		}
	}
	return values
}

func zeroFlagBound(summary Summary) ZeroFlagBound {
	result := ZeroFlagBound{Status: "not_computed", Population: "independent_groups_with_eligible_negatives"}
	for _, group := range summary.Groups {
		c := group.Metrics.Counts
		if c.FP+c.TN+c.AbstainedNegative > 0 {
			result.Groups++
			if c.FP > 0 {
				result.Flagged++
			}
		}
	}
	switch {
	case result.Groups == 0:
		result.Reason = "no_negative_groups"
	case result.Flagged != 0:
		result.Reason = "zero_event_formula_not_applicable"
	default:
		upper := -math.Expm1(math.Log(0.05) / float64(result.Groups))
		result.Upper, result.Status = &upper, "conditional_exact_zero_event"
		result.Reason = "requires_independent_group_sampling_not_micro_fpr_evidence"
	}
	return result
}
