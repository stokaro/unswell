package model_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func fitOptions() model.FitOptions {
	return model.FitOptions{L2: 1 / (1 + math.E), Tolerance: 1e-10, MaxIterations: 100, MaxOperations: 1000000}
}

// These numeric fixtures have scripted binary targets, not human editorial labels.
func symmetricData() []model.Example {
	return []model.Example{{Values: []float64{8, 7}, Label: 0}, {Values: []float64{12, 7}, Label: 1}}
}

func TestFitLogisticKnownRegularizedOptimum(t *testing.T) {
	c := qt.New(t)
	data, options := symmetricData(), fitOptions()
	result, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	p := result.Model.Parameters()
	c.Assert(p.Means, qt.DeepEquals, []float64{10, 7})
	c.Assert(p.Scales, qt.DeepEquals, []float64{2, 1})
	// At w=1, mean loss derivative is -1/(1+e); L2*w cancels it.
	c.Assert(math.Abs(p.Weights[0]-(1.0)) <= 1e-8, qt.IsTrue)
	c.Assert(math.Abs(p.Weights[1]-(0.0)) <= 1e-12, qt.IsTrue)
	c.Assert(math.Abs(p.Intercept-(0.0)) <= 1e-12, qt.IsTrue)
	c.Assert(math.Abs(result.Loss-(math.Log1p(math.Exp(-1))+options.L2/2)) <= 1e-12, qt.IsTrue)
	c.Assert(result.GradientNorm <= options.Tolerance, qt.IsTrue)
	c.Assert(result.Algorithm, qt.Equals, model.Algorithm)
	c.Assert(result.Options, qt.DeepEquals, options)
	c.Assert(result.InputSHA256, qt.HasLen, 64)
	again, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	c.Assert(again.Model.Parameters(), qt.DeepEquals, result.Model.Parameters())
	metadata := result
	again.Model, metadata.Model = nil, nil
	c.Assert(again, qt.DeepEquals, metadata)
	data[0].Values[0] = -500
	prediction, err := result.Model.Evaluate(t.Context(), []float64{12, 7})
	c.Assert(err, qt.IsNil)
	c.Assert(math.Abs(prediction.Response-(1/(1+math.Exp(-1)))) <= 1e-8, qt.IsTrue)
}

func TestFitInterceptIsNotPenalized(t *testing.T) {
	c := qt.New(t)
	data := []model.Example{
		{Values: []float64{7}, Label: 0}, {Values: []float64{7}, Label: 1},
		{Values: []float64{7}, Label: 1}, {Values: []float64{7}, Label: 1},
	}
	options := fitOptions()
	options.L2 = 100
	result, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	c.Assert(math.Abs(result.Model.Parameters().Intercept-(math.Log(3))) <= 1e-12, qt.IsTrue)
	c.Assert(result.Model.Parameters().Weights, qt.DeepEquals, []float64{0})
	value, err := result.Model.Evaluate(t.Context(), []float64{7})
	c.Assert(err, qt.IsNil)
	c.Assert(math.Abs(value.Response-(0.75)) <= 1e-12, qt.IsTrue)
}

func TestRegularizationAndTrainingHash(t *testing.T) {
	c := qt.New(t)
	data, options := symmetricData(), fitOptions()
	weak, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	options.L2 = 10
	strong, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	c.Assert(strong.Model.Parameters().Weights[0] < weak.Model.Parameters().Weights[0], qt.IsTrue)
	c.Assert(strong.InputSHA256, qt.Equals, weak.InputSHA256)
	data[0], data[1] = data[1], data[0]
	reordered, err := model.FitLogistic(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	c.Assert(reordered.InputSHA256, qt.Not(qt.Equals), strong.InputSHA256)
}

func TestFitNeverReturnsPartialModels(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*model.FitOptions)
		want error
	}{
		{"work limit", func(o *model.FitOptions) { o.MaxOperations = 1 }, model.ErrBudget},
		{"Newton work limit", func(o *model.FitOptions) { o.MaxOperations = 43 }, model.ErrBudget},
		{"iteration limit", func(o *model.FitOptions) { o.MaxIterations = 1 }, model.ErrConvergence},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			options := fitOptions()
			row.edit(&options)
			result, err := model.FitLogistic(t.Context(), symmetricData(), options)
			c.Assert(err, qt.ErrorIs, row.want)
			c.Assert(result, qt.DeepEquals, model.FitResult{})
		})
	}
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := model.FitLogistic(ctx, symmetricData(), fitOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, model.FitResult{})
}
