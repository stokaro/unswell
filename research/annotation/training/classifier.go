package training

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell/model"
)

// Forest records the fitted trees and bounded work on the selected training rows.
type Forest struct {
	Algorithm   string                 `json:"algorithm"`
	InputSHA256 string                 `json:"input_sha256"`
	Parameters  model.ForestParameters `json:"parameters"`
	Operations  int64                  `json:"operations"`
	Nodes       int                    `json:"nodes"`
}

type classifierValue struct {
	score, response          float64
	linear, logistic, forest *float64
}

type numericalClassifier struct {
	scoreKind string
	evaluate  func(context.Context, []float64) (classifierValue, error)
}

func logisticClassifier(m *model.Logistic) numericalClassifier {
	return numericalClassifier{"linear_score", func(ctx context.Context, values []float64) (classifierValue, error) {
		value, err := m.Evaluate(ctx, values)
		return classifierValue{score: value.LinearScore, response: value.Response,
			linear: &value.LinearScore, logistic: &value.Response}, err
	}}
}

func forestClassifier(m *model.Forest) numericalClassifier {
	return numericalClassifier{"forest_response", func(ctx context.Context, values []float64) (classifierValue, error) {
		value, err := m.Evaluate(ctx, values)
		return classifierValue{score: value.Response, response: value.Response, forest: &value.Response}, err
	}}
}

func validateEstimatorOptions(o Options) error {
	switch o.Estimator {
	case "logistic":
		if o.Fit == nil || o.Forest != nil {
			return fmt.Errorf("logistic requires only logistic fit options")
		}
		return o.Fit.numerical().Validate()
	case "forest":
		if o.Forest == nil || o.Fit != nil {
			return fmt.Errorf("forest requires only forest fit options")
		}
		return o.Forest.Validate()
	default:
		return fmt.Errorf("estimator must be logistic or forest")
	}
}

func copyEstimatorOptions(o Options) Options {
	if o.Fit != nil {
		fit := *o.Fit
		o.Fit = &fit
	}
	if o.Forest != nil {
		forest := *o.Forest
		o.Forest = &forest
	}
	return o
}

type fittedClassifier struct {
	classifier numericalClassifier
	logistic   *Logistic
	forest     *Forest
}

func fitClassifier(ctx context.Context, examples []model.Example, options Options) (fittedClassifier, error) {
	if options.Estimator == "forest" {
		fit, err := model.FitForest(ctx, examples, *options.Forest)
		if err != nil {
			return fittedClassifier{}, err
		}
		return fittedClassifier{classifier: forestClassifier(fit.Model), forest: &Forest{
			Algorithm: fit.Algorithm, InputSHA256: fit.InputSHA256, Parameters: fit.Model.Parameters(),
			Operations: fit.Operations, Nodes: fit.Nodes}}, nil
	}
	fit, err := model.FitLogistic(ctx, examples, options.Fit.numerical())
	if err != nil {
		return fittedClassifier{}, err
	}
	parameters := logisticResult(fit)
	return fittedClassifier{classifier: logisticClassifier(fit.Model), logistic: &parameters}, nil
}
