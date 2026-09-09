package model_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

// The ordered isotonic fit also equals max(lower) min(upper) interval means.
// Enumerating all intervals supplies an independent small-input oracle for PAV.
func intervalOracle(rows []model.CalibrationSample, target float64) float64 {
	result := 0.0
	for _, lower := range rows {
		if lower.Score > target {
			continue
		}
		candidate := 1.0
		for _, upper := range rows {
			if upper.Score >= target {
				candidate = math.Min(candidate, intervalMean(rows, lower.Score, upper.Score))
			}
		}
		result = math.Max(result, candidate)
	}
	return result
}

func intervalMean(rows []model.CalibrationSample, lower, upper float64) float64 {
	count, positive := 0, 0
	for _, row := range rows {
		if row.Score >= lower && row.Score <= upper {
			count++
			positive += row.Label
		}
	}
	return float64(positive) / float64(count)
}

func TestIsotonicMatchesIntervalOracle(t *testing.T) {
	c := qt.New(t)
	// Exhaustive scripted labels, including tied scores; no human corpus is used.
	scores := []float64{-2, -2, 0, 1, 1, 3}
	for labels := range 1 << len(scores) {
		rows := make([]model.CalibrationSample, len(scores))
		for i, score := range scores {
			rows[i] = model.CalibrationSample{Score: score, Label: (labels >> i) & 1}
		}
		fit, err := model.FitIsotonic(t.Context(), rows)
		c.Assert(err, qt.IsNil)
		for _, score := range scores {
			got, err := fit.Model.Evaluate(t.Context(), score)
			c.Assert(err, qt.IsNil)
			c.Assert(math.Abs(got-intervalOracle(rows, score)) < 1e-15, qt.IsTrue,
				qt.Commentf("labels=%d, score=%g", labels, score))
		}
	}
}

func BenchmarkFitIsotonic(b *testing.B) {
	rows := make([]model.CalibrationSample, 5000)
	for i := range rows {
		rows[i] = model.CalibrationSample{Score: float64((i * 17) % 1001), Label: i % 2}
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := model.FitIsotonic(b.Context(), rows); err != nil {
			b.Fatal(err)
		}
	}
}
