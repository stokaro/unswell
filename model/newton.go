package model

import (
	"context"
	"fmt"
	"math"
)

func newtonStep(ctx context.Context, data trainingData, theta []float64, current derivatives, l2 float64, budget *workBudget) (
	[]float64, error,
) {
	width := int64(len(theta))
	if err := budget.charge(width*width*width + 2*width*width); err != nil {
		return nil, err
	}
	direction, err := solvePositive(current.hessian, current.gradient)
	if err != nil {
		return nil, err
	}
	slope := dot(current.gradient, direction)
	if !finite(slope) || slope <= 0 {
		return nil, fmt.Errorf("%w: Newton direction is not descent", ErrNumerical)
	}
	candidate := make([]float64, len(theta))
	step := 1.0
	for range 30 {
		for i := range theta {
			candidate[i] = theta[i] - step*direction[i]
		}
		next, err := objective(ctx, data, candidate, l2, false, budget)
		if err != nil {
			return nil, err
		}
		if acceptableStep(current, next, step*slope) {
			return candidate, nil
		}
		step /= 2
	}
	return nil, fmt.Errorf("%w: line search exhausted", ErrConvergence)
}

func acceptableStep(current, next derivatives, decrease float64) bool {
	if next.loss <= current.loss-1e-4*decrease {
		return true
	}
	// Near a stationary point, the expected loss decrease can be below rounding
	// resolution. Permit only that band, with a separate gradient reduction.
	roundoff := 16 * (math.Nextafter(current.loss, math.Inf(1)) - current.loss)
	return finite(roundoff) && decrease <= roundoff && next.loss-current.loss <= roundoff &&
		infinityNorm(next.gradient) < infinityNorm(current.gradient)/2
}

func solvePositive(matrix, right []float64) ([]float64, error) {
	width := len(right)
	lower, err := cholesky(matrix, width)
	if err != nil {
		return nil, err
	}
	result := make([]float64, width)
	for i := range width {
		value := right[i]
		for j := range i {
			value -= lower[i*width+j] * result[j]
		}
		result[i] = value / lower[i*width+i]
	}
	for i := width - 1; i >= 0; i-- {
		value := result[i]
		for j := i + 1; j < width; j++ {
			value -= lower[j*width+i] * result[j]
		}
		result[i] = value / lower[i*width+i]
	}
	if !allFinite(result) {
		return nil, fmt.Errorf("%w: Cholesky solve overflow", ErrNumerical)
	}
	return result, nil
}

func cholesky(matrix []float64, width int) ([]float64, error) {
	lower := make([]float64, len(matrix))
	for i := range width {
		for j := 0; j <= i; j++ {
			value := matrix[i*width+j]
			for k := 0; k < j; k++ {
				value -= lower[i*width+k] * lower[j*width+k]
			}
			if i != j {
				lower[i*width+j] = value / lower[j*width+j]
				continue
			}
			if !finite(value) || value <= 0 {
				return nil, fmt.Errorf("%w: Hessian is not positive definite", ErrNumerical)
			}
			lower[i*width+j] = math.Sqrt(value)
		}
	}
	return lower, nil
}
