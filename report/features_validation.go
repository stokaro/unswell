package report

import (
	"encoding/hex"
	"fmt"
	"math"
	"reflect"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func validateFeatures(result unswell.RunResult) error {
	collection := result.Features
	if collection == nil {
		return nil
	}
	if collection.Version != unswell.FeatureCollectionVersion || collection.BlockContract != feature.Contract {
		return fmt.Errorf("unsupported feature collection contract")
	}
	definitions, err := requestedDefinitions(collection.Requested)
	if err != nil {
		return err
	}
	if result.Manifest.Complete && len(collection.Sources) != len(result.Documents) {
		return fmt.Errorf("complete feature collection must cover every analyzed source")
	}
	return validateFeatureSources(result, definitions)
}

func requestedDefinitions(ids []string) ([]feature.Descriptor, error) {
	catalog := feature.Catalog()
	if len(ids) == 0 || len(ids) > len(catalog) || !slices.IsSorted(ids) {
		return nil, fmt.Errorf("invalid feature request order or size")
	}
	var definitions []feature.Descriptor
	for i, id := range ids {
		if i > 0 && id == ids[i-1] {
			return nil, fmt.Errorf("duplicate requested feature")
		}
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		if index < 0 {
			return nil, fmt.Errorf("unknown requested feature")
		}
		definitions = append(definitions, catalog[index])
	}
	return definitions, nil
}

func validateFeatureSources(result unswell.RunResult, definitions []feature.Descriptor) error {
	documents := make(map[string]unswell.DocumentResult, len(result.Documents))
	for _, doc := range result.Documents {
		if _, duplicate := documents[doc.Name]; duplicate {
			return fmt.Errorf("duplicate analyzed source in feature collection")
		}
		documents[doc.Name] = doc
	}
	seen := make(map[string]bool)
	for _, source := range result.Features.Sources {
		doc, exists := documents[source.Path]
		if seen[source.Path] || !exists {
			return fmt.Errorf("duplicate or unknown feature source")
		}
		seen[source.Path] = true
		if err := validateFeatureIdentity(source, doc, result.Manifest.NLP); err != nil {
			return err
		}
		if err := validateFeatureCapabilities(source, definitions); err != nil {
			return err
		}
		if err := validateFeatureUnits(source, doc, definitions); err != nil {
			return err
		}
	}
	return nil
}

func validateFeatureIdentity(source unswell.FeatureSource, doc unswell.DocumentResult, identity nlp.Identity) error {
	if source.SourceHash != doc.SourceHash || source.PolicyHash != doc.ConfigHash ||
		source.VocabularyHash != source.PolicyHash || source.Preprocessing != "extracted-provider-tokens-v1" {
		return fmt.Errorf("inconsistent feature input identity")
	}
	if !hashString(source.SourceHash) || !hashString(source.PolicyHash) || !reflect.DeepEqual(source.NLP, identity) {
		return fmt.Errorf("invalid feature source, policy, or NLP identity")
	}
	return nil
}

func validateFeatureCapabilities(source unswell.FeatureSource, definitions []feature.Descriptor) error {
	if len(source.Capabilities) > 7 || !slices.IsSorted(source.Capabilities) {
		return fmt.Errorf("invalid feature capability order or size")
	}
	for i, capability := range source.Capabilities {
		if !slices.Contains(source.NLP.Capabilities, capability) || (i > 0 && capability == source.Capabilities[i-1]) {
			return fmt.Errorf("inconsistent feature capability")
		}
	}
	for _, definition := range definitions {
		for _, required := range definition.Requires {
			if !slices.Contains(source.Capabilities, required) {
				return fmt.Errorf("requested feature requires missing capability %s", required)
			}
		}
	}
	return nil
}

func validateFeatureUnits(source unswell.FeatureSource, doc unswell.DocumentResult, definitions []feature.Descriptor) error {
	if len(source.Units) != doc.Blocks {
		return fmt.Errorf("feature source does not cover its extracted blocks")
	}
	for i, unit := range source.Units {
		if unit.UnitID != i || unit.Scope != "block" || !unit.Span.Valid(doc.Bytes) ||
			len(unit.Kind) > 128 || !hashString(unit.ContextHash) {
			return fmt.Errorf("invalid feature unit identity or range")
		}
		if err := validateFeatureSegments(unit); err != nil {
			return err
		}
		if err := validateFeatureValues(unit, definitions); err != nil {
			return err
		}
	}
	return nil
}

func validateFeatureSegments(unit unswell.FeatureUnit) error {
	if !feature.SupportsBlock(unit.Kind) {
		if unit.InputHash != "" || len(unit.Segments) != 0 {
			return fmt.Errorf("unsupported feature unit cannot have measured input")
		}
		return nil
	}
	if !hashString(unit.InputHash) {
		return fmt.Errorf("invalid feature input hash")
	}
	last := unit.Span.Start
	for _, span := range unit.Segments {
		if span.Start < last || !span.Valid(unit.Span.End) {
			return fmt.Errorf("invalid counted feature segment")
		}
		last = span.End
	}
	return nil
}

func validateFeatureValues(unit unswell.FeatureUnit, definitions []feature.Descriptor) error {
	if len(unit.Values) != len(definitions) {
		return fmt.Errorf("feature unit does not match its request")
	}
	for i, value := range unit.Values {
		d := definitions[i]
		if value.ID != d.ID || value.Version != d.Version || value.Unit != d.Unit {
			return fmt.Errorf("feature value has an incompatible definition")
		}
		if err := validateFeatureNumber(value, feature.SupportsBlock(unit.Kind), d.MinWords); err != nil {
			return err
		}
	}
	return nil
}

func validateFeatureNumber(value feature.Value, supported bool, minimum int) error {
	if !supported {
		if value.Number != nil || value.Reason != "unsupported_unit" {
			return fmt.Errorf("unsupported feature value must be absent")
		}
		return nil
	}
	if value.Number == nil {
		if value.Reason != "no_prose_words" || minimum == 0 {
			return fmt.Errorf("invalid absent feature value")
		}
		return nil
	}
	if value.Reason != "" || math.IsNaN(*value.Number) || math.IsInf(*value.Number, 0) {
		return fmt.Errorf("invalid numeric feature value")
	}
	return nil
}

func hashString(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}
