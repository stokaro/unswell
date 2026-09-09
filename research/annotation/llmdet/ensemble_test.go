package llmdet_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/llmdet"
)

func smallEnsemble() llmdet.EnsembleSpec {
	return llmdet.EnsembleSpec{
		Version: llmdet.EnsembleVersion, FeatureCount: 1, Classes: []string{"first", "second"}, Trees: []llmdet.TreeSpec{
			{Class: 0, Comparison: "numeric-le", Splits: []llmdet.Split{{Feature: 0, Threshold: 1, Left: -1, Right: -2}}, Leaves: []float64{2, -2}},
			{Class: 1, Comparison: "numeric-le", Splits: []llmdet.Split{}, Leaves: []float64{0}},
		},
	}
}

func TestEnsembleThresholdBoundaryAndSnapshotOwnership(t *testing.T) {
	c := qt.New(t)
	spec := smallEnsemble()
	model, err := llmdet.NewEnsemble(t.Context(), spec)
	c.Assert(err, qt.IsNil)
	spec.Trees[0].Splits[0].Threshold = 5
	spec.Trees[0].Leaves[0] = 10
	spec.Classes[0] = "changed"
	for _, test := range []struct{ value, want float64 }{{math.Nextafter(1, 0), 2}, {1, 2}, {math.Nextafter(1, 2), -2}} {
		result, err := model.Predict(t.Context(), []float64{test.value})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Classes, qt.DeepEquals, []string{"first", "second"})
		c.Assert(result.Raw, qt.DeepEquals, []float64{test.want, 0})
		c.Assert(math.Abs(result.Responses[0]+result.Responses[1]-1) < 1e-15, qt.IsTrue)
		c.Assert(result.Responses[0] > result.Responses[1], qt.Equals, test.want > 0)
		result.Classes[0] = "mutated output"
		result.Raw[0] = 50
	}
}

func TestEnsemblePreservesSummationOrder(t *testing.T) {
	c := qt.New(t)
	spec := smallEnsemble()
	spec.Trees = nil
	for _, value := range []float64{1e16, 1, -1e16} {
		spec.Trees = append(spec.Trees, llmdet.TreeSpec{Comparison: "numeric-le", Leaves: []float64{value}})
	}
	model, err := llmdet.NewEnsemble(t.Context(), spec)
	c.Assert(err, qt.IsNil)
	result, err := model.Predict(t.Context(), []float64{0})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Raw, qt.DeepEquals, []float64{0, 0})
	c.Assert(result.Responses, qt.DeepEquals, []float64{0.5, 0.5})
}

func TestEnsembleRejectsMissingInputsAndOverflow(t *testing.T) {
	c := qt.New(t)
	model, err := llmdet.NewEnsemble(t.Context(), smallEnsemble())
	c.Assert(err, qt.IsNil)
	for _, values := range [][]float64{nil, {1, 2}, {math.NaN()}, {math.Inf(1)}} {
		_, err := model.Predict(t.Context(), values)
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = model.Predict(ctx, []float64{1})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	spec := smallEnsemble()
	spec.Trees = []llmdet.TreeSpec{
		{Comparison: "numeric-le", Leaves: []float64{math.MaxFloat64}},
		{Comparison: "numeric-le", Leaves: []float64{math.MaxFloat64}},
	}
	model, err = llmdet.NewEnsemble(t.Context(), spec)
	c.Assert(err, qt.IsNil)
	_, err = model.Predict(t.Context(), []float64{0})
	c.Assert(err, qt.ErrorMatches, "ensemble margin overflow")
}

func TestEnsembleRejectsMalformedTrees(t *testing.T) {
	for _, test := range []struct {
		name string
		edit func(*llmdet.EnsembleSpec)
	}{
		{"version", func(s *llmdet.EnsembleSpec) { s.Version = "unknown" }},
		{"class", func(s *llmdet.EnsembleSpec) { s.Trees[0].Class = 2 }},
		{"feature", func(s *llmdet.EnsembleSpec) { s.Trees[0].Splits[0].Feature = 1 }},
		{"comparison", func(s *llmdet.EnsembleSpec) { s.Trees[0].Comparison = "categorical" }},
		{"leaf", func(s *llmdet.EnsembleSpec) { s.Trees[0].Leaves[0] = math.NaN() }},
		{"threshold", func(s *llmdet.EnsembleSpec) { s.Trees[0].Splits[0].Threshold = math.Inf(1) }},
		{"cycle", func(s *llmdet.EnsembleSpec) { s.Trees[0].Splits[0].Left = 0 }},
		{"child", func(s *llmdet.EnsembleSpec) { s.Trees[0].Splits[0].Right = -3 }},
		{"shared leaf", func(s *llmdet.EnsembleSpec) { s.Trees[0].Splits[0].Right = -1 }},
		{"repeated class", func(s *llmdet.EnsembleSpec) { s.Classes[1] = s.Classes[0] }},
		{"extra leaf", func(s *llmdet.EnsembleSpec) { s.Trees[0].Leaves = append(s.Trees[0].Leaves, 3) }},
		{"deep", func(s *llmdet.EnsembleSpec) { s.Trees[0] = deepTree(65) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			spec := smallEnsemble()
			test.edit(&spec)
			_, err := llmdet.NewEnsemble(t.Context(), spec)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func deepTree(depth int) llmdet.TreeSpec {
	tree := llmdet.TreeSpec{Comparison: "numeric-le", Leaves: make([]float64, depth+1)}
	for i := range depth {
		tree.Splits = append(tree.Splits, llmdet.Split{Left: -i - 1, Right: i + 1})
	}
	tree.Splits[depth-1].Right = -depth - 1
	return tree
}
