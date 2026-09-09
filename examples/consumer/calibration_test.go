package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestSeparateCalibrationPublicAPI(t *testing.T) {
	c := qt.New(t)
	// Disjoint scripted numeric inputs demonstrate the two stages, not editorial quality.
	training := []model.Example{{Values: []float64{-1}, Label: 0}, {Values: []float64{1}, Label: 1}}
	fit, err := model.FitLogistic(t.Context(), training,
		model.FitOptions{L2: 0.1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 100000})
	c.Assert(err, qt.IsNil)
	calibration := []model.CalibrationSample{}
	for i, input := range []float64{-0.5, 0.5} {
		value, err := fit.Model.Evaluate(t.Context(), []float64{input})
		c.Assert(err, qt.IsNil)
		calibration = append(calibration, model.CalibrationSample{Score: value.LinearScore, Label: i})
	}
	mapping, err := model.FitIsotonic(t.Context(), calibration)
	c.Assert(err, qt.IsNil)
	restored, err := model.NewIsotonic(mapping.Model.Parameters())
	c.Assert(err, qt.IsNil)
	value, err := fit.Model.Evaluate(t.Context(), []float64{0})
	c.Assert(err, qt.IsNil)
	response, err := restored.Evaluate(t.Context(), value.LinearScore)
	c.Assert(err, qt.IsNil)
	c.Assert(response, qt.Equals, 0.5)
}
