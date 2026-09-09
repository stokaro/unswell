package feature_test

import (
	"crypto/sha256"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
)

func TestCompressionStoredBlockControl(t *testing.T) {
	c := qt.New(t)
	reference := "An independent reference paragraph."
	m, err := feature.NewCompression(t.Context(), reference, feature.CompressionOptions{Level: 0, MaxInputBytes: 1 << 20})
	c.Assert(err, qt.IsNil)
	unit := featureUnits(t, testBlock("The cache cannot retry."), false)[0]
	result, err := m.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Counts.TargetBytes, qt.Equals, len(unit.Block().Text))
	// Small nonempty stored streams use the same framing, so only target bytes differ.
	c.Assert(result.Counts.SeedTargetBytes-result.Counts.SeedBytes, qt.Equals, result.Counts.TargetBytes)
	c.Assert(result.Counts.ControlTargetBytes-result.Counts.ControlBytes, qt.Equals, result.Counts.TargetBytes)
	c.Assert(*result.Values[0].Number, qt.Equals, 1.0)
	c.Assert(*result.Values[1].Number, qt.Equals, 0.0)
	c.Assert(result.Identity.Contract, qt.Equals, feature.CompressionContract)
	c.Assert(result.Identity.GoVersion, qt.Equals, runtime.Version())
	c.Assert(result.Identity.ReferenceSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(reference+"\n"))))
	c.Assert(result.Binding, qt.DeepEquals, unit.Binding())
	c.Assert(result.Identity, qt.Equals, m.Identity())
	descriptors, err := feature.CompressionCatalog(unit.Binding().Kind)
	c.Assert(err, qt.IsNil)
	for i, value := range result.Values {
		c.Assert(value.ID, qt.Equals, descriptors[i].ID)
		c.Assert(value.Unit, qt.Equals, descriptors[i].Unit)
		c.Assert(value.Version, qt.Equals, descriptors[i].Version)
	}
	_, err = feature.CompressionCatalog("document")
	c.Assert(err, qt.IsNotNil)
}

func TestCompressionBindsTargetContextAndSettings(t *testing.T) {
	c := qt.New(t)
	m, err := feature.NewCompression(t.Context(), "Reference prose.", feature.CompressionOptions{Level: 9, MaxInputBytes: 10000})
	c.Assert(err, qt.IsNil)
	first := featureUnits(t, testBlock("The cache may retry. The service cannot wait."), false)[0]
	second := featureUnits(t, testBlock("The cache may retry. The service must wait."), false)[0]
	a, err := m.Measure(t.Context(), first, unitIdentity(first), testLimits())
	c.Assert(err, qt.IsNil)
	b, err := m.Measure(t.Context(), second, unitIdentity(second), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(a.Values, qt.DeepEquals, b.Values)
	c.Assert(a.Hash, qt.Not(qt.Equals), b.Hash)
	identity := unitIdentity(first)
	slices.Reverse(identity.Capabilities)
	slices.Reverse(identity.NLP.Capabilities)
	reordered, err := m.Measure(t.Context(), first, identity, testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(reordered.Hash, qt.Equals, a.Hash)
	identity.Source = "different-source"
	changed, err := m.Measure(t.Context(), first, identity, testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(changed.Hash, qt.Not(qt.Equals), a.Hash)
	other, err := feature.NewCompression(t.Context(), "Different reference.", m.Identity().Options)
	c.Assert(err, qt.IsNil)
	changed, err = other.Measure(t.Context(), first, unitIdentity(first), testLimits())
	c.Assert(err, qt.IsNil)
	c.Assert(changed.Hash, qt.Not(qt.Equals), a.Hash)
}

func TestCompressionPreservesPreparedBoundaries(t *testing.T) {
	c := qt.New(t)
	m, err := feature.NewCompression(t.Context(), "An API cache reference.", feature.CompressionOptions{Level: 0, MaxInputBytes: 10000})
	c.Assert(err, qt.IsNil)
	for _, text := range []string{
		"café café HTTP_500", "404", "cache_key cache_key",
		"Return nil. `SECRET_CODE` Never concatenate this.", "CRLF text.\r\nMore text.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "case.md", Format: document.Markdown, Bytes: []byte(text)},
				extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			units := featureUnits(t, doc.Blocks[0], false)
			c.Assert(len(units) > 0, qt.IsTrue)
			for _, unit := range units {
				got, err := m.Measure(t.Context(), unit, unitIdentity(unit), testLimits())
				c.Assert(err, qt.IsNil)
				c.Assert(got.Counts.TargetBytes, qt.Equals, len(unit.Block().Text))
				c.Assert(strings.ContainsRune(unit.Block().Text, 0), qt.IsFalse)
				c.Assert(unit.Block().Text, qt.Not(qt.Contains), "SECRET_CODE")
				c.Assert(*got.Values[0].Number, qt.Equals, 1.0)
			}
		})
	}
}
