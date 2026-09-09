package report

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func validatePreparedFeatures(result unswell.RunResult) error {
	c := result.PreparedFeatures
	if c == nil {
		return nil
	}
	if c.Version != unswell.PreparedFeatureCollectionVersion ||
		c.FeatureContract != feature.UnitContract || c.UnitContract != nlp.UnitContract {
		return fmt.Errorf("unsupported prepared feature collection contract")
	}
	definitions, err := preparedDefinitions(c)
	if err != nil {
		return err
	}
	if result.Manifest.Complete && len(c.Sources) != len(result.Documents) {
		return fmt.Errorf("complete prepared collection must cover every analyzed source")
	}
	return validatePreparedSources(result, definitions)
}

func validatePreparedSources(result unswell.RunResult, definitions []feature.Descriptor) error {
	documents := make(map[string]unswell.DocumentResult)
	for _, doc := range result.Documents {
		if _, exists := documents[doc.Name]; exists {
			return fmt.Errorf("duplicate prepared document")
		}
		documents[doc.Name] = doc
	}
	for _, source := range result.PreparedFeatures.Sources {
		doc, exists := documents[source.Path]
		if !exists {
			return fmt.Errorf("unknown or duplicate prepared source")
		}
		delete(documents, source.Path)
		if err := validatePreparedSource(source, doc, result.Manifest.NLP, definitions, result.PreparedFeatures.Kinds); err != nil {
			return err
		}
	}
	return nil
}

func preparedDefinitions(c *unswell.PreparedFeatureCollection) ([]feature.Descriptor, error) {
	if err := validateSavedKinds(c.Kinds); err != nil {
		return nil, err
	}
	catalog := feature.Catalog()
	if len(c.Requested) == 0 || len(c.Requested) > len(catalog) || !canonicalPreparedSet(c.Requested) {
		return nil, fmt.Errorf("invalid prepared feature request")
	}
	var result []feature.Descriptor
	for _, id := range c.Requested {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		if index < 0 {
			return nil, fmt.Errorf("unknown prepared feature")
		}
		result = append(result, catalog[index])
	}
	return result, nil
}

func canonicalPreparedSet(values []string) bool {
	for i, value := range values {
		if i > 0 && value <= values[i-1] {
			return false
		}
	}
	return true
}

func validatePreparedSource(source unswell.PreparedFeatureSource, doc unswell.DocumentResult, identity nlp.Identity,
	definitions []feature.Descriptor, kinds []string,
) error {
	if err := validatePreparedIdentity(source, doc, identity); err != nil {
		return err
	}
	if err := validateFeatureCapabilities(unswell.FeatureSource{NLP: source.NLP, Capabilities: source.Capabilities}, definitions); err != nil {
		return err
	}
	if source.TargetCount != len(source.Units) {
		return fmt.Errorf("prepared target count does not match collection")
	}
	seen := make(map[string]bool)
	for _, unit := range source.Units {
		if seen[unit.InputHash] {
			return fmt.Errorf("duplicate prepared target")
		}
		seen[unit.InputHash] = true
		if err := validatePreparedUnit(unit, doc, definitions, kinds); err != nil {
			return err
		}
	}
	return nil
}

func validatePreparedUnit(unit unswell.PreparedFeatureUnit, doc unswell.DocumentResult,
	definitions []feature.Descriptor, kinds []string,
) error {
	if err := validatePreparedBinding(unit, doc, kinds); err != nil {
		return err
	}
	b := unit.Binding
	if !validPreparedSpans(b.Segments, doc.Bytes, true) || !validPreparedSpans(b.ContextSpans, doc.Bytes, true) ||
		!validPreparedSpans(unit.Segments, doc.Bytes, false) || !preparedSubset(b.Segments, b.ContextSpans) ||
		!preparedSubset(unit.Segments, b.Segments) {
		return fmt.Errorf("invalid prepared target, context, or counted segments")
	}
	return validatePreparedValues(unit.Values, definitions)
}

func validatePreparedValues(values []feature.Value, definitions []feature.Descriptor) error {
	if len(values) != len(definitions) {
		return fmt.Errorf("prepared values do not match request")
	}
	for i, value := range values {
		d := definitions[i]
		if value.ID != d.ID || value.Version != d.Version || value.Unit != d.Unit {
			return fmt.Errorf("incompatible prepared feature definition")
		}
		if err := validateFeatureNumber(value, true, d.MinWords); err != nil {
			return err
		}
	}
	return nil
}

func validPreparedSpans(spans []document.Span, size int, required bool) bool {
	if required && (len(spans) == 0 || len(spans) > 4096) {
		return false
	}
	last := 0
	for _, span := range spans {
		if span.Start < last || !span.Valid(size) {
			return false
		}
		last = span.End
	}
	return true
}

func preparedSubset(inner, outer []document.Span) bool {
	index := 0
	for _, span := range inner {
		for index < len(outer) && outer[index].End <= span.Start {
			index++
		}
		if index == len(outer) || span.Start < outer[index].Start || span.End > outer[index].End {
			return false
		}
	}
	return true
}

func validateSavedKinds(kinds []string) error {
	if len(kinds) == 0 || len(kinds) > 3 || !canonicalPreparedSet(kinds) {
		return fmt.Errorf("invalid prepared kinds")
	}
	for _, kind := range kinds {
		if _, err := feature.UnitCatalog(kind); err != nil {
			return err
		}
	}
	return nil
}

func validatePreparedIdentity(source unswell.PreparedFeatureSource, doc unswell.DocumentResult, identity nlp.Identity) error {
	if source.SourceHash != doc.SourceHash || source.PolicyHash != doc.ConfigHash || source.VocabularyHash != source.PolicyHash ||
		!preparedHashes(source.SourceHash, source.PolicyHash, source.ExtractionPolicyHash, source.PreparationHash) ||
		!reflect.DeepEqual(source.NLP, identity) {
		return fmt.Errorf("inconsistent prepared input identity")
	}
	return nil
}

func validatePreparedBinding(unit unswell.PreparedFeatureUnit, doc unswell.DocumentResult, kinds []string) error {
	b := unit.Binding
	if b.Contract != nlp.UnitContract || !slices.Contains(kinds, b.Kind) || b.BlockID < 0 || b.BlockID >= doc.Blocks ||
		b.BlockKind == "" || len(b.BlockKind) > 128 || !preparedHashes(b.TextSHA256, b.ContextSHA256, b.GrammarSHA256, unit.InputHash) {
		return fmt.Errorf("invalid prepared target binding")
	}
	return nil
}

func preparedHashes(values ...string) bool {
	for _, value := range values {
		if !hashString(value) {
			return false
		}
	}
	return true
}
