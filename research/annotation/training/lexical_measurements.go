package training

import (
	"context"
	"fmt"
	"math"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

const lexicalPreprocessing = "shared-prepared-target/log1p-count-v1"

func lexicalCounts(ctx context.Context, unit nlp.PreparedUnit, options feature.LexicalOptions) ([]feature.LexicalTerm, error) {
	identity := feature.Identity{NLP: unit.Identity(), Capabilities: unit.Capabilities(), Source: "verified-corpus",
		Policy: "frozen-corpus-extraction", Vocabulary: "count-all-keys", Preprocessing: lexicalPreprocessing}
	limits := feature.Limits{MaxTokens: 1000000, MaxUniqueWords: 65536, MaxBytes: corpus.MaxSourceBytes, MaxBlocks: 1}
	return feature.CountLexical(ctx, unit, identity, limits, options)
}

func lexicalIdentity(candidates corpus.Artifact, prepared corpus.Prepared, decisions annotation.DecisionSet,
	options Options, v *Vocabulary,
) (Identity, error) {
	policy, err := hashJSON(prepared.Policy.Extraction)
	if err != nil {
		return Identity{}, err
	}
	columns := []feature.Descriptor{{ID: "lexical-selection", Version: "1", Unit: "selection"}}
	vocabulary := "not_fitted"
	if v != nil {
		columns, vocabulary = lexicalColumns(*v, options.Kind), v.SHA256
	}
	preparation, err := nlp.PreparationHash(policy, false, false)
	if err != nil {
		return Identity{}, err
	}
	columnHash, err := hashJSON(columns)
	if err != nil {
		return Identity{}, err
	}
	return Identity{FeatureSource: "lexical_ngrams", Context: "prepared_piece", Preprocessing: lexicalPreprocessing,
		Task: "editorial_needs_revision", Rubric: decisions.Rubric, ProfileSHA256: decisions.ProfileSHA256,
		Kind: options.Kind, FeatureContract: feature.LexicalCountContract, UnitContract: nlp.UnitContract,
		Columns: columns, ColumnsSHA256: columnHash, NLP: candidates.Pipeline.NLP,
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, PolicyHash: prepared.Policy.Hash,
		VocabularyHash: prepared.Policy.Hash, LexicalVocabularyHash: vocabulary,
		ExtractionPolicyHash: policy, PreparationHash: preparation}, nil
}

func lexicalSelector(ctx context.Context, candidates corpus.Artifact, prepared corpus.Prepared, decisions annotation.DecisionSet,
	options Options, vocabulary *Vocabulary, histograms map[string][]feature.LexicalTerm,
) (rowSelector, []corpus.FeatureBinding, error) {
	identity, err := lexicalIdentity(candidates, prepared, decisions, options, vocabulary)
	if err != nil {
		return rowSelector{}, nil, err
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
			return rowSelector{}, nil, fmt.Errorf("lexical target order mismatch")
		}
		values := lexicalZeroValues(identity.Columns)
		if terms, measured := histograms[target.UnitID]; measured && vocabulary != nil {
			values = lexicalValues(terms, *vocabulary, identity.Columns)
		}
		hash, err := hashJSON(struct {
			Source   string
			Binding  nlp.UnitBinding
			Identity Identity
		}{candidate.Unit.Source.SHA256, target.Unit.Binding(), identity})
		if err != nil {
			return rowSelector{}, nil, err
		}
		binding := corpus.FeatureBinding{UnitID: target.UnitID, SourceID: candidate.SourceID, GroupID: candidate.GroupID,
			Path: paths[candidate.SourceID], Partition: candidate.Partition, FeatureInputHash: hash}
		bindings = append(bindings, binding)
		units[measurementKey(binding.Path, hash)] = measurement{kind: candidate.Unit.Kind, identity: identity, values: values}
	}
	selector := rowSelector{options: options, measure: func(binding corpus.FeatureBinding) (measurement, bool) {
		unit, exists := units[measurementKey(binding.Path, binding.FeatureInputHash)]
		return unit, exists
	}}
	return selector, bindings, ctx.Err()
}

func lexicalZeroValues(columns []feature.Descriptor) []feature.Value {
	values := make([]feature.Value, len(columns))
	for i, column := range columns {
		values[i] = feature.Value{ID: column.ID, Version: column.Version, Unit: column.Unit, Number: new(float64)}
	}
	return values
}

func lexicalValues(terms []feature.LexicalTerm, vocabulary Vocabulary,
	columns []feature.Descriptor,
) []feature.Value {
	counts := make(map[string]int, len(terms))
	for _, term := range terms {
		counts[term.Key] = term.Count
	}
	values := lexicalZeroValues(columns)
	for i, term := range vocabulary.Terms {
		*values[i].Number = math.Log1p(float64(counts[term.Key]))
	}
	return values
}

func lexicalPredictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, configuration []byte, partition string,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	if len(configuration) != 0 {
		return rowSelector{}, nil, corpus.Verification{}, fmt.Errorf("lexical prediction rejects rule configuration")
	}
	prepared, err := corpus.Prepare(ctx, candidates, files)
	if err != nil {
		return rowSelector{}, nil, corpus.Verification{}, err
	}
	selected := make(map[string]bool)
	for _, candidate := range candidates.Units {
		if candidate.Partition == partition && candidate.Unit.Kind == fitted.Options.Kind {
			selected[candidate.Unit.ID] = true
		}
	}
	counts, err := collectLexical(ctx, prepared, selected, fitted.Lexical.Options.Counts)
	if err != nil {
		return rowSelector{}, nil, corpus.Verification{}, err
	}
	decisions := annotation.DecisionSet{Rubric: fitted.Identity.Rubric, ProfileSHA256: fitted.Identity.ProfileSHA256}
	selector, bindings, err := lexicalSelector(ctx, candidates, prepared, decisions, fitted.Options, fitted.Lexical, counts)
	return selector, bindings, prepared.Verification, err
}

func collectLexical(ctx context.Context, prepared corpus.Prepared, selected map[string]bool,
	options feature.LexicalOptions,
) (map[string][]feature.LexicalTerm, error) {
	counts := make(map[string][]feature.LexicalTerm, len(selected))
	var retained int64
	for _, target := range prepared.Targets {
		if !selected[target.UnitID] {
			continue
		}
		terms, err := lexicalCounts(ctx, target.Unit, options)
		if err != nil {
			return nil, err
		}
		for _, term := range terms {
			retained += int64(len(term.Key)) + 64
		}
		if retained > 64<<20 {
			return nil, fmt.Errorf("selected lexical counts exceed retained-byte budget")
		}
		counts[target.UnitID] = terms
	}
	if len(counts) != len(selected) {
		return nil, fmt.Errorf("selected lexical target has no prepared unit")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}
