package generation_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/generation"
)

// The floor is a study variable: the origin experiment separated nothing on
// units of 12 to 49 words, so a study of longer prose raises it. A value below
// the built-in floor cannot weaken it.
func TestSamplerWordFloorOnlyRises(t *testing.T) {
	c := qt.New(t)
	c.Assert(generation.MinWords, qt.Equals, 12)
	for _, test := range []struct{ option, want int }{
		{0, 12}, {5, 12}, {12, 12}, {100, 100},
	} {
		got := generation.EligibleFloor(generation.Options{MinWords: test.option})
		c.Assert(got, qt.Equals, test.want, qt.Commentf("option %d", test.option))
	}
}
