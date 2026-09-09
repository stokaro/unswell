package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func TestPrepareAndMeasureSameTarget(t *testing.T) {
	c := qt.New(t)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "example.md", Format: document.Markdown,
		Bytes: []byte("The cache may retry. The service cannot wait.")}, extract.Options{MaxBytes: 4096, MaxBlocks: 10})
	c.Assert(err, qt.IsNil)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	units, err := nlp.PrepareUnits(t.Context(), doc.Blocks[0], provider, nlp.UnitOptions{
		Kinds: []string{"sentence", "paragraph"}, Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Limits: nlp.UnitLimits{MaxBytes: 4096, MaxContextBytes: 4096, MaxUnits: 10, MaxTokens: 100, MaxSegments: 100}})
	c.Assert(err, qt.IsNil)
	c.Assert(units, qt.HasLen, 3)
	for i, words := range []int{4, 4, 8} {
		unit := units[i]
		measurements, err := feature.MeasureUnit(t.Context(), unit, feature.Identity{
			NLP: unit.Identity(), Capabilities: unit.Capabilities(), Source: doc.Hash,
			Policy: "consumer-policy-v1", Vocabulary: "none", Preprocessing: nlp.UnitContract},
			feature.Limits{MaxBytes: 4096, MaxTokens: 100, MaxUniqueWords: 100, MaxBlocks: 10})
		c.Assert(err, qt.IsNil)
		c.Assert(measurements.Counts().Words, qt.Equals, words)
		c.Assert(unit.Binding().Segments, qt.DeepEquals, unit.Block().Spans(0, len(unit.Block().Text)))
	}
}

func TestCollectPreparedTargetsThroughPublicEngine(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{PreparedFeatures: []string{"prose-words"},
		PreparedKinds: []string{"sentence", "paragraph"}})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "example.txt", Format: document.Plain,
		Bytes: []byte("The cache may retry. The service cannot wait.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.PreparedFeatures.Sources[0].Units, qt.HasLen, 3)
	for i, words := range []float64{4, 4, 8} {
		c.Assert(*result.PreparedFeatures.Sources[0].Units[i].Values[0].Number, qt.Equals, words)
	}
}
