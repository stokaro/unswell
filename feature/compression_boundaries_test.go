package feature_test

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func TestCompressionRejectsInvalidReferences(t *testing.T) {
	for _, row := range []struct {
		name, reference string
		level, budget   int
	}{
		{"empty", "", 9, 10000},
		{"UTF8", "\xff", 9, 10000},
		{"protected", "text\x00text", 9, 10000},
		{"prefix limit", strings.Repeat("x", 32<<10), 9, 1 << 20},
		{"default level is not pinned", "text", -1, 10000},
		{"level limit", "text", 10, 10000},
		{"missing budget", "text", 9, 0},
		{"budget maximum", "text", 9, 1<<20 + 1},
		{"baseline budget", "text", 9, 5},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, err := feature.NewCompression(t.Context(), row.reference,
				feature.CompressionOptions{Level: row.level, MaxInputBytes: row.budget})
			c.Assert(err, qt.IsNotNil)
			c.Assert(m, qt.IsNil)
		})
	}
}

func TestCompressionRejectsMissingTargetAndWrongIdentity(t *testing.T) {
	c := qt.New(t)
	m, err := feature.NewCompression(t.Context(), "Reference.", feature.CompressionOptions{Level: 9, MaxInputBytes: 1000})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, testBlock("Cache retries."), false)[0]
	_, err = m.Measure(t.Context(), nlp.PreparedUnit{}, unitIdentity(unit), testLimits())
	c.Assert(err, qt.ErrorMatches, ".*prepared target.*")
	identity := unitIdentity(unit)
	identity.NLP.Name = "another-provider"
	_, err = m.Measure(t.Context(), unit, identity, testLimits())
	c.Assert(err, qt.ErrorMatches, ".*NLP identity.*")
	limited, err := feature.NewCompression(t.Context(), "Reference.", feature.CompressionOptions{Level: 9, MaxInputBytes: 12})
	c.Assert(err, qt.IsNil)
	result, err := limited.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.ErrorMatches, ".*input-byte budget.*")
	c.Assert(result, qt.DeepEquals, feature.CompressionResult{})
	var missing *feature.Compression
	_, err = missing.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.IsNotNil)
	c.Assert(missing.Identity(), qt.Equals, feature.CompressionIdentity{})
}

type compressionCancelContext struct {
	context.Context
	cancel context.CancelFunc
	checks atomic.Int64
}

func (c *compressionCancelContext) Err() error {
	if c.checks.Add(-1) == 0 {
		c.cancel()
	}
	return c.Context.Err()
}

func TestCompressionCancellationReturnsNoPartialCounts(t *testing.T) {
	c := qt.New(t)
	m, err := feature.NewCompression(t.Context(), strings.Repeat("Reference. ", 2500),
		feature.CompressionOptions{Level: 9, MaxInputBytes: 1 << 20})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, testBlock("Cache retries."), false)[0]
	parent, cancel := context.WithCancel(t.Context())
	defer cancel()
	ctx := &compressionCancelContext{Context: parent, cancel: cancel}
	ctx.checks.Store(10)
	result, err := m.Measure(ctx, unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, feature.CompressionResult{})
	_, err = feature.NewCompression(parent, "Reference.", m.Identity().Options)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestCompressionConcurrentCallsOwnTheirResults(t *testing.T) {
	c := qt.New(t)
	m, err := feature.NewCompression(t.Context(), "Reference.", feature.CompressionOptions{Level: 9, MaxInputBytes: 10000})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, testBlock("Cache retries."), false)[0]
	want, err := m.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.IsNil)
	for range 12 {
		t.Run("owned values", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := m.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
			got.Binding.Segments[0].End++
			got.Values[0].Reason = "changed"
		})
	}
}
