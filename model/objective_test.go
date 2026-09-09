package model

import (
	"math"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestAnalyticDerivativesMatchFiniteDifferences(t *testing.T) {
	c := qt.New(t)
	budget := workBudget{limit: 1000000}
	data, err := prepareTraining(t.Context(), []Example{
		{Values: []float64{-2, 3}, Label: 0}, {Values: []float64{1, 1}, Label: 1},
		{Values: []float64{3, -4}, Label: 0}, {Values: []float64{4, 2}, Label: 1},
	}, &budget)
	c.Assert(err, qt.IsNil)
	theta := []float64{0.3, -0.8, 1.2}
	want, err := objective(t.Context(), data, theta, 0.2, true, &budget)
	c.Assert(err, qt.IsNil)
	const step = 1e-5
	for i := range theta {
		plus, minus := slices.Clone(theta), slices.Clone(theta)
		plus[i] += step
		minus[i] -= step
		a, err := objective(t.Context(), data, plus, 0.2, false, &budget)
		c.Assert(err, qt.IsNil)
		b, err := objective(t.Context(), data, minus, 0.2, false, &budget)
		c.Assert(err, qt.IsNil)
		c.Assert(math.Abs(want.gradient[i]-((a.loss-b.loss)/(2*step))) <= 1e-8, qt.IsTrue)
		for j := range theta {
			c.Assert(math.Abs(want.hessian[j*len(theta)+i]-(a.gradient[j]-b.gradient[j])/(2*step)) <= 1e-8, qt.IsTrue)
		}
	}
}

func TestCholeskySolvesPositiveSystemAndRejectsIndefinite(t *testing.T) {
	c := qt.New(t)
	// [4 2; 2 3] * [1; 2] = [8; 8].
	result, err := solvePositive([]float64{4, 2, 2, 3}, []float64{8, 8})
	c.Assert(err, qt.IsNil)
	c.Assert(math.Abs(result[0]-(1.0)) <= 1e-12, qt.IsTrue)
	c.Assert(math.Abs(result[1]-(2.0)) <= 1e-12, qt.IsTrue)
	_, err = solvePositive([]float64{1, 2, 2, 1}, []float64{1, 1})
	c.Assert(err, qt.ErrorIs, ErrNumerical)
	_, err = solvePositive([]float64{math.NaN()}, []float64{1})
	c.Assert(err, qt.ErrorIs, ErrNumerical)
}

func TestRoundoffAcceptanceKeepsLossAndGradientLimits(t *testing.T) {
	for _, row := range []struct {
		name                     string
		loss, gradient, decrease float64
		accepted                 bool
	}{
		{"roundoff with gradient improvement", math.Nextafter(1, 2), 1e-10, 1e-16, true},
		{"no gradient improvement", math.Nextafter(1, 2), 1e-8, 1e-16, false},
		{"material loss increase", 1.01, 0, 1e-16, false},
		{"resolvable expected decrease", math.Nextafter(1, 2), 1e-10, 0.01, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			current := derivatives{loss: 1, gradient: []float64{1e-8}}
			next := derivatives{loss: row.loss, gradient: []float64{row.gradient}}
			c.Assert(acceptableStep(current, next, row.decrease), qt.Equals, row.accepted)
		})
	}
}
