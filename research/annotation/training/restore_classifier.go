package training

import (
	"fmt"

	"github.com/stokaro/unswell/model"
)

func restoreModels(a Artifact) (numericalClassifier, *model.Isotonic, error) {
	classifier, err := restoreClassifier(a)
	if err != nil {
		return numericalClassifier{}, nil, err
	}
	if a.Options.Calibration == "none" && a.Calibration == nil {
		return classifier, nil, nil
	}
	c := a.Calibration
	if err := validateCalibration(a, classifier.scoreKind); err != nil {
		return numericalClassifier{}, nil, err
	}
	calibration, err := model.NewIsotonic(model.IsotonicParameters{Scores: c.Scores, Responses: c.Responses})
	return classifier, calibration, err
}

func validateCalibration(a Artifact, scoreKind string) error {
	c := a.Calibration
	if a.Options.Calibration != "isotonic" || c == nil || c.Algorithm != model.IsotonicAlgorithm ||
		c.ScoreKind != scoreKind || !validDigest(c.InputSHA256) {
		return fmt.Errorf("training artifact has incompatible calibration")
	}
	if a.Options.Estimator == "forest" {
		for _, score := range c.Scores {
			if !validResponse(&score) {
				return fmt.Errorf("forest calibration requires scores within [0,1]")
			}
		}
	}
	return nil
}

func restoreClassifier(a Artifact) (numericalClassifier, error) {
	if a.Options.Estimator == "forest" {
		return restoreForest(a)
	}
	p := a.Logistic
	if p == nil || a.Forest != nil || p.Algorithm != model.Algorithm || !validDigest(p.InputSHA256) ||
		len(p.Weights) != len(a.Identity.Columns) {
		return numericalClassifier{}, fmt.Errorf("training artifact has incompatible logistic parameters")
	}
	classifier, err := model.NewLogistic(model.Parameters{Means: p.Means, Scales: p.Scales, Weights: p.Weights, Intercept: p.Intercept})
	return logisticClassifier(classifier), err
}

func restoreForest(a Artifact) (numericalClassifier, error) {
	p := a.Forest
	if p == nil || a.Logistic != nil || p.Algorithm != model.ForestAlgorithm || !validDigest(p.InputSHA256) ||
		p.Parameters.Features != len(a.Identity.Columns) {
		return numericalClassifier{}, fmt.Errorf("training artifact has incompatible forest parameters")
	}
	classifier, err := model.NewForest(p.Parameters)
	if err != nil {
		return numericalClassifier{}, err
	}
	if err := validateForestFit(a); err != nil {
		return numericalClassifier{}, err
	}
	return forestClassifier(classifier), nil
}

func validateForestFit(a Artifact) error {
	p, o := a.Forest, a.Options.Forest
	if err := validateForestFitLimits(*p, *o); err != nil {
		return err
	}
	rows, positive, err := forestTrainingCounts(a.Partitions[0])
	if err != nil {
		return err
	}
	nodes := 0
	for _, tree := range p.Parameters.Trees {
		nodes += len(tree)
		if tree[0].Samples != rows || (!o.Bootstrap && tree[0].Positive != positive) {
			return fmt.Errorf("forest root counts disagree with training rows")
		}
		if err := validateFittedTree(tree, 0, 0, *o); err != nil {
			return err
		}
	}
	if nodes != p.Nodes {
		return fmt.Errorf("forest fitted node count mismatch")
	}
	return nil
}

func forestTrainingCounts(partition Partition) (int, int, error) {
	rows, positive := len(partition.Rows), partition.Classes["needs_revision"]
	if rows < 2 || positive == 0 || positive == rows {
		return 0, 0, fmt.Errorf("forest fitting requires both training classes")
	}
	return rows, positive, nil
}

func validateForestFitLimits(p Forest, o model.ForestOptions) error {
	if o.FeaturesPerSplit > p.Parameters.Features || len(p.Parameters.Trees) != o.Trees ||
		p.Nodes > o.MaxNodes || p.Operations < 1 || p.Operations > o.MaxOperations {
		return fmt.Errorf("forest parameters exceed recorded fitting limits")
	}
	return nil
}

func validateFittedTree(tree []model.ForestNode, index, depth int, options model.ForestOptions) error {
	node := tree[index]
	if node.Feature == -1 {
		return nil
	}
	if depth >= options.MaxDepth || node.Positive == 0 || node.Positive == node.Samples ||
		tree[node.Left].Samples < options.MinLeaf || tree[node.Right].Samples < options.MinLeaf {
		return fmt.Errorf("forest split violates fitted depth, class, or leaf limits")
	}
	if err := validateFittedTree(tree, node.Left, depth+1, options); err != nil {
		return err
	}
	return validateFittedTree(tree, node.Right, depth+1, options)
}
