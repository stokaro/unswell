package model

import (
	"context"
	"fmt"
	"math/rand/v2"
)

type forestBuilder struct {
	ctx     context.Context
	data    forestData
	options ForestOptions
	budget  *workBudget
	random  *rand.Rand
	tree    []ForestNode
	nodes   int
}

func (b *forestBuilder) sample() ([]int, error) {
	if err := b.budget.charge(int64(len(b.data.rows))); err != nil {
		return nil, err
	}
	indices := make([]int, len(b.data.rows))
	for i := range indices {
		if err := b.ctx.Err(); err != nil {
			return nil, err
		}
		indices[i] = i
		if b.options.Bootstrap {
			indices[i] = b.random.IntN(len(indices))
		}
	}
	return indices, nil
}

func (b *forestBuilder) node(rows []int, depth int) (int, error) {
	if err := b.ctx.Err(); err != nil {
		return 0, err
	}
	if b.nodes >= b.options.MaxNodes {
		return 0, fmt.Errorf("%w: forest exceeds max_nodes", ErrBudget)
	}
	positive, err := b.positives(rows)
	if err != nil {
		return 0, err
	}
	index := len(b.tree)
	b.tree = append(b.tree, ForestNode{Feature: -1, Left: -1, Right: -1, Positive: positive, Samples: len(rows)})
	b.nodes++
	if b.terminal(len(rows), positive, depth) {
		return index, nil
	}
	split, err := b.bestSplit(rows, positive)
	if err != nil {
		return 0, err
	}
	if split.feature == -1 {
		return index, nil
	}
	if err := b.children(index, rows, depth, split); err != nil {
		return 0, err
	}
	return index, nil
}

func (b *forestBuilder) terminal(samples, positive, depth int) bool {
	return positive == 0 || positive == samples || depth == b.options.MaxDepth || samples < 2*b.options.MinLeaf
}

func (b *forestBuilder) positives(rows []int) (int, error) {
	if err := b.budget.charge(int64(len(rows))); err != nil {
		return 0, err
	}
	total := 0
	for _, row := range rows {
		if err := b.ctx.Err(); err != nil {
			return 0, err
		}
		total += b.data.labels[row]
	}
	return total, nil
}

func (b *forestBuilder) children(index int, rows []int, depth int, split forestSplit) error {
	left, right, err := b.partition(rows, split)
	if err != nil {
		return err
	}
	leftIndex, err := b.node(left, depth+1)
	if err != nil {
		return err
	}
	rightIndex, err := b.node(right, depth+1)
	if err != nil {
		return err
	}
	b.tree[index].Feature, b.tree[index].Threshold = split.feature, split.threshold
	b.tree[index].Left, b.tree[index].Right = leftIndex, rightIndex
	return nil
}

func (b *forestBuilder) partition(rows []int, split forestSplit) ([]int, []int, error) {
	if err := b.budget.charge(int64(len(rows))); err != nil {
		return nil, nil, err
	}
	storage := make([]int, len(rows))
	left, right := storage[:0:split.left], storage[split.left:split.left]
	for _, row := range rows {
		if err := b.ctx.Err(); err != nil {
			return nil, nil, err
		}
		if b.data.rows[row][split.feature] <= split.threshold {
			left = append(left, row)
		} else {
			right = append(right, row)
		}
	}
	return left, right, nil
}
