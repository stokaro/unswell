package evaluation

import (
	"context"
	"fmt"
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

func BenchmarkPairedSourceGroups(b *testing.B) {
	for _, groups := range []int{100, 1000} {
		b.Run(fmt.Sprint(groups), func(b *testing.B) {
			rows := make([]PairedObservation, groups*2)
			for i := range rows {
				value, positive := 0.8, true
				observation := Observation{UnitID: fmt.Sprint(i), GroupID: fmt.Sprint(i / 2), Label: i % 2,
					Response: &value, Positive: &positive}
				rows[i] = PairedObservation{observation, observation}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for b.Loop() {
				if _, err := Compare(b.Context(), rows, [2]float64{0.5, 0.5}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
