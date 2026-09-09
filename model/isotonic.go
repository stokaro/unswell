package model

import (
	"context"
	"errors"
	"fmt"
	"math"
	"slices"
)

// MaxCalibrationSamples bounds fitting rows and restored isotonic knots.
const MaxCalibrationSamples = 100000

// ErrCalibrationRange means a score is outside the fitted knot range.
// This numerical boundary does not establish prose-domain applicability.
var ErrCalibrationRange = errors.New("score outside fitted calibration range")

// IsotonicParameters describes an increasing piecewise-linear numerical mapping.
// Scores increase strictly; Responses are nondecreasing and within [0,1].
type IsotonicParameters struct {
	Scores, Responses []float64
}

// Isotonic owns calibration knots. Its zero value cannot evaluate a score.
// A numerical fit does not qualify editorial probabilities or a model pack.
type Isotonic struct {
	parameters IsotonicParameters
}

// NewIsotonic validates and copies an isotonic parameter snapshot without I/O.
func NewIsotonic(parameters IsotonicParameters) (*Isotonic, error) {
	count := len(parameters.Scores)
	if count < 1 || count > MaxCalibrationSamples || len(parameters.Responses) != count {
		return nil, fmt.Errorf("isotonic parameters require matching dimensions within 1..%d", MaxCalibrationSamples)
	}
	for i, score := range parameters.Scores {
		response := parameters.Responses[i]
		if !validCalibrationKnot(score, response) {
			return nil, fmt.Errorf("isotonic knot %d requires a finite score and response within [0,1]", i)
		}
		if i > 0 && (score <= parameters.Scores[i-1] || response < parameters.Responses[i-1]) {
			return nil, fmt.Errorf("isotonic knots require increasing scores and nondecreasing responses")
		}
	}
	return &Isotonic{parameters: cloneIsotonic(parameters)}, nil
}

func validCalibrationKnot(score, response float64) bool {
	return finite(score) && finite(response) && response >= 0 && response <= 1
}

// Parameters returns an owned numerical snapshot, without qualification metadata.
func (m *Isotonic) Parameters() IsotonicParameters {
	if m == nil {
		return IsotonicParameters{}
	}
	return cloneIsotonic(m.parameters)
}

func cloneIsotonic(p IsotonicParameters) IsotonicParameters {
	return IsotonicParameters{Scores: slices.Clone(p.Scores), Responses: slices.Clone(p.Responses)}
}

// Evaluate interpolates a finite score within the observed knot range.
// Errors return zero, which is not an available estimate. Concurrent calls are safe.
func (m *Isotonic) Evaluate(ctx context.Context, score float64) (float64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if m == nil || len(m.parameters.Scores) == 0 || !finite(score) {
		return 0, fmt.Errorf("isotonic evaluation requires a model and a finite score")
	}
	p := m.parameters
	if score < p.Scores[0] || score > p.Scores[len(p.Scores)-1] {
		return 0, ErrCalibrationRange
	}
	index, exact := slices.BinarySearch(p.Scores, score)
	response := p.Responses[index]
	if !exact {
		fraction := intervalFraction(p.Scores[index-1], p.Scores[index], score)
		response = p.Responses[index-1] + fraction*(response-p.Responses[index-1])
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	return response, nil
}

func intervalFraction(left, right, value float64) float64 {
	width := right - left
	if math.IsInf(width, 1) {
		return (value/2 - left/2) / (right/2 - left/2)
	}
	return (value - left) / width
}
