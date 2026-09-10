package probability

import (
	"context"
	"errors"
	"fmt"

	"github.com/stokaro/unswell/model"
)

// Estimate decides applicability and returns a calibrated probability only for
// an applicable unit. Abstention carries a machine-readable status and no value.
// A vector that does not match the pack's columns is a caller error, not an
// abstention. Errors and cancellation return no partial estimate.
func (p *Pack) Estimate(ctx context.Context, unit Unit) (Estimate, error) {
	if err := ctx.Err(); err != nil {
		return Estimate{}, err
	}
	if unit.Kind != p.file.Kind {
		return abstain(StatusUnsupportedUnit, p.file.Kind), nil
	}
	if unit.Words < p.file.Limits.MinWords {
		return abstain(StatusInsufficientEvidence, fmt.Sprintf("min_words=%d", p.file.Limits.MinWords)), nil
	}
	values, missing, err := p.vector(unit)
	if err != nil {
		return Estimate{}, err
	}
	if missing != "" {
		return abstain(StatusMissingFeature, missing), nil
	}
	return p.evaluate(ctx, values)
}

func (p *Pack) evaluate(ctx context.Context, values []float64) (Estimate, error) {
	evaluation, err := p.classifier.Evaluate(ctx, values)
	if err != nil {
		return Estimate{}, err
	}
	score := evaluation.LinearScore
	response, err := p.calibration.Evaluate(ctx, score)
	if errors.Is(err, model.ErrCalibrationRange) {
		estimate := abstain(StatusCalibrationRange, "")
		estimate.LinearScore = &score
		return estimate, nil
	}
	if err != nil {
		return Estimate{}, err
	}
	if err := ctx.Err(); err != nil {
		return Estimate{}, err
	}
	return Estimate{Status: StatusAvailable, Probability: &response, LinearScore: &score}, nil
}

// vector returns the ordered numeric inputs, or the first unavailable column.
func (p *Pack) vector(unit Unit) ([]float64, string, error) {
	columns := p.file.Contract.Columns
	if len(unit.Values) != len(columns) {
		return nil, "", fmt.Errorf("probability pack %s requires %d measured columns", p.file.ID, len(columns))
	}
	values := make([]float64, 0, len(columns))
	for i, column := range columns {
		value := unit.Values[i]
		if value.ID != column.ID || value.Version != column.Version {
			return nil, "", fmt.Errorf("probability pack %s expects column %s version %s at position %d",
				p.file.ID, column.ID, column.Version, i)
		}
		if value.Number == nil {
			return nil, missingDetail(column.ID, value.Reason), nil
		}
		values = append(values, *value.Number)
	}
	return values, "", nil
}

func missingDetail(id, reason string) string {
	if reason == "" {
		return id
	}
	return id + "/" + reason
}

func abstain(status, detail string) Estimate {
	return Estimate{Status: status, Detail: detail}
}
