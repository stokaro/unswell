package reviewbaseline

import (
	"context"
	"math"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
)

type columns struct {
	kind       string
	ids, terms []string
}

func fit(ctx context.Context, rows []row, fold int, kind string) (fitted, error) {
	spec, err := selectColumns(ctx, rows, fold, kind)
	if err != nil {
		return fitted{}, err
	}
	examples, groups := trainingExamples(rows, fold, spec)
	trained, err := model.FitLogistic(ctx, examples, model.FitOptions{
		L2: 0.01, Tolerance: 1e-6, MaxIterations: 100, MaxOperations: 1e12})
	if err != nil {
		return fitted{}, err
	}
	development, err := predict(ctx, trained.Model, spec, rows, (fold+1)%5, true)
	if err != nil {
		return fitted{}, err
	}
	point, err := threshold(development)
	if err != nil {
		return fitted{}, err
	}
	predictions, err := predict(ctx, trained.Model, spec, rows, fold, false)
	if err != nil {
		return fitted{}, err
	}
	for i := range predictions {
		predictions[i].Selected = point.Available && predictions[i].Score >= *point.Threshold
	}
	result := fitted{Kind: kind, Fold: fold, Features: spec.ids, Parameters: trained.Model.Parameters(),
		TrainingHash: trained.InputSHA256, TrainingRows: len(examples), Options: trained.Options,
		Iterations: trained.Iterations, Operations: trained.Operations, OperatingPoint: point, Predictions: predictions}
	for group := range groups {
		result.TrainingGroups = append(result.TrainingGroups, group)
	}
	slices.Sort(result.TrainingGroups)
	return result, nil
}

func trainingExamples(rows []row, fold int, spec columns) ([]model.Example, map[string]bool) {
	var examples []model.Example
	groups := make(map[string]bool)
	for _, row := range rows {
		if trainingRow(row, fold) {
			examples = append(examples, model.Example{Values: vector(row, spec), Label: *row.label})
			groups[row.group] = true
		}
	}
	return examples, groups
}

func selectColumns(ctx context.Context, rows []row, fold int, kind string) (columns, error) {
	spec := columns{kind: kind}
	if kind == "S" || kind == "SW" {
		for _, descriptor := range feature.Catalog() {
			spec.ids = append(spec.ids, descriptor.ID, "missing/"+descriptor.ID)
		}
	} else {
		spec.ids = []string{"log1p-prose-words"}
	}
	if strings.Contains(kind, "W") {
		terms, err := vocabulary(ctx, rows, fold, model.MaxFeatures-len(spec.ids))
		if err != nil {
			return columns{}, err
		}
		spec.terms = terms
		spec.ids = append(spec.ids, terms...)
	}
	return spec, nil
}

func vector(row row, spec columns) []float64 {
	var values []float64
	if spec.kind == "S" || spec.kind == "SW" {
		values = slices.Clone(row.numeric)
	} else {
		values = []float64{math.Log1p(float64(row.words))}
	}
	for _, term := range spec.terms {
		value := 0.0
		if row.terms[term] {
			value = 1
		}
		values = append(values, value)
	}
	return values
}

func predict(ctx context.Context, classifier *model.Logistic, spec columns, rows []row, fold int, labeledOnly bool) ([]prediction, error) {
	var predictions []prediction
	for _, row := range rows {
		if row.fold != fold || labeledOnly && row.label == nil {
			continue
		}
		score, err := classifier.Evaluate(ctx, vector(row, spec))
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction{Page: row.page, Unit: row.unit, Words: row.words,
			Cohort: row.cohort, Label: row.label, Score: score.LinearScore})
	}
	return predictions, nil
}
