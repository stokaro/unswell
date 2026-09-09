package model

import (
	"cmp"
	"context"
	"crypto/sha256"
	"fmt"
	"math"
	"slices"
)

// IsotonicAlgorithm versions tie handling, pooling, and interpolation semantics.
const IsotonicAlgorithm = "unswell-isotonic-pav-linear-v1"

// CalibrationSample pairs a frozen classifier's score with a binary target.
// Callers supply an independent calibration partition and verify target identity.
type CalibrationSample struct {
	Score float64
	Label int
}

// CalibrationFit records a numerical fit, not held-out quality or qualification.
// MeanSquaredError is measured on the supplied calibration rows only.
type CalibrationFit struct {
	Model                          *Isotonic
	Algorithm, InputSHA256         string
	Samples, DistinctScores, Pools int
	MeanSquaredError               float64
}

type calibrationGroup struct {
	score           float64
	count, positive int
}

// FitIsotonic fits a nondecreasing squared-error mapping with separate input rows.
// It does not choose partitions, retrain a classifier, or measure held-out quality.
// The input is copied; cancellation and errors return no partial fit.
func FitIsotonic(ctx context.Context, samples []CalibrationSample) (CalibrationFit, error) {
	rows, hash, err := calibrationRows(ctx, samples)
	if err != nil {
		return CalibrationFit{}, err
	}
	groups, err := groupCalibration(ctx, rows)
	if err != nil {
		return CalibrationFit{}, err
	}
	parameters, pools, loss, err := poolCalibration(ctx, groups)
	if err != nil {
		return CalibrationFit{}, err
	}
	m, err := NewIsotonic(parameters)
	if err != nil {
		return CalibrationFit{}, err
	}
	if err := ctx.Err(); err != nil {
		return CalibrationFit{}, err
	}
	return CalibrationFit{Model: m, Algorithm: IsotonicAlgorithm, InputSHA256: hash, Samples: len(samples),
		DistinctScores: len(groups), Pools: pools, MeanSquaredError: loss / float64(len(samples))}, nil
}

func calibrationRows(ctx context.Context, samples []CalibrationSample) ([]CalibrationSample, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	if len(samples) < 1 || len(samples) > MaxCalibrationSamples {
		return nil, "", fmt.Errorf("isotonic fitting requires 1..%d samples", MaxCalibrationSamples)
	}
	rows := make([]CalibrationSample, len(samples))
	digest := sha256.New()
	_, _ = digest.Write([]byte("unswell-isotonic-input-v1\x00"))
	hashInteger(digest, uint64(len(samples)))
	for i, row := range samples {
		if err := ctx.Err(); err != nil {
			return nil, "", err
		}
		if !validCalibrationSample(row) {
			return nil, "", fmt.Errorf("calibration row %d requires a finite score and binary label", i)
		}
		hashInteger(digest, math.Float64bits(row.Score))
		label := uint64(0)
		if row.Label == 1 {
			label = 1
		}
		hashInteger(digest, label)
		if row.Score == 0 {
			row.Score = 0 // Canonicalize negative zero in fitted knots, not the input hash.
		}
		rows[i] = row
	}
	return rows, fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func validCalibrationSample(row CalibrationSample) bool {
	return finite(row.Score) && (row.Label == 0 || row.Label == 1)
}

func groupCalibration(ctx context.Context, rows []CalibrationSample) ([]calibrationGroup, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	slices.SortFunc(rows, func(a, b CalibrationSample) int { return cmp.Compare(a.Score, b.Score) })
	groups := make([]calibrationGroup, 0, len(rows))
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		last := len(groups) - 1
		if last >= 0 && groups[last].score == row.Score {
			groups[last].count++
			groups[last].positive += row.Label
		} else {
			groups = append(groups, calibrationGroup{score: row.Score, count: 1, positive: row.Label})
		}
	}
	return groups, nil
}
