package sparsemodel

import (
	"context"
	"fmt"
	"math"
)

type statistics struct {
	counts []int
	first  []float64
	varies []bool
}

func prepare(ctx context.Context, rows []Example, features int) (data, int, error) {
	if err := ctx.Err(); err != nil {
		return data{}, 0, err
	}
	if invalidDimensions(rows, features) {
		return data{}, 0, fmt.Errorf("sparse fit dimensions exceed supported bounds")
	}
	result := data{rows: rows, means: make([]float64, features), scales: make([]float64, features), constant: make([]bool, features)}
	stats := statistics{counts: make([]int, features), first: make([]float64, features), varies: make([]bool, features)}
	positive, entries := 0, 0
	for _, row := range rows {
		if err := validateExample(ctx, row, features); err != nil {
			return data{}, 0, err
		}
		positive += row.Label
		entries += len(row.Values)
		if entries > 4000000 {
			return data{}, 0, fmt.Errorf("sparse fit exceeds retained entry budget")
		}
		accumulate(&result, stats, row)
	}
	if positive == 0 || positive == len(rows) {
		return data{}, 0, fmt.Errorf("sparse fit requires both classes")
	}
	if err := normalize(&result, stats); err != nil {
		return data{}, 0, err
	}
	return result, positive, nil
}

func accumulate(d *data, stats statistics, row Example) {
	for _, entry := range row.Values {
		j := entry.Index
		if stats.counts[j] == 0 {
			stats.first[j] = entry.Value
		} else if stats.first[j] != entry.Value {
			stats.varies[j] = true
		}
		stats.counts[j]++
		delta := entry.Value - d.means[j]
		d.means[j] += delta / float64(stats.counts[j])
		d.scales[j] += delta * (entry.Value - d.means[j])
	}
}

func invalidDimensions(rows []Example, features int) bool {
	return features < 1 || features > MaxFeatures || len(rows) < 2 || len(rows) > 100000
}

func validateExample(ctx context.Context, row Example, features int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if row.Label != 0 && row.Label != 1 {
		return fmt.Errorf("invalid sparse binary label")
	}
	return validateRow(row.Values, features)
}

func validateRow(values []Entry, features int) error {
	previous := -1
	for _, entry := range values {
		if entry.Index <= previous || entry.Index >= features || !finite(entry.Value) ||
			entry.Value == 0 || math.Abs(entry.Value) > 1e12 {
			return fmt.Errorf("sparse entries must be sorted, unique, in range and finite nonzero values")
		}
		previous = entry.Index
	}
	return nil
}

func normalize(d *data, stats statistics) error {
	n := float64(len(d.rows))
	for j, nonzeroMean := range d.means {
		fraction := float64(stats.counts[j]) / n
		mean := nonzeroMean * fraction
		// Combine nonzero Welford statistics with the omitted all-zero group.
		variance := d.scales[j]/n + nonzeroMean*nonzeroMean*fraction*(1-fraction)
		if !finite(mean) || !finite(variance) || variance < 0 {
			return fmt.Errorf("sparse normalization overflow")
		}
		d.means[j] = mean
		d.scales[j] = math.Sqrt(math.Max(variance, 0))
		if d.scales[j] == 0 {
			if stats.varies[j] || stats.counts[j] > 0 && stats.counts[j] < len(d.rows) {
				return fmt.Errorf("nonconstant sparse feature variance underflow")
			}
			d.scales[j] = 1
			d.constant[j] = true
		}
	}
	return nil
}
