package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func measureRuleSources(ctx context.Context, engine *unswell.Engine, plan Plan,
	files map[string][]byte,
) (unswell.FeatureCollection, error) {
	configuration, err := measurementConfig(plan)
	if err != nil {
		return unswell.FeatureCollection{}, err
	}
	expected, _, err := config.CompileBundle(config.Bundle{Root: ".unswell.yaml",
		Files: map[string][]byte{".unswell.yaml": configuration}}, engine.Catalog())
	if err != nil {
		return unswell.FeatureCollection{}, err
	}
	var collection unswell.FeatureCollection
	bytes, blocks := 0, 0
	for _, source := range plan.Manifest.Sources {
		part, err := measureRuleSource(ctx, engine, expected, source, files[source.Path])
		if err != nil {
			return unswell.FeatureCollection{}, err
		}
		if collection.Version == "" {
			collection = *part
			collection.Sources = nil
		}
		record := part.Sources[0]
		encoded, err := json.Marshal(record)
		if err != nil {
			return unswell.FeatureCollection{}, err
		}
		bytes, blocks = bytes+len(encoded), blocks+len(record.Units)
		if bytes > MaxArtifactBytes/2 || blocks > MaxUnits {
			return unswell.FeatureCollection{}, fmt.Errorf("rule corpus exceeds byte or block limit")
		}
		collection.Sources = append(collection.Sources, record)
	}
	slices.SortFunc(collection.Sources, func(a, b unswell.FeatureSource) int { return strings.Compare(a.Path, b.Path) })
	return collection, ctx.Err()
}

func measureRuleSource(ctx context.Context, engine *unswell.Engine, expected *config.Plan, source Source,
	data []byte,
) (*unswell.FeatureCollection, error) {
	policy, err := engine.PolicyForFile(source.Path)
	if err != nil {
		return nil, err
	}
	frozen, err := expected.ForFile(source.Path)
	if err != nil {
		return nil, err
	}
	if policy.Analysis.IncludeQuotes || !sameRuleExtraction(policy.Extraction, frozen.Extraction) {
		return nil, fmt.Errorf("rule extraction differs from frozen corpus policy for %s", source.ID)
	}
	result, err := engine.AnalyzeAll(ctx, []document.Source{{Name: source.Path, Format: source.Format, Bytes: data}})
	if err != nil {
		return nil, err
	}
	if !result.Manifest.Complete || result.Features == nil || len(result.Features.Sources) != 1 {
		return nil, fmt.Errorf("incomplete rule collection for %s", source.ID)
	}
	return result.Features, nil
}

func sameRuleExtraction(actual, frozen extract.Policy) bool {
	// Empty override maps and exception lists mean no entries. Context sets are
	// different: nil inherits defaults, while an empty set disables extraction.
	if len(actual.Languages) == 0 {
		actual.Languages = nil
	}
	if len(frozen.Languages) == 0 {
		frozen.Languages = nil
	}
	if len(actual.Exceptions) == 0 {
		actual.Exceptions = nil
	}
	if len(frozen.Exceptions) == 0 {
		frozen.Exceptions = nil
	}
	return reflect.DeepEqual(actual, frozen)
}
