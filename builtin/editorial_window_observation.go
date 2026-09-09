package builtin

import (
	"slices"

	"github.com/stokaro/unswell/document"
)

type windowBlockObservation struct {
	visited, meetsMinimum, evaluated bool
}

type editorialWindowObservations struct {
	blocks           map[int]windowBlockObservation
	patternsRequired bool
	lengths          []int
}

func (m *editorialMatcher) collectWindowObservations(patternsRequired bool) {
	if m.view.Observer == nil {
		return
	}
	m.windowObservations = &editorialWindowObservations{
		blocks: make(map[int]windowBlockObservation), patternsRequired: patternsRequired,
	}
	for _, patterns := range m.patterns {
		for _, pattern := range patterns {
			m.windowObservations.lengths = append(m.windowObservations.lengths, len(pattern))
		}
	}
	slices.Sort(m.windowObservations.lengths)
	m.windowObservations.lengths = slices.Compact(m.windowObservations.lengths)
}

func (m *editorialMatcher) visitWindowSentence(sentence document.Sentence) {
	if m.windowObservations == nil {
		return
	}
	state := m.windowObservations.blocks[sentence.BlockID]
	state.visited = true
	state.meetsMinimum = state.meetsMinimum || sentence.Words >= m.view.Parameters.MinWords
	m.windowObservations.blocks[sentence.BlockID] = state
}

func (m *editorialMatcher) observeWindowCandidate(blockID int) {
	if m.windowObservations == nil {
		return
	}
	state := m.windowObservations.blocks[blockID]
	state.evaluated = true
	m.windowObservations.blocks[blockID] = state
}

func (m *editorialMatcher) observeWindowEvents(events []editorialEvent) {
	if m.windowObservations == nil {
		return
	}
	for _, event := range events {
		if event.last-event.first >= m.view.Parameters.WindowSentences {
			continue
		}
		for _, occurrence := range event.occurrences {
			m.observeWindowCandidate(occurrence.BlockID)
		}
	}
}

func (m *editorialMatcher) observeWindowWord(sentence document.Sentence, index int) {
	if m.windowObservations != nil && sentence.Tokens[index].Word && !m.view.Exempts(sentence, index, index+1) {
		m.observeWindowCandidate(sentence.BlockID)
	}
}

func (m *editorialMatcher) observeWindowBlocks() error {
	if m.windowObservations == nil {
		return m.ctx.Err()
	}
	for _, block := range m.view.Document.Blocks {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		if err := observeBlock(m.view, block, m.windowBlockReason(block)); err != nil {
			return err
		}
	}
	return m.ctx.Err()
}

func (m *editorialMatcher) windowBlockReason(block document.Block) string {
	state := m.windowObservations.blocks[block.ID]
	switch {
	case !proseBlock(block):
		return "unsupported_unit"
	case m.windowObservations.patternsRequired && len(m.patterns) == 0:
		return "no_patterns"
	case len(block.Sentences) == 0:
		return "no_sentences"
	case state.visited && !state.meetsMinimum:
		return "insufficient_words"
	case !state.evaluated:
		return "no_eligible_window"
	default:
		return ""
	}
}
