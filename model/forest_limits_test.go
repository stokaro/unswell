package model_test

import (
	"context"
	"sync/atomic"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestForestOptionLimits(t *testing.T) {
	for _, edit := range []func(*model.ForestOptions){
		func(o *model.ForestOptions) { o.Trees = 0 },
		func(o *model.ForestOptions) { o.Trees = model.MaxForestTrees + 1 },
		func(o *model.ForestOptions) { o.MaxDepth = 0 },
		func(o *model.ForestOptions) { o.MaxDepth = model.MaxForestDepth + 1 },
		func(o *model.ForestOptions) { o.MinLeaf = 0 },
		func(o *model.ForestOptions) { o.MinLeaf = 100001 },
		func(o *model.ForestOptions) { o.FeaturesPerSplit = -1 },
		func(o *model.ForestOptions) { o.FeaturesPerSplit = model.MaxFeatures + 1 },
		func(o *model.ForestOptions) { o.MaxNodes = 0 },
		func(o *model.ForestOptions) { o.MaxNodes = model.MaxForestNodes + 1 },
		func(o *model.ForestOptions) { o.MaxOperations = 0 },
		func(o *model.ForestOptions) { o.MaxOperations = 1e12 + 1 },
	} {
		c := qt.New(t)
		options := forestOptions()
		edit(&options)
		c.Assert(options.Validate(), qt.IsNotNil)
		fit, err := model.FitForest(t.Context(), xorExamples(), options)
		c.Assert(err, qt.IsNotNil)
		c.Assert(fit, qt.DeepEquals, model.ForestFitResult{})
	}
	c := qt.New(t)
	options := forestOptions()
	options.FeaturesPerSplit = 3
	c.Assert(options.Validate(), qt.IsNil)
	fit, err := model.FitForest(t.Context(), xorExamples(), options)
	c.Assert(err, qt.ErrorMatches, "forest features per split exceeds input width")
	c.Assert(fit, qt.DeepEquals, model.ForestFitResult{})
}

func chainForest(depth int) model.ForestParameters {
	var nodes []model.ForestNode
	for i := range depth {
		index := len(nodes)
		nodes = append(nodes, model.ForestNode{Feature: 0, Threshold: float64(i), Left: index + 1, Right: index + 2,
			Positive: 1, Samples: depth - i + 1}, forestLeaf(0, 1))
	}
	nodes = append(nodes, forestLeaf(1, 1))
	return model.ForestParameters{Features: 1, Trees: [][]model.ForestNode{nodes}}
}

func TestForestImportedDepthLimit(t *testing.T) {
	c := qt.New(t)
	classifier, err := model.NewForest(chainForest(model.MaxForestDepth))
	c.Assert(err, qt.IsNil)
	value, err := classifier.Evaluate(t.Context(), []float64{1000})
	c.Assert(err, qt.IsNil)
	c.Assert(value.Response, qt.Equals, 1.0)
	c.Assert(value.Trees[0].Path, qt.HasLen, model.MaxForestDepth+1)
	classifier, err = model.NewForest(chainForest(model.MaxForestDepth + 1))
	c.Assert(err, qt.ErrorMatches, ".*tree depth.*")
	c.Assert(classifier, qt.IsNil)
}

type cancelAfterChecks struct {
	context.Context
	cancel    context.CancelFunc
	remaining atomic.Int64
}

func (c *cancelAfterChecks) Err() error {
	if c.remaining.Add(-1) == 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestForestCancellationAfterWorkStarts(t *testing.T) {
	c := qt.New(t)
	parent, cancel := context.WithCancel(t.Context())
	defer cancel()
	ctx := &cancelAfterChecks{Context: parent, cancel: cancel}
	ctx.remaining.Store(10)
	fit, err := model.FitForest(ctx, xorExamples(), forestOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(fit, qt.DeepEquals, model.ForestFitResult{})
	c.Assert(parent.Err(), qt.ErrorIs, context.Canceled)
}

func TestForestConcurrentSnapshots(t *testing.T) {
	c := qt.New(t)
	classifier, err := model.NewForest(forestSnapshot())
	c.Assert(err, qt.IsNil)
	for range 16 {
		t.Run("owned result", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			value, err := classifier.Evaluate(t.Context(), []float64{10, 1})
			c.Assert(err, qt.IsNil)
			c.Assert(value.Response, qt.Equals, 0.75)
			value.Trees[0].Path[0] = 999
			snapshot := classifier.Parameters()
			snapshot.Trees[0][0].Threshold = -999
		})
	}
}
