package model_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestForestRejectsCorruptSnapshots(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*model.ForestParameters)
	}{
		{"no features", func(p *model.ForestParameters) { p.Features = 0 }},
		{"too many features", func(p *model.ForestParameters) { p.Features = model.MaxFeatures + 1 }},
		{"no trees", func(p *model.ForestParameters) { p.Trees = nil }},
		{"empty tree", func(p *model.ForestParameters) { p.Trees[0] = nil }},
		{"too many trees", func(p *model.ForestParameters) { p.Trees = make([][]model.ForestNode, model.MaxForestTrees+1) }},
		{"too many nodes", func(p *model.ForestParameters) { p.Trees[0] = make([]model.ForestNode, model.MaxForestNodes+1) }},
		{"unknown feature", func(p *model.ForestParameters) { p.Trees[0][0].Feature = 2 }},
		{"negative feature", func(p *model.ForestParameters) { p.Trees[0][0].Feature = -2 }},
		{"NaN", func(p *model.ForestParameters) { p.Trees[0][0].Threshold = math.NaN() }},
		{"infinity", func(p *model.ForestParameters) { p.Trees[0][0].Threshold = math.Inf(-1) }},
		{"cycle", func(p *model.ForestParameters) { p.Trees[0][0].Left = 0 }},
		{"shared child", func(p *model.ForestParameters) { p.Trees[0][0].Right = 1 }},
		{"invalid child", func(p *model.ForestParameters) { p.Trees[0][0].Right = 99 }},
		{"leaf has child", func(p *model.ForestParameters) { p.Trees[0][1].Left = 2 }},
		{"leaf has threshold", func(p *model.ForestParameters) { p.Trees[0][1].Threshold = 1 }},
		{"zero samples", func(p *model.ForestParameters) { p.Trees[0][1].Samples = 0 }},
		{"negative count", func(p *model.ForestParameters) { p.Trees[0][1].Positive = -1 }},
		{"excess positive", func(p *model.ForestParameters) { p.Trees[0][1].Positive = 3 }},
		{"excess samples", func(p *model.ForestParameters) { p.Trees[0][1].Samples = 100001 }},
		{"parent totals", func(p *model.ForestParameters) { p.Trees[0][0].Positive = 1 }},
		{"unreachable node", func(p *model.ForestParameters) { p.Trees[0] = append(p.Trees[0], forestLeaf(0, 1)) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			p := forestSnapshot()
			row.edit(&p)
			classifier, err := model.NewForest(p)
			c.Assert(err, qt.IsNotNil)
			c.Assert(classifier, qt.IsNil)
		})
	}
}

func TestForestFitRejectsInvalidRows(t *testing.T) {
	for _, rows := range [][]model.Example{
		nil, {{Values: []float64{1}, Label: 0}},
		{{Values: []float64{1}, Label: 0}, {Values: []float64{2}, Label: 0}},
		{{Values: []float64{1}, Label: -1}, {Values: []float64{2}, Label: 1}},
		{{Values: nil, Label: 0}, {Values: nil, Label: 1}},
		{{Values: []float64{1}, Label: 0}, {Values: []float64{2, 3}, Label: 1}},
		{{Values: []float64{math.NaN()}, Label: 0}, {Values: []float64{2}, Label: 1}},
		{{Values: []float64{1}, Label: 0}, {Values: []float64{math.Inf(1)}, Label: 1}},
		{{Values: []float64{1}, Label: 0}, {Values: []float64{1e13}, Label: 1}},
	} {
		c := qt.New(t)
		result, err := model.FitForest(t.Context(), rows, forestOptions())
		c.Assert(err, qt.IsNotNil)
		c.Assert(result, qt.DeepEquals, model.ForestFitResult{})
	}
}

func TestForestBudgetsReturnNoPartialModel(t *testing.T) {
	for _, edit := range []func(*model.ForestOptions){
		func(o *model.ForestOptions) { o.MaxOperations = 1 },
		func(o *model.ForestOptions) { o.MaxOperations = 279 },
		func(o *model.ForestOptions) { o.MaxNodes = 6 },
	} {
		c := qt.New(t)
		options := forestOptions()
		edit(&options)
		result, err := model.FitForest(t.Context(), xorExamples(), options)
		c.Assert(err, qt.ErrorIs, model.ErrBudget)
		c.Assert(result, qt.DeepEquals, model.ForestFitResult{})
	}
	c := qt.New(t)
	options := forestOptions()
	options.MaxOperations = 280
	result, err := model.FitForest(t.Context(), xorExamples(), options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Operations, qt.Equals, int64(280))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = model.FitForest(ctx, xorExamples(), options)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, model.ForestFitResult{})
}

func TestForestAdjacentFloatThreshold(t *testing.T) {
	c := qt.New(t)
	a, b := math.SmallestNonzeroFloat64, 2*math.SmallestNonzeroFloat64
	options := forestOptions()
	options.FeaturesPerSplit = 1
	result, err := model.FitForest(t.Context(), []model.Example{
		{Values: []float64{a}, Label: 0}, {Values: []float64{b}, Label: 1},
	}, options)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Model.Parameters().Trees[0][0].Threshold, qt.Equals, a)
	left, err := result.Model.Evaluate(t.Context(), []float64{a})
	c.Assert(err, qt.IsNil)
	right, err := result.Model.Evaluate(t.Context(), []float64{b})
	c.Assert(err, qt.IsNil)
	c.Assert(left.Response, qt.Equals, 0.0)
	c.Assert(right.Response, qt.Equals, 1.0)
}
