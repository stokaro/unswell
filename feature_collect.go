package unswell

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func (e *Engine) selectFeatures(ids []string) error {
	catalog := feature.Catalog()
	for _, d := range e.descriptors {
		definition := feature.ActivationDescriptor(d.ID)
		definition.Requires = slices.Clone(d.Requires)
		catalog = append(catalog, definition)
	}
	if len(ids) > len(catalog) {
		return fmt.Errorf("too many requested features")
	}
	e.featureIDs = slices.Clone(ids)
	e.activationIndices = make(map[string]int)
	slices.Sort(e.featureIDs)
	for i, id := range e.featureIDs {
		if i > 0 && id == e.featureIDs[i-1] {
			return fmt.Errorf("duplicate requested feature %q", id)
		}
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		if index < 0 {
			return fmt.Errorf("unknown requested feature %q", id)
		}
		e.featureDefinitions = append(e.featureDefinitions, catalog[index])
		if ruleID, activation := strings.CutPrefix(id, "activation/"); activation {
			e.activationIndices[ruleID] = i
		}
	}
	return nil
}

// FeatureIDs returns owned, canonical IDs explicitly requested for collection.
func (e *Engine) FeatureIDs() []string { return slices.Clone(e.featureIDs) }

func (e *Engine) featureCollection() *FeatureCollection {
	if len(e.featureIDs) == 0 {
		return nil
	}
	collection := &FeatureCollection{Version: FeatureCollectionVersion, BlockContract: feature.Contract, Requested: e.FeatureIDs()}
	if len(e.activationIndices) > 0 {
		collection.ActivationContract = feature.ActivationContract
	}
	return collection
}

func (e *Engine) captureFeatures(ctx context.Context, doc *document.Document, set *feature.Set, result *RunResult) error {
	if result.Features == nil {
		return nil
	}
	if len(doc.Blocks) > e.policy.Analysis.MaxCandidates/len(e.featureIDs) {
		return fmt.Errorf("collected feature values exceed max_candidates")
	}
	identity := e.featureIdentity(doc)
	source := FeatureSource{Path: doc.Name, SourceHash: doc.Hash, PolicyHash: identity.Policy,
		VocabularyHash: identity.Vocabulary, Preprocessing: identity.Preprocessing, NLP: identity.NLP,
		Capabilities: identity.Capabilities}
	if len(e.activationIndices) > 0 {
		source.RulesetHash = e.rulesetHash
	}
	for _, block := range doc.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		unit, err := e.captureBlock(block, set)
		if err != nil {
			return err
		}
		source.Units = append(source.Units, unit)
	}
	result.Features.Sources = append(result.Features.Sources, source)
	return nil
}

func (e *Engine) captureBlock(block document.Block, set *feature.Set) (FeatureUnit, error) {
	unit := FeatureUnit{Scope: "block", UnitID: block.ID, Kind: block.Kind, Span: block.Span}
	contextBytes, err := json.Marshal(block.Context)
	if err != nil {
		return unit, err
	}
	unit.ContextHash = fmt.Sprintf("%x", sha256.Sum256(contextBytes))
	var m feature.Measurements
	if feature.SupportsBlock(block.Kind) {
		m, err = set.Block(block.ID)
		if err != nil {
			return unit, err
		}
		unit.InputHash, unit.Segments = m.Hash(), m.Spans()
	}
	for _, d := range e.featureDefinitions {
		value, err := e.initialFeatureValue(d, block, m)
		if err != nil {
			return unit, err
		}
		unit.Values = append(unit.Values, value)
	}
	return unit, nil
}

func (e *Engine) initialFeatureValue(d feature.Descriptor, block document.Block, m feature.Measurements) (feature.Value, error) {
	value := feature.Value{ID: d.ID, Version: d.Version, Unit: d.Unit}
	if ruleID, activation := strings.CutPrefix(d.ID, "activation/"); activation {
		value.Reason = "not_evaluated"
		if !e.policy.Rules[ruleID].Enabled {
			value.Reason = "disabled"
		}
		return value, nil
	}
	if !feature.SupportsBlock(block.Kind) {
		value.Reason = "unsupported_unit"
		return value, nil
	}
	return m.Value(d.ID)
}

func (r *RunResult) appendFeatureSources(part *FeatureCollection) {
	if r.Features != nil && part != nil {
		r.Features.Sources = append(r.Features.Sources, part.Sources...)
	}
}
