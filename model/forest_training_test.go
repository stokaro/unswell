package model_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

// Scripted numeric labels test the algorithm, not human editorial judgments.
func xorExamples() []model.Example {
	return []model.Example{
		{Values: []float64{0, 0}, Label: 0}, {Values: []float64{0, 1}, Label: 1},
		{Values: []float64{1, 0}, Label: 1}, {Values: []float64{1, 1}, Label: 0},
	}
}

func forestOptions() model.ForestOptions {
	return model.ForestOptions{Seed: 7, Trees: 1, MaxDepth: 2, MinLeaf: 1, FeaturesPerSplit: 2,
		MaxNodes: 7, MaxOperations: 10000}
}

func TestFitForestLearnsFeatureInteraction(t *testing.T) {
	c := qt.New(t)
	data, options := xorExamples(), forestOptions()
	fit, err := model.FitForest(t.Context(), data, options)
	c.Assert(err, qt.IsNil)
	// Root gains tie at zero. Feature 0 and then feature 1 separate all four corners.
	want := []model.ForestNode{
		{Feature: 0, Threshold: 0.5, Left: 1, Right: 4, Positive: 2, Samples: 4},
		{Feature: 1, Threshold: 0.5, Left: 2, Right: 3, Positive: 1, Samples: 2},
		forestLeaf(0, 1), forestLeaf(1, 1),
		{Feature: 1, Threshold: 0.5, Left: 5, Right: 6, Positive: 1, Samples: 2},
		forestLeaf(1, 1), forestLeaf(0, 1),
	}
	c.Assert(fit.Model.Parameters(), qt.DeepEquals, model.ForestParameters{Features: 2, Trees: [][]model.ForestNode{want}})
	c.Assert(fit.Algorithm, qt.Equals, model.ForestAlgorithm)
	c.Assert(fit.InputSHA256, qt.Equals, "18e5cb1c231417b56bd2cdf51a85d14f0d330724b63d8f4d38ca0428b4103650")
	c.Assert(fit.Nodes, qt.Equals, 7)
	c.Assert(fit.Operations, qt.Equals, int64(280))
	c.Assert(fit.Options, qt.DeepEquals, options)
	for _, example := range data {
		value, err := fit.Model.Evaluate(t.Context(), example.Values)
		c.Assert(err, qt.IsNil)
		c.Assert(value.Response, qt.Equals, float64(example.Label))
	}
	data[0].Values[0] = 999
	value, err := fit.Model.Evaluate(t.Context(), []float64{0, 0})
	c.Assert(err, qt.IsNil)
	c.Assert(value.Response, qt.Equals, 0.0)
}

func TestForestDepthAndLeafSize(t *testing.T) {
	for _, edit := range []func(*model.ForestOptions){
		func(o *model.ForestOptions) { o.MaxDepth = 1 },
		func(o *model.ForestOptions) { o.MinLeaf = 2 },
	} {
		c := qt.New(t)
		options := forestOptions()
		edit(&options)
		fit, err := model.FitForest(t.Context(), xorExamples(), options)
		c.Assert(err, qt.IsNil)
		c.Assert(fit.Nodes, qt.Equals, 3)
		value, err := fit.Model.Evaluate(t.Context(), []float64{0, 0})
		c.Assert(err, qt.IsNil)
		c.Assert(value.Response, qt.Equals, 0.5)
	}
}

func TestForestBootstrapAndSeedReproducibility(t *testing.T) {
	c := qt.New(t)
	options := forestOptions()
	options.Bootstrap, options.Trees, options.MaxNodes, options.FeaturesPerSplit = true, 16, 112, 0
	first, err := model.FitForest(t.Context(), xorExamples(), options)
	c.Assert(err, qt.IsNil)
	second, err := model.FitForest(t.Context(), xorExamples(), options)
	c.Assert(err, qt.IsNil)
	c.Assert(second.Model.Parameters(), qt.DeepEquals, first.Model.Parameters())
	c.Assert(second.InputSHA256, qt.Equals, first.InputSHA256)
	c.Assert(second.Operations, qt.Equals, first.Operations)
	options.Seed++
	third, err := model.FitForest(t.Context(), xorExamples(), options)
	c.Assert(err, qt.IsNil)
	c.Assert(third.Model.Parameters(), qt.Not(qt.DeepEquals), first.Model.Parameters())
	c.Assert(third.InputSHA256, qt.Equals, first.InputSHA256)
	for _, tree := range first.Model.Parameters().Trees {
		c.Assert(tree[0].Samples, qt.Equals, 4)
	}
}
