package feature_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
)

func TestExcludedFeatureInputIsUnavailableAndValidated(t *testing.T) {
	c := qt.New(t)
	block := testBlock("Excluded prose.")
	block.Sentences = nil
	_, err := feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorMatches, "feature tokens omit extracted prose")
	block.Excluded = true
	set, err := feature.NewSet(t.Context(), []document.Block{block}, testIdentity(), testLimits())
	c.Assert(err, qt.IsNil)
	m, err := set.Block(block.ID)
	c.Assert(err, qt.IsNil)
	c.Assert(m.Counts().Available, qt.IsFalse)
	c.Assert(m.Spans(), qt.HasLen, 0)
	for _, descriptor := range feature.Catalog() {
		value, err := m.Value(descriptor.ID)
		c.Assert(err, qt.IsNil)
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, "excluded_unit")
	}
	block.Words = 1
	_, err = feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorMatches, "excluded feature block must not contain NLP data")
	block.Words, block.Map = 0, nil
	_, err = feature.Measure(t.Context(), block, testIdentity(), testLimits())
	c.Assert(err, qt.ErrorMatches, "feature input has an invalid mapped text")
}
