package training

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/llmdet"
)

// llmdetPreprocessing names what an LLMDet measurement reads: the UTF-8 text
// of the prepared target, tokenized by each model's byte-level tokenizer.
const llmdetPreprocessing = "shared-prepared-target/byte-level-bpe-v1"

// LLMDetContract names the feature contract of the proxy columns.
const LLMDetContract = "unswell-llmdet-proxy-features-v1"

// llmdetSource is the reference source of a table pack: a proxy perplexity
// and a context coverage per model.
func llmdetSource(kind string, pack LLMDetPack) referenceSource {
	return referenceSource{source: "llmdet_tables", contract: LLMDetContract, preprocessing: llmdetPreprocessing,
		hash: pack.SHA256, columns: llmdetColumns(kind, pack),
		measure: func(ctx context.Context, prepared corpus.Prepared, kind string) (map[string]map[string]feature.Value, error) {
			return measureLLMDet(ctx, prepared, kind, pack)
		}}
}

// llmdetColumns describes the two measurements of every model, with the
// model name as a column suffix.
func llmdetColumns(kind string, pack LLMDetPack) []feature.Descriptor {
	columns := make([]feature.Descriptor, 0, 2*len(pack.Models))
	for _, model := range pack.Models {
		columns = append(columns,
			feature.Descriptor{ID: "llmdet.proxy-perplexity/" + model.Model, Version: "1", Family: "llmdet", Type: "number",
				Unit: "bits/matched-position", Scope: kind, Normalization: llmdet.ProxyVersion,
				Formula:  "Negative sum of the base-2 log likelihoods of the matched positions, divided by the matched positions plus one.",
				Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}, MinWords: 1,
				Missing:     "No value with fewer than four tokens, with no matching context, or with no evaluated likelihood.",
				Limitations: "A numerical proxy of the published calculation, not a probability of origin or a judgment of quality."},
			feature.Descriptor{ID: "llmdet.context-coverage/" + model.Model, Version: "1", Family: "llmdet", Type: "number",
				Unit: "matched/possible", Scope: kind, Normalization: llmdet.ProxyVersion,
				Formula:  "Positions with a matching context of any order, divided by the positions the reference examines.",
				Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}, MinWords: 1,
				Missing:     "No value with fewer than four tokens.",
				Limitations: "Coverage of one model's dictionary, not evidence of origin."})
	}
	return columns
}

// measureLLMDet measures every prepared target of the kind against each
// model's table, one table in memory at a time.
func measureLLMDet(ctx context.Context, prepared corpus.Prepared, kind string,
	pack LLMDetPack,
) (map[string]map[string]feature.Value, error) {
	result := make(map[string]map[string]feature.Value)
	for _, model := range pack.Models {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if err := measureLLMDetModel(ctx, prepared, kind, pack, model, result); err != nil {
			return nil, fmt.Errorf("LLMDet model %s: %w", model.Model, err)
		}
	}
	return result, ctx.Err()
}

func measureLLMDetModel(ctx context.Context, prepared corpus.Prepared, kind string, pack LLMDetPack,
	model LLMDetModel, result map[string]map[string]feature.Value,
) error {
	tokenizer, table, err := loadLLMDetModel(ctx, pack, model)
	if err != nil {
		return err
	}
	for _, target := range prepared.Targets {
		if target.Unit.Binding().Kind != kind {
			continue
		}
		tokens, err := tokenizer.Encode(ctx, target.Unit.Block().Text)
		if err != nil {
			return fmt.Errorf("unit %s: %w", target.UnitID, err)
		}
		measured, err := table.Measure(ctx, tokens)
		if err != nil {
			return fmt.Errorf("unit %s: %w", target.UnitID, err)
		}
		values := result[target.UnitID]
		if values == nil {
			values = make(map[string]feature.Value)
			result[target.UnitID] = values
		}
		for _, value := range llmdetValues(model.Model, measured) {
			values[value.ID] = value
		}
	}
	return nil
}

func loadLLMDetModel(ctx context.Context, pack LLMDetPack, model LLMDetModel) (*llmdet.BPE, *llmdet.Table, error) {
	vocabulary, err := pack.read(model.Vocabulary)
	if err != nil {
		return nil, nil, err
	}
	merges, err := pack.read(model.Merges)
	if err != nil {
		return nil, nil, err
	}
	tokenizer, err := llmdet.LoadBPE(ctx, vocabulary, merges)
	if err != nil {
		return nil, nil, err
	}
	path, err := pack.path(model.Table)
	if err != nil {
		return nil, nil, err
	}
	table, err := llmdet.LoadTable(ctx, path)
	if err != nil {
		return nil, nil, err
	}
	if table.Header().Model != model.Model {
		return nil, nil, fmt.Errorf("table names model %s", table.Header().Model)
	}
	return tokenizer, table, nil
}

// llmdetValues turns one proxy result into the model's two values, with the
// reference's reasons where no value exists.
func llmdetValues(model string, measured llmdet.ProxyResult) []feature.Value {
	perplexity := feature.Value{ID: "llmdet.proxy-perplexity/" + model, Version: "1", Unit: "bits/matched-position"}
	coverage := feature.Value{ID: "llmdet.context-coverage/" + model, Version: "1", Unit: "matched/possible"}
	if measured.Value != nil {
		value := *measured.Value
		perplexity.Number = &value
	} else {
		perplexity.Reason = measured.Reason
	}
	if measured.Possible > 0 {
		value := measured.Coverage
		coverage.Number = &value
	} else {
		coverage.Reason = "insufficient_tokens"
	}
	return []feature.Value{perplexity, coverage}
}

// llmdetPredictionMeasurements measures every target the same way the fit
// did; the predictor keeps the planned partition.
func llmdetPredictionMeasurements(ctx context.Context, candidates corpus.Artifact, files map[string][]byte,
	fitted Artifact, pack *LLMDetPack,
) (rowSelector, []corpus.FeatureBinding, corpus.Verification, error) {
	if pack == nil || pack.SHA256 != fitted.Identity.LLMDetPackHash {
		return rowSelector{}, nil, corpus.Verification{}, fmt.Errorf("LLMDet prediction requires the fitted table pack")
	}
	return referencePredictionMeasurements(ctx, candidates, files, fitted, llmdetSource(fitted.Options.Kind, *pack))
}
