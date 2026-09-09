package model_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
)

func TestTrainingShapeLimits(t *testing.T) {
	for _, row := range []struct {
		name        string
		rows, width int
	}{
		{"row limit", 100001, 1},
		{"feature limit", 2, 129},
		{"cell limit", 31251, 128},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			values := make([]float64, row.width)
			data := make([]model.Example, row.rows)
			for i := range data {
				data[i] = model.Example{Values: values, Label: i % 2}
			}
			result, err := model.FitLogistic(t.Context(), data, fitOptions())
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, model.FitResult{})
		})
	}
}

func TestCorrelatedColumnsRemainTrainable(t *testing.T) {
	c := qt.New(t)
	data := []model.Example{
		{Values: []float64{-1, -1}, Label: 0}, {Values: []float64{1, 1}, Label: 1},
	}
	result, err := model.FitLogistic(t.Context(), data, fitOptions())
	c.Assert(err, qt.IsNil)
	p := result.Model.Parameters()
	c.Assert(math.Abs(p.Weights[0]-p.Weights[1]) < 1e-10, qt.IsTrue)
	c.Assert(p.Weights[0] > 0, qt.IsTrue)
}

func BenchmarkFitLogistic(b *testing.B) {
	// Numerical throughput only: these labels do not represent an editorial corpus.
	data := make([]model.Example, 5000)
	for i := range data {
		values := make([]float64, 40)
		for j := range values {
			values[j] = float64((i*17+j*13)%101) / 100
		}
		data[i] = model.Example{Values: values, Label: i % 2}
	}
	options := model.FitOptions{L2: 0.1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 1000000000}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := model.FitLogistic(b.Context(), data, options); err != nil {
			b.Fatal(err)
		}
	}
}
