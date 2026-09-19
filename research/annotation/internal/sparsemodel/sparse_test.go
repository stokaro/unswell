package sparsemodel_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/internal/sparsemodel"
)

func sparse(values []float64) []sparsemodel.Entry {
	var entries []sparsemodel.Entry
	for i, value := range values {
		if value != 0 {
			entries = append(entries, sparsemodel.Entry{Index: i, Value: value})
		}
	}
	return entries
}

func TestSparseMatchesDenseObjective(t *testing.T) {
	c := qt.New(t)
	var dense []model.Example
	var rows []sparsemodel.Example
	for i := range 80 {
		values := []float64{float64(i%9 - 4), float64(i%3 - 1), 7, 0}
		if i%11 == 0 {
			values[3] = 2
		}
		label := 0
		if i%7 > 2 {
			label = 1
		}
		dense = append(dense, model.Example{Values: values, Label: label})
		rows = append(rows, sparsemodel.Example{Values: sparse(values), Label: label})
	}
	reference, err := model.FitLogistic(t.Context(), dense, model.FitOptions{
		L2: 0.01, Tolerance: 1e-9, MaxIterations: 100, MaxOperations: 100000000})
	c.Assert(err, qt.IsNil)
	result, err := sparsemodel.Fit(t.Context(), rows, 4)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gradient <= 1e-8, qt.IsTrue)
	for _, values := range [][]float64{{0, 0, 7, 0}, {-3, 1, 7, 2}, {4, -1, 7, 0}} {
		want, err := reference.Model.Evaluate(t.Context(), values)
		c.Assert(err, qt.IsNil)
		got, err := result.Score(t.Context(), sparse(values))
		c.Assert(err, qt.IsNil)
		c.Assert(math.Abs(got-want.LinearScore) < 1e-4, qt.IsTrue)
	}
	again, err := sparsemodel.Fit(t.Context(), rows, 4)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
}

func TestLargeVocabularyAndConstantColumn(t *testing.T) {
	c := qt.New(t)
	rows := []sparsemodel.Example{
		{Values: []sparsemodel.Entry{{Index: 0, Value: 1e12}}, Label: 0},
		{Values: []sparsemodel.Entry{{Index: 0, Value: 1e12}, {Index: 8191, Value: 1}}, Label: 1},
	}
	result, err := sparsemodel.Fit(t.Context(), rows, sparsemodel.MaxFeatures)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Parameters.Scales[0], qt.Equals, 1.0)
	c.Assert(result.Parameters.Weights[0], qt.Equals, 0.0)
	negative, err := result.Score(t.Context(), rows[0].Values)
	c.Assert(err, qt.IsNil)
	positive, err := result.Score(t.Context(), rows[1].Values)
	c.Assert(err, qt.IsNil)
	c.Assert(negative < -2 && positive > 2, qt.IsTrue)
}

func TestInterceptIsNotPenalized(t *testing.T) {
	c := qt.New(t)
	rows := []sparsemodel.Example{{Label: 0}, {Label: 1}, {Label: 1}, {Label: 1}}
	result, err := sparsemodel.Fit(t.Context(), rows, 3)
	c.Assert(err, qt.IsNil)
	c.Assert(math.Abs(result.Parameters.Intercept-math.Log(3)) < 1e-12, qt.IsTrue)
	c.Assert(result.Parameters.Weights, qt.DeepEquals, []float64{0, 0, 0})
}

func TestInvalidDataReturnsNoModel(t *testing.T) {
	for _, values := range [][]sparsemodel.Entry{
		{{Index: -1, Value: 1}}, {{Index: 2, Value: 1}},
		{{Index: 1, Value: 1}, {Index: 1, Value: 2}},
		{{Index: 1, Value: 1}, {Index: 0, Value: 2}},
		{{Index: 0, Value: 0}}, {{Index: 0, Value: math.NaN()}},
	} {
		t.Run("invalid sparse row", func(t *testing.T) {
			c := qt.New(t)
			result, err := sparsemodel.Fit(t.Context(), []sparsemodel.Example{{Values: values}, {Label: 1}}, 2)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, sparsemodel.Result{})
		})
	}
	c := qt.New(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := sparsemodel.Fit(ctx, nil, 0)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, sparsemodel.Result{})
	result, err = sparsemodel.Fit(t.Context(), []sparsemodel.Example{{Label: 0}, {Label: 1}}, sparsemodel.MaxFeatures+1)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, sparsemodel.Result{})
}

func TestUnrepresentableVarianceReturnsNoModel(t *testing.T) {
	c := qt.New(t)
	rows := []sparsemodel.Example{
		{Values: []sparsemodel.Entry{{Index: 0, Value: 1e-300}}, Label: 0},
		{Values: []sparsemodel.Entry{{Index: 0, Value: 2e-300}}, Label: 1},
	}
	result, err := sparsemodel.Fit(t.Context(), rows, 1)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, sparsemodel.Result{})
}
