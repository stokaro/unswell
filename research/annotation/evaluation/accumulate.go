package evaluation

import "math"

type accumulator struct {
	counts               Counts
	bins                 []Bin
	sums, positives      [10]float64
	brier, constantBrier float64
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

func (a *accumulator) finish(groups int) Metrics {
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
	return Metrics{Counts: c, Coverage: finitePointer(rates[0]),
		Precision: finitePointer(rates[1]), Recall: finitePointer(rates[2]),
		FPR: finitePointer(rates[3]), FullRecall: finitePointer(rates[4]),
		Brier: ratio(a.brier, c.Covered), ConstantBrier: ratio(a.constantBrier, c.Covered),
		ECE: ratio(ece, c.Covered), Bins: a.bins}
}

func ratio(numerator float64, denominator int) *float64 {
	if denominator == 0 {
		return nil
	}
	value := numerator / float64(denominator)
	return &value
}
