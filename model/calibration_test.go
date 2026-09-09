package model_test

import (
	"math"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestIsotonicPoolingAndTies(t *testing.T) {
	for _, row := range []struct {
		name      string
		samples   []model.CalibrationSample
		scores    []float64
		responses []float64
		pools     int
		loss      float64
	}{
		{"ordered", []model.CalibrationSample{{Score: -1, Label: 0}, {Score: 1, Label: 1}},
			[]float64{-1, 1}, []float64{0, 1}, 2, 0},
		{"inverted", []model.CalibrationSample{{Score: 0, Label: 1}, {Score: 1, Label: 0}},
			[]float64{0, 1}, []float64{0.5, 0.5}, 1, 0.25},
		{"cascade", []model.CalibrationSample{{Score: 0, Label: 0}, {Score: 1, Label: 1}, {Score: 2, Label: 1},
			{Score: 3, Label: 0}, {Score: 4, Label: 0}, {Score: 5, Label: 1}},
			[]float64{0, 1, 2, 3, 4, 5}, []float64{0, 0.5, 0.5, 0.5, 0.5, 1}, 3, 1.0 / 6},
		{"weighted ties", []model.CalibrationSample{{Score: 0, Label: 0}, {Score: 0, Label: 1}, {Score: 0, Label: 1},
			{Score: 1, Label: 0}, {Score: 2, Label: 1}, {Score: 2, Label: 1}},
			[]float64{0, 1, 2}, []float64{0.5, 0.5, 1}, 2, 1.0 / 6},
		{"constant score", []model.CalibrationSample{{Score: 3, Label: 1}, {Score: 3, Label: 0}, {Score: 3, Label: 1}},
			[]float64{3}, []float64{2.0 / 3}, 1, 2.0 / 9},
		{"equal pools", []model.CalibrationSample{{Score: 0, Label: 0}, {Score: 1, Label: 0},
			{Score: 2, Label: 1}, {Score: 3, Label: 1}}, []float64{0, 1, 2, 3}, []float64{0, 0, 1, 1}, 2, 0},
		{"only negative", []model.CalibrationSample{{Score: 0, Label: 0}, {Score: 1, Label: 0}},
			[]float64{0, 1}, []float64{0, 0}, 1, 0},
		{"one positive", []model.CalibrationSample{{Score: 3, Label: 1}}, []float64{3}, []float64{1}, 1, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			fit, err := model.FitIsotonic(t.Context(), row.samples)
			c.Assert(err, qt.IsNil)
			c.Assert(fit.Model.Parameters(), qt.DeepEquals, model.IsotonicParameters{Scores: row.scores, Responses: row.responses})
			c.Assert(fit.Algorithm, qt.Equals, model.IsotonicAlgorithm)
			c.Assert(fit.Samples, qt.Equals, len(row.samples))
			c.Assert(fit.DistinctScores, qt.Equals, len(row.scores))
			c.Assert(fit.Pools, qt.Equals, row.pools)
			c.Assert(math.Abs(fit.MeanSquaredError-row.loss) < 1e-15, qt.IsTrue)
			measured := 0.0
			for _, sample := range row.samples {
				value, err := fit.Model.Evaluate(t.Context(), sample.Score)
				c.Assert(err, qt.IsNil)
				delta := value - float64(sample.Label)
				measured += delta * delta / float64(len(row.samples))
			}
			c.Assert(math.Abs(measured-fit.MeanSquaredError) < 1e-15, qt.IsTrue)
		})
	}
}

func TestIsotonicOrderingHashAndInputOwnership(t *testing.T) {
	c := qt.New(t)
	rows := []model.CalibrationSample{{Score: 2, Label: 1}, {Score: 0, Label: 1}, {Score: 1, Label: 0}}
	original := slices.Clone(rows)
	first, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(rows, qt.DeepEquals, original)
	again, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(again.Model.Parameters(), qt.DeepEquals, first.Model.Parameters())
	expected := first
	again.Model, expected.Model = nil, nil
	c.Assert(again, qt.DeepEquals, expected)
	slices.Reverse(rows)
	reversed, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(reversed.Model.Parameters(), qt.DeepEquals, first.Model.Parameters())
	c.Assert(reversed.InputSHA256, qt.Not(qt.Equals), first.InputSHA256)
	rows[0] = model.CalibrationSample{Score: -999, Label: 1}
	c.Assert(first.Model.Parameters().Scores, qt.DeepEquals, []float64{0, 1, 2})
}

func TestIsotonicCanonicalizesZeroKnots(t *testing.T) {
	c := qt.New(t)
	negativeZero := math.Copysign(0, -1)
	rows := []model.CalibrationSample{{Score: negativeZero, Label: 1}, {Score: 0, Label: 0}}
	fit, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(fit.DistinctScores, qt.Equals, 1)
	c.Assert(math.Signbit(fit.Model.Parameters().Scores[0]), qt.IsFalse)
	rows[0].Score = 0
	positive, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(fit.Model.Parameters(), qt.DeepEquals, positive.Model.Parameters())
	c.Assert(fit.InputSHA256, qt.Not(qt.Equals), positive.InputSHA256)
}
