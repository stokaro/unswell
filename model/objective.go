package model

import (
	"context"
	"fmt"
	"math"
)

type derivatives struct {
	loss              float64
	gradient, hessian []float64
}

func objective(ctx context.Context, data trainingData, theta []float64, l2 float64, second bool, budget *workBudget) (derivatives, error) {
	width := len(theta)
	work := width * 2
	if second {
		work += width * (width + 1) / 2
	}
	if err := budget.charge(int64(len(data.rows)) * int64(work)); err != nil {
		return derivatives{}, err
	}
	result := derivatives{gradient: make([]float64, width)}
	if second {
		result.hessian = make([]float64, width*width)
	}
	x := make([]float64, width)
	x[0] = 1
	for i, row := range data.rows {
		if err := ctx.Err(); err != nil {
			return derivatives{}, err
		}
		copy(x[1:], row)
		if err := result.addRow(theta, x, data.labels[i]); err != nil {
			return derivatives{}, err
		}
	}
	result.regularize(theta, l2, float64(len(data.rows)))
	if err := result.validate(); err != nil {
		return derivatives{}, err
	}
	if err := ctx.Err(); err != nil {
		return derivatives{}, err
	}
	return result, nil
}

func (d *derivatives) addRow(theta, x []float64, label float64) error {
	z := dot(theta, x)
	if !finite(z) {
		return fmt.Errorf("%w: linear score overflow", ErrNumerical)
	}
	d.loss += math.Max(z, 0) - label*z + math.Log1p(math.Exp(-math.Abs(z)))
	p := sigmoid(z)
	for j, value := range x {
		d.gradient[j] += (p - label) * value
	}
	if len(d.hessian) > 0 {
		addCurvature(d.hessian, x, p*(1-p))
	}
	return nil
}

func (d derivatives) validate() error {
	if !finite(d.loss) || !allFinite(d.gradient) || !allFinite(d.hessian) {
		return fmt.Errorf("%w: objective overflow", ErrNumerical)
	}
	return nil
}

func addCurvature(hessian, x []float64, curvature float64) {
	width := len(x)
	for j, left := range x {
		for k := 0; k <= j; k++ {
			hessian[j*width+k] += curvature * left * x[k]
		}
	}
}

func (d *derivatives) regularize(theta []float64, l2, rows float64) {
	d.loss /= rows
	for i := range theta {
		d.gradient[i] /= rows
		if i > 0 {
			d.loss += l2 * theta[i] * theta[i] / 2
			d.gradient[i] += l2 * theta[i]
		}
	}
	width := len(theta)
	for i := 0; i < width && len(d.hessian) > 0; i++ {
		for j := 0; j <= i; j++ {
			value := d.hessian[i*width+j] / rows
			if i == j && i > 0 {
				value += l2
			}
			d.hessian[i*width+j], d.hessian[j*width+i] = value, value
		}
	}
}

func allFinite(values []float64) bool {
	for _, value := range values {
		if !finite(value) {
			return false
		}
	}
	return true
}

func dot(a, b []float64) float64 {
	value := 0.0
	for i, v := range a {
		value += v * b[i]
	}
	return value
}
