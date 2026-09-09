package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func headingEcho(ctx context.Context, view rule.View, emit rule.Emitter) error {
	observations := newCandidateObservations(view, headingProseBlock)
	budget := repetitionBudget{ctx, view.MaxCandidates}
	gaps := newProseGapIndex(view.Document.Excluded)
	var previous *document.Block
	for i := range view.Document.Blocks {
		block := &view.Document.Blocks[i]
		blocked := gaps.between(previous, block)
		if previous != nil && previous.Kind == "heading" && block.Kind == "paragraph" && !blocked &&
			adjacentProse(view.Document, previous, block, true) {
			if err := compareHeading(view, *previous, *block, &budget, emit, observations); err != nil {
				return err
			}
		}
		previous = block
		if err := budget.spend(1); err != nil {
			return err
		}
	}
	return observations.finish(ctx, view)
}

func headingProseBlock(block document.Block) bool {
	return block.Kind == "heading" || block.Kind == "paragraph"
}

func compareHeading(view rule.View, heading, paragraph document.Block, budget *repetitionBudget, emit rule.Emitter,
	observations *candidateObservations) error {
	left, ok, err := makeLexicalUnit(view, heading, 0, budget, observations)
	if err != nil || !ok {
		return err
	}
	right, ok, err := makeLexicalUnit(view, paragraph, 1, budget, observations)
	if err != nil || !ok {
		return err
	}
	if err := budget.spend(left.words.Len() + right.words.Len()); err != nil {
		return err
	}
	observations.advance(heading.ID, candidateEvaluated)
	observations.advance(paragraph.ID, candidateEvaluated)
	if left.signature != right.signature {
		return nil
	}
	overlap, err := feature.CompareWords(budget.ctx, left.words, right.words)
	if err != nil {
		return err
	}
	shared, union := overlap.Intersection(), overlap.Union()
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
