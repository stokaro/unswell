package training_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

// A fit on arms of unequal length can reach a high recall by learning where the
// long units are. A band holds length fixed so that what the fit keeps is the
// prose, and the counts it reports have to say which units the band removed.
func TestWordBandExcludesUnitsOutsideItAndSaysSo(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)

	options := fittingOptions()
	open, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)

	options.MinUnitWords = 1_000_000
	narrow, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNotNil, qt.Commentf("a band that admits nothing leaves no rows to fit"))
	c.Assert(narrow, qt.DeepEquals, training.Artifact{})

	options.MinUnitWords, options.MaxUnitWords = 0, 1_000_000
	wide, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	for i, partition := range wide.Partitions {
		c.Assert(partition.Rows, qt.HasLen, len(open.Partitions[i].Rows),
			qt.Commentf("a band wider than every unit removes none of them"))
	}
}

func TestWordBandNeedsBoundsAndACountedFeature(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*training.Options)
	}{
		{"negative minimum", func(o *training.Options) { o.MinUnitWords = -1 }},
		{"negative maximum", func(o *training.Options) { o.MaxUnitWords = -1 }},
		{"inverted band", func(o *training.Options) { o.MinUnitWords, o.MaxUnitWords = 40, 20 }},
		{"no counted feature", func(o *training.Options) {
			o.MinUnitWords, o.Features = 5, []string{"type-token-ratio"}
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			input := testfixture.Load(t, "testdata")
			candidates, round := input.Compile(t)
			options := fittingOptions()
			row.edit(&options)
			result, err := training.Run(t.Context(), candidates, round, input.Files, options)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, training.Artifact{})
		})
	}
}
