package sparsemodel

import (
	"context"
	"fmt"
	"math"
)

type objectiveValue struct {
	loss     float64
	gradient []float64
}

func objective(ctx context.Context, d data, theta []float64, work *budget) (objectiveValue, error) {
	result := objectiveValue{gradient: make([]float64, len(theta))}
	if err := work.charge(int64(len(theta)) * 10); err != nil {
		return objectiveValue{}, err
	}
	offset := theta[0]
	for j, mean := range d.means {
		offset -= theta[j+1] * mean / d.scales[j]
	}
	for _, row := range d.rows {
		if err := ctx.Err(); err != nil {
			return objectiveValue{}, err
		}
		if err := work.charge(int64(len(row.Values))*8 + 10); err != nil {
			return objectiveValue{}, err
		}
		addRow(&result, d, row, theta, offset)
	}
	regularize(&result, d, theta)
	if !finite(result.loss) {
		return objectiveValue{}, fmt.Errorf("sparse objective overflow")
	}
	for _, g := range result.gradient {
		if !finite(g) {
			return objectiveValue{}, fmt.Errorf("sparse gradient overflow")
		}
	}
	return result, nil
}

func addRow(result *objectiveValue, d data, row Example, theta []float64, offset float64) {
	z := offset
	for _, entry := range row.Values {
		if !d.constant[entry.Index] {
			z += theta[entry.Index+1] * entry.Value / d.scales[entry.Index]
		}
	}
	label := float64(row.Label)
	result.loss += math.Max(z, 0) - label*z + math.Log1p(math.Exp(-math.Abs(z)))
	probability := math.Exp(-math.Abs(z))
	if z >= 0 {
		probability = 1 / (1 + probability)
	} else {
		probability /= 1 + probability
	}
	residual := probability - label
	result.gradient[0] += residual
	for _, entry := range row.Values {
		if !d.constant[entry.Index] {
			result.gradient[entry.Index+1] += residual * entry.Value / d.scales[entry.Index]
		}
	}
}

func regularize(result *objectiveValue, d data, theta []float64) {
	n := float64(len(d.rows))
	result.loss /= n
	result.gradient[0] /= n
	for j, weight := range theta[1:] {
		result.loss += 0.01 * weight * weight / 2
		if d.constant[j] {
			result.gradient[j+1] = 0.01 * weight
		} else {
			result.gradient[j+1] = result.gradient[j+1]/n - d.means[j]/d.scales[j]*result.gradient[0] + 0.01*weight
		}
	}
}
