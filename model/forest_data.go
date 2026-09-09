package model

import (
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"math"
)

type forestData struct {
	rows   [][]float64
	labels []int
	width  int
	hash   string
}

func prepareForest(ctx context.Context, examples []Example, budget *workBudget) (forestData, error) {
	if err := ctx.Err(); err != nil {
		return forestData{}, err
	}
	width, err := trainingWidth(examples)
	if err != nil {
		return forestData{}, err
	}
	if err := budget.charge(int64(len(examples)) * int64(width+1)); err != nil {
		return forestData{}, err
	}
	data := forestData{rows: make([][]float64, len(examples)), labels: make([]int, len(examples)), width: width}
	storage := make([]float64, len(examples)*width)
	digest := sha256.New()
	_, _ = digest.Write([]byte("unswell-forest-input-v1\x00"))
	hashInteger(digest, uint64(len(examples)))
	hashInteger(digest, uint64(len(examples[0].Values)))
	positive := 0
	for i, example := range examples {
		if err := ctx.Err(); err != nil {
			return forestData{}, err
		}
		if err := validateExample(example, width); err != nil {
			return forestData{}, fmt.Errorf("forest training row %d: %w", i, err)
		}
		data.rows[i], data.labels[i] = storage[i*width:(i+1)*width], example.Label
		copy(data.rows[i], example.Values)
		positive += example.Label
		hashForestRow(digest, example)
	}
	if positive == 0 || positive == len(examples) {
		return forestData{}, fmt.Errorf("forest training requires both binary classes")
	}
	data.hash = fmt.Sprintf("%x", digest.Sum(nil))
	return data, nil
}

func hashForestRow(digest hash.Hash, example Example) {
	label := uint64(0)
	if example.Label == 1 {
		label = 1
	}
	hashInteger(digest, label)
	for _, value := range example.Values {
		hashInteger(digest, math.Float64bits(value))
	}
}
