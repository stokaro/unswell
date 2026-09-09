package feature_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func lexicalTermCounts(terms []feature.LexicalTerm) map[string]int {
	result := make(map[string]int)
	for _, term := range terms {
		result[term.Key] = term.Count
	}
	return result
}

func TestLexicalWordCountsKeepNegationAndSentenceBoundaries(t *testing.T) {
	c := qt.New(t)
	units := featureUnits(t, testBlock("Clients must retry, clients may not retry. Clients must wait."), false)
	unit := units[len(units)-1]
	options := feature.LexicalOptions{WordMin: 1, WordMax: 3}
	terms, err := feature.CountLexical(t.Context(), unit, unitIdentity(unit), testLimits(), options)
	c.Assert(err, qt.IsNil)
	counts := lexicalTermCounts(terms)
	c.Assert(counts[`w:["clients"]`], qt.Equals, 3)
	c.Assert(counts[`w:["must"]`], qt.Equals, 2)
	c.Assert(counts[`w:["may","not","retry"]`], qt.Equals, 1)
	c.Assert(counts[`w:["retry","clients"]`], qt.Equals, 0)
	c.Assert(counts[`w:["clients","must"]`], qt.Equals, 2)
	for _, term := range terms {
		c.Assert(feature.ValidLexicalKey(term.Key, options), qt.IsTrue)
	}
	before := slices.Clone(terms)
	terms[0].Key = "changed"
	again, err := feature.CountLexical(t.Context(), unit, unitIdentity(unit), testLimits(), options)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, before)
}

func TestLexicalCharactersCountRunesAndCollapseWhitespace(t *testing.T) {
	c := qt.New(t)
	doc, err := extract.Parse(t.Context(), document.Source{Name: "unicode.txt", Format: document.Plain, Bytes: []byte("ÉÉ\tÉ.")},
		extract.Options{})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, doc.Blocks[0], false)[0]
	options := feature.LexicalOptions{CharMin: 1, CharMax: 2}
	terms, err := feature.CountLexical(t.Context(), unit, unitIdentity(unit), testLimits(), options)
	c.Assert(err, qt.IsNil)
	counts := lexicalTermCounts(terms)
	c.Assert(counts["c:é"], qt.Equals, 3)
	c.Assert(counts["c:éé"], qt.Equals, 1)
	c.Assert(counts["c:é "], qt.Equals, 1)
	c.Assert(counts["c: é"], qt.Equals, 1)
	c.Assert(counts["c:é."], qt.Equals, 1)
	for _, term := range terms {
		c.Assert(feature.ValidLexicalKey(term.Key, options), qt.IsTrue)
	}
}

func TestLexicalCountsDoNotCrossProtectedPieces(t *testing.T) {
	c := qt.New(t)
	units := featureUnits(t, testBlock("Must \x00 retry."), false)
	options := feature.LexicalOptions{WordMin: 2, WordMax: 2, CharMin: 6, CharMax: 6}
	for _, unit := range units {
		terms, err := feature.CountLexical(t.Context(), unit, unitIdentity(unit), testLimits(), options)
		c.Assert(err, qt.IsNil)
		counts := lexicalTermCounts(terms)
		c.Assert(counts[`w:["must","retry"]`], qt.Equals, 0)
		c.Assert(counts["c:must r"], qt.Equals, 0)
	}
}

func TestLexicalCountFailures(t *testing.T) {
	c := qt.New(t)
	unit := featureUnits(t, testBlock("Clients must retry."), false)[0]
	options := feature.LexicalOptions{WordMin: 1, WordMax: 2}
	limits := testLimits()
	limits.MaxUniqueWords = 1
	terms, err := feature.CountLexical(t.Context(), unit, unitIdentity(unit), limits, options)
	c.Assert(err, qt.ErrorMatches, ".*budget exceeded")
	c.Assert(terms, qt.IsNil)
	_, err = feature.CountLexical(t.Context(), nlp.PreparedUnit{}, unitIdentity(unit), testLimits(), options)
	c.Assert(err, qt.IsNotNil)
	_, err = feature.CountLexical(t.Context(), unit, unitIdentity(unit), testLimits(), feature.LexicalOptions{})
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = feature.CountLexical(ctx, unit, unitIdentity(unit), testLimits(), options)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	for _, key := range []string{"x:bad", `w:["UPPER"]`, `w:[]`, `w:["ok", "spaces"]`, "c:x", "w:\xff"} {
		c.Assert(feature.ValidLexicalKey(key, options), qt.IsFalse)
	}
}

func FuzzLexicalKey(f *testing.F) {
	for _, key := range []string{`w:["must","retry"]`, "c:éé", `w:["a\u0000b"]`, "c:\x00", `w:["30","seconds"]`} {
		f.Add(key)
	}
	f.Fuzz(func(t *testing.T, key string) {
		if len(key) > 65536 {
			t.Skip()
		}
		_ = feature.ValidLexicalKey(key, feature.LexicalOptions{WordMin: 1, WordMax: 3, CharMin: 1, CharMax: 6})
	})
}
