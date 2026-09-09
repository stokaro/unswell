package model

import (
	"context"
	"errors"
	"fmt"
	"math"
)

// Algorithm identifies normalization, objective, initialization, and optimizer order.
const Algorithm = "unswell-logistic-newton-v1"

// ErrConvergence means the requested gradient tolerance was not reached.
var ErrConvergence = errors.New("logistic optimization did not converge")

// ErrBudget means the requested operation budget was exhausted.
var ErrBudget = errors.New("logistic operation budget exhausted")

// ErrNumerical means finite arithmetic or a positive-definite Newton step failed.
var ErrNumerical = errors.New("logistic numerical failure")

// Example is one ordered complete numeric vector and a binary 0/1 target.
// The caller establishes feature compatibility, label validity, and training rights.
type Example struct {
	Values []float64
	Label  int
}

// FitOptions makes convergence and computational limits explicit.
// L2 penalizes weights only. MaxOperations counts the documented logical work.
type FitOptions struct {
	L2, Tolerance float64
	MaxIterations int
	MaxOperations int64
}

// Validate checks the shared fitting limits without reading or preparing rows.
func (o FitOptions) Validate() error { return validateOptions(o) }

// FitResult records reproducible numerical training, not corpus qualification.
// InputSHA256 binds ordered raw vectors and labels; Options records tuning inputs.
type FitResult struct {
	Model                  *Logistic
	Algorithm, InputSHA256 string
	Options                FitOptions
	Iterations             int
	Operations             int64
	Loss, GradientNorm     float64
}

type workBudget struct {
	used, limit int64
}

func (b *workBudget) charge(cost int64) error {
	if cost > b.limit-b.used {
		return ErrBudget
	}
	b.used += cost
	return nil
}

// FitLogistic learns a regularized binary model using only supplied training rows.
// It standardizes with training statistics and requires both classes. Input order
// is fixed, no randomness is used, and errors return no partial model or result.
func FitLogistic(ctx context.Context, examples []Example, options FitOptions) (FitResult, error) {
	if err := options.Validate(); err != nil {
		return FitResult{}, err
	}
	budget := workBudget{limit: options.MaxOperations}
	data, err := prepareTraining(ctx, examples, &budget)
	if err != nil {
		return FitResult{}, err
	}
	theta := make([]float64, len(data.means)+1)
	theta[0] = math.Log(float64(data.positive) / float64(len(data.rows)-data.positive))
	for iteration := 0; iteration <= options.MaxIterations; iteration++ {
		current, err := objective(ctx, data, theta, options.L2, true, &budget)
		if err != nil {
			return FitResult{}, err
		}
		norm := infinityNorm(current.gradient)
		if norm <= options.Tolerance {
			return fitted(data, theta, options, budget.used, iteration, current.loss, norm)
		}
		if iteration == options.MaxIterations {
			return FitResult{}, ErrConvergence
		}
		theta, err = newtonStep(ctx, data, theta, current, options.L2, &budget)
		if err != nil {
			return FitResult{}, err
		}
	}
	return FitResult{}, ErrConvergence
}

func validateOptions(options FitOptions) error {
	if !finite(options.L2) || options.L2 < 1e-8 || options.L2 > 1e3 {
		return fmt.Errorf("logistic L2 must be within 1e-8..1e3")
	}
	if !finite(options.Tolerance) || options.Tolerance < 1e-12 || options.Tolerance > 0.1 {
		return fmt.Errorf("logistic tolerance must be within 1e-12..0.1")
	}
	return validateFitLimits(options)
}

func validateFitLimits(options FitOptions) error {
	if options.MaxIterations < 1 || options.MaxIterations > 1000 || options.MaxOperations < 1 || options.MaxOperations > 1e12 {
		return fmt.Errorf("logistic iteration or operation limits are invalid")
	}
	return nil
}

func fitted(data trainingData, theta []float64, options FitOptions, operations int64, iterations int, loss, norm float64) (
	FitResult, error,
) {
	m, err := NewLogistic(Parameters{Means: data.means, Scales: data.scales, Weights: theta[1:], Intercept: theta[0]})
	if err != nil {
		return FitResult{}, err
	}
	return FitResult{Model: m, Algorithm: Algorithm, InputSHA256: data.inputHash, Options: options,
		Iterations: iterations, Operations: operations, Loss: loss, GradientNorm: norm}, nil
}

func infinityNorm(values []float64) float64 {
	norm := 0.0
	for _, value := range values {
		norm = math.Max(norm, math.Abs(value))
	}
	return norm
}
