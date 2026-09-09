package training

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func selectRuleRows(ctx context.Context, plan corpus.Plan, joined corpus.RuleJoinedArtifact,
	options Options, configHash string,
) (selection, error) {
	columns, err := ruleColumnIdentity(joined, options.Kind, configHash)
	if err != nil {
		return selection{}, err
	}
	sources := make(map[string]unswell.FeatureSource)
	for _, source := range joined.Features.Sources {
		sources[source.Path] = source
	}
	bindings := make([]corpus.FeatureBinding, 0, len(joined.Bindings))
	units := make(map[string]measurement, len(joined.Bindings))
	for _, binding := range joined.Bindings {
		unit, err := ruleMeasurement(binding, sources, columns)
		if err != nil {
			return selection{}, err
		}
		units[binding.UnitID] = unit
		bindings = append(bindings, corpus.FeatureBinding{UnitID: binding.UnitID, SourceID: binding.SourceID,
			Path: binding.Path, GroupID: binding.GroupID, Partition: binding.Partition, FeatureInputHash: binding.InputHash})
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[binding.UnitID]
		return unit, exists
	}}
	return selectMeasuredRows(ctx, plan, joined.Decisions, bindings, selector)
}

func ruleMeasurement(binding corpus.RuleFeatureBinding, sources map[string]unswell.FeatureSource, columns Identity) (measurement, error) {
	unit := measurement{kind: binding.Kind, reason: binding.Reason}
	if binding.BlockID == nil {
		if binding.Reason != "no_complete_block_match" || binding.InputHash != "" {
			return measurement{}, fmt.Errorf("invalid unavailable block binding for %s", binding.UnitID)
		}
		return unit, nil
	}
	source, exists := sources[binding.Path]
	index := *binding.BlockID
	if !exists || index < 0 || index >= len(source.Units) || binding.InputHash == "" || binding.Reason != "" {
		return measurement{}, fmt.Errorf("invalid measured block binding for %s", binding.UnitID)
	}
	unit.values = source.Units[index].Values
	unit.identity = ruleSourceIdentity(columns, source)
	return unit, nil
}
