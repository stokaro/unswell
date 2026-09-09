package model

import (
	"cmp"
	"math"
	"math/bits"
	"slices"
)

type forestSplit struct {
	feature, left   int
	threshold, gain float64
}

func (b *forestBuilder) bestSplit(rows []int, positive int) (forestSplit, error) {
	if err := b.budget.charge(int64(b.data.width)); err != nil {
		return forestSplit{}, err
	}
	features := b.subset()
	ordered := make([]int, len(rows))
	best := forestSplit{feature: -1, gain: -1}
	for _, feature := range features {
		if err := b.ctx.Err(); err != nil {
			return forestSplit{}, err
		}
		// Charge a fixed conservative sorting/scan cost, independent of sort internals.
		cost := int64(len(rows)) * int64(4*bits.Len(uint(len(rows)))+4)
		if err := b.budget.charge(cost); err != nil {
			return forestSplit{}, err
		}
		copy(ordered, rows)
		slices.SortFunc(ordered, func(a, c int) int { return cmp.Compare(b.data.rows[a][feature], b.data.rows[c][feature]) })
		if err := b.ctx.Err(); err != nil {
			return forestSplit{}, err
		}
		best = b.scanSplits(ordered, feature, positive, best)
	}
	return best, nil
}

func (b *forestBuilder) subset() []int {
	features := make([]int, b.data.width)
	for i := range features {
		features[i] = i
	}
	count := b.options.FeaturesPerSplit
	if count == 0 {
		count = int(math.Sqrt(float64(b.data.width)))
	}
	if count == b.data.width {
		return features
	}
	for i := range count {
		other := i + b.random.IntN(b.data.width-i)
		features[i], features[other] = features[other], features[i]
	}
	features = features[:count]
	slices.Sort(features)
	return features
}

func (b *forestBuilder) scanSplits(rows []int, feature, positive int, best forestSplit) forestSplit {
	leftPositive := 0
	for i := 0; i+1 < len(rows); i++ {
		leftPositive += b.data.labels[rows[i]]
		left, right := i+1, len(rows)-i-1
		a, next := b.data.rows[rows[i]][feature], b.data.rows[rows[i+1]][feature]
		if left < b.options.MinLeaf || right < b.options.MinLeaf || a == next {
			continue
		}
		gain := binaryGiniGain(left, right, leftPositive, positive-leftPositive)
		if gain <= best.gain {
			continue
		}
		threshold := a + (next-a)/2
		if threshold >= next {
			threshold = a
		}
		best = forestSplit{feature: feature, left: left, threshold: threshold, gain: gain}
	}
	return best
}

func binaryGiniGain(left, right, leftPositive, rightPositive int) float64 {
	l, r := float64(left), float64(right)
	difference := float64(leftPositive)*r - float64(rightPositive)*l
	return 2 * difference * difference / ((l + r) * (l + r) * l * r)
}
