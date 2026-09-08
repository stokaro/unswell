package feature_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func sequenceLimits() feature.SequenceLimits {
	return feature.SequenceLimits{MaxTokens: 100, MaxBytes: 1000}
}

func ngramTokens() []document.Token {
	return []document.Token{
		{Normal: "the", Tag: "DT", Word: true},
		{Normal: "clear", Tag: "JJ", Word: true},
		{Normal: "release", Tag: "NN", Word: true},
		{Normal: "notes", Tag: "NNS", Word: true},
		{Normal: ",", Tag: ","},
	}
}

func TestNgramOrderRangesAndLogicalVisits(t *testing.T) {
	c := qt.New(t)
	tokens := ngramTokens()
	want := []feature.Ngram{
		{Start: 0, End: 3, Key: "the clear release"},
		{Start: 0, End: 4, Key: "the clear release notes"},
		{Start: 1, End: 4, Key: "clear release notes"},
	}
	for _, limit := range []int{14, 13} {
		var got []feature.Ngram
		visits, err := feature.ScanNgrams(t.Context(), tokens,
			feature.NgramOptions{MinWords: 3, MaxWords: 4, MaxVisits: limit}, sequenceLimits(),
			func(candidate feature.Ngram) error { got = append(got, candidate); return nil })
		c.Assert(visits, qt.Equals, limit)
		c.Assert(got, qt.DeepEquals, want)
		if limit == 14 {
			c.Assert(err, qt.IsNil)
		} else {
			// Producing every candidate is still incomplete until boundary visits finish.
			c.Assert(err, qt.ErrorIs, feature.ErrTokenLimit)
		}
	}
}

func TestNgramSelectionDoesNotCrossProtectedTokens(t *testing.T) {
	c := qt.New(t)
	tokens := ngramTokens()
	tokens[0].Protected = true
	var got []feature.Ngram
	_, err := feature.ScanNgrams(t.Context(), tokens,
		feature.NgramOptions{MinWords: 3, MaxWords: 4, MaxVisits: 100}, sequenceLimits(),
		func(candidate feature.Ngram) error { got = append(got, candidate); return nil })
	c.Assert(err, qt.IsNil)
	c.Assert(got, qt.DeepEquals, []feature.Ngram{{Start: 1, End: 4, Key: "clear release notes"}})
	tokens[1].Normal = "changed"
	c.Assert(got[0].Key, qt.Equals, "clear release notes")
	for _, words := range []string{"the and with", "cache cache cache", "the café timeout"} {
		var input []document.Token
		for word := range strings.FieldsSeq(words) {
			input = append(input, document.Token{Normal: word, Word: true})
		}
		count := 0
		_, err := feature.ScanNgrams(t.Context(), input,
			feature.NgramOptions{MinWords: 3, MaxWords: 3, MaxVisits: 100}, sequenceLimits(),
			func(feature.Ngram) error { count++; return nil })
		c.Assert(err, qt.IsNil)
		want := 0
		if words == "the café timeout" {
			want = 1
		}
		c.Assert(count, qt.Equals, want)
	}
}

func TestNgramCallbackFailureAndCancellation(t *testing.T) {
	c := qt.New(t)
	failure := errors.New("consumer failed")
	for _, cancelDuringCallback := range []bool{false, true} {
		ctx, cancel := context.WithCancel(t.Context())
		calls := 0
		visits, err := feature.ScanNgrams(ctx, ngramTokens(),
			feature.NgramOptions{MinWords: 3, MaxWords: 4, MaxVisits: 100}, sequenceLimits(), func(feature.Ngram) error {
				calls++
				if cancelDuringCallback {
					cancel()
					return nil
				}
				return failure
			})
		cancel()
		c.Assert(calls, qt.Equals, 1)
		c.Assert(visits, qt.Equals, 3)
		if cancelDuringCallback {
			c.Assert(err, qt.ErrorIs, context.Canceled)
		} else {
			c.Assert(err, qt.ErrorIs, failure)
		}
	}
}

func TestNgramRejectsInvalidOptionsAndRepresentations(t *testing.T) {
	c := qt.New(t)
	noop := func(feature.Ngram) error { return nil }
	for _, options := range []feature.NgramOptions{{}, {MinWords: 2, MaxWords: 4}, {MinWords: 3, MaxWords: 9},
		{MinWords: 5, MaxWords: 4}, {MinWords: 3, MaxWords: 4, MaxVisits: -1}} {
		_, err := feature.ScanNgrams(t.Context(), ngramTokens(), options, sequenceLimits(), noop)
		c.Assert(err, qt.IsNotNil)
	}
	options := feature.NgramOptions{MinWords: 3, MaxWords: 4, MaxVisits: 100}
	_, err := feature.ScanNgrams(t.Context(), ngramTokens(), options, sequenceLimits(), nil)
	c.Assert(err, qt.IsNotNil)
	for _, limits := range []feature.SequenceLimits{{}, {MaxTokens: 1, MaxBytes: 1000}, {MaxTokens: 100, MaxBytes: 1}} {
		_, err := feature.ScanNgrams(t.Context(), ngramTokens(), options, limits, noop)
		c.Assert(err, qt.IsNotNil)
	}
	for _, token := range []document.Token{{Word: true}, {Normal: "\xff"}, {Normal: "hidden\x00boundary"},
		{Normal: "word", Tag: strings.Repeat("N", 129)}, {Normal: "word", Tag: "\xff"}, {Normal: "word", Tag: "N\x00N"}} {
		_, err := feature.ScanNgrams(t.Context(), []document.Token{token}, options, sequenceLimits(), noop)
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = feature.ScanNgrams(ctx, ngramTokens(), options, sequenceLimits(), noop)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	options.MaxVisits = 0
	visits, err := feature.ScanNgrams(t.Context(), nil, options, sequenceLimits(), noop)
	c.Assert(err, qt.IsNil)
	c.Assert(visits, qt.Equals, 0)
}

func templateSentence() document.Sentence {
	return document.Sentence{Text: "The servers cached stable replies quickly.", Tokens: []document.Token{
		{Normal: "the", Tag: "DT", Word: true}, {Normal: "servers", Tag: "NNS", Word: true},
		{Normal: "cached", Tag: "VBD", Word: true}, {Normal: "stable", Tag: "JJ", Word: true},
		{Normal: "replies", Tag: "NNS", Word: true}, {Normal: "quickly", Tag: "RB", Word: true},
		{Normal: ".", Tag: "."},
	}}
}

func TestPOSPatternKeepsSelectedLiteralsAndOwnsItsKey(t *testing.T) {
	c := qt.New(t)
	sentence := templateSentence()
	pattern, err := feature.PreparePOSPattern(t.Context(), sentence, nil, sequenceLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(pattern.Available(), qt.IsTrue)
	c.Assert(pattern.Reason(), qt.Equals, "")
	c.Assert(pattern.Key(), qt.Equals, "DT|NN|VB|JJ|NN|RB|.")
	literal := make([]bool, len(sentence.Tokens))
	literal[1] = true
	pattern, err = feature.PreparePOSPattern(t.Context(), sentence, literal, sequenceLimits())
	c.Assert(err, qt.IsNil)
	sentence.Tokens[1].Normal, literal[1] = "changed", false
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() { c.Check(pattern.Key(), qt.Equals, "DT|servers|VB|JJ|NN|RB|.") })
	}
	workers.Wait()
}

func TestPOSPatternAbsenceIsExplicit(t *testing.T) {
	c := qt.New(t)
	c.Assert((feature.POSPattern{}).Reason(), qt.Equals, "not_computed")
	for _, row := range []struct {
		reason string
		edit   func(*document.Sentence)
	}{
		{"empty", func(s *document.Sentence) { s.Tokens = nil }},
		{"imperative", func(s *document.Sentence) { s.Tokens[0].Tag = "VB" }},
		{"question", func(s *document.Sentence) { s.Text = "Does the server cache replies?\n" }},
		{"protected_token", func(s *document.Sentence) { s.Tokens[1].Protected = true }},
		{"missing_pos", func(s *document.Sentence) { s.Tokens[1].Tag = "" }},
		{"missing_pos", func(s *document.Sentence) { s.Tokens[6].Tag = "" }},
	} {
		sentence := templateSentence()
		row.edit(&sentence)
		pattern, err := feature.PreparePOSPattern(t.Context(), sentence, nil, sequenceLimits())
		c.Assert(err, qt.IsNil)
		c.Assert(pattern.Available(), qt.IsFalse)
		c.Assert(pattern.Key(), qt.Equals, "")
		c.Assert(pattern.Reason(), qt.Equals, row.reason)
	}
}

func TestPOSPatternLimitsAndCatalogOwnership(t *testing.T) {
	c := qt.New(t)
	_, err := feature.PreparePOSPattern(t.Context(), templateSentence(), []bool{}, sequenceLimits())
	c.Assert(err, qt.IsNotNil)
	_, err = feature.PreparePOSPattern(t.Context(), document.Sentence{Text: strings.Repeat("x", 1001)}, nil, sequenceLimits())
	c.Assert(err, qt.IsNotNil)
	_, err = feature.PreparePOSPattern(t.Context(), document.Sentence{Text: "\xff"}, nil, sequenceLimits())
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = feature.PreparePOSPattern(ctx, templateSentence(), nil, sequenceLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	catalog := feature.PatternCatalog()
	c.Assert(catalog, qt.HasLen, 2)
	c.Assert(catalog[0].ID, qt.Equals, "prose-ngram")
	c.Assert(catalog[1].ID, qt.Equals, "surface-pos-template")
	capabilities := slices.Clone(catalog[1].Requires)
	catalog[1].Requires[0] = "changed"
	c.Assert(feature.PatternCatalog()[1].Requires, qt.DeepEquals, capabilities)
}
