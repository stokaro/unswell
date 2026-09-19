package reviewbaseline

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/research/annotation/internal/sparsemodel"
)

// RunSparse compares fixed vocabulary capacities using the same exposed units.
// The wider models are research artifacts, not supported product model packs.
func RunSparse(ctx context.Context, reader io.Reader, writer io.Writer) error {
	data, err := commandio.Await(ctx, func() ([]byte, error) { return io.ReadAll(io.LimitReader(reader, (32<<20)+1)) })
	if err != nil {
		return err
	}
	in, err := decode(data)
	if err != nil {
		return err
	}
	rows, provider, err := prepare(ctx, in)
	if err != nil {
		return err
	}
	out := output{Version: 2, Basis: "exposed-assistant-development; sparse capacity comparison",
		InputHash: fmt.Sprintf("%x", sha256.Sum256(data)), ModelAlgorithm: sparsemodel.Algorithm,
		UnitContract: nlp.UnitContract, FeatureContract: feature.UnitContract,
		LexicalContract: feature.LexicalCountContract, Provider: provider, Rows: len(rows)}
	for fold := range 5 {
		for _, capacity := range []int{128, 1024, 8192} {
			result, err := fitSparse(ctx, rows, fold, capacity)
			if err != nil {
				return fmt.Errorf("fold %d capacity %d: %w", fold, capacity, err)
			}
			out.Models = append(out.Models, result)
		}
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}

type sparseColumns struct {
	ids   []string
	terms map[string]int
}

func selectSparseColumns(ctx context.Context, rows []row, fold, capacity int) (sparseColumns, error) {
	spec := sparseColumns{terms: make(map[string]int)}
	for _, descriptor := range feature.Catalog() {
		spec.ids = append(spec.ids, descriptor.ID, "missing/"+descriptor.ID)
	}
	terms, err := vocabulary(ctx, rows, fold, capacity-len(spec.ids))
	if err != nil {
		return sparseColumns{}, err
	}
	for _, term := range terms {
		spec.terms[term] = len(spec.ids)
		spec.ids = append(spec.ids, term)
	}
	return spec, nil
}

func sparseVector(row row, spec sparseColumns) []sparsemodel.Entry {
	var result []sparsemodel.Entry
	for i, value := range row.numeric {
		if value != 0 {
			result = append(result, sparsemodel.Entry{Index: i, Value: value})
		}
	}
	for term := range row.terms {
		if index, ok := spec.terms[term]; ok {
			result = append(result, sparsemodel.Entry{Index: index, Value: 1})
		}
	}
	slices.SortFunc(result, func(a, b sparsemodel.Entry) int { return a.Index - b.Index })
	return result
}

func sparseTraining(rows []row, fold int, spec sparseColumns) ([]sparsemodel.Example, []string) {
	var examples []sparsemodel.Example
	groups := make(map[string]bool)
	for _, row := range rows {
		if trainingRow(row, fold) {
			examples = append(examples, sparsemodel.Example{Values: sparseVector(row, spec), Label: *row.label})
			groups[row.group] = true
		}
	}
	var names []string
	for name := range groups {
		names = append(names, name)
	}
	slices.Sort(names)
	return examples, names
}

func fitSparse(ctx context.Context, rows []row, fold, capacity int) (fitted, error) {
	spec, err := selectSparseColumns(ctx, rows, fold, capacity)
	if err != nil {
		return fitted{}, err
	}
	examples, groups := sparseTraining(rows, fold, spec)
	trained, err := sparsemodel.Fit(ctx, examples, len(spec.ids))
	if err != nil {
		return fitted{}, err
	}
	development, err := predictSparse(ctx, trained, spec, rows, (fold+1)%5, true)
	if err != nil {
		return fitted{}, err
	}
	point, err := threshold(development)
	if err != nil {
		return fitted{}, err
	}
	predictions, err := predictSparse(ctx, trained, spec, rows, fold, false)
	if err != nil {
		return fitted{}, err
	}
	for i := range predictions {
		predictions[i].Selected = point.Available && predictions[i].Score >= *point.Threshold
	}
	encoded, err := json.Marshal(examples)
	if err != nil {
		return fitted{}, err
	}
	return fitted{Kind: fmt.Sprintf("SW%d", capacity), Fold: fold, Features: spec.ids, Parameters: trained.Parameters,
		TrainingHash: fmt.Sprintf("%x", sha256.Sum256(encoded)), TrainingRows: len(examples), TrainingGroups: groups,
		Options:    model.FitOptions{L2: 0.01, Tolerance: 1e-8, MaxIterations: 500, MaxOperations: 3000000000},
		Iterations: trained.Iterations, Operations: trained.Operations, Gradient: trained.Gradient,
		OperatingPoint: point, Predictions: predictions}, nil
}

func predictSparse(ctx context.Context, classifier sparsemodel.Result, spec sparseColumns, rows []row, fold int,
	labeledOnly bool,
) ([]prediction, error) {
	var predictions []prediction
	for _, row := range rows {
		if row.fold != fold || labeledOnly && row.label == nil {
			continue
		}
		score, err := classifier.Score(ctx, sparseVector(row, spec))
		if err != nil {
			return nil, err
		}
		predictions = append(predictions, prediction{Page: row.page, Unit: row.unit, Words: row.words,
			Cohort: row.cohort, Label: row.label, Score: score})
	}
	return predictions, nil
}
