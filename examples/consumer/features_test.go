package main

import (
	"context"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type featureRule struct {
	id   string
	sets chan *feature.Set
}

// Descriptor requests common measurements with the existing token capabilities.
func (r featureRule) Descriptor() rule.Descriptor {
	d := teamRule{}.Descriptor()
	d.ID, d.SharedFeatures = r.id, true
	d.BlockObservations = true
	return d
}

// Evaluate consumes the public feature set without rerunning extraction or NLP.
func (r featureRule) Evaluate(ctx context.Context, view rule.View, _ rule.Emitter) error {
	measurements, err := view.Features.Block(0)
	if err != nil {
		return err
	}
	if err := validateConsumerMeasurements(measurements); err != nil {
		return err
	}
	if err := view.Observe(feature.BlockObservation{BlockID: 0, Status: "evaluated"}); err != nil {
		return err
	}
	select {
	case r.sets <- view.Features:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func validateConsumerMeasurements(measurements feature.Measurements) error {
	value, err := measurements.Value("prose-words")
	if err != nil {
		return err
	}
	if value.Number == nil || *value.Number != 3 {
		return fmt.Errorf("expected three prose words")
	}
	pos, err := measurements.Value("noun-token-ratio")
	if err != nil {
		return err
	}
	if pos.Number != nil || pos.Reason != "capability_missing" {
		return fmt.Errorf("unrequested POS must remain unavailable")
	}
	return nil
}

func TestSharedFeaturesThroughPublicEngine(t *testing.T) {
	c := qt.New(t)
	sets := make(chan *feature.Set, 4)
	engine, err := unswell.New(unswell.Options{Features: []string{"prose-words", "activation/team.features-a"}, Rules: []rule.Rule{
		featureRule{id: "team.features-a", sets: sets}, featureRule{id: "team.features-b", sets: sets},
	}})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("One **two** three `hidden prose`.\r\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	first, second := <-sets, <-sets
	c.Assert(first, qt.Equals, second)
	_, err = engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	next, repeated := <-sets, <-sets
	c.Assert(next, qt.Equals, repeated)
	c.Assert(next, qt.Not(qt.Equals), first)
	m, err := first.Block(0)
	c.Assert(err, qt.IsNil)
	c.Assert(m.Spans(), qt.DeepEquals, []document.Span{{Start: 0, End: 3}, {Start: 6, End: 9}, {Start: 12, End: 17}})
	c.Assert(result.Features.Sources[0].Units[0].InputHash, qt.Equals, m.Hash())
	c.Assert(result.Features.Sources[0].Units[0].Segments, qt.DeepEquals, m.Spans())
	c.Assert(*result.Features.Sources[0].Units[0].Values[0].Number, qt.Equals, float64(0))
	c.Assert(*result.Features.Sources[0].Units[0].Values[1].Number, qt.Equals, float64(3))
}

func TestPublicLexicalMeasurements(t *testing.T) {
	c := qt.New(t)
	limits := feature.WordLimits{MaxWords: 10, MaxUniqueWords: 10, MaxBytes: 100}
	left, err := feature.NewWordSet(t.Context(), []string{"the", "cache", "entry"}, limits)
	c.Assert(err, qt.IsNil)
	right, err := feature.NewWordSet(t.Context(), []string{"the", "cache", "record"}, limits)
	c.Assert(err, qt.IsNil)
	value, err := feature.CompareWords(t.Context(), left, right)
	c.Assert(err, qt.IsNil)
	c.Assert(value.Available(), qt.IsTrue)
	c.Assert(value.Intersection(), qt.Equals, 2)
	c.Assert(value.Union(), qt.Equals, 4)
	c.Assert(*value.Values()[2].Number, qt.Equals, 0.5)
}

func TestPublicPatternPreparation(t *testing.T) {
	c := qt.New(t)
	sentence := document.Sentence{Text: "The cache expires.", Tokens: []document.Token{
		{Normal: "the", Word: true, Tag: "DT"}, {Normal: "cache", Word: true, Tag: "NN"},
		{Normal: "expires", Word: true, Tag: "VBZ"}, {Normal: ".", Tag: "."},
	}}
	limits := feature.SequenceLimits{MaxTokens: 20, MaxBytes: 1000}
	var candidates []feature.Ngram
	_, err := feature.ScanNgrams(t.Context(), sentence.Tokens,
		feature.NgramOptions{MinWords: 3, MaxWords: 3, MaxVisits: 100}, limits,
		func(candidate feature.Ngram) error { candidates = append(candidates, candidate); return nil })
	c.Assert(err, qt.IsNil)
	c.Assert(candidates, qt.DeepEquals, []feature.Ngram{{Start: 0, End: 3, Key: "the cache expires"}})
	pattern, err := feature.PreparePOSPattern(t.Context(), sentence, nil, limits)
	c.Assert(err, qt.IsNil)
	c.Assert(pattern.Available(), qt.IsTrue)
	c.Assert(pattern.Key(), qt.Equals, "DT|NN|VB|.")
}
