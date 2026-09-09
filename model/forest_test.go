package model_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func forestLeaf(positive, samples int) model.ForestNode {
	return model.ForestNode{Feature: -1, Left: -1, Right: -1, Positive: positive, Samples: samples}
}

func forestSnapshot() model.ForestParameters {
	return model.ForestParameters{Features: 2, Trees: [][]model.ForestNode{
		{{Feature: 0, Threshold: 10, Left: 1, Right: 2, Positive: 3, Samples: 4}, forestLeaf(1, 2), forestLeaf(2, 2)},
		{{Feature: 1, Threshold: 0, Left: 1, Right: 2, Positive: 2, Samples: 4}, forestLeaf(0, 2), forestLeaf(2, 2)},
	}}
}

func TestForestClassFractionsAndPaths(t *testing.T) {
	c := qt.New(t)
	parameters := forestSnapshot()
	classifier, err := model.NewForest(parameters)
	c.Assert(err, qt.IsNil)
	result, err := classifier.Evaluate(t.Context(), []float64{10, 1})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Response, qt.Equals, 0.75)
	c.Assert(result.Trees, qt.DeepEquals, []model.TreeEvaluation{
		{Path: []int{0, 1}, Leaf: 1, Positive: 1, Samples: 2, Response: 0.5},
		{Path: []int{0, 2}, Leaf: 2, Positive: 2, Samples: 2, Response: 1},
	})
	right, err := classifier.Evaluate(t.Context(), []float64{math.Nextafter(10, math.Inf(1)), 1})
	c.Assert(err, qt.IsNil)
	c.Assert(right.Response, qt.Equals, 1.0)
	parameters.Trees[0][0].Threshold = -999
	snapshot := classifier.Parameters()
	snapshot.Trees[0][1].Positive = 0
	result.Trees[0].Path[0] = 999
	again, err := classifier.Evaluate(t.Context(), []float64{10, 1})
	c.Assert(err, qt.IsNil)
	c.Assert(again.Response, qt.Equals, 0.75)
	c.Assert(again.Trees[0].Path, qt.DeepEquals, []int{0, 1})
	loaded, err := model.NewForest(classifier.Parameters())
	c.Assert(err, qt.IsNil)
	fromSnapshot, err := loaded.Evaluate(t.Context(), []float64{10, 1})
	c.Assert(err, qt.IsNil)
	c.Assert(fromSnapshot, qt.DeepEquals, again)
}

func TestForestRejectsMissingInputsAndCancellation(t *testing.T) {
	c := qt.New(t)
	classifier, err := model.NewForest(forestSnapshot())
	c.Assert(err, qt.IsNil)
	for _, values := range [][]float64{nil, {1}, {1, 2, 3}, {math.NaN(), 1}, {1, math.Inf(1)}, {math.Inf(-1), 1}} {
		result, err := classifier.Evaluate(t.Context(), values)
		c.Assert(err, qt.IsNotNil)
		c.Assert(result, qt.DeepEquals, model.ForestEvaluation{})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := classifier.Evaluate(ctx, []float64{1, 2})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, model.ForestEvaluation{})
	for _, empty := range []*model.Forest{nil, {}} {
		result, err = empty.Evaluate(t.Context(), []float64{1, 2})
		c.Assert(err, qt.IsNotNil)
		c.Assert(result, qt.DeepEquals, model.ForestEvaluation{})
	}
}
