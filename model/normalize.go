package model

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
)

type trainingData struct {
	rows          [][]float64
	labels        []float64
	means, scales []float64
	inputHash     string
	positive      int
}

func prepareTraining(ctx context.Context, examples []Example, budget *workBudget) (trainingData, error) {
	if err := ctx.Err(); err != nil {
		return trainingData{}, err
	}
	width, err := trainingWidth(examples)
	if err != nil {
		return trainingData{}, err
	}
	if err := budget.charge(int64(len(examples)) * int64(4*width+1)); err != nil {
		return trainingData{}, err
	}
	data, err := trainingStatistics(ctx, examples, width)
	if err != nil {
		return trainingData{}, err
	}
	if data.positive == 0 || data.positive == len(examples) {
		return trainingData{}, fmt.Errorf("logistic training requires both binary classes")
	}
	if err := data.normalize(ctx, examples); err != nil {
		return trainingData{}, err
	}
	return data, nil
}

func trainingWidth(examples []Example) (int, error) {
	if len(examples) < 2 || len(examples) > 100000 {
		return 0, fmt.Errorf("logistic training requires 2..100000 rows")
	}
	width := len(examples[0].Values)
	if width < 1 || width > MaxFeatures || len(examples)*width > 4000000 {
		return 0, fmt.Errorf("logistic training exceeds feature or cell limits")
	}
	return width, nil
}

func trainingStatistics(ctx context.Context, examples []Example, width int) (trainingData, error) {
	data := trainingData{means: make([]float64, width), scales: make([]float64, width)}
	digest := sha256.New()
	_, _ = digest.Write([]byte("unswell-logistic-input-v1\x00"))
	hashInteger(digest, uint64(len(examples)))
	hashInteger(digest, uint64(len(data.means)))
	for i, example := range examples {
		if err := ctx.Err(); err != nil {
			return trainingData{}, err
		}
		if err := validateExample(example, width); err != nil {
			return trainingData{}, fmt.Errorf("training row %d: %w", i, err)
		}
		data.positive += example.Label
		labelBits := uint64(0)
		if example.Label == 1 {
			labelBits = 1
		}
		hashInteger(digest, labelBits)
		for j, value := range example.Values {
			hashInteger(digest, math.Float64bits(value))
			delta := value - data.means[j]
			data.means[j] += delta / float64(i+1)
			data.scales[j] += delta * (value - data.means[j])
		}
	}
	data.inputHash = fmt.Sprintf("%x", digest.Sum(nil))
	if err := checkVariance(ctx, examples, data.scales); err != nil {
		return trainingData{}, err
	}
	return data, nil
}

func checkVariance(ctx context.Context, examples []Example, sums []float64) error {
	for j, sum := range sums {
		if sum != 0 {
			continue
		}
		for _, row := range examples {
			if err := ctx.Err(); err != nil {
				return err
			}
			if row.Values[j] != examples[0].Values[j] {
				return fmt.Errorf("%w: nonconstant feature variance underflow", ErrNumerical)
			}
		}
	}
	return nil
}

func validateExample(example Example, width int) error {
	if len(example.Values) != width || (example.Label != 0 && example.Label != 1) {
		return fmt.Errorf("expected a matching vector and a binary 0/1 label")
	}
	for _, value := range example.Values {
		if !finite(value) || math.Abs(value) > 1e12 {
			return fmt.Errorf("training values must be finite and within -1e12..1e12; missing values are not imputed")
		}
	}
	return nil
}

func (d *trainingData) normalize(ctx context.Context, examples []Example) error {
	for i, sum := range d.scales {
		if !finite(sum) || sum < 0 {
			return fmt.Errorf("%w: invalid normalization variance", ErrNumerical)
		}
		d.scales[i] = math.Sqrt(sum) / math.Sqrt(float64(len(examples)))
		if d.scales[i] == 0 {
			d.scales[i] = 1
		}
	}
	width := len(d.means)
	storage := make([]float64, len(examples)*width)
	d.rows, d.labels = make([][]float64, len(examples)), make([]float64, len(examples))
	for i, example := range examples {
		if err := ctx.Err(); err != nil {
			return err
		}
		d.rows[i], d.labels[i] = storage[i*width:(i+1)*width], float64(example.Label)
		for j, value := range example.Values {
			d.rows[i][j] = (value - d.means[j]) / d.scales[j]
			if !finite(d.rows[i][j]) {
				return fmt.Errorf("%w: normalization overflow", ErrNumerical)
			}
		}
	}
	return nil
}

func hashInteger(digest hash.Hash, value uint64) {
	var bytes [8]byte
	binary.LittleEndian.PutUint64(bytes[:], value)
	_, _ = digest.Write(bytes[:]) // SHA-256 writes never fail.
}
