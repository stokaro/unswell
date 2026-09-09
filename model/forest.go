package model

import (
	"context"
	"fmt"
	"slices"
)

// ForestAlgorithm identifies the specified bootstrap/Gini/PCG fitting protocol.
const ForestAlgorithm = "unswell-binary-forest-v1"

// MaxForestTrees bounds one immutable ensemble.
const MaxForestTrees = 64

// MaxForestDepth bounds splits on any root-to-leaf path.
const MaxForestDepth = 12

// MaxForestNodes bounds all nodes across one ensemble.
const MaxForestNodes = 65535

// ForestNode holds a split or leaf and its bootstrap class counts.
// Leaves use Feature, Left, Right = -1 and Threshold = 0. Splits take the left
// child when input[Feature] <= Threshold. Trees use complete preorder layout.
type ForestNode struct {
	Feature   int     `json:"feature"`
	Left      int     `json:"left"`
	Right     int     `json:"right"`
	Threshold float64 `json:"threshold"`
	Positive  int     `json:"positive"`
	Samples   int     `json:"samples"`
}

// ForestParameters is a numerical snapshot, without prose or feature identities.
type ForestParameters struct {
	Features int            `json:"features"`
	Trees    [][]ForestNode `json:"trees"`
}

// Forest owns validated immutable trees. Its zero value cannot evaluate inputs.
type Forest struct {
	parameters ForestParameters
}

// TreeEvaluation records one path and the reached leaf's class fraction.
// Paths index the corresponding tree snapshot; these are not causal attributions.
type TreeEvaluation struct {
	Path              []int
	Leaf              int
	Positive, Samples int
	Response          float64
}

// ForestEvaluation contains owned tree paths and their mean uncalibrated response.
type ForestEvaluation struct {
	Trees    []TreeEvaluation
	Response float64
}

// NewForest validates and copies a complete numerical forest without loading files.
func NewForest(parameters ForestParameters) (*Forest, error) {
	if err := validateForest(parameters); err != nil {
		return nil, err
	}
	return &Forest{parameters: cloneForest(parameters)}, nil
}

// Parameters returns an independent snapshot of all trees.
func (m *Forest) Parameters() ForestParameters {
	if m == nil {
		return ForestParameters{}
	}
	return cloneForest(m.parameters)
}

func cloneForest(p ForestParameters) ForestParameters {
	result := ForestParameters{Features: p.Features, Trees: make([][]ForestNode, len(p.Trees))}
	for i, tree := range p.Trees {
		result.Trees[i] = slices.Clone(tree)
	}
	return result
}

// Evaluate traverses each tree without altering model or input buffers.
// Missing/nonfinite values and cancellation return no partial result.
func (m *Forest) Evaluate(ctx context.Context, values []float64) (ForestEvaluation, error) {
	if err := ctx.Err(); err != nil {
		return ForestEvaluation{}, err
	}
	if m == nil || len(m.parameters.Trees) == 0 || len(values) != m.parameters.Features {
		return ForestEvaluation{}, fmt.Errorf("forest evaluation requires a model and matching input dimensions")
	}
	for i, value := range values {
		if !finite(value) {
			return ForestEvaluation{}, fmt.Errorf("forest input %d must be finite; missing values are not imputed", i)
		}
	}
	result := ForestEvaluation{Trees: make([]TreeEvaluation, 0, len(m.parameters.Trees))}
	for _, tree := range m.parameters.Trees {
		if err := ctx.Err(); err != nil {
			return ForestEvaluation{}, err
		}
		value := evaluateTree(tree, values)
		result.Trees = append(result.Trees, value)
		result.Response += value.Response
	}
	result.Response /= float64(len(result.Trees))
	if err := ctx.Err(); err != nil {
		return ForestEvaluation{}, err
	}
	return result, nil
}

func evaluateTree(tree []ForestNode, values []float64) TreeEvaluation {
	var result TreeEvaluation
	index := 0
	for {
		result.Path = append(result.Path, index)
		node := tree[index]
		if node.Feature == -1 {
			result.Leaf, result.Positive, result.Samples = index, node.Positive, node.Samples
			result.Response = float64(node.Positive) / float64(node.Samples)
			return result
		}
		if values[node.Feature] <= node.Threshold {
			index = node.Left
		} else {
			index = node.Right
		}
	}
}
