package builtin

import (
	"slices"

	"github.com/stokaro/unswell/document"
)

func (m *editorialMatcher) observeNotOnlyStart(sentence document.Sentence, start int) {
	if m.windowObservations == nil || start+3 > len(sentence.Tokens) {
		return
	}
	// The smallest complete pattern is "not only but"; semicolons reset its start.
	for _, token := range sentence.Tokens[start : start+3] {
		if token.Normal == ";" {
			return
		}
	}
	m.observeWindowCandidate(sentence.BlockID)
}

func (m *editorialMatcher) observeContrastPair(first, second document.Sentence, a, b editorialPrefix) {
	if m.windowObservations != nil && m.view.Parameters.WindowSentences >= 2 && a.eligible && b.eligible && !question(second) {
		m.observeWindowCandidate(first.BlockID)
	}
}

func (m *editorialMatcher) observeWhetherStart(sentence document.Sentence, phrases ...string) {
	if m.windowObservations == nil {
		return
	}
	for _, phrase := range phrases {
		start := len(m.fixedPattern(phrase))
		for end := start; end < min(len(sentence.Tokens), 32); end++ {
			if sentence.Tokens[end].Normal != "," {
				continue
			}
			if end > start && !m.view.Exempts(sentence, 0, end+1) {
				m.observeWindowCandidate(sentence.BlockID)
			}
			break
		}
	}
}

func (m *editorialMatcher) observeQuestionPair(first, answer document.Sentence) {
	if m.windowObservations == nil || m.view.Parameters.WindowSentences < 2 {
		return
	}
	if slices.Contains(m.windowObservations.lengths, len(first.Tokens)) &&
		!m.view.Exempts(first, 0, len(first.Tokens)) && !m.view.Exempts(answer, 0, len(answer.Tokens)) {
		m.observeWindowCandidate(first.BlockID)
	}
}
