package unswell_test

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

type countedFeatureNLP struct {
	nlp.Provider
	calls atomic.Int64
}

func (p *countedFeatureNLP) Analyze(
	ctx context.Context, text document.MappedText, capabilities []nlp.Capability,
) ([]document.Sentence, error) {
	p.calls.Add(1)
	return p.Provider.Analyze(ctx, text, capabilities)
}

func TestFeatureCollectionUsesOneEnrichmentAndOriginalSegments(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	counted := &countedFeatureNLP{Provider: provider}
	ids := []string{"type-token-ratio", "prose-words"}
	engine, err := unswell.New(unswell.Options{Features: ids, NLP: counted})
	c.Assert(err, qt.IsNil)
	ids[0] = "changed"
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Heading\r\n\r\nOne **two** three `hidden words`.\r\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(counted.calls.Load(), qt.Equals, int64(result.Documents[0].Blocks))
	collection := result.Features
	c.Assert(collection, qt.IsNotNil)
	c.Assert(collection.Version, qt.Equals, unswell.FeatureCollectionVersion)
	c.Assert(collection.BlockContract, qt.Equals, feature.Contract)
	c.Assert(collection.Requested, qt.DeepEquals, []string{"prose-words", "type-token-ratio"})
	c.Assert(collection.Sources, qt.HasLen, 1)
	input := collection.Sources[0]
	c.Assert(input.SourceHash, qt.Equals, result.Documents[0].SourceHash)
	c.Assert(input.PolicyHash, qt.Equals, result.Documents[0].ConfigHash)
	c.Assert(input.Units, qt.HasLen, 2)
	c.Assert(input.Units[0].Values[0].Number, qt.IsNil)
	c.Assert(input.Units[0].Values[0].Reason, qt.Equals, "unsupported_unit")
	paragraph := input.Units[1]
	c.Assert(*paragraph.Values[0].Number, qt.Equals, float64(3))
	c.Assert(*paragraph.Values[1].Number, qt.Equals, float64(1))
	c.Assert(paragraph.InputHash, qt.HasLen, 64)
	var words []string
	for _, span := range paragraph.Segments {
		words = append(words, string(source.Bytes[span.Start:span.End]))
	}
	c.Assert(words, qt.DeepEquals, []string{"One", "two", "three"})
	data, err := json.Marshal(collection)
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Not(qt.Contains), "hidden words")
	c.Assert(string(data), qt.Not(qt.Contains), "Heading")
	engine.FeatureIDs()[0] = "changed"
	c.Assert(engine.FeatureIDs(), qt.DeepEquals, collection.Requested)
}

func TestFeatureRequestsValidateIDsAndActualCapabilities(t *testing.T) {
	c := qt.New(t)
	for _, ids := range [][]string{{"unknown"}, {"prose-words", "prose-words"}, make([]string, 100)} {
		_, err := unswell.New(unswell.Options{Features: ids})
		c.Assert(err, qt.IsNotNil)
	}
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	_, err = unswell.New(unswell.Options{Features: []string{"noun-token-ratio"}, NLP: limitedPolicyNLP{Provider: provider},
		Config: []byte("version: 1\nextends: [builtin:custom]\n")})
	c.Assert(err, qt.ErrorMatches, "feature noun-token-ratio requires unavailable capability pos")
	engine, err := unswell.New(unswell.Options{Features: []string{"noun-token-ratio"},
		Config: []byte("version: 1\nextends: [builtin:custom]\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "plain.txt", Format: document.Plain, Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Features.Sources[0].Capabilities, qt.Contains, nlp.POS)
	c.Assert(result.Features.Sources[0].Units[0].Values[0].Number, qt.IsNotNil)
}

func TestFeatureCollectionPreservesRuleAndPolicyEvidence(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\nrules:\n  filler.announced-importance: {enabled: true, gate: forbid}\n")
	engine, err := unswell.New(unswell.Options{Config: policy})
	c.Assert(err, qt.IsNil)
	collector, err := unswell.New(unswell.Options{Config: policy, Features: []string{"prose-words", "type-token-ratio"}})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Heading\n\nIt is important to note that the cache expires.\n")}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	got, err := collector.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(got.Features, qt.IsNotNil)
	c.Assert(got.Gate.Passed, qt.IsFalse)
	got.Features = nil
	c.Assert(got, qt.DeepEquals, want)
}

func TestFeatureCollectionOwnsConcurrentResults(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"prose-words"}})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "plain.txt", Format: document.Plain, Bytes: []byte("The cache expires.")}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
			got.Features.Requested[0] = "changed"
			got.Features.Sources[0].Capabilities[0] = "changed"
			*got.Features.Sources[0].Units[0].Values[0].Number = -100
			got.Features.Sources[0].Units[0].Segments[0].Start = 999
			got.Features.Sources[0].Units[0].Binding.TextSHA256 = "changed"
			got.Features.Sources[0].Units[0].Binding.Segments[0].Start = 999
			got.Features.Sources[0].Units[0].Binding.TrimmedSegments[0].Start = 888
		})
	}
}

func TestFeatureCollectionPreservesFilePoliciesAndPartialFailure(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\noverrides:\n  - files: [reference.md]\n" +
		"    rules: {policy.banned-phrases: {enabled: false}}\n")
	sources := []document.Source{
		{Name: "reference.md", Format: document.Markdown, Bytes: []byte("The cache expires.")},
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The cache expires.")},
	}
	engine, err := unswell.New(unswell.Options{Features: []string{"prose-words"}, Config: policy, Jobs: 2})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Features.Sources[0].Path, qt.Equals, "guide.md")
	c.Assert(result.Features.Sources[1].Path, qt.Equals, "reference.md")
	c.Assert(result.Features.Sources[0].PolicyHash, qt.Not(qt.Equals), result.Features.Sources[1].PolicyHash)
	sources[1] = document.Source{Name: "broken.cs", Format: document.CSharp, Bytes: []byte("class Sample { string value = \"unfinished")}
	partial, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNotNil)
	c.Assert(partial.Manifest.Complete, qt.IsFalse)
	c.Assert(partial.Gate.Passed, qt.IsFalse)
	c.Assert(partial.Features.Sources, qt.HasLen, 1)
	c.Assert(partial.Features.Sources[0].Path, qt.Equals, "reference.md")
}
