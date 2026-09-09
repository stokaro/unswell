package training_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func TestFeatureOrderAndReturnedOwnership(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := fittingOptions()
	options.Features = []string{"prose-words", "noun-token-ratio"}
	first, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	slices.Reverse(options.Features)
	second, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(first, qt.DeepEquals, second)
	c.Assert(first.Options.Features, qt.DeepEquals, []string{"noun-token-ratio", "prose-words"})
	c.Assert(first.Identity.Columns[0].ID, qt.Equals, "noun-token-ratio")
	c.Assert(first.Identity.Columns[1].ID, qt.Equals, "prose-words")
	c.Assert(first.Logistic.Means[1], qt.Equals, float64(6))
	options.Features[0] = "changed"
	c.Assert(first.Options.Features[0], qt.Equals, "noun-token-ratio")
	first.Identity.Columns[0].ID = "changed"
	first.Identity.Capabilities[0] = "changed"
	first.Logistic.Weights[0] = 99
	first.Calibration.Scores[0] = 99
	first.Partitions[0].Rows[0].UnitID = "changed"
	options = second.Options
	again, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, second)
}
