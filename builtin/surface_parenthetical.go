package builtin

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type wordPositions struct {
	starts []int
	counts []int
}

func nonexemptWordPositions(m *editorialMatcher, block document.Block) (wordPositions, error) {
	index := wordPositions{counts: []int{0}}
	for _, sentence := range block.Sentences {
		for i, token := range sentence.Tokens {
			if err := m.spend(); err != nil {
				return index, err
			}
			if !token.Word || token.Protected {
				continue
			}
			count := 0
			if !m.view.Exempts(sentence, i, i+1) {
				count = 1
			}
			index.starts = append(index.starts, token.Start)
			index.counts = append(index.counts, index.counts[len(index.counts)-1]+count)
		}
	}
	return index, nil
}

func (w wordPositions) between(start, end int) int {
	left, _ := slices.BinarySearch(w.starts, start)
	right, _ := slices.BinarySearch(w.starts, end)
	return w.counts[right] - w.counts[left]
}

type insertionFrame struct {
	start, pairs, depth int
	opening             rune
}

type insertionLoad struct {
	words, pairs, depth int
	ranges              []document.Span
}

func parentheticalLoad(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) || block.Words < view.Parameters.MinWords {
			continue
		}
		index, err := nonexemptWordPositions(m, block)
		if err != nil {
			return err
		}
		load, err := balancedInsertions(m, block.Text, index)
		if err != nil {
			return err
		}
		if err := emitInsertionLoad(view.Parameters, block, load, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func balancedInsertions(m *editorialMatcher, text string, words wordPositions) (insertionLoad, error) {
	var stack []insertionFrame
	var load insertionLoad
	for at, r := range text {
		if err := m.spend(); err != nil {
			return load, err
		}
		switch r {
		case 0:
			stack = nil
		case '(', '[':
			if len(stack) >= 256 {
				return load, fmt.Errorf("parenthetical nesting exceeds 256 levels")
			}
			stack = append(stack, insertionFrame{start: at, opening: r})
		case ')', ']':
			stack = closeInsertion(stack, &load, words, at, r)
		}
	}
	return load, nil
}

func closeInsertion(stack []insertionFrame, load *insertionLoad, words wordPositions, at int, closing rune) []insertionFrame {
	if len(stack) == 0 {
		return stack
	}
	frame := stack[len(stack)-1]
	if frame.opening == '(' && closing != ')' || frame.opening == '[' && closing != ']' {
		return nil
	}
	stack = stack[:len(stack)-1]
	count := words.between(frame.start, at)
	if count == 0 {
		return stack
	}
	frame.pairs++
	frame.depth++
	if len(stack) > 0 {
		parent := &stack[len(stack)-1]
		parent.pairs += frame.pairs
		parent.depth = max(parent.depth, frame.depth)
		return stack
	}
	load.words += count
	load.pairs += frame.pairs
	load.depth = max(load.depth, frame.depth)
	load.ranges = append(load.ranges, document.Span{Start: frame.start, End: at + 1})
	return stack
}

func emitInsertionLoad(p rule.Parameters, block document.Block, load insertionLoad, emit rule.Emitter) error {
	level := max(activation(load.words, p.Onset, p.Saturation),
		activation(load.pairs, p.AllowedOccurrences, p.SaturationOccurrences), activation(load.depth, p.AllowedDepth, p.SaturationDepth))
	if level == 0 {
		return nil
	}
	var occurrences []rule.Occurrence
	mapping := newMappedRangeIndex(block)
	for _, span := range load.ranges {
		occurrences = append(occurrences, mapping.occurrences(span.Start, span.End)...)
	}
	return emit.Emit(rule.Evidence{Kind: "heuristic", Activation: level, Occurrences: occurrences,
		Metrics: []rule.Metric{
			{Name: "parenthetical-words", Value: float64(load.words), Unit: "nonexempt-prose-words",
				Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
			{Name: "balanced-insertions", Value: float64(load.pairs), Unit: "pairs",
				Onset: float64(p.AllowedOccurrences), Saturation: float64(p.SaturationOccurrences)},
			{Name: "maximum-insertion-depth", Value: float64(load.depth), Unit: "levels",
				Onset: float64(p.AllowedDepth), Saturation: float64(p.SaturationDepth)},
		}})
}
