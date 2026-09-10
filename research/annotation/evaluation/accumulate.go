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
	response   float64
	positive   bool
	wrong      bool
}

// falsePositiveLimits and prevalences are the fixed operating points the
// evaluation reports; a record reader compares runs at the same points.
var (
	falsePositiveLimits = []float64{0.01, 0.02, 0.05, 0.10}
	prevalences         = []float64{0.01, 0.05, 0.10, 0.25, 0.50}
)

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
	a.selected = append(a.selected, selection{unitID: row.UnitID, confidence: max(response, 1-response),
		response: response, positive: row.Label == 1, wrong: *row.Positive != (row.Label == 1)})
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
	var recallLimits []RecallPoint
	var prevalence []PrevalencePoint
	if curve {
		risk = a.riskCoverage()
		recallLimits = a.recallAtLimits()
		prevalence = prevalenceSensitivity(finitePointer(rates[2]), finitePointer(rates[3]))
	}
	return Metrics{Counts: c, Coverage: finitePointer(rates[0]), RiskCoverage: risk,
		RecallAtFalsePositiveLimits: recallLimits, PrevalenceSensitivity: prevalence,
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

// recallAtLimits sweeps a response threshold downward and keeps, for each
// false-positive limit, the highest recall reached without exceeding it. A
// threshold can only fall between distinct responses, so equal responses are
// accepted together.
func (a *accumulator) recallAtLimits() []RecallPoint {
	ordered := slices.Clone(a.selected)
	slices.SortFunc(ordered, func(x, y selection) int {
		if x.response != y.response {
			return cmp.Compare(y.response, x.response)
		}
		return strings.Compare(x.unitID, y.unitID)
	})
	positives, negatives := classCounts(ordered)
	points := make([]RecallPoint, len(falsePositiveLimits))
	for i, limit := range falsePositiveLimits {
		points[i] = RecallPoint{FalsePositiveLimit: limit}
	}
	if positives == 0 || negatives == 0 {
		return points
	}
	truePositives, falsePositives := 0, 0
	for index, entry := range ordered {
		if entry.positive {
			truePositives++
		} else {
			falsePositives++
		}
		if index+1 < len(ordered) && ordered[index+1].response == entry.response {
			continue
		}
		improveRecallPoints(points, float64(falsePositives)/float64(negatives),
			float64(truePositives)/float64(positives), entry.response)
	}
	return points
}

func classCounts(ordered []selection) (int, int) {
	positives := 0
	for _, entry := range ordered {
		if entry.positive {
			positives++
		}
	}
	return positives, len(ordered) - positives
}

// improveRecallPoints records one threshold for every limit it satisfies with a
// higher recall than the limit has seen so far.
func improveRecallPoints(points []RecallPoint, rate, recall, threshold float64) {
	for i := range points {
		if rate > points[i].FalsePositiveLimit || (points[i].Recall != nil && recall <= *points[i].Recall) {
			continue
		}
		rate, recall, threshold := rate, recall, threshold
		points[i].Recall, points[i].FalsePositiveRate, points[i].Threshold = &recall, &rate, &threshold
	}
}

// prevalenceSensitivity applies Bayes' rule to the covered recall and
// false-positive rate at fixed assumed prevalences. It needs both rates; a set
// without negatives has no false-positive rate and reports no precision.
func prevalenceSensitivity(recall, falsePositiveRate *float64) []PrevalencePoint {
	points := make([]PrevalencePoint, len(prevalences))
	for i, prevalence := range prevalences {
		points[i] = PrevalencePoint{Prevalence: prevalence}
		if recall == nil || falsePositiveRate == nil {
			continue
		}
		flagged := prevalence**recall + (1-prevalence)**falsePositiveRate
		if flagged == 0 {
			continue
		}
		precision := prevalence * *recall / flagged
		points[i].Precision = &precision
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
