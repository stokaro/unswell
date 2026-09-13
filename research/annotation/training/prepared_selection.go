package training

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func selectRows(ctx context.Context, plan corpus.Plan, joined corpus.JoinedArtifact, options Options) (selection, error) {
	selector, err := preparedSelector(joined, options)
	if err != nil {
		return selection{}, err
	}
	return selectMeasuredRows(ctx, plan, joined.Decisions, joined.Bindings, selector)
}

func preparedSelector(joined corpus.JoinedArtifact, options Options) (rowSelector, error) {
	columns, err := columnIdentity(joined, options.Kind)
	if err != nil {
		return rowSelector{}, err
	}
	words, err := wordColumn(joined.Features.Requested, options)
	if err != nil {
		return rowSelector{}, err
	}
	units := make(map[string]measurement)
	for _, source := range joined.Features.Sources {
		identity := sourceIdentity(columns, source)
		for _, unit := range source.Units {
			units[measurementKey(source.Path, unit.InputHash)] = measurement{
				kind: unit.Binding.Kind, words: unitWords(unit.Values, words), identity: identity, values: unit.Values}
		}
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		return unit, exists
	}}
	return selector, nil
}

// wordColumn locates the prose-words column a word band reads, and reports -1
// when no band is set. The band counts the same prose words the corpus counts
// when it records a candidate's length, so the feature must be requested for a
// band to mean anything.
func wordColumn(requested []string, options Options) (int, error) {
	if options.MinUnitWords == 0 && options.MaxUnitWords == 0 {
		return -1, nil
	}
	index := slices.Index(requested, "prose-words")
	if index < 0 {
		return 0, fmt.Errorf("a word band needs the prose-words feature among the requested set")
	}
	return index, nil
}

// unitWords returns the counted prose words of one prepared unit, or zero when
// the band is off or the value is missing.
func unitWords(values []feature.Value, column int) int {
	if column < 0 || column >= len(values) || values[column].Number == nil {
		return 0
	}
	return int(*values[column].Number)
}
