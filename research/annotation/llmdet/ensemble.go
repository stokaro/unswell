package llmdet

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
)

const (
	// EnsembleVersion identifies finite numeric trees with ordered leaf summation.
	EnsembleVersion = "unswell-llmdet-ensemble-v1"
	maxTrees        = 5000
	maxLeaves       = 1024
	maxNodes        = 100_000
	maxFeatures     = 128
	maxClasses      = 32
	maxDepth        = 64
)

// Split routes values less than or equal to Threshold to Left, otherwise Right.
// Nonnegative children index splits; -1 identifies leaf zero, -2 leaf one, etc.
type Split struct {
	Feature   int     `json:"feature"`
	Threshold float64 `json:"threshold"`
	Left      int     `json:"left"`
	Right     int     `json:"right"`
}

// TreeSpec stores a numeric tree for one output class. Leaves already include
// the reference learning rate; inference must not apply shrinkage again.
type TreeSpec struct {
	Class      int       `json:"class"`
	Comparison string    `json:"comparison"`
	Splits     []Split   `json:"splits"`
	Leaves     []float64 `json:"leaves"`
}

// EnsembleSpec supplies explicit feature dimensions, class order, and tree order.
type EnsembleSpec struct {
	Version      string     `json:"version"`
	FeatureCount int        `json:"feature_count"`
	Classes      []string   `json:"classes"`
	Trees        []TreeSpec `json:"trees"`
}

// Ensemble owns a validated snapshot for finite complete feature vectors.
type Ensemble struct{ spec EnsembleSpec }

// EnsembleResult retains margins and normalized multiclass responses separately.
// Responses are not calibrated probabilities of authorship or editorial quality.
type EnsembleResult struct {
	Classes   []string  `json:"classes"`
	Raw       []float64 `json:"raw"`
	Responses []float64 `json:"responses"`
}

// NewEnsemble checks resource limits and every tree before copying its data.
func NewEnsemble(ctx context.Context, spec EnsembleSpec) (*Ensemble, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validEnsembleDimensions(spec) {
		return nil, fmt.Errorf("unsupported ensemble version or dimensions")
	}
	if err := validClasses(spec.Classes); err != nil {
		return nil, err
	}
	result := &Ensemble{spec: spec}
	result.spec.Classes = slices.Clone(spec.Classes)
	result.spec.Trees = make([]TreeSpec, len(spec.Trees))
	count := 0
	for i, tree := range spec.Trees {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		count += len(tree.Splits) + len(tree.Leaves)
		if count > maxNodes {
			return nil, fmt.Errorf("ensemble exceeds %d nodes", maxNodes)
		}
		if err := validateTree(tree, spec.FeatureCount, len(spec.Classes)); err != nil {
			return nil, fmt.Errorf("tree %d: %w", i, err)
		}
		tree.Splits = slices.Clone(tree.Splits)
		tree.Leaves = slices.Clone(tree.Leaves)
		result.spec.Trees[i] = tree
	}
	return result, nil
}

func validEnsembleDimensions(spec EnsembleSpec) bool {
	return spec.Version == EnsembleVersion && spec.FeatureCount >= 1 && spec.FeatureCount <= maxFeatures &&
		len(spec.Classes) >= 2 && len(spec.Classes) <= maxClasses && len(spec.Trees) > 0 && len(spec.Trees) <= maxTrees
}

func validClasses(classes []string) error {
	seen := make(map[string]bool, len(classes))
	for _, name := range classes {
		if strings.TrimSpace(name) != name || len(name) == 0 || len(name) > 64 || seen[name] {
			return fmt.Errorf("invalid or duplicate ensemble class")
		}
		seen[name] = true
	}
	return nil
}

// Predict evaluates a complete finite vector and returns a detached result.
func (e *Ensemble) Predict(ctx context.Context, values []float64) (EnsembleResult, error) {
	var result EnsembleResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if e == nil || e.spec.FeatureCount < 1 || len(values) != e.spec.FeatureCount {
		return result, fmt.Errorf("missing ensemble or incorrect feature count")
	}
	for _, value := range values {
		if !finite(value) {
			return result, fmt.Errorf("ensemble requires finite complete features")
		}
	}
	raw := make([]float64, len(e.spec.Classes))
	for _, tree := range e.spec.Trees {
		if err := ctx.Err(); err != nil {
			return EnsembleResult{}, err
		}
		raw[tree.Class] += treeValue(tree, values)
		if !finite(raw[tree.Class]) {
			return EnsembleResult{}, fmt.Errorf("ensemble margin overflow")
		}
	}
	result.Classes = slices.Clone(e.spec.Classes)
	result.Raw = raw
	result.Responses = softmax(raw)
	return result, nil
}

func treeValue(tree TreeSpec, values []float64) float64 {
	if len(tree.Splits) == 0 {
		return tree.Leaves[0]
	}
	child := 0
	for child >= 0 {
		split := tree.Splits[child]
		child = split.Right
		if values[split.Feature] <= split.Threshold {
			child = split.Left
		}
	}
	return tree.Leaves[-child-1]
}

func softmax(raw []float64) []float64 {
	values := make([]float64, len(raw))
	maximum, total := slices.Max(raw), 0.0
	for i, value := range raw {
		values[i] = math.Exp(value - maximum)
		total += values[i]
	}
	for i := range values {
		values[i] /= total
	}
	return values
}
