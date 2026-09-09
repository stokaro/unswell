package model_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestInvalidLogisticParameters(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*model.Parameters)
	}{
		{"dimensions", func(p *model.Parameters) { p.Weights = []float64{1, 2} }},
		{"mean", func(p *model.Parameters) { p.Means[0] = math.NaN() }},
		{"scale zero", func(p *model.Parameters) { p.Scales[0] = 0 }},
		{"scale negative", func(p *model.Parameters) { p.Scales[0] = -1 }},
		{"scale infinity", func(p *model.Parameters) { p.Scales[0] = math.Inf(1) }},
		{"weight", func(p *model.Parameters) { p.Weights[0] = math.NaN() }},
		{"intercept", func(p *model.Parameters) { p.Intercept = math.Inf(-1) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			p := model.Parameters{Means: []float64{0}, Scales: []float64{1}, Weights: []float64{1}}
			row.edit(&p)
			m, err := model.NewLogistic(p)
			c.Assert(err, qt.IsNotNil)
			c.Assert(m, qt.IsNil)
		})
	}
}

func TestLogisticOverflowReturnsNoEvaluation(t *testing.T) {
	c := qt.New(t)
	m, err := model.NewLogistic(model.Parameters{Means: []float64{0}, Scales: []float64{1}, Weights: []float64{math.MaxFloat64}})
	c.Assert(err, qt.IsNil)
	result, err := m.Evaluate(t.Context(), []float64{2})
	c.Assert(err, qt.ErrorMatches, "logistic evaluation overflow.*")
	c.Assert(result, qt.DeepEquals, model.Evaluation{})
}

func TestInvalidTrainingInputs(t *testing.T) {
	for _, row := range []struct {
		name string
		data []model.Example
	}{
		{"empty", nil},
		{"single", []model.Example{{Values: []float64{1}, Label: 1}}},
		{"no features", []model.Example{{Label: 0}, {Label: 1}}},
		{"nonbinary", []model.Example{{Values: []float64{1}, Label: 2}, {Values: []float64{2}, Label: 0}}},
		{"missing class", []model.Example{{Values: []float64{1}, Label: 0}, {Values: []float64{2}, Label: 0}}},
		{"dimensions", []model.Example{{Values: []float64{1}, Label: 0}, {Values: []float64{2, 3}, Label: 1}}},
		{"missing", []model.Example{{Values: []float64{math.NaN()}, Label: 0}, {Values: []float64{2}, Label: 1}}},
		{"nonfinite", []model.Example{{Values: []float64{math.Inf(1)}, Label: 0}, {Values: []float64{2}, Label: 1}}},
		{"range", []model.Example{{Values: []float64{1e13}, Label: 0}, {Values: []float64{2}, Label: 1}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := model.FitLogistic(t.Context(), row.data, fitOptions())
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, model.FitResult{})
		})
	}
}

func TestVarianceUnderflowIsNotAConstantFeature(t *testing.T) {
	c := qt.New(t)
	data := []model.Example{{Values: []float64{-1e-200}, Label: 0}, {Values: []float64{1e-200}, Label: 1}}
	result, err := model.FitLogistic(t.Context(), data, fitOptions())
	c.Assert(err, qt.ErrorIs, model.ErrNumerical)
	c.Assert(result, qt.DeepEquals, model.FitResult{})
}

func TestInvalidFitOptions(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*model.FitOptions)
	}{
		{"unregularized", func(o *model.FitOptions) { o.L2 = 0 }},
		{"nonfinite penalty", func(o *model.FitOptions) { o.L2 = math.NaN() }},
		{"negative tolerance", func(o *model.FitOptions) { o.Tolerance = -1 }},
		{"nonfinite tolerance", func(o *model.FitOptions) { o.Tolerance = math.Inf(1) }},
		{"iterations", func(o *model.FitOptions) { o.MaxIterations = 0 }},
		{"operations", func(o *model.FitOptions) { o.MaxOperations = 0 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			options := fitOptions()
			row.edit(&options)
			result, err := model.FitLogistic(t.Context(), symmetricData(), options)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, model.FitResult{})
		})
	}
}
