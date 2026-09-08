package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type windowPhraseObservations struct {
	lengths   []int
	evaluated map[int]bool
}

func windowPhrases(opening, sectionOpenings bool) func(context.Context, rule.View, rule.Emitter) error {
	return func(ctx context.Context, view rule.View, emit rule.Emitter) error {
		matcher := newEditorialMatcher(ctx, view)
		if view.Observer != nil {
			matcher.observations = matcher.windowObservations()
		}
		if err := matcher.evaluateWindows(phraseEvents(opening), "phrase-patterns", sectionOpenings, emit); err != nil {
			return err
		}
		return matcher.observeWindowBlocks()
	}
}

func (m *editorialMatcher) windowObservations() *windowPhraseObservations {
	result := &windowPhraseObservations{evaluated: make(map[int]bool)}
	for _, patterns := range m.patterns {
		for _, pattern := range patterns {
			result.lengths = append(result.lengths, len(pattern))
		}
	}
	slices.Sort(result.lengths)
	result.lengths = slices.Compact(result.lengths)
	return result
}

func (m *editorialMatcher) observePhraseStart(sentence document.Sentence, start int) error {
	if m.observations == nil || m.observations.evaluated[sentence.BlockID] {
		return nil
	}
	// The window traversal supplies only sentences without protected tokens.
	// A failed first-token lookup still evaluates an eligible dictionary window.
	for _, length := range m.observations.lengths {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		end := start + length
		if end > len(sentence.Tokens) {
			break
		}
		if !m.view.Exempts(sentence, start, end) {
			m.observations.evaluated[sentence.BlockID] = true
			break
		}
	}
	return nil
}

func (m *editorialMatcher) observeWindowBlocks() error {
	if m.observations == nil {
		return m.ctx.Err()
	}
	for _, block := range m.view.Document.Blocks {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		observation := phraseObservation(block, len(m.observations.lengths), m.observations.evaluated[block.ID])
		if !proseBlock(block) {
			observation.Reason = "unsupported_unit"
		}
		if err := m.view.Observe(observation); err != nil {
			return err
		}
	}
	return m.ctx.Err()
}
