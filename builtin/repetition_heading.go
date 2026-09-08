package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func headingEcho(ctx context.Context, view rule.View, emit rule.Emitter) error {
	budget := repetitionBudget{ctx, view.MaxCandidates}
	gaps := newProseGapIndex(view.Document.Excluded)
	var previous *document.Block
	for i := range view.Document.Blocks {
		block := &view.Document.Blocks[i]
		blocked := gaps.between(previous, block)
		if previous != nil && previous.Kind == "heading" && block.Kind == "paragraph" && !blocked &&
			adjacentProse(view.Document, previous, block, true) {
			if err := compareHeading(view, *previous, *block, &budget, emit); err != nil {
				return err
			}
		}
		previous = block
		if err := budget.spend(1); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func compareHeading(view rule.View, heading, paragraph document.Block, budget *repetitionBudget, emit rule.Emitter) error {
	left, ok, err := makeLexicalUnit(view, heading, 0, budget)
	if err != nil || !ok {
		return err
	}
	right, ok, err := makeLexicalUnit(view, paragraph, 1, budget)
	if err != nil || !ok {
		return err
	}
	if err := budget.spend(len(left.words) + len(right.words)); err != nil {
		return err
	}
	if left.signature != right.signature {
		return nil
	}
	shared, union := wordOverlap(left.words, right.words)
	value := float64(shared) / float64(union)
	if value < view.Parameters.Similarity {
		return nil
	}
	occurrences := append(unitOccurrences(right), unitOccurrences(left)...)
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: 1000, Occurrences: occurrences,
		Metrics: []rule.Metric{
			{Name: "heading-paragraph-jaccard", Value: value, Unit: "ratio", Onset: view.Parameters.Similarity, Saturation: 1},
			{Name: "shared-prose-words", Value: float64(shared), Unit: "distinct-words"},
			{Name: "union-prose-words", Value: float64(union), Unit: "distinct-words"},
		}})
}
