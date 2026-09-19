package representation

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
)

type tokenBlock struct {
	ID       int             `json:"block_id"`
	Kind     string          `json:"kind"`
	Excluded bool            `json:"excluded"`
	Segments []document.Span `json:"segments"`
}

type tokenSource struct {
	Path       string       `json:"path"`
	SourceHash string       `json:"source_hash"`
	Blocks     []tokenBlock `json:"blocks"`
}

// Original token collection also covers list items and headings. The older
// block feature catalog deliberately does not measure every extraction kind.
func originalTokens(ctx context.Context, sources []document.Source, engine *unswell.Engine,
	provider nlp.Provider, features *unswell.FeatureCollection,
) ([]tokenSource, error) {
	bindings := make(map[string]unswell.FeatureSource)
	for _, source := range features.Sources {
		bindings[source.Path] = source
	}
	result := make([]tokenSource, 0, len(sources))
	for _, source := range sources {
		row, err := originalSourceTokens(ctx, source, engine, provider, bindings[source.Name])
		if err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, nil
}

func originalSourceTokens(ctx context.Context, source document.Source, engine *unswell.Engine,
	provider nlp.Provider, binding unswell.FeatureSource,
) (tokenSource, error) {
	policy, err := engine.ResolvedPolicy(source.Name)
	if err != nil {
		return tokenSource{}, err
	}
	doc, err := extract.Parse(ctx, source, extract.Options{
		IncludeStructure: true, IncludeQuotes: policy.Analysis.IncludeQuotes,
		MaxBytes: policy.Analysis.MaxFileBytes, MaxBlocks: policy.Analysis.MaxBlocks, Policy: policy.Extraction,
	})
	if err != nil {
		return tokenSource{}, err
	}
	if binding.SourceHash != doc.Hash || len(binding.Units) != len(doc.Blocks) {
		return tokenSource{}, fmt.Errorf("original source binding mismatch")
	}
	row := tokenSource{Path: source.Name, SourceHash: doc.Hash, Blocks: make([]tokenBlock, 0, len(doc.Blocks))}
	for i, block := range doc.Blocks {
		unit := binding.Units[i]
		if unit.UnitID != block.ID || unit.Binding == nil ||
			unit.Binding.TextSHA256 != fmt.Sprintf("%x", sha256.Sum256([]byte(block.Text))) {
			return tokenSource{}, fmt.Errorf("original block binding mismatch")
		}
		// Respect the engine's post-extraction exclusions, including non-Latin prose.
		block.Excluded = unit.Excluded
		value, err := blockTokens(ctx, block, provider)
		if err != nil {
			return tokenSource{}, err
		}
		row.Blocks = append(row.Blocks, value)
	}
	return row, nil
}

func blockTokens(ctx context.Context, block document.Block, provider nlp.Provider) (tokenBlock, error) {
	row := tokenBlock{ID: block.ID, Kind: block.Kind, Excluded: block.Excluded, Segments: []document.Span{}}
	if block.Excluded {
		return row, nil
	}
	sentences, err := provider.Analyze(ctx, block.MappedText, []nlp.Capability{nlp.Tokens, nlp.Sentences})
	if err != nil {
		return tokenBlock{}, err
	}
	for _, sentence := range sentences {
		for _, token := range sentence.Tokens {
			if token.Word && !token.Protected {
				row.Segments = append(row.Segments, token.Spans...)
			}
		}
	}
	return row, nil
}
