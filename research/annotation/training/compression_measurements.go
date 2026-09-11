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

// compressionColumns describes the two measurements of every reference
// cohort, with the cohort ID as a column suffix.
func compressionColumns(kind string, bank corpus.CompressionBank) ([]feature.Descriptor, error) {
	catalog, err := feature.CompressionCatalog(kind)
	if err != nil {
		return nil, err
	}
	columns := make([]feature.Descriptor, 0, len(catalog)*len(bank.Cohorts))
	for _, cohort := range bank.Cohorts {
		for _, descriptor := range catalog {
			column := descriptor
			column.ID = descriptor.ID + "/" + cohort.ID
			column.Requires = slices.Clone(descriptor.Requires)
			columns = append(columns, column)
		}
	}
	return columns, nil
}

// measureCompression measures every prepared target of the kind against each
// reference cohort. The result maps a unit to its values, keyed by column ID.
func measureCompression(ctx context.Context, prepared corpus.Prepared, kind string,
	bank corpus.CompressionBank,
) (map[string]map[string]feature.Value, error) {
	compressors, err := bankCompressors(ctx, bank)
	if err != nil {
		return nil, err
	}
	limits := feature.Limits{MaxTokens: 1000000, MaxUniqueWords: 65536, MaxBytes: corpus.MaxSourceBytes, MaxBlocks: 1}
	result := make(map[string]map[string]feature.Value)
	for _, target := range prepared.Targets {
		if target.Unit.Binding().Kind != kind {
			continue
		}
		values, err := measureTarget(ctx, compressors, bank, target, limits)
		if err != nil {
			return nil, err
		}
		result[target.UnitID] = values
	}
	return result, ctx.Err()
}

// bankCompressors initializes one compressor per reference cohort and checks
// each against the identity the bank recorded.
func bankCompressors(ctx context.Context, bank corpus.CompressionBank) ([]*feature.Compression, error) {
	compressors := make([]*feature.Compression, len(bank.Cohorts))
	for i, cohort := range bank.Cohorts {
		compressor, err := feature.NewCompression(ctx, cohort.Reference, bank.Options.Compression)
		if err != nil {
			return nil, err
		}
		if compressor.Identity().ReferenceSHA256 != cohort.Identity.ReferenceSHA256 {
			return nil, fmt.Errorf("compression reference %s does not match the bank's identity", cohort.ID)
		}
		compressors[i] = compressor
	}
	return compressors, nil
}

func measureTarget(ctx context.Context, compressors []*feature.Compression, bank corpus.CompressionBank,
	target corpus.PreparedTarget, limits feature.Limits,
) (map[string]feature.Value, error) {
	identity := feature.Identity{NLP: target.Unit.Identity(), Capabilities: target.Unit.Capabilities(),
		Source: "verified-corpus", Policy: "frozen-corpus-extraction", Vocabulary: "reference-bank",
		Preprocessing: compressionPreprocessing}
	values := make(map[string]feature.Value)
	for i, cohort := range bank.Cohorts {
		measured, err := compressors[i].Measure(ctx, target.Unit, identity, limits)
		if err != nil {
			return nil, fmt.Errorf("unit %s against reference %s: %w", target.UnitID, cohort.ID, err)
		}
		for _, value := range measured.Values {
			value.ID += "/" + cohort.ID
			values[value.ID] = value
		}
	}
	return values, nil
}

// compressionSelector builds the rows of a compression fit: reference columns
// alone, or prepared features joined with them, in one sorted column order.
func compressionSelector(ctx context.Context, candidates corpus.Artifact, prepared corpus.Prepared,
	decisions annotation.DecisionSet, options Options, bank corpus.CompressionBank, joined *corpus.JoinedArtifact,
) (rowSelector, []corpus.FeatureBinding, []feature.Descriptor, error) {
	identity, err := compressionIdentity(candidates, prepared, decisions, options, bank, joined)
	if err != nil {
		return rowSelector{}, nil, nil, err
	}
	measured, err := measureCompression(ctx, prepared, options.Kind, bank)
	if err != nil {
		return rowSelector{}, nil, nil, err
	}
	if joined != nil {
		selector := joinedCompressionSelector(*joined, identity, options, measured)
		return selector, joined.Bindings, identity.Columns, nil
	}
	paths := make(map[string]string)
	for _, source := range candidates.Plan.Manifest.Sources {
		paths[source.ID] = source.Path
	}
	units := make(map[string]measurement)
	bindings := make([]corpus.FeatureBinding, 0, len(prepared.Targets))
	for i, target := range prepared.Targets {
		candidate := candidates.Units[i]
		if target.UnitID != candidate.Unit.ID {
			return rowSelector{}, nil, nil, fmt.Errorf("compression targets are out of order")
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

// joinedCompressionSelector joins prepared measurements with the reference
// columns of the same unit.
func joinedCompressionSelector(joined corpus.JoinedArtifact, identity Identity, options Options,
	measured map[string]map[string]feature.Value,
) rowSelector {
	units := make(map[string]measurement)
	for _, source := range joined.Features.Sources {
		sourceID := sourceIdentity(identity, source)
		for _, unit := range source.Units {
			units[measurementKey(source.Path, unit.InputHash)] = measurement{
				kind: unit.Binding.Kind, identity: sourceID, values: unit.Values}
		}
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		if !exists {
			return measurement{}, false
		}
		unit.values = orderedValues(identity.Columns, unit.values, measured[binding.UnitID])
		return unit, true
	}}
	return selector
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

func compressionIdentity(candidates corpus.Artifact, prepared corpus.Prepared, decisions annotation.DecisionSet,
	options Options, bank corpus.CompressionBank, joined *corpus.JoinedArtifact,
) (Identity, error) {
	references, err := compressionColumns(options.Kind, bank)
	if err != nil {
		return Identity{}, err
	}
	var identity Identity
	if joined != nil {
		identity, err = columnIdentity(*joined, options.Kind)
	} else {
		identity, err = preparedTargetIdentity(candidates, prepared, decisions, options.Kind, feature.CompressionContract)
	}
	if err != nil {
		return Identity{}, err
	}
	identity.Columns = append(slices.Clone(identity.Columns), references...)
	slices.SortFunc(identity.Columns, func(a, b feature.Descriptor) int { return strings.Compare(a.ID, b.ID) })
	identity.ColumnsSHA256, err = hashJSON(identity.Columns)
	if err != nil {
		return Identity{}, err
	}
	identity.FeatureSource, identity.Context = "compression_bank", "prepared_piece"
	identity.Preprocessing, identity.CompressionBankHash = compressionPreprocessing, bank.SHA256
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

// compressionPredictionMeasurements measures every target the same way the
// fit did; the predictor keeps the planned partition.
func compressionPredictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, bank *corpus.CompressionBank,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	if bank == nil || bank.SHA256 != fitted.Identity.CompressionBankHash {
		return rowSelector{}, nil, corpus.Verification{}, fmt.Errorf("compression prediction requires the fitted reference bank")
	}
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
	selector, bindings, _, err := compressionSelector(ctx, candidates, prepared, decisions, options, *bank, joined)
	return selector, bindings, verification, err
}
