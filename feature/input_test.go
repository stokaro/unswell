package feature_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func TestInvalidInputsDoNotProduceMeasurements(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*document.Block, *feature.Identity, *feature.Limits)
	}{
		{"bounds", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences[0].Tokens[0].End = 100 }},
		{"mapping", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences[0].Tokens[0].Spans = nil }},
		{"pos", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences[0].Tokens[0].Tag = "" }},
		{"normal", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences[0].Tokens[0].Normal = "" }},
		{"omitted-all", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences = nil }},
		{"omitted-prefix", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) {
			b.Sentences[0].Tokens = b.Sentences[0].Tokens[1:]
		}},
		{"omitted-tail", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) {
			b.Sentences[0].Tokens = b.Sentences[0].Tokens[:1]
		}},
		{"utf8-tag", func(b *document.Block, _ *feature.Identity, _ *feature.Limits) { b.Sentences[0].Tokens[0].Tag = "\xff" }},
		{"identity", func(_ *document.Block, i *feature.Identity, _ *feature.Limits) { i.Preprocessing = "" }},
		{"capabilities", func(_ *document.Block, i *feature.Identity, _ *feature.Limits) { i.NLP.Capabilities = nil }},
		{"unknown", func(_ *document.Block, i *feature.Identity, _ *feature.Limits) {
			i.Capabilities = []nlp.Capability{"unknown"}
		}},
		{"tokens", func(_ *document.Block, _ *feature.Identity, l *feature.Limits) { l.MaxTokens, l.MaxUniqueWords = 1, 1 }},
		{"unique", func(_ *document.Block, _ *feature.Identity, l *feature.Limits) { l.MaxUniqueWords = 1 }},
		{"bytes", func(_ *document.Block, _ *feature.Identity, l *feature.Limits) { l.MaxBytes = 1 }},
		{"limits", func(_ *document.Block, _ *feature.Identity, l *feature.Limits) { l.MaxTokens = 0 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			block, identity, limits := testBlock("a b"), testIdentity(), testLimits()
			row.change(&block, &identity, &limits)
			m, err := feature.Measure(t.Context(), block, identity, limits)
			c.Assert(err, qt.IsNotNil)
			c.Assert(m.Hash(), qt.Equals, "")
			c.Assert(m.Counts().Available, qt.IsFalse)
		})
	}
}

func TestSetBoundsAndBlockIdentity(t *testing.T) {
	c := qt.New(t)
	first, second := testBlock("a b"), testBlock("c d")
	second.ID = 1
	limits := testLimits()
	limits.MaxTokens, limits.MaxUniqueWords = 3, 3
	set, err := feature.NewSet(t.Context(), []document.Block{first, second}, testIdentity(), limits)
	c.Assert(err, qt.ErrorIs, feature.ErrTokenLimit)
	c.Assert(set, qt.IsNil)
	_, err = feature.NewSet(t.Context(), []document.Block{first, first}, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorMatches, ".*duplicate.*")
	limits = testLimits()
	limits.MaxBlocks = 1
	_, err = feature.NewSet(t.Context(), []document.Block{first, second}, testIdentity(), limits)
	c.Assert(err, qt.ErrorMatches, ".*max_blocks")
	set, err = feature.NewSet(t.Context(), []document.Block{second, first}, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	m, err := set.Block(0)
	c.Assert(err, qt.IsNil)
	c.Assert(m.SentenceLengths(), qt.DeepEquals, []int{2})
	_, err = set.Block(9)
	c.Assert(err, qt.ErrorMatches, ".*was not measured")
}

func TestCapabilityAbsenceAndUnsupportedUnit(t *testing.T) {
	c := qt.New(t)
	block, identity := testBlock("a b"), testIdentity()
	identity.Capabilities = nil
	m, err := feature.Measure(t.Context(), block, identity, testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(m.Counts().Available, qt.IsFalse)
	for _, value := range m.Values() {
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, "capability_missing")
	}
	block.Kind = "heading"
	m, err = feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	for _, value := range m.Values() {
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, "unsupported_unit")
	}
}

func TestCatalogOwnershipAndPresentZero(t *testing.T) {
	c := qt.New(t)
	first := feature.Catalog()
	c.Assert(first, qt.HasLen, 14)
	first[0].Requires[0] = nlp.Dependencies
	first[0].Formula = "changed"
	c.Assert(feature.Catalog()[0].Requires, qt.DeepEquals, []nlp.Capability{nlp.Tokens, nlp.Sentences})
	c.Assert(feature.Catalog()[0].Formula, qt.Not(qt.Equals), "changed")
	m, err := feature.Measure(t.Context(), testBlock("a b"), testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(numeric(t, m, "verb-token-ratio"), qt.Equals, float64(0))
	var zero feature.Measurements
	c.Assert(zero.Counts().Available, qt.IsFalse)
	c.Assert(zero.Values()[0].Reason, qt.Equals, "not_computed")
}
