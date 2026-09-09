package corpus

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

type ruleTargetKey struct {
	Path, Source, Text, Context string
	Segments                    []document.Span
}

type ruleTarget struct {
	block int
	input string
}

func bindRuleCandidates(ctx context.Context, artifact Artifact,
	collection unswell.FeatureCollection,
) ([]RuleFeatureBinding, error) {
	measured, err := ruleTargets(ctx, collection)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]string, len(artifact.Plan.Manifest.Sources))
	for _, source := range artifact.Plan.Manifest.Sources {
		paths[source.ID] = source.Path
	}
	bindings := make([]RuleFeatureBinding, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		unit := candidate.Unit
		binding := RuleFeatureBinding{UnitID: unit.ID, SourceID: candidate.SourceID, Path: paths[candidate.SourceID],
			GroupID: candidate.GroupID, Partition: candidate.Partition, Kind: unit.Kind, Reason: "no_complete_block_match"}
		key, err := digest(ruleTargetKey{Path: binding.Path, Source: unit.Source.SHA256, Text: hashText(unit.Text),
			Context: hashText(unit.Context), Segments: unit.Source.Segments})
		if err != nil {
			return nil, err
		}
		if target, found := measured[key]; found {
			binding.BlockID, binding.InputHash, binding.Reason = &target.block, target.input, ""
		}
		bindings = append(bindings, binding)
	}
	return bindings, ctx.Err()
}

func ruleTargets(ctx context.Context, collection unswell.FeatureCollection) (map[string]ruleTarget, error) {
	result := make(map[string]ruleTarget)
	for _, source := range collection.Sources {
		identity := source
		identity.Units = nil
		sourceHash, err := digest(identity)
		if err != nil {
			return nil, err
		}
		for _, unit := range source.Units {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			if err := addRuleTarget(result, source, unit, sourceHash); err != nil {
				return nil, err
			}
		}
	}
	return result, ctx.Err()
}

func addRuleTarget(result map[string]ruleTarget, source unswell.FeatureSource, unit unswell.FeatureUnit, sourceHash string) error {
	if unit.Binding == nil || unit.Binding.Contract != unswell.FeatureBlockBindingContract {
		return fmt.Errorf("rule collection lacks a compatible complete block binding")
	}
	key, err := digest(ruleTargetKey{Path: source.Path, Source: source.SourceHash, Text: unit.Binding.TrimmedSHA256,
		Context: unit.Binding.TrimmedSHA256, Segments: unit.Binding.TrimmedSegments})
	if err != nil {
		return err
	}
	if _, found := result[key]; found {
		return fmt.Errorf("rule collection contains an ambiguous original block")
	}
	unit.Values = nil // Bind inputs and policy; observed outcomes are not input identity.
	inputHash, err := digest(struct {
		Contract, Source string
		Unit             unswell.FeatureUnit
	}{RuleJoinedVersion, sourceHash, unit})
	if err != nil {
		return err
	}
	result[key] = ruleTarget{block: unit.UnitID, input: inputHash}
	return nil
}
