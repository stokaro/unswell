package reviewbaseline

import (
	"context"
	"fmt"
	"reflect"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func prepare(ctx context.Context, in input) ([]row, nlp.Identity, error) {
	provider, err := english.New()
	if err != nil {
		return nil, nlp.Identity{}, err
	}
	engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:technical-v1]\n"), NLP: provider})
	if err != nil {
		return nil, nlp.Identity{}, err
	}
	var rows []row
	for _, page := range in.Pages {
		prepared, err := preparePage(ctx, page, engine, provider)
		if err != nil {
			return nil, nlp.Identity{}, fmt.Errorf("%s: %w", page.ID, err)
		}
		rows = append(rows, prepared...)
	}
	return rows, provider.Identity(), validateReuse(rows)
}

func preparePage(ctx context.Context, page page, engine *unswell.Engine, provider nlp.Provider) ([]row, error) {
	policy, err := engine.ResolvedPolicy(page.ID)
	if err != nil {
		return nil, err
	}
	doc, err := extract.Parse(ctx, document.Source{Name: page.ID, Format: page.Format, Bytes: []byte(page.Text)}, extract.Options{
		IncludeStructure: true, IncludeQuotes: policy.Analysis.IncludeQuotes,
		MaxBytes: policy.Analysis.MaxFileBytes, MaxBlocks: policy.Analysis.MaxBlocks, Policy: policy.Extraction})
	if err != nil {
		return nil, err
	}
	options := nlp.UnitOptions{Kinds: []string{"paragraph", "fragment"},
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS},
		Limits: nlp.UnitLimits{MaxBytes: policy.Analysis.MaxFileBytes, MaxContextBytes: 65536,
			MaxUnits: 10000, MaxTokens: min(policy.Analysis.MaxTokens, 1000000), MaxSegments: 4096}}
	var rows []row
	for _, block := range doc.Blocks {
		if slices.Contains(page.Excluded, block.ID) {
			block.Excluded = true
		}
		units, err := nlp.PrepareUnits(ctx, block, provider, options)
		if err != nil {
			return nil, err
		}
		for _, unit := range units {
			value, err := measureRow(ctx, page, len(rows), unit, policy.Hash)
			if err != nil {
				return nil, err
			}
			rows = append(rows, value)
		}
	}
	if len(rows) != len(page.Units) {
		return nil, fmt.Errorf("prepared unit count changed")
	}
	return rows, nil
}

func measureRow(ctx context.Context, page page, ordinal int, unit nlp.PreparedUnit, policy string) (row, error) {
	if ordinal >= len(page.Units) || !reflect.DeepEqual(unit.Binding(), page.Units[ordinal].Binding) {
		return row{}, fmt.Errorf("prepared unit binding changed at %d", ordinal)
	}
	label := page.Units[ordinal].Label
	if label != nil && *label != 0 && *label != 1 {
		return row{}, fmt.Errorf("invalid binary label at %d", ordinal)
	}
	identity := feature.Identity{NLP: unit.Identity(), Capabilities: unit.Capabilities(), Source: page.Hash,
		Policy: policy, Vocabulary: "review-baseline-unfitted", Preprocessing: nlp.UnitContract}
	limits := feature.Limits{MaxTokens: 1000000, MaxUniqueWords: 65536, MaxBytes: 2 << 20, MaxBlocks: 1}
	measurements, err := feature.MeasureUnit(ctx, unit, identity, limits)
	if err != nil {
		return row{}, err
	}
	terms, err := feature.CountLexical(ctx, unit, identity, limits, feature.LexicalOptions{WordMin: 1, WordMax: 3})
	if err != nil {
		return row{}, err
	}
	value := row{page: page.ID, group: page.Group, cohort: page.Cohort, textHash: unit.Binding().TextSHA256,
		text: unit.Block().Text, unit: ordinal, fold: page.Fold, words: measurements.Counts().Words,
		label: label, terms: make(map[string]bool)}
	value.numeric = numericValues(measurements)
	for _, term := range terms {
		value.terms[term.Key] = true
	}
	return value, nil
}

func numericValues(measurements feature.Measurements) []float64 {
	var values []float64
	for _, number := range measurements.Values() {
		if number.Number == nil {
			values = append(values, 0, 1)
		} else {
			values = append(values, *number.Number, 0)
		}
	}
	return values
}

func validateReuse(rows []row) error {
	groups := make(map[string]string)
	for _, row := range rows {
		if row.words < 12 {
			continue
		}
		if previous, ok := groups[row.textHash]; ok && previous != row.group {
			return fmt.Errorf("reused long prose crosses source groups")
		}
		groups[row.textHash] = row.group
	}
	return nil
}
