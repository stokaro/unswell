// Package model trains numerical classifiers and separate calibration mappings in Go.
// Numerical outputs alone do not qualify prose models or editorial probabilities.
package model

import (
	"context"
	"fmt"
	"math"
	"slices"
)

// MaxFeatures bounds numerical vector width and the dense logistic Hessian.
const MaxFeatures = 128

// Parameters describes a logistic model over centered and scaled numeric inputs.
// Feature identities, data permissions, and calibration are separate contracts.
type Parameters struct {
	Means, Scales, Weights []float64
	Intercept              float64
}

// Logistic owns immutable parameters. Its zero value cannot evaluate inputs.
type Logistic struct {
	parameters Parameters
}

// Evaluation explains a linear score and its uncalibrated sigmoid response.
// Contributions are in linear-score units, not probability percentage points.
type Evaluation struct {
	Normalized, Contributions []float64
	Intercept, LinearScore    float64
	Response                  float64
}

// NewLogistic validates and copies numerical parameters without loading resources.
func NewLogistic(parameters Parameters) (*Logistic, error) {
	if !validParameterShape(parameters) {
		return nil, fmt.Errorf("logistic parameter dimensions must match and be within 1..%d", MaxFeatures)
	}
	if !finite(parameters.Intercept) {
		return nil, fmt.Errorf("logistic intercept must be finite")
	}
	for i := range parameters.Weights {
		if !finite(parameters.Weights[i]) || !finite(parameters.Means[i]) ||
			!finite(parameters.Scales[i]) || parameters.Scales[i] <= 0 {
			return nil, fmt.Errorf("logistic column %d requires finite parameters and a positive scale", i)
		}
	}
	return &Logistic{parameters: cloneParameters(parameters)}, nil
}

func validParameterShape(p Parameters) bool {
	width := len(p.Weights)
	return width >= 1 && width <= MaxFeatures && len(p.Means) == width && len(p.Scales) == width
}

// Parameters returns an independent numerical snapshot, not a qualified model pack.
func (m *Logistic) Parameters() Parameters {
	if m == nil {
		return Parameters{}
	}
	return cloneParameters(m.parameters)
}

// Evaluate computes owned contributions in feature order. Missing inputs are errors.
// Errors and cancellation return no partial evaluation. Concurrent calls are safe.
func (m *Logistic) Evaluate(ctx context.Context, values []float64) (Evaluation, error) {
	if err := ctx.Err(); err != nil {
		return Evaluation{}, err
	}
	if m == nil || len(m.parameters.Weights) == 0 || len(values) != len(m.parameters.Weights) {
		return Evaluation{}, fmt.Errorf("logistic evaluation requires a model and matching input dimensions")
	}
	result, err := m.evaluate(values)
	if err != nil {
		return Evaluation{}, err
	}
	if err := ctx.Err(); err != nil {
		return Evaluation{}, err
	}
	return result, nil
}

func (m *Logistic) evaluate(values []float64) (Evaluation, error) {
	p := m.parameters
	result := Evaluation{Normalized: make([]float64, len(values)), Contributions: make([]float64, len(values)),
		Intercept: p.Intercept, LinearScore: p.Intercept}
	for i, value := range values {
		if !finite(value) {
			return Evaluation{}, fmt.Errorf("logistic input %d must be finite; missing values require an explicit contract", i)
		}
		result.Normalized[i] = (value - p.Means[i]) / p.Scales[i]
		result.Contributions[i] = result.Normalized[i] * p.Weights[i]
		result.LinearScore += result.Contributions[i]
		if !finite(result.Normalized[i]) || !finite(result.Contributions[i]) || !finite(result.LinearScore) {
			return Evaluation{}, fmt.Errorf("logistic evaluation overflow at column %d", i)
		}
	}
	result.Response = sigmoid(result.LinearScore)
	return result, nil
}

func cloneParameters(p Parameters) Parameters {
	return Parameters{Means: slices.Clone(p.Means), Scales: slices.Clone(p.Scales),
		Weights: slices.Clone(p.Weights), Intercept: p.Intercept}
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func sigmoid(value float64) float64 {
	if value >= 0 {
		return 1 / (1 + math.Exp(-value))
	}
	exp := math.Exp(value)
	return exp / (1 + exp)
}
