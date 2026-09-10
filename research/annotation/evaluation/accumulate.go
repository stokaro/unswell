package evaluation

import (
	"cmp"
	"math"
	"slices"
	"strings"
)

type accumulator struct {
	counts               Counts
	bins                 []Bin
	sums, positives      [10]float64
	brier, constantBrier float64
	selected             []selection
}

// selection retains one covered decision for the risk-coverage curve. Ordering
// by confidence is what makes selective risk meaningful; the unit ID breaks ties
// so the same predictions always produce the same curve.
type selection struct {
	unitID     string
	confidence float64
	wrong      bool
}

func newAccumulator() *accumulator {
	a := &accumulator{bins: make([]Bin, 10)}
	for i := range a.bins {
		a.bins[i] = Bin{Lower: float64(i) / 10, Upper: float64(i+1) / 10}
	}
	return a
}

func (a *accumulator) add(row Observation, constant float64) {
	a.counts.Eligible++
	if row.Response == nil {
		if row.Label == 1 {
			a.counts.AbstainedPositive++
		} else {
			a.counts.AbstainedNegative++
		}
		return
	}
	a.counts.Covered++
	a.addClassification(row)
	response := *row.Response
	a.selected = append(a.selected, selection{unitID: row.UnitID,
		confidence: max(response, 1-response), wrong: *row.Positive != (row.Label == 1)})
	label := float64(row.Label)
	a.brier += (response - label) * (response - label)
	a.constantBrier += (constant - label) * (constant - label)
	bin := min(int(response*10), 9)
	a.bins[bin].Count++
	a.sums[bin] += response
	a.positives[bin] += label
}

func (a *accumulator) addClassification(row Observation) {
	switch {
	case *row.Positive && row.Label == 1:
		a.counts.TP++
	case *row.Positive:
		a.counts.FP++
	case row.Label == 1:
		a.counts.FN++
	default:
		a.counts.TN++
	}
}

func (a *accumulator) finish(groups int, curve bool) Metrics {
	c := a.counts
	c.Groups = groups
	rates := classificationValues(c)
	ece := 0.0
	for i := range a.bins {
		bin := &a.bins[i]
		bin.MeanResponse = ratio(a.sums[i], bin.Count)
		bin.PositiveRate = ratio(a.positives[i], bin.Count)
		if bin.Count != 0 {
			ece += math.Abs(a.sums[i] - a.positives[i])
		}
	}
	var risk []RiskPoint
	if curve {
		risk = a.riskCoverage()
	}
	return Metrics{Counts: c, Coverage: finitePointer(rates[0]), RiskCoverage: risk,
		Precision: finitePointer(rates[1]), Recall: finitePointer(rates[2]),
		FPR: finitePointer(rates[3]), FullRecall: finitePointer(rates[4]),
		Brier: ratio(a.brier, c.Covered), ConstantBrier: ratio(a.constantBrier, c.Covered),
		ECE: ratio(ece, c.Covered), Bins: a.bins}
}

// riskCoverage reports the error rate among the most confident decisions at
// thresholds a caller could actually apply. A prefix that splits two equally
// confident decisions is skipped, because no threshold separates them.
func (a *accumulator) riskCoverage() []RiskPoint {
	ordered := slices.Clone(a.selected)
	slices.SortFunc(ordered, func(x, y selection) int {
		if x.confidence != y.confidence {
			return cmp.Compare(y.confidence, x.confidence)
		}
		return strings.Compare(x.unitID, y.unitID)
	})
	points := []RiskPoint{}
	errors := 0
	step := max((len(ordered)+19)/20, 1)
	for index, entry := range ordered {
		if entry.wrong {
			errors++
		}
		accepted := index + 1
		if accepted < len(ordered) && ordered[accepted].confidence == entry.confidence {
			continue
		}
		if accepted != len(ordered) && accepted%step != 0 {
			continue
		}
		points = append(points, RiskPoint{Coverage: float64(accepted) / float64(a.counts.Eligible),
			Accepted: accepted, Errors: errors, Risk: float64(errors) / float64(accepted),
			MinimumConfidence: entry.confidence})
	}
	return points
}

func ratio(numerator float64, denominator int) *float64 {
	if denominator == 0 {
		return nil
	}
	value := numerator / float64(denominator)
	return &value
}
