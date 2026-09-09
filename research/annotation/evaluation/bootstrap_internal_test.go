package evaluation

// White-box tests: Check private percentile formulas and cancel during a specific bootstrap draw;
// Compare exposes final intervals without a hook for individual resampling steps.

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestPercentileAndZeroEventBound(t *testing.T) {
	c := qt.New(t)
	c.Assert(percentile([]float64{1, 2, 5, 10}, 0.25), qt.Equals, 1.75)
	c.Assert(percentile([]float64{1, 2, 5, 10}, 0.975), qt.Equals, 9.625)
	groups := make([]GroupMetrics, 1000)
	for i := range groups {
		groups[i].Metrics.Counts.TN = 1
	}
	bound := zeroFlagBound(Summary{Groups: groups})
	c.Assert(bound.Upper, qt.IsNotNil)
	c.Assert(math.Abs(*bound.Upper-0.0029912495450953314) < 1e-14, qt.IsTrue)
	c.Assert(bound.Groups, qt.Equals, 1000)
}

type cancelDuringDraw struct {
	context.Context
	calls int
}

func (c *cancelDuringDraw) Err() error {
	c.calls++
	if c.calls >= 3 {
		return context.Canceled
	}
	return nil
}

func TestBootstrapCancellation(t *testing.T) {
	c := qt.New(t)
	ctx := &cancelDuringDraw{Context: t.Context()}
	result, err := bootstrap(ctx, make([]pairedGroup, 2))
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.IsNil)
}
