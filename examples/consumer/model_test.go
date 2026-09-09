package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestNumericalModelPublicAPI(t *testing.T) {
	c := qt.New(t)
	// Scripted numeric labels exercise the API; they are not editorial annotations.
	data := []model.Example{{Values: []float64{-1}, Label: 0}, {Values: []float64{1}, Label: 1}}
	options := model.FitOptions{L2: 0.1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 100000}
	c.Assert(options.Validate(), qt.IsNil)
	c.Assert((model.FitOptions{}).Validate(), qt.IsNotNil)
	fit, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	restored, err := model.NewLogistic(fit.Model.Parameters())
	c.Assert(err, qt.IsNil)
	value, err := restored.Evaluate(t.Context(), []float64{1})
	c.Assert(err, qt.IsNil)
	c.Assert(value.LinearScore > 0, qt.IsTrue)
	c.Assert(value.Contributions, qt.HasLen, 1)
}
