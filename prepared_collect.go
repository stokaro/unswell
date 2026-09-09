package unswell

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func (e *Engine) capturePreparedFeatures(ctx context.Context, doc *document.Document, result *RunResult, structure bool) error {
	if result.PreparedFeatures == nil {
		return nil
	}
	source, err := e.preparedSource(doc, structure)
	if err != nil {
		return err
	}
	identity := feature.Identity{NLP: source.NLP, Capabilities: source.Capabilities, Source: doc.Hash,
		Policy: source.PolicyHash, Vocabulary: source.VocabularyHash, Preprocessing: source.PreparationHash}
	budget := e.policy.Analysis.MaxCandidates
	for _, block := range doc.Blocks {
		if err := e.capturePreparedBlock(ctx, block, identity, &source, &budget); err != nil {
			return fmt.Errorf("prepared collection: %w", err)
		}
	}
	source.TargetCount = len(source.Units)
	result.PreparedFeatures.Sources = append(result.PreparedFeatures.Sources, source)
	return nil
}

func (e *Engine) preparedSource(doc *document.Document, structure bool) (PreparedFeatureSource, error) {
	source := PreparedFeatureSource{Path: doc.Name, SourceHash: doc.Hash, PolicyHash: e.policy.Hash,
		VocabularyHash: e.policy.Hash, IncludeQuotes: e.policy.Analysis.IncludeQuotes, IncludeStructure: structure,
		NLP: e.nlp.Identity(), Capabilities: slices.Clone(e.preparedCapabilities), Units: []PreparedFeatureUnit{}}
	source.NLP.Capabilities = slices.Clone(source.NLP.Capabilities)
	policy, err := json.Marshal(e.policy.Extraction)
	if err != nil {
		return source, err
	}
	source.ExtractionPolicyHash = fmt.Sprintf("%x", sha256.Sum256(policy))
	preparation, err := json.Marshal(struct {
		Contract, Policy                string
		IncludeQuotes, IncludeStructure bool
	}{nlp.UnitContract, source.ExtractionPolicyHash, source.IncludeQuotes, source.IncludeStructure})
	if err != nil {
		return source, err
	}
	source.PreparationHash = fmt.Sprintf("%x", sha256.Sum256(preparation))
	return source, nil
}

func (e *Engine) capturePreparedBlock(ctx context.Context, block document.Block, identity feature.Identity,
	source *PreparedFeatureSource, budget *int,
) error {
	if *budget < len(e.preparedIDs) {
		return fmt.Errorf("measurements exceed max_candidates")
	}
	limits := nlp.UnitLimits{MaxBytes: min(e.policy.Analysis.MaxFileBytes, 16<<20), MaxContextBytes: 65536,
		MaxUnits: min(*budget/len(e.preparedIDs), 10000), MaxTokens: min(e.policy.Analysis.MaxTokens, 1000000), MaxSegments: 4096}
	units, err := nlp.PrepareUnits(ctx, block, e.nlp, nlp.UnitOptions{
		Kinds: e.preparedKinds, Capabilities: e.preparedCapabilities, Limits: limits})
	if err != nil {
		return err
	}
	for _, prepared := range units {
		unit, err := e.measurePrepared(ctx, prepared, identity, budget)
		if err != nil {
			return err
		}
		source.Units = append(source.Units, unit)
	}
	return ctx.Err()
}

func (e *Engine) measurePrepared(ctx context.Context, prepared nlp.PreparedUnit, identity feature.Identity,
	budget *int,
) (PreparedFeatureUnit, error) {
	unit := PreparedFeatureUnit{Binding: prepared.Binding()}
	cost := len(e.preparedIDs) + len(unit.Binding.Segments) + len(unit.Binding.ContextSpans)
	if cost >= *budget {
		return unit, fmt.Errorf("measurements exceed max_candidates")
	}
	*budget -= cost
	m, err := feature.MeasureUnit(ctx, prepared, identity, feature.Limits{MaxTokens: *budget, MaxUniqueWords: *budget,
		MaxBytes: e.policy.Analysis.MaxFileBytes, MaxBlocks: 1})
	if err != nil {
		return unit, err
	}
	unit.InputHash, unit.Segments = m.Hash(), m.Spans()
	*budget -= m.Counts().TokenVisits + len(unit.Segments)
	if *budget < 0 {
		return unit, fmt.Errorf("measurements exceed max_candidates")
	}
	for _, id := range e.preparedIDs {
		value, err := m.Value(id)
		if err != nil {
			return unit, err
		}
		unit.Values = append(unit.Values, value)
	}
	return unit, nil
}

func (r *RunResult) appendPreparedSources(part *PreparedFeatureCollection) {
	if r.PreparedFeatures != nil && part != nil {
		r.PreparedFeatures.Sources = append(r.PreparedFeatures.Sources, part.Sources...)
	}
}
