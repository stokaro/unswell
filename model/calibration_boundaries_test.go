package model_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestIsotonicRejectsInvalidSnapshots(t *testing.T) {
	for _, row := range []struct {
		name string
		p    model.IsotonicParameters
	}{
		{"empty", model.IsotonicParameters{}},
		{"missing response", model.IsotonicParameters{Scores: []float64{1}}},
		{"extra response", model.IsotonicParameters{Scores: []float64{1}, Responses: []float64{0, 1}}},
		{"duplicate score", model.IsotonicParameters{Scores: []float64{1, 1}, Responses: []float64{0, 1}}},
		{"descending score", model.IsotonicParameters{Scores: []float64{1, 0}, Responses: []float64{0, 1}}},
		{"descending response", model.IsotonicParameters{Scores: []float64{0, 1}, Responses: []float64{1, 0}}},
		{"nonfinite score", model.IsotonicParameters{Scores: []float64{math.Inf(1)}, Responses: []float64{0}}},
		{"nan response", model.IsotonicParameters{Scores: []float64{0}, Responses: []float64{math.NaN()}}},
		{"negative response", model.IsotonicParameters{Scores: []float64{0}, Responses: []float64{-0.1}}},
		{"excess response", model.IsotonicParameters{Scores: []float64{0}, Responses: []float64{1.1}}},
		{"too many knots", model.IsotonicParameters{Scores: make([]float64, model.MaxCalibrationSamples+1),
			Responses: make([]float64, model.MaxCalibrationSamples+1)}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, err := model.NewIsotonic(row.p)
			c.Assert(err, qt.IsNotNil)
			c.Assert(m, qt.IsNil)
		})
	}
}

func TestIsotonicRejectsInvalidSamples(t *testing.T) {
	for _, row := range []struct {
		name    string
		samples []model.CalibrationSample
	}{
		{"empty", nil},
		{"too many", make([]model.CalibrationSample, model.MaxCalibrationSamples+1)},
		{"nonbinary", []model.CalibrationSample{{Score: 0, Label: 2}}},
		{"negative label", []model.CalibrationSample{{Score: 0, Label: -1}}},
		{"nan", []model.CalibrationSample{{Score: math.NaN(), Label: 1}}},
		{"infinity", []model.CalibrationSample{{Score: math.Inf(1), Label: 1}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			fit, err := model.FitIsotonic(t.Context(), row.samples)
			c.Assert(err, qt.IsNotNil)
			c.Assert(fit, qt.DeepEquals, model.CalibrationFit{})
		})
	}
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	fit, err := model.FitIsotonic(ctx, []model.CalibrationSample{{Score: 0, Label: 1}})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(fit, qt.DeepEquals, model.CalibrationFit{})
}

func TestIsotonicFitsAtTheSampleLimit(t *testing.T) {
	c := qt.New(t)
	rows := make([]model.CalibrationSample, model.MaxCalibrationSamples)
	for i := range rows {
		rows[i] = model.CalibrationSample{Score: float64(i), Label: 1 - i%2}
	}
	fit, err := model.FitIsotonic(t.Context(), rows)
	c.Assert(err, qt.IsNil)
	c.Assert(fit.Samples, qt.Equals, model.MaxCalibrationSamples)
	c.Assert(fit.Pools, qt.Equals, 1)
	c.Assert(fit.MeanSquaredError, qt.Equals, 0.25)
	value, err := fit.Model.Evaluate(t.Context(), 0)
	c.Assert(err, qt.IsNil)
	c.Assert(value, qt.Equals, 0.5)
}
