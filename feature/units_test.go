package feature_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func featureUnits(t *testing.T, block document.Block, pos bool) []nlp.PreparedUnit {
	t.Helper()
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	options := nlp.UnitOptions{Kinds: []string{"sentence", "paragraph", "fragment"},
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Limits:       nlp.UnitLimits{MaxBytes: 65536, MaxContextBytes: 65536, MaxUnits: 1000, MaxTokens: 1000, MaxSegments: 1024}}
	if pos {
		options.Capabilities = append(options.Capabilities, nlp.POS)
	}
	units, err := nlp.PrepareUnits(t.Context(), block, provider, options)
	c.Assert(err, qt.IsNil)
	return units
}

func unitIdentity(unit nlp.PreparedUnit) feature.Identity {
	i := testIdentity()
	i.NLP, i.Capabilities, i.Preprocessing = unit.Identity(), unit.Capabilities(), nlp.UnitContract
	return i
}

func TestUnitMeasurementsReuseDescriptiveFormulas(t *testing.T) {
	c := qt.New(t)
	units := featureUnits(t, testBlock("The cache may retry. The service cannot wait."), true)
	c.Assert(units, qt.HasLen, 3)
	for i, words := range []float64{4, 4, 8} {
		m, err := feature.MeasureUnit(t.Context(), units[i], unitIdentity(units[i]), testLimits())
		c.Assert(err, qt.IsNil)
		c.Assert(numeric(t, m, "prose-words"), qt.Equals, words)
		legacy, err := feature.Measure(t.Context(), units[i].Block(), unitIdentity(units[i]), testLimits())
		c.Assert(err, qt.IsNil)
		c.Assert(m.Values(), qt.DeepEquals, legacy.Values())
		c.Assert(m.Hash(), qt.Not(qt.Equals), legacy.Hash())
		descriptors, err := feature.UnitCatalog(units[i].Binding().Kind)
		c.Assert(err, qt.IsNil)
		for _, descriptor := range descriptors {
			c.Assert(descriptor.Scope, qt.Equals, units[i].Binding().Kind)
		}
	}
	block := testBlock("A heading message.")
	block.Kind = "heading"
	unit := featureUnits(t, block, false)[0]
	m, err := feature.MeasureUnit(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(numeric(t, m, "prose-words"), qt.Equals, float64(3))
	c.Assert(unit.Binding().Kind, qt.Equals, "fragment")
	value, err := m.Value("noun-token-ratio")
	c.Assert(err, qt.IsNil)
	c.Assert(value.Number, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "capability_missing")
}

func TestUnitMeasurementBindsContextAndRepresentation(t *testing.T) {
	c := qt.New(t)
	first := featureUnits(t, testBlock("The cache may retry. The service cannot wait."), true)[0]
	second := featureUnits(t, testBlock("The cache may retry. The service must wait."), true)[0]
	c.Assert(first.Binding().Segments, qt.DeepEquals, second.Binding().Segments)
	c.Assert(first.Binding().TextSHA256, qt.Equals, second.Binding().TextSHA256)
	c.Assert(first.Binding().ContextSHA256, qt.Not(qt.Equals), second.Binding().ContextSHA256)
	m, err := feature.MeasureUnit(t.Context(), first, unitIdentity(first), testLimits())
	c.Assert(err, qt.IsNil)
	again, err := feature.MeasureUnit(t.Context(), second, unitIdentity(second), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(m.Values(), qt.DeepEquals, again.Values())
	c.Assert(m.Hash(), qt.Not(qt.Equals), again.Hash())
	identity := unitIdentity(first)
	slices.Reverse(identity.Capabilities)
	same, err := feature.MeasureUnit(t.Context(), first, identity, testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(same.Hash(), qt.Equals, m.Hash())
	identity.NLP.Version = "changed"
	_, err = feature.MeasureUnit(t.Context(), first, identity, testLimits())
	c.Assert(err, qt.ErrorMatches, ".*NLP identity or capabilities.*")
	identity = unitIdentity(first)
	identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
	_, err = feature.MeasureUnit(t.Context(), first, identity, testLimits())
	c.Assert(err, qt.ErrorMatches, ".*NLP identity or capabilities.*")
}

func TestUnitMeasurementFailures(t *testing.T) {
	c := qt.New(t)
	_, err := feature.MeasureUnit(t.Context(), nlp.PreparedUnit{}, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorMatches, "invalid prepared feature unit")
	_, err = feature.UnitCatalog("document")
	c.Assert(err, qt.IsNotNil)
	unit := featureUnits(t, testBlock("The cache may retry."), false)[0]
	limits := testLimits()
	limits.MaxTokens = 1
	_, err = feature.MeasureUnit(t.Context(), unit, unitIdentity(unit), limits)
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = feature.MeasureUnit(ctx, unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
