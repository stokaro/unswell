package sparsemodel

import (
	"context"
	"fmt"
	"math"
	"slices"
)

type update struct {
	s, y []float64
	rho  float64
}

func optimize(ctx context.Context, d data, theta []float64, work *budget) ([]float64, int, float64, error) {
	current, err := objective(ctx, d, theta, work)
	if err != nil {
		return nil, 0, 0, err
	}
	var history []update
	for iteration := range 500 {
		norm := maxAbs(current.gradient)
		if norm <= 1e-8 {
			return theta, iteration, norm, nil
		}
		if err := work.charge(int64(len(theta)) * int64(20+20*len(history))); err != nil {
			return nil, 0, 0, err
		}
		direction := direction(current.gradient, history)
		nextTheta, next, err := search(ctx, d, theta, current, direction, work)
		if err != nil {
			return nil, 0, 0, err
		}
		history = remember(history, theta, nextTheta, current.gradient, next.gradient)
		theta, current = nextTheta, next
	}
	return nil, 0, 0, fmt.Errorf("sparse fit did not converge in 500 iterations")
}

func direction(gradient []float64, history []update) []float64 {
	result := slices.Clone(gradient)
	alpha := make([]float64, len(history))
	for i := len(history) - 1; i >= 0; i-- {
		u := history[i]
		alpha[i] = u.rho * dot(u.s, result)
		add(result, u.y, -alpha[i])
	}
	if len(history) > 0 {
		u := history[len(history)-1]
		scale := dot(u.s, u.y) / dot(u.y, u.y)
		for i := range result {
			result[i] *= scale
		}
	}
	for i, u := range history {
		beta := u.rho * dot(u.y, result)
		add(result, u.s, alpha[i]-beta)
	}
	for i := range result {
		result[i] = -result[i]
	}
	return result
}

func search(ctx context.Context, d data, theta []float64, current objectiveValue, direction []float64,
	work *budget,
) ([]float64, objectiveValue, error) {
	slope := dot(current.gradient, direction)
	if !finite(slope) || slope >= 0 {
		return nil, objectiveValue{}, fmt.Errorf("sparse fit has no finite descent direction")
	}
	step := 1.0
	for range 40 {
		candidate := slices.Clone(theta)
		add(candidate, direction, step)
		next, err := objective(ctx, d, candidate, work)
		if err != nil {
			return nil, objectiveValue{}, err
		}
		if next.loss <= current.loss+1e-4*step*slope {
			return candidate, next, nil
		}
		step /= 2
	}
	return nil, objectiveValue{}, fmt.Errorf("sparse fit line search did not converge")
}

func remember(history []update, theta, next, gradient, nextGradient []float64) []update {
	s, y := slices.Clone(next), slices.Clone(nextGradient)
	add(s, theta, -1)
	add(y, gradient, -1)
	curvature := dot(s, y)
	if curvature <= 1e-12*math.Sqrt(dot(s, s)*dot(y, y)) || !finite(curvature) {
		return history
	}
	if len(history) == 10 {
		history = history[1:]
	}
	return append(history, update{s: s, y: y, rho: 1 / curvature})
}

func dot(a, b []float64) float64 {
	result := 0.0
	for i, value := range a {
		result += value * b[i]
	}
	return result
}

func add(a, b []float64, scale float64) {
	for i, value := range b {
		a[i] += scale * value
	}
}

func maxAbs(values []float64) float64 {
	result := 0.0
	for _, value := range values {
		result = math.Max(result, math.Abs(value))
	}
	return result
}
