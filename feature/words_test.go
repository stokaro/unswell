package feature_test

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
)

func wordLimits() feature.WordLimits {
	return feature.WordLimits{MaxWords: 100, MaxUniqueWords: 100, MaxBytes: 1000}
}

func TestWordOverlapKeepsCandidateStopwords(t *testing.T) {
	c := qt.New(t)
	left, err := feature.NewWordSet(t.Context(), []string{"the", "cache", "cache", "timeout"}, wordLimits())
	c.Assert(err, qt.IsNil)
	right, err := feature.NewWordSet(t.Context(), []string{"the", "cache", "budget"}, wordLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(left.Len(), qt.Equals, 3)
	c.Assert(left.ContentKeys(), qt.DeepEquals, []string{"cache", "timeout"})
	c.Assert(right.ContentKeys(), qt.DeepEquals, []string{"budget", "cache"})
	value, err := feature.CompareWords(t.Context(), left, right)
	c.Assert(err, qt.IsNil)
	c.Assert(value.Available(), qt.IsTrue)
	c.Assert(value.Intersection(), qt.Equals, 2)
	c.Assert(value.Union(), qt.Equals, 4)
	c.Assert(*value.Values()[2].Number, qt.Equals, 0.5)
	reverse, err := feature.CompareWords(t.Context(), right, left)
	c.Assert(err, qt.IsNil)
	c.Assert(reverse.Values(), qt.DeepEquals, value.Values())
	for i, descriptor := range feature.LexicalCatalog() {
		c.Assert(value.Values()[i].ID, qt.Equals, descriptor.ID)
		c.Assert(value.Values()[i].Version, qt.Equals, descriptor.Version)
		c.Assert(value.Values()[i].Unit, qt.Equals, descriptor.Unit)
	}
}

func TestEmptyOverlapIsNotSimilarityEvidence(t *testing.T) {
	c := qt.New(t)
	empty, err := feature.NewWordSet(t.Context(), nil, wordLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(empty.Available(), qt.IsTrue)
	value, err := feature.CompareWords(t.Context(), empty, empty)
	c.Assert(err, qt.IsNil)
	c.Assert(value.Available(), qt.IsTrue)
	c.Assert(value.Union(), qt.Equals, 0)
	c.Assert(*value.Values()[0].Number, qt.Equals, float64(0))
	c.Assert(value.Values()[2].Number, qt.IsNil)
	c.Assert(value.Values()[2].Reason, qt.Equals, "empty_union")
	data, err := json.Marshal(value.Values())
	c.Assert(err, qt.IsNil)
	c.Assert(string(data), qt.Contains, `"number":null`)
	var absent feature.WordSet
	c.Assert(absent.Available(), qt.IsFalse)
	value, err = feature.CompareWords(t.Context(), empty, absent)
	c.Assert(err, qt.ErrorMatches, ".*two available word sets")
	c.Assert(value.Available(), qt.IsFalse)
	for _, item := range value.Values() {
		c.Assert(item.Number, qt.IsNil)
		c.Assert(item.Reason, qt.Equals, "not_computed")
	}
}

func TestWordSetsOwnTheirContents(t *testing.T) {
	c := qt.New(t)
	input := []string{"café", "server", "café"}
	set, err := feature.NewWordSet(t.Context(), input, wordLimits())
	c.Assert(err, qt.IsNil)
	input[0] = "changed"
	var workers sync.WaitGroup
	for range 8 {
		workers.Go(func() {
			words := set.Words()
			c.Check(words, qt.DeepEquals, []string{"café", "server"})
			words[0] = "changed"
			set.ContentKeys()[0] = "changed"
			value, compareErr := feature.CompareWords(t.Context(), set, set)
			c.Check(compareErr, qt.IsNil)
			c.Check(value.Intersection(), qt.Equals, 2)
			c.Check(value.Union(), qt.Equals, 2)
			*value.Values()[0].Number = 100
		})
	}
	workers.Wait()
	c.Assert(set.Words(), qt.DeepEquals, []string{"café", "server"})
}

func TestWordSetBoundsAndCancellation(t *testing.T) {
	c := qt.New(t)
	for _, row := range []struct {
		words  []string
		limits feature.WordLimits
	}{
		{[]string{"one", "two"}, feature.WordLimits{MaxWords: 1, MaxUniqueWords: 1, MaxBytes: 100}},
		{[]string{"one", "two"}, feature.WordLimits{MaxWords: 3, MaxUniqueWords: 1, MaxBytes: 100}},
		{[]string{"one", "two"}, feature.WordLimits{MaxWords: 3, MaxUniqueWords: 3, MaxBytes: 5}},
		{[]string{"one"}, feature.WordLimits{}},
		{[]string{""}, wordLimits()},
		{[]string{"\xff"}, wordLimits()},
		{[]string{"hidden\x00boundary"}, wordLimits()},
	} {
		value, err := feature.NewWordSet(t.Context(), row.words, row.limits)
		c.Assert(err, qt.IsNotNil)
		c.Assert(value.Available(), qt.IsFalse)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := feature.NewWordSet(ctx, []string{"one"}, wordLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	set, err := feature.NewWordSet(t.Context(), []string{"one"}, wordLimits())
	c.Assert(err, qt.IsNil)
	_, err = feature.CompareWords(ctx, set, set)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestLexicalNormalizationIsExplicit(t *testing.T) {
	c := qt.New(t)
	set, err := feature.NewWordSet(t.Context(), []string{"Cache", "cache", "not", "no", "may", "must"}, wordLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(set.Len(), qt.Equals, 6)
	c.Assert(set.Words(), qt.DeepEquals, []string{"Cache", "cache", "may", "must", "no", "not"})
	c.Assert(set.ContentKeys(), qt.DeepEquals, []string{"Cache", "cache"})
	c.Assert(feature.InformativeWord("é"), qt.IsFalse)
	c.Assert(feature.InformativeWord("界"), qt.IsTrue)
}
