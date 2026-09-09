package model

import (
	"context"
	"fmt"
	"math/rand/v2"
)

// ForestOptions declares randomization, tree complexity, and total work limits.
// FeaturesPerSplit=0 selects floor(sqrt(width)); otherwise it selects that many
// features without replacement at each node. Bootstrap draws one row per input
// row with replacement; false uses the ordered input directly in every tree.
type ForestOptions struct {
	Seed             uint64 `json:"seed"`
	Trees            int    `json:"trees"`
	MaxDepth         int    `json:"max_depth"`
	MinLeaf          int    `json:"min_leaf"`
	FeaturesPerSplit int    `json:"features_per_split"`
	MaxNodes         int    `json:"max_nodes"`
	MaxOperations    int64  `json:"max_operations"`
	Bootstrap        bool   `json:"bootstrap"`
}

// Validate checks limits independent of training vector width.
func (o ForestOptions) Validate() error {
	if o.Trees < 1 || o.Trees > MaxForestTrees || o.MaxDepth < 1 || o.MaxDepth > MaxForestDepth {
		return fmt.Errorf("forest tree count or depth is out of range")
	}
	if o.MinLeaf < 1 || o.MinLeaf > maxForestSamples || o.FeaturesPerSplit < 0 || o.FeaturesPerSplit > MaxFeatures {
		return fmt.Errorf("forest leaf size or features per split is out of range")
	}
	return o.validateBudgets()
}

func (o ForestOptions) validateBudgets() error {
	if o.MaxNodes < o.Trees || o.MaxNodes > MaxForestNodes || o.MaxOperations < 1 || o.MaxOperations > 1e12 {
		return fmt.Errorf("forest node or operation budget is out of range")
	}
	return nil
}

// ForestFitResult records numerical fitting, without editorial qualification.
// InputSHA256 binds the ordered raw vectors and labels; Options binds the seed.
type ForestFitResult struct {
	Model                  *Forest
	Algorithm, InputSHA256 string
	Options                ForestOptions
	Operations             int64
	Nodes                  int
}

// FitForest fits only the supplied training rows. It copies input vectors, uses
// a private PCG generator, and returns no partial result on cancellation or error.
// Callers establish feature identities, label permissions, and split independence.
func FitForest(ctx context.Context, examples []Example, options ForestOptions) (ForestFitResult, error) {
	if err := options.Validate(); err != nil {
		return ForestFitResult{}, err
	}
	budget := workBudget{limit: options.MaxOperations}
	data, err := prepareForest(ctx, examples, &budget)
	if err != nil {
		return ForestFitResult{}, err
	}
	if options.FeaturesPerSplit > data.width {
		return ForestFitResult{}, fmt.Errorf("forest features per split exceeds input width")
	}
	builder := forestBuilder{ctx: ctx, data: data, options: options, budget: &budget,
		// #nosec G404 -- Fixed PCG seeds make numerical experiments reproducible; no secrets are generated.
		random: rand.New(rand.NewPCG(options.Seed, 0x9e3779b97f4a7c15))}
	parameters := ForestParameters{Features: data.width, Trees: make([][]ForestNode, 0, options.Trees)}
	for range options.Trees {
		indices, err := builder.sample()
		if err != nil {
			return ForestFitResult{}, err
		}
		builder.tree = nil
		if _, err := builder.node(indices, 0); err != nil {
			return ForestFitResult{}, err
		}
		parameters.Trees = append(parameters.Trees, builder.tree)
	}
	if err := budget.charge(int64(builder.nodes) * 2); err != nil {
		return ForestFitResult{}, err
	}
	classifier, err := NewForest(parameters)
	if err != nil {
		return ForestFitResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return ForestFitResult{}, err
	}
	return ForestFitResult{Model: classifier, Algorithm: ForestAlgorithm, InputSHA256: data.hash, Options: options,
		Operations: budget.used, Nodes: builder.nodes}, nil
}
