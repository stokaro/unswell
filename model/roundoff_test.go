package model_test

import (
	"fmt"
	"math"
	"math/rand/v2"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestFitNumericalRoundoffAndBacktracking(t *testing.T) {
	for _, mode := range []string{"balanced", "rare-class"} {
		t.Run(mode, func(t *testing.T) {
			c := qt.New(t)
			// #nosec G404 -- Fixed seeds reproduce numerical regressions; no security values are generated.
			random := rand.New(rand.NewPCG(1, 2))
			backtracked := 0
			for scenario := range 200 {
				data := numericFixture(random, mode)
				t.Run(fmt.Sprint(scenario), func(t *testing.T) {
					c := qt.New(t)
					options := model.FitOptions{L2: math.Pow(10, float64(scenario%8-6)),
						Tolerance: 1e-10, MaxIterations: 100, MaxOperations: 100000000}
					result, err := model.FitLogistic(t.Context(), data, options)
					c.Assert(err, qt.IsNil)
					c.Assert(result.GradientNorm <= options.Tolerance, qt.IsTrue)
					if extraLineSearches(result, len(data), len(data[0].Values)) {
						backtracked++
					}
				})
			}
			if mode == "rare-class" {
				c.Assert(backtracked > 0, qt.IsTrue)
			}
		})
	}
}

// This generator supplies numerical stress cases, not editorial training labels.
func numericFixture(random *rand.Rand, mode string) []model.Example {
	n, width := 10+random.IntN(100), 1+random.IntN(8)
	data := make([]model.Example, n)
	for i := range data {
		values := make([]float64, width)
		for j := range values {
			values[j] = random.NormFloat64()
		}
		label := i % 2
		if mode == "rare-class" {
			label = min(i, 1)
		}
		data[i] = model.Example{Values: values, Label: label}
	}
	return data
}

func extraLineSearches(result model.FitResult, rows, width int) bool {
	n, d, p := int64(rows), int64(width), int64(width+1)
	iterations := int64(result.Iterations)
	hessian := n * (2*p + p*(p+1)/2)
	minimum := n*(4*d+1) + (iterations+1)*hessian + iterations*(p*p*p+2*p*p+n*2*p)
	return result.Operations > minimum
}
