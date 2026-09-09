package unswell_test

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func preparedOptions() unswell.Options {
	return unswell.Options{PreparedFeatures: []string{"prose-words", "noun-token-ratio"},
		PreparedKinds: []string{"sentence", "paragraph", "fragment"}}
}

func TestPreparedCollectionSeparatesTargetsAndSharesPieceAnalysis(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	counted := &countedFeatureNLP{Provider: provider}
	options := preparedOptions()
	options.NLP = counted
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	options.PreparedKinds[0], options.PreparedFeatures[0] = "changed", "changed"
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Private heading\r\n\r\nThe **cache** expires. The client retries.\r\n\r\nFirst `hidden` tail.\r\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	collection := result.PreparedFeatures
	c.Assert(collection.Version, qt.Equals, unswell.PreparedFeatureCollectionVersion)
	c.Assert(collection.FeatureContract, qt.Equals, feature.UnitContract)
	c.Assert(collection.UnitContract, qt.Equals, nlp.UnitContract)
	c.Assert(collection.Kinds, qt.DeepEquals, []string{"fragment", "paragraph", "sentence"})
	c.Assert(collection.Requested, qt.DeepEquals, []string{"noun-token-ratio", "prose-words"})
	input := collection.Sources[0]
	c.Assert(input.Units, qt.HasLen, 6)
	c.Assert(counted.calls.Load(), qt.Equals, int64(result.Documents[0].Blocks+4))
	for i, kind := range []string{"fragment", "sentence", "sentence", "paragraph", "fragment", "fragment"} {
		c.Assert(input.Units[i].Binding.Kind, qt.Equals, kind)
	}
	a, b, paragraph := input.Units[1], input.Units[2], input.Units[3]
	c.Assert(*a.Values[1].Number, qt.Equals, float64(3))
	c.Assert(*paragraph.Values[1].Number, qt.Equals, float64(6))
	c.Assert(a.Binding.BlockID, qt.Equals, paragraph.Binding.BlockID)
	c.Assert(a.Binding.ContextSHA256, qt.Equals, b.Binding.ContextSHA256)
	c.Assert(a.Binding.ContextSHA256, qt.Equals, paragraph.Binding.TextSHA256)
	c.Assert(a.Binding.TextSHA256, qt.Not(qt.Equals), b.Binding.TextSHA256)
	c.Assert(a.InputHash, qt.Not(qt.Equals), paragraph.InputHash)
	c.Assert(a.Binding.Segments, qt.Not(qt.DeepEquals), a.Segments)
	c.Assert(input.Units[4].Binding.ContextSHA256, qt.Not(qt.Equals), input.Units[5].Binding.ContextSHA256)
	data, err := json.Marshal(collection)
	c.Assert(err, qt.IsNil)
	for _, secret := range []string{"Private heading", "hidden", "expires", "retries"} {
		c.Assert(string(data), qt.Not(qt.Contains), secret)
	}
	engine.PreparedFeatureIDs()[0], engine.PreparedUnitKinds()[0] = "changed", "changed"
	c.Assert(engine.PreparedFeatureIDs(), qt.DeepEquals, collection.Requested)
	c.Assert(engine.PreparedUnitKinds(), qt.DeepEquals, collection.Kinds)
}

func TestPreparedCollectionPreservesExistingResults(t *testing.T) {
	c := qt.New(t)
	options := preparedOptions()
	options.Features = []string{"prose-words", "activation/filler.announced-importance"}
	collector, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	options.PreparedFeatures, options.PreparedKinds = nil, nil
	plain, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Heading\n\nIt is important to note that the cache expires.\n")}
	want, err := plain.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	got, err := collector.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(got.PreparedFeatures, qt.IsNotNil)
	got.PreparedFeatures = nil
	c.Assert(got, qt.DeepEquals, want)
}

func TestPreparedCollectionValidatesRequestsAndCapabilities(t *testing.T) {
	c := qt.New(t)
	for _, options := range []unswell.Options{
		{PreparedFeatures: []string{"prose-words"}}, {PreparedKinds: []string{"sentence"}},
		{PreparedFeatures: []string{"unknown"}, PreparedKinds: []string{"sentence"}},
		{PreparedFeatures: []string{"activation/filler.announced-importance"}, PreparedKinds: []string{"sentence"}},
		{PreparedFeatures: []string{"prose-words", "prose-words"}, PreparedKinds: []string{"sentence"}},
		{PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"unknown"}},
		{PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"sentence", "sentence"}},
	} {
		_, err := unswell.New(options)
		c.Assert(err, qt.IsNotNil)
	}
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	options := preparedOptions()
	options.Config = []byte("version: 1\nextends: [builtin:custom]\n")
	options.NLP = limitedPolicyNLP{Provider: provider}
	_, err = unswell.New(options)
	c.Assert(err, qt.ErrorMatches, "prepared feature noun-token-ratio requires unavailable capability pos")
	options.PreparedFeatures = []string{"prose-words"}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "plain.txt", Format: document.Plain, Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.PreparedFeatures.Sources[0].Capabilities, qt.DeepEquals, []nlp.Capability{nlp.Sentences, nlp.Tokens})
}

func TestPreparedCollectionFailsWithoutPartialSourceOrSuccessfulGate(t *testing.T) {
	c := qt.New(t)
	for _, config := range []string{
		"version: 1\nanalysis: {max_candidates: 10}\n",
		"version: 1\nanalysis: {max_tokens: 2}\n",
	} {
		options := preparedOptions()
		options.Config, options.NoGate = []byte(config), true
		engine, err := unswell.New(options)
		c.Assert(err, qt.IsNil)
		result, err := engine.Analyze(t.Context(), document.Source{Name: "plain.txt", Format: document.Plain,
			Bytes: []byte("The cache expires. The client retries.")})
		c.Assert(err, qt.IsNotNil)
		c.Assert(result.Status, qt.Equals, "incomplete")
		c.Assert(result.Gate.Passed, qt.IsFalse)
		c.Assert(result.PreparedFeatures.Sources, qt.HasLen, 0)
	}
	engine, err := unswell.New(preparedOptions())
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := engine.Analyze(ctx, document.Source{Name: "plain.txt", Format: document.Plain, Bytes: []byte("The cache expires.")})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestPreparedCollectionOwnsConcurrentResults(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(preparedOptions())
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
			got.PreparedFeatures.Requested[0] = "changed"
			got.PreparedFeatures.Sources[0].NLP.Capabilities[0] = "changed"
			unit := &got.PreparedFeatures.Sources[0].Units[0]
			unit.Binding.ContextSpans[0].Start = 999
			unit.Binding.Segments[0].Start = 999
			*unit.Values[0].Number = 999
		})
	}
}
