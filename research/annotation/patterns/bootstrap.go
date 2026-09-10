package patterns

import (
	"context"
	"math"
	"math/rand/v2"
	"slices"
)

// Resampling parameters are part of the version: 10,000 replicates of the
// components with replacement, PCG seed 17, and percentile bounds.
const (
	replicates   = 10000
	seed         = 17
	minimumCount = 2
)

// drawComponents draws every replicate once, so all rules and cohorts share
// the same joint resampling of components.
func drawComponents(ctx context.Context, components int) ([][]int, error) {
	if components < minimumCount {
		return nil, nil
	}
	// #nosec G404 -- Fixed research resampling seed; this generator is not used for security.
	random := rand.New(rand.NewPCG(seed, 0))
	draws := make([][]int, replicates)
	for r := range draws {
		if r%256 == 0 {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
		}
		multiplicity := make([]int, components)
		for range components {
			multiplicity[random.IntN(components)]++
		}
		draws[r] = multiplicity
	}
	return draws, nil
}

// intervalFor estimates a prevalence, or with a baseline the difference of
// two prevalences, from the observed tallies and the shared draws.
func intervalFor(draws [][]int, tallies, baseline []counts, components int) Estimate {
	point, ok := statistic(nil, tallies, baseline)
	estimate := Estimate{Status: "insufficient_evidence"}
	if ok {
		estimate.Value = &point
	}
	if !ok {
		estimate.Status = "nothing_counted"
		return estimate
	}
	if components < minimumCount {
		estimate.Status = "fewer_than_two_components"
		return estimate
	}
	values := make([]float64, 0, len(draws))
	for _, multiplicity := range draws {
		if value, valid := statistic(multiplicity, tallies, baseline); valid {
			values = append(values, value)
		}
	}
	estimate.Replicates = len(values)
	if len(values) < minimumCount {
		estimate.Status = "fewer_than_two_valid_replicates"
		return estimate
	}
	slices.Sort(values)
	lower, upper := percentile(values, 0.025), percentile(values, 0.975)
	estimate.Lower, estimate.Upper, estimate.Status = &lower, &upper, "cluster_percentile"
	return estimate
}

// statistic returns the weighted prevalence, or the difference from the
// baseline prevalence, over components weighted by multiplicity (nil means
// the original sample). It reports false when a denominator is zero.
func statistic(multiplicity []int, tallies, baseline []counts) (float64, bool) {
	value, ok := prevalence(multiplicity, tallies)
	if !ok || baseline == nil {
		return value, ok
	}
	reference, ok := prevalence(multiplicity, baseline)
	if !ok {
		return math.NaN(), false
	}
	return value - reference, true
}

func prevalence(multiplicity []int, tallies []counts) (float64, bool) {
	counted, withFinding := 0, 0
	for i, tally := range tallies {
		weight := 1
		if multiplicity != nil {
			weight = multiplicity[i]
		}
		counted += weight * tally.counted
		withFinding += weight * tally.withFinding
	}
	if counted == 0 {
		return math.NaN(), false
	}
	return float64(withFinding) / float64(counted), true
}

// percentile interpolates with h=(n-1)*p between sorted observations, the
// convention the evaluation harness fixed.
func percentile(sorted []float64, probability float64) float64 {
	h := float64(len(sorted)-1) * probability
	lower := int(h)
	upper := min(lower+1, len(sorted)-1)
	return sorted[lower] + (h-float64(lower))*(sorted[upper]-sorted[lower])
}
