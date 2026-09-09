package model_test

import (
	"context"
	"math"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestLogisticExplainsNormalizedContributions(t *testing.T) {
	c := qt.New(t)
	parameters := model.Parameters{Means: []float64{10, 5}, Scales: []float64{2, 1}, Weights: []float64{3, -2}, Intercept: 0.5}
	m, err := model.NewLogistic(parameters)
	c.Assert(err, qt.IsNil)
	parameters.Weights[0] = 99
	result, err := m.Evaluate(t.Context(), []float64{12, 7})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Normalized, qt.DeepEquals, []float64{1, 2})
	c.Assert(result.Contributions, qt.DeepEquals, []float64{3, -4})
	c.Assert(result.Intercept, qt.Equals, 0.5)
	c.Assert(result.LinearScore, qt.Equals, -0.5)
	c.Assert(math.Abs(result.Response-(1/(1+math.Exp(0.5)))) <= 1e-15, qt.IsTrue)
	copy := m.Parameters()
	copy.Means[0], copy.Scales[0], copy.Weights[0] = -100, 100, -100
	again, err := m.Evaluate(t.Context(), []float64{12, 7})
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
}

func TestLogisticExtremeScoresAndInvalidInputs(t *testing.T) {
	for _, row := range []struct {
		name                 string
		value, score, output float64
	}{
		{"negative", -1000, -1000, 0},
		{"positive", 1000, 1000, 1},
		{"zero", 0, 0, 0.5},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, err := model.NewLogistic(model.Parameters{Means: []float64{0}, Scales: []float64{1}, Weights: []float64{1}})
			c.Assert(err, qt.IsNil)
			result, err := m.Evaluate(t.Context(), []float64{row.value})
			c.Assert(err, qt.IsNil)
			c.Assert(result.LinearScore, qt.Equals, row.score)
			c.Assert(result.Response, qt.Equals, row.output)
		})
	}
	for _, values := range [][]float64{nil, {1, 2}, {math.NaN()}, {math.Inf(1)}} {
		c := qt.New(t)
		m, err := model.NewLogistic(model.Parameters{Means: []float64{0}, Scales: []float64{1}, Weights: []float64{1}})
		c.Assert(err, qt.IsNil)
		result, err := m.Evaluate(t.Context(), values)
		c.Assert(err, qt.IsNotNil)
		c.Assert(result, qt.DeepEquals, model.Evaluation{})
	}
}

func TestLogisticOwnershipAndCancellation(t *testing.T) {
	c := qt.New(t)
	m, err := model.NewLogistic(model.Parameters{Means: []float64{0}, Scales: []float64{1}, Weights: []float64{2}})
	c.Assert(err, qt.IsNil)
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			result, err := m.Evaluate(t.Context(), []float64{2})
			if !c.Check(err, qt.IsNil) {
				return
			}
			c.Check(result.LinearScore, qt.Equals, 4.0)
			result.Contributions[0], result.Normalized[0] = 0, 0
			m.Parameters().Weights[0] = 0
		})
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := m.Evaluate(ctx, []float64{2})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, model.Evaluation{})
	_, err = (&model.Logistic{}).Evaluate(t.Context(), []float64{2})
	c.Assert(err, qt.IsNotNil)
	var absent *model.Logistic
	_, err = absent.Evaluate(t.Context(), []float64{2})
	c.Assert(err, qt.IsNotNil)
}
