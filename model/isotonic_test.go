package model_test

import (
	"context"
	"math"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestIsotonicInterpolationAndRange(t *testing.T) {
	for _, row := range []struct {
		name            string
		left, right, at float64
		response        float64
	}{
		{"left endpoint", -2, 2, -2, 0},
		{"right endpoint", -2, 2, 2, 1},
		{"middle", -2, 2, 0, 0.5},
		{"quarter", -2, 2, -1, 0.25},
		{"extreme interval", -math.MaxFloat64, math.MaxFloat64, 0, 0.5},
		{"subnormal interval", 0, 2 * math.SmallestNonzeroFloat64, math.SmallestNonzeroFloat64, 0.5},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, err := model.NewIsotonic(model.IsotonicParameters{Scores: []float64{row.left, row.right}, Responses: []float64{0, 1}})
			c.Assert(err, qt.IsNil)
			got, err := m.Evaluate(t.Context(), row.at)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.Equals, row.response)
		})
	}
	for _, score := range []float64{math.Nextafter(-1, math.Inf(-1)), math.Nextafter(1, math.Inf(1))} {
		c := qt.New(t)
		m, err := model.NewIsotonic(model.IsotonicParameters{Scores: []float64{-1, 1}, Responses: []float64{0.2, 0.8}})
		c.Assert(err, qt.IsNil)
		value, err := m.Evaluate(t.Context(), score)
		c.Assert(err, qt.ErrorIs, model.ErrCalibrationRange)
		c.Assert(value, qt.Equals, 0.0)
	}
}

func TestIsotonicOwnershipAndConcurrentCalls(t *testing.T) {
	c := qt.New(t)
	p := model.IsotonicParameters{Scores: []float64{-1, 1}, Responses: []float64{0, 1}}
	m, err := model.NewIsotonic(p)
	c.Assert(err, qt.IsNil)
	p.Scores[0], p.Responses[0] = 99, 99
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			copy := m.Parameters()
			copy.Scores[0], copy.Responses[0] = -100, 1
			got, err := m.Evaluate(t.Context(), 0)
			c.Check(err, qt.IsNil)
			c.Check(got, qt.Equals, 0.5)
		})
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	value, err := m.Evaluate(ctx, 0)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(value, qt.Equals, 0.0)
}

func TestIsotonicRejectsInvalidEvaluation(t *testing.T) {
	c := qt.New(t)
	m, err := model.NewIsotonic(model.IsotonicParameters{Scores: []float64{1}, Responses: []float64{0.5}})
	c.Assert(err, qt.IsNil)
	for _, score := range []float64{math.NaN(), math.Inf(-1), math.Inf(1)} {
		value, err := m.Evaluate(t.Context(), score)
		c.Assert(err, qt.IsNotNil)
		c.Assert(value, qt.Equals, 0.0)
	}
	for _, empty := range []*model.Isotonic{nil, {}} {
		value, err := empty.Evaluate(t.Context(), 1)
		c.Assert(err, qt.IsNotNil)
		c.Assert(value, qt.Equals, 0.0)
		c.Assert(empty.Parameters(), qt.DeepEquals, model.IsotonicParameters{})
	}
}
