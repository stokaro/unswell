package training

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// referenceSource is a feature source that measures prepared targets against
// an external reference: the cohorts of a compression bank or the dictionary
// tables of the LLMDet port. It names its columns, its contract, and the
// digest of the reference, and it measures every target of the fitted kind.
type referenceSource struct {
	source, contract, preprocessing, hash string
	columns                               []feature.Descriptor
	measure                               func(context.Context, corpus.Prepared, string) (map[string]map[string]feature.Value, error)
}

// referenceColumnPrefixes are the column families reference sources add
// beside prepared features.
var referenceColumnPrefixes = []string{"compression.", "llmdet."}

// preparedFeatureIDs returns the prepared features of a reference fit,
// without its reference columns.
func preparedFeatureIDs(features []string) []string {
	result := make([]string, 0, len(features))
	for _, id := range features {
		if !isReferenceColumn(id) {
			result = append(result, id)
		}
	}
	return result
}

func isReferenceColumn(id string) bool {
	for _, prefix := range referenceColumnPrefixes {
		if strings.HasPrefix(id, prefix) {
			return true
		}
	}
	return false
}

// validateReferenceFeatures admits prepared features beside the reference
// columns and nothing else.
func validateReferenceFeatures(features []string) error {
	for _, id := range features {
		if strings.HasPrefix(id, "activation/") || isReferenceColumn(id) {
			return fmt.Errorf("reference training adds reference columns to prepared features only")
		}
	}
	return nil
}

func referenceRoundInputs(ctx context.Context, candidates corpus.Artifact, round *annotation.Round,
	files map[string][]byte, features []string,
) (*corpus.JoinedArtifact, annotation.DecisionSet, error) {
	if len(features) != 0 {
		joined, err := corpus.Join(ctx, candidates, round, files, features)
		if err != nil {
			return nil, annotation.DecisionSet{}, err
		}
		return &joined, joined.Decisions, nil
	}
	expected := make([]annotation.Unit, len(candidates.Units))
	for i, candidate := range candidates.Units {
		expected[i] = candidate.Unit
	}
	if err := round.MatchTargets(ctx, expected); err != nil {
		return nil, annotation.DecisionSet{}, err
	}
	decisions, err := round.Decisions(ctx)
	return nil, decisions, err
}

func referenceDecisionInputs(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, features []string,
) (*corpus.JoinedArtifact, error) {
	if len(features) != 0 {
		joined, err := corpus.JoinDecisions(ctx, candidates, decisions, files, features)
		if err != nil {
			return nil, err
		}
		return &joined, nil
	}
	return nil, corpus.MatchDecisionTargets(ctx, candidates, decisions)
}

func fitReference(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	decisions annotation.DecisionSet, options Options, source referenceSource, joined *corpus.JoinedArtifact,
) (Artifact, error) {
	if decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return Artifact{}, err
	}
	selector, bindings, columns, err := referenceSelector(ctx, candidates, prepared, decisions, options, source, joined)
	if err != nil {
		return Artifact{}, err
	}
	options.Features = make([]string, 0, len(columns))
	for _, column := range columns {
		options.Features = append(options.Features, column.ID)
	}
	selected, err := selectMeasuredRows(ctx, candidates.Plan, decisions, bindings, selector)
	if err != nil {
		return Artifact{}, err
	}
	joinedHash, err := hashJSON(struct {
		Corpus, Round, Reference string
		Bindings                 []corpus.FeatureBinding
	}{candidates.SHA256, decisions.RoundSHA256, source.hash, bindings})
	if err != nil {
		return Artifact{}, err
	}
	return fitSelected(ctx, candidates, decisions, joinedHash, options, selected)
}

// referenceSelector builds the rows of a reference fit: reference columns
// alone, or prepared features joined with them, in one sorted column order.
func referenceSelector(ctx context.Context, candidates corpus.Artifact, prepared corpus.Prepared,
	decisions annotation.DecisionSet, options Options, source referenceSource, joined *corpus.JoinedArtifact,
) (rowSelector, []corpus.FeatureBinding, []feature.Descriptor, error) {
	identity, err := referenceIdentity(candidates, prepared, decisions, options, source, joined)
	if err != nil {
		return rowSelector{}, nil, nil, err
	}
	measured, err := source.measure(ctx, prepared, options.Kind)
	if err != nil {
		return rowSelector{}, nil, nil, err
	}
	if joined != nil {
		selector := joinedReferenceSelector(*joined, identity, options, measured)
		return selector, joined.Bindings, identity.Columns, nil
	}
	paths := make(map[string]string)
	for _, s := range candidates.Plan.Manifest.Sources {
		paths[s.ID] = s.Path
	}
	units := make(map[string]measurement)
	bindings := make([]corpus.FeatureBinding, 0, len(prepared.Targets))
	for i, target := range prepared.Targets {
		candidate := candidates.Units[i]
		if target.UnitID != candidate.Unit.ID {
			return rowSelector{}, nil, nil, fmt.Errorf("reference targets are out of order")
		}
		hash, err := hashJSON(struct {
			Source   string
			Binding  nlp.UnitBinding
			Identity Identity
		}{candidate.Unit.Source.SHA256, target.Unit.Binding(), identity})
		if err != nil {
			return rowSelector{}, nil, nil, err
		}
		binding := corpus.FeatureBinding{UnitID: target.UnitID, SourceID: candidate.SourceID, GroupID: candidate.GroupID,
			Path: paths[candidate.SourceID], Partition: candidate.Partition, FeatureInputHash: hash}
		bindings = append(bindings, binding)
		units[measurementKey(binding.Path, hash)] = measurement{kind: candidate.Unit.Kind, identity: identity,
			values: orderedValues(identity.Columns, nil, measured[target.UnitID])}
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		return unit, exists
	}}
	return selector, bindings, identity.Columns, ctx.Err()
}

// joinedReferenceSelector joins prepared measurements with the reference
// columns of the same unit.
func joinedReferenceSelector(joined corpus.JoinedArtifact, identity Identity, options Options,
	measured map[string]map[string]feature.Value,
) rowSelector {
	units := make(map[string]measurement)
	for _, s := range joined.Features.Sources {
		sourceID := sourceIdentity(identity, s)
		for _, unit := range s.Units {
			units[measurementKey(s.Path, unit.InputHash)] = measurement{
				kind: unit.Binding.Kind, identity: sourceID, values: unit.Values}
		}
	}
	return rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		if !exists {
			return measurement{}, false
		}
		unit.values = orderedValues(identity.Columns, unit.values, measured[binding.UnitID])
		return unit, true
	}}
}

// orderedValues returns one value per column. It takes the prepared values
// first and the reference measurements second. A column that has neither
// stays unavailable, with a reason.
func orderedValues(columns []feature.Descriptor, prepared []feature.Value, references map[string]feature.Value) []feature.Value {
	byID := make(map[string]feature.Value, len(prepared)+len(references))
	for _, value := range prepared {
		byID[value.ID] = value
	}
	maps.Copy(byID, references)
	values := make([]feature.Value, len(columns))
	for i, column := range columns {
		value, exists := byID[column.ID]
		if !exists {
			value = feature.Value{ID: column.ID, Version: column.Version, Unit: column.Unit, Reason: "unmeasured"}
		}
		values[i] = value
	}
	return values
}

// referenceIdentity is the identity of a reference fit: the prepared
// identity, or the prepared-target identity when no prepared feature joins,
// with the reference columns merged in sorted order and the source named.
func referenceIdentity(candidates corpus.Artifact, prepared corpus.Prepared, decisions annotation.DecisionSet,
	options Options, source referenceSource, joined *corpus.JoinedArtifact,
) (Identity, error) {
	var identity Identity
	var err error
	if joined != nil {
		identity, err = columnIdentity(*joined, options.Kind)
	} else {
		identity, err = preparedTargetIdentity(candidates, prepared, decisions, options.Kind, source.contract)
	}
	if err != nil {
		return Identity{}, err
	}
	identity.Columns = append(slices.Clone(identity.Columns), source.columns...)
	slices.SortFunc(identity.Columns, func(a, b feature.Descriptor) int { return strings.Compare(a.ID, b.ID) })
	identity.ColumnsSHA256, err = hashJSON(identity.Columns)
	if err != nil {
		return Identity{}, err
	}
	identity.FeatureSource, identity.Context, identity.Preprocessing = source.source, "prepared_piece", source.preprocessing
	switch source.source {
	case "compression_bank":
		identity.CompressionBankHash = source.hash
	case "llmdet_tables":
		identity.LLMDetPackHash = source.hash
	}
	return identity, nil
}

// preparedTargetIdentity is the identity of rows measured on prepared targets
// without the engine's feature join, as the lexical baseline records it.
func preparedTargetIdentity(candidates corpus.Artifact, prepared corpus.Prepared, decisions annotation.DecisionSet,
	kind, contract string,
) (Identity, error) {
	policy, err := hashJSON(prepared.Policy.Extraction)
	if err != nil {
		return Identity{}, err
	}
	preparation, err := nlp.PreparationHash(policy, false, false)
	if err != nil {
		return Identity{}, err
	}
	task, err := annotation.TaskForRubric(decisions.Rubric)
	if err != nil {
		return Identity{}, err
	}
	return Identity{Task: task, Rubric: decisions.Rubric, ProfileSHA256: decisions.ProfileSHA256,
		Kind: kind, FeatureContract: contract, UnitContract: nlp.UnitContract, NLP: candidates.Pipeline.NLP,
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, PolicyHash: prepared.Policy.Hash,
		VocabularyHash: prepared.Policy.Hash, ExtractionPolicyHash: policy, PreparationHash: preparation}, nil
}

// referencePredictionMeasurements measures every target the same way the
// fit did; the predictor keeps the planned partition.
func referencePredictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, source referenceSource,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return rowSelector{}, nil, corpus.Verification{}, err
	}
	options := fitted.Options
	options.Features = preparedFeatureIDs(fitted.Options.Features)
	decisions := annotation.DecisionSet{Rubric: fitted.Identity.Rubric, ProfileSHA256: fitted.Identity.ProfileSHA256}
	var joined *corpus.JoinedArtifact
	verification := prepared.Verification
	if len(options.Features) != 0 {
		measured, err := corpus.Measure(ctx, candidates, files, options.Features)
		if err != nil {
			return rowSelector{}, nil, corpus.Verification{}, err
		}
		joined = &corpus.JoinedArtifact{Features: measured.Features, Bindings: measured.Bindings, Decisions: decisions}
		verification = measured.Verification
	}
	selector, bindings, _, err := referenceSelector(ctx, candidates, prepared, decisions, options, source, joined)
	return selector, bindings, verification, err
}

// referenceColumns counts the reference columns of a fitted artifact.
func referenceColumns(a Artifact) int {
	return len(a.Options.Features) - len(preparedFeatureIDs(a.Options.Features))
}
