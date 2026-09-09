package evaluation

import (
	"context"
	"math"
	"math/rand/v2"
	"slices"
)

const bootstrapReplicates = 10000

type bootstrapSamples [2][3][metricCount][]float64

func bootstrap(ctx context.Context, groups []pairedGroup) ([]MetricComparison, error) {
	var original pairedAggregate
	for _, group := range groups {
		original.add(group)
	}
	var samples bootstrapSamples
	if len(groups) >= 2 {
		if err := drawSamples(ctx, groups, &samples); err != nil {
			return nil, err
		}
	}
	result := bootstrapResults(original, samples, len(groups))
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func bootstrapResults(original pairedAggregate, samples bootstrapSamples, groups int) []MetricComparison {
	points := original.values()
	result := make([]MetricComparison, metricCount)
	for metric, id := range pairedMetricIDs {
		result[metric] = MetricComparison{ID: id}
		for aggregation, destination := range []*Contrast{&result[metric].Micro, &result[metric].GroupMacro} {
			for side, estimate := range []*Estimate{&destination.Candidate, &destination.Comparator, &destination.Difference} {
				value := points[aggregation][side][metric]
				estimate.Value = finitePointer(value)
				estimate.Groups = original.groups[side][metric]
				zeroRisk := side != 2 && (metric == 3 || metric == 5) && value == 0
				estimate.Interval = sampleInterval(samples[aggregation][side][metric], groups, estimate.Groups, zeroRisk)
			}
		}
	}
	return result
}

func drawSamples(ctx context.Context, groups []pairedGroup, samples *bootstrapSamples) error {
	samples.allocate()
	// #nosec G404 -- Fixed research resampling seed; this generator is not used for security.
	random := rand.New(rand.NewPCG(17, 0))
	for range bootstrapReplicates {
		var draw pairedAggregate
		for i := range groups {
			if i%256 == 0 {
				if err := ctx.Err(); err != nil {
					return err
				}
			}
			draw.add(groups[random.IntN(len(groups))])
		}
		samples.add(draw.values())
	}
	return ctx.Err()
}

func (samples *bootstrapSamples) allocate() {
	for aggregation := range samples {
		for side := range samples[aggregation] {
			for metric := range samples[aggregation][side] {
				samples[aggregation][side][metric] = make([]float64, 0, bootstrapReplicates)
			}
		}
	}
}

func (samples *bootstrapSamples) add(values [2][3][metricCount]float64) {
	for aggregation, sides := range values {
		for side, metrics := range sides {
			for metric, value := range metrics {
				if !math.IsNaN(value) {
					samples[aggregation][side][metric] = append(samples[aggregation][side][metric], value)
				}
			}
		}
	}
}

func sampleInterval(values []float64, sourceGroups, definedGroups int, zeroRisk bool) Interval {
	result := Interval{Status: "insufficient_evidence", Valid: len(values)}
	if sourceGroups >= 2 {
		result.Invalid = bootstrapReplicates - len(values)
	}
	switch {
	case definedGroups < 2:
		result.Reason = "fewer_than_two_defined_groups"
	case len(values) < 2:
		result.Reason = "fewer_than_two_valid_replicates"
	case zeroRisk:
		result.Reason = "zero_observed_false_flags_require_separate_risk_bound"
	default:
		slices.Sort(values)
		lower, upper := percentile(values, 0.025), percentile(values, 0.975)
		result.Lower, result.Upper = &lower, &upper
		result.Status = "numerical_percentile"
	}
	return result
}

// Interpolation uses h=(n-1)*p between adjacent sorted observations. This fixes
// the quantile convention as part of the comparison method's version.
func percentile(sorted []float64, probability float64) float64 {
	h := float64(len(sorted)-1) * probability
	lower := int(h)
	upper := min(lower+1, len(sorted)-1)
	return sorted[lower] + (h-float64(lower))*(sorted[upper]-sorted[lower])
}

func finitePointer(value float64) *float64 {
	if math.IsNaN(value) {
		return nil
	}
	return &value
}
