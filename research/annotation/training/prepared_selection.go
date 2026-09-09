package training

import (
	"context"

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
	units := make(map[string]measurement)
	for _, source := range joined.Features.Sources {
		identity := sourceIdentity(columns, source)
		for _, unit := range source.Units {
			units[measurementKey(source.Path, unit.InputHash)] = measurement{
				kind: unit.Binding.Kind, identity: identity, values: unit.Values}
		}
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		return unit, exists
	}}
	return selector, nil
}
