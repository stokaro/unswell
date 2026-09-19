package reviewbaseline

import (
	"context"
	"slices"
	"strings"
)

type termCount struct {
	positive, negative int
	groups             map[string]bool
}

type rankedTerm struct {
	key   string
	score float64
}

func trainingRow(row row, fold int) bool {
	return row.label != nil && row.fold != fold && row.fold != (fold+1)%5
}

func vocabulary(ctx context.Context, rows []row, fold, limit int) ([]string, error) {
	counts := make(map[string]*termCount)
	positive, negative := 0, 0
	for _, row := range rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !trainingRow(row, fold) {
			continue
		}
		if *row.label == 1 {
			positive++
		} else {
			negative++
		}
		addTerms(counts, row)
	}
	terms := rankTerms(counts, positive, negative)
	result := make([]string, 0, min(limit, len(terms)))
	for _, term := range terms[:min(limit, len(terms))] {
		result = append(result, term.key)
	}
	// Selection rank does not determine the model's column order.
	slices.Sort(result)
	return result, nil
}

func addTerms(counts map[string]*termCount, row row) {
	for key := range row.terms {
		value := counts[key]
		if value == nil {
			value = &termCount{groups: make(map[string]bool)}
			counts[key] = value
		}
		if *row.label == 1 {
			value.positive++
		} else {
			value.negative++
		}
		value.groups[row.group] = true
	}
}

func rankTerms(counts map[string]*termCount, positive, negative int) []rankedTerm {
	var terms []rankedTerm
	for key, count := range counts {
		if len(count.groups) < 3 {
			continue
		}
		a, b := float64(count.positive), float64(count.negative)
		c, d := float64(positive-count.positive), float64(negative-count.negative)
		denominator := (a + b) * (c + d) * (a + c) * (b + d)
		if denominator <= 0 {
			continue
		}
		delta := a*d - b*c
		terms = append(terms, rankedTerm{key, float64(positive+negative) * delta * delta / denominator})
	}
	slices.SortFunc(terms, func(a, b rankedTerm) int {
		if a.score > b.score {
			return -1
		}
		if a.score < b.score {
			return 1
		}
		return strings.Compare(a.key, b.key)
	})
	return terms
}
