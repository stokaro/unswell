// Package sparsemodel fits a bounded sparse logistic research baseline.
// It does not qualify probabilities or produce an installable model pack.
package sparsemodel

import (
	"context"
	"fmt"
	"math"

	"github.com/stokaro/unswell/model"
)

// Algorithm identifies the fixed objective and deterministic optimizer.
const Algorithm = "unswell-research-sparse-logistic-lbfgs-v1"

// MaxFeatures bounds the research vocabulary independently of dense model packs.
const MaxFeatures = 8192

// Entry is one nonzero column in a sparse row.
type Entry struct {
	Index int
	Value float64
}

// Example contains sorted, unique columns and an explicit binary label.
type Example struct {
	Values []Entry
	Label  int
}

// Result owns fitted parameters and the actual convergence and operation counts.
type Result struct {
	Parameters model.Parameters
	Iterations int
	Operations int64
	Gradient   float64
}

type data struct {
	rows          []Example
	means, scales []float64
	constant      []bool
}

type budget struct{ operations int64 }

func (b *budget) charge(n int64) error {
	if n < 0 || b.operations > 3_000_000_000-n {
		return fmt.Errorf("sparse fit exceeds operation budget")
	}
	b.operations += n
	return nil
}

// Fit minimizes mean logistic loss plus 0.01/2 times squared standardized slopes.
// The intercept is unpenalized; only training rows determine normalization.
func Fit(ctx context.Context, examples []Example, features int) (Result, error) {
	prepared, positive, err := prepare(ctx, examples, features)
	if err != nil {
		return Result{}, err
	}
	theta := make([]float64, features+1)
	theta[0] = math.Log(float64(positive) / float64(len(examples)-positive))
	work := new(budget)
	theta, iterations, gradient, err := optimize(ctx, prepared, theta, work)
	if err != nil {
		return Result{}, err
	}
	return Result{Parameters: model.Parameters{Means: prepared.means, Scales: prepared.scales,
		Weights: theta[1:], Intercept: theta[0]}, Iterations: iterations, Operations: work.operations, Gradient: gradient}, nil
}

// Score returns a raw linear score, not a calibrated probability.
func (r Result) Score(ctx context.Context, values []Entry) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	p := r.Parameters
	if err := validateParameters(p); err != nil {
		return 0, err
	}
	if err := validateRow(values, len(p.Weights)); err != nil {
		return 0, err
	}
	value := p.Intercept
	for j, weight := range p.Weights {
		value -= weight * p.Means[j] / p.Scales[j]
	}
	for _, entry := range values {
		value += p.Weights[entry.Index] * entry.Value / p.Scales[entry.Index]
	}
	if !finite(value) {
		return 0, fmt.Errorf("nonfinite sparse score")
	}
	return value, nil
}

func validateParameters(p model.Parameters) error {
	if len(p.Weights) == 0 || len(p.Weights) > MaxFeatures || len(p.Weights) != len(p.Means) ||
		len(p.Weights) != len(p.Scales) || !finite(p.Intercept) {
		return fmt.Errorf("invalid sparse model dimensions or intercept")
	}
	for j, weight := range p.Weights {
		if !validColumn(weight, p.Means[j], p.Scales[j]) {
			return fmt.Errorf("invalid sparse model parameters")
		}
	}
	return nil
}

func validColumn(weight, mean, scale float64) bool {
	return finite(weight) && finite(mean) && finite(scale) && scale > 0
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
