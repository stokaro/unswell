package reviewbaseline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/internal/sparsemodel"
)

func initializeEncodedOutput(out *encodedOutput) {
	out.ModelAlgorithm = sparsemodel.Algorithm
	out.UnitContract = nlp.UnitContract
	out.FeatureContract = feature.UnitContract
	out.LexicalContract = "not used"
}

func encodedColumns(kind string) []string {
	var ids []string
	for i := range 128 {
		ids = append(ids, fmt.Sprintf("encoder/%03d", i))
	}
	if kind == "ES" {
		for _, d := range feature.Catalog() {
			ids = append(ids, d.ID, "missing/"+d.ID)
		}
	}
	return ids
}

func encodedVector(r row, e encodedRow, kind string) []sparsemodel.Entry {
	var result []sparsemodel.Entry
	for i, v := range e.Values {
		if v != 0 {
			result = append(result, sparsemodel.Entry{Index: i, Value: v})
		}
	}
	if kind == "ES" {
		for i, v := range r.numeric {
			if v != 0 {
				result = append(result, sparsemodel.Entry{Index: 128 + i, Value: v})
			}
		}
	}
	return result
}

func encodedTraining(rows []row, enc []encodedRow, fold int, kind string) ([]sparsemodel.Example, []string) {
	var examples []sparsemodel.Example
	groups := map[string]bool{}
	for i, r := range rows {
		if !trainingRow(r, fold) || enc[i].Values == nil {
			continue
		}
		examples = append(examples, sparsemodel.Example{Values: encodedVector(r, enc[i], kind), Label: *r.label})
		groups[r.group] = true
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	slices.Sort(names)
	return examples, names
}

func fitEncoded(ctx context.Context, rows []row, enc []encodedRow, fold int, kind string) (encodedFit, error) {
	ids := encodedColumns(kind)
	examples, groups := encodedTraining(rows, enc, fold, kind)
	trained, err := sparsemodel.Fit(ctx, examples, len(ids))
	if err != nil {
		return encodedFit{}, err
	}
	development, err := predictEncoded(ctx, trained, rows, enc, (fold+1)%5, kind)
	if err != nil {
		return encodedFit{}, err
	}
	point, err := encodedThreshold(development)
	if err != nil {
		return encodedFit{}, err
	}
	predictions, err := predictEncoded(ctx, trained, rows, enc, fold, kind)
	if err != nil {
		return encodedFit{}, err
	}
	selectEncoded(predictions, point)
	raw, err := json.Marshal(examples)
	if err != nil {
		return encodedFit{}, err
	}
	return encodedFit{fitted: fitted{Kind: kind, Fold: fold, Features: ids, Parameters: trained.Parameters,
		TrainingHash: fmt.Sprintf("%x", sha256.Sum256(raw)), TrainingRows: len(examples), TrainingGroups: groups,
		Options:    model.FitOptions{L2: 0.01, Tolerance: 1e-8, MaxIterations: 500, MaxOperations: 3000000000},
		Iterations: trained.Iterations, Operations: trained.Operations, Gradient: trained.Gradient, OperatingPoint: point},
		Predictions: predictions}, nil
}

func predictEncoded(ctx context.Context, trained sparsemodel.Result, rows []row, enc []encodedRow, fold int,
	kind string,
) ([]encodedPrediction, error) {
	var result []encodedPrediction
	for i, r := range rows {
		if r.fold != fold {
			continue
		}
		p := encodedPrediction{prediction: prediction{Page: r.page, Unit: r.unit, Words: r.words, Cohort: r.cohort, Label: r.label},
			Reason: enc[i].Reason}
		if enc[i].Values != nil {
			score, err := trained.Score(ctx, encodedVector(r, enc[i], kind))
			if err != nil {
				return nil, err
			}
			p.Score = &score
			p.prediction.Score = score
		}
		result = append(result, p)
	}
	return result, nil
}

func encodedThreshold(development []encodedPrediction) (operatingPoint, error) {
	var rows []prediction
	for _, r := range development {
		if r.Label != nil && r.Score != nil {
			rows = append(rows, r.prediction)
		}
	}
	return threshold(rows)
}

func selectEncoded(predictions []encodedPrediction, point operatingPoint) {
	for i := range predictions {
		r := &predictions[i]
		r.Selected = r.Score != nil && point.Available && *r.Score >= *point.Threshold
	}
}
