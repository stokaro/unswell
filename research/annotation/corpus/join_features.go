package corpus

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func measureCandidates(ctx context.Context, plan Plan, files map[string][]byte,
	features []string,
) (unswell.PreparedFeatureCollection, error) {
	configuration, err := measurementConfig(plan)
	if err != nil {
		return unswell.PreparedFeatureCollection{}, err
	}
	engine, err := unswell.New(unswell.Options{Config: configuration, PreparedFeatures: features,
		PreparedKinds: plan.Manifest.UnitKinds, Jobs: 1, AllowEmpty: true})
	if err != nil {
		return unswell.PreparedFeatureCollection{}, err
	}
	var collection unswell.PreparedFeatureCollection
	bytes, targets := 0, 0
	for _, source := range plan.Manifest.Sources {
		part, err := measureSource(ctx, engine, source, files[source.Path])
		if err != nil {
			return unswell.PreparedFeatureCollection{}, err
		}
		if collection.Version == "" {
			collection = *part
			collection.Sources = nil
		}
		record := part.Sources[0]
		encoded, err := json.Marshal(record)
		if err != nil {
			return unswell.PreparedFeatureCollection{}, err
		}
		bytes += len(encoded)
		targets += record.TargetCount
		if bytes > MaxArtifactBytes/2 || targets > MaxUnits {
			return unswell.PreparedFeatureCollection{}, fmt.Errorf("prepared corpus exceeds byte or target limit")
		}
		collection.Sources = append(collection.Sources, record)
	}
	slices.SortFunc(collection.Sources, func(a, b unswell.PreparedFeatureSource) int { return strings.Compare(a.Path, b.Path) })
	return collection, ctx.Err()
}

func measureSource(ctx context.Context, engine *unswell.Engine, source Source,
	data []byte,
) (*unswell.PreparedFeatureCollection, error) {
	result, err := engine.AnalyzeAll(ctx, []document.Source{{Name: source.Path, Format: source.Format, Bytes: data}})
	if err != nil {
		return nil, err
	}
	if !result.Manifest.Complete || result.PreparedFeatures == nil || len(result.PreparedFeatures.Sources) != 1 {
		return nil, fmt.Errorf("incomplete prepared collection for %s", source.ID)
	}
	return result.PreparedFeatures, nil
}

// Config rejects explicit null overrides. Omit nil fields while retaining empty
// context sets, which disable extraction rather than inheriting the defaults.
func measurementConfig(plan Plan) ([]byte, error) {
	encoded, err := json.Marshal(plan.Manifest.Policy)
	if err != nil {
		return nil, err
	}
	var policy map[string]any
	if err := json.Unmarshal(encoded, &policy); err != nil {
		return nil, err
	}
	omitNullFields(policy)
	return json.Marshal(map[string]any{"version": 1, "extends": []string{"builtin:custom"}, "extraction": policy})
}

func omitNullFields(value any) {
	switch object := value.(type) {
	case map[string]any:
		for key, child := range object {
			if child == nil {
				delete(object, key)
			} else {
				omitNullFields(child)
			}
		}
	case []any:
		for _, child := range object {
			omitNullFields(child)
		}
	}
}

// A complete source map is required: a bounding interval can cross protected text.
type targetKey struct {
	Path, Source, Kind, Text, Context string
	Segments                          []document.Span
}

func bindCandidates(ctx context.Context, artifact Artifact,
	collection unswell.PreparedFeatureCollection,
) ([]FeatureBinding, error) {
	measured, err := measuredTargets(ctx, collection)
	if err != nil {
		return nil, err
	}
	paths := make(map[string]string, len(artifact.Plan.Manifest.Sources))
	for _, source := range artifact.Plan.Manifest.Sources {
		paths[source.ID] = source.Path
	}
	bindings := make([]FeatureBinding, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		unit := candidate.Unit
		key, err := digest(targetKey{Path: paths[candidate.SourceID], Source: unit.Source.SHA256,
			Kind: unit.Kind, Text: hashText(unit.Text), Context: hashText(unit.Context), Segments: unit.Source.Segments})
		if err != nil {
			return nil, err
		}
		inputHash, exists := measured[key]
		if !exists {
			return nil, fmt.Errorf("candidate %s has no exact prepared measurement", unit.ID)
		}
		delete(measured, key)
		bindings = append(bindings, FeatureBinding{UnitID: unit.ID, SourceID: candidate.SourceID,
			Path: paths[candidate.SourceID], GroupID: candidate.GroupID, Partition: candidate.Partition, FeatureInputHash: inputHash})
	}
	if len(measured) != 0 {
		return nil, fmt.Errorf("prepared collection contains targets outside the reproduced corpus")
	}
	return bindings, ctx.Err()
}

func measuredTargets(ctx context.Context, collection unswell.PreparedFeatureCollection) (map[string]string, error) {
	result := make(map[string]string)
	for _, source := range collection.Sources {
		for _, unit := range source.Units {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			binding := unit.Binding
			key, err := digest(targetKey{Path: source.Path, Source: source.SourceHash, Kind: binding.Kind,
				Text: binding.TextSHA256, Context: binding.ContextSHA256, Segments: binding.Segments})
			if err != nil {
				return nil, err
			}
			if _, exists := result[key]; exists {
				return nil, fmt.Errorf("prepared collection contains an ambiguous target")
			}
			result[key] = unit.InputHash
		}
	}
	return result, nil
}

func hashText(text string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(text))) }
