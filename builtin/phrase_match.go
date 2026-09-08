package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type phraseMatcher struct {
	view     rule.View
	emit     rule.Emitter
	patterns [][]string
	index    int
	total    int
}

func phraseEvaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := phraseMatcher{view: view, emit: emit}
	for _, phrase := range view.Parameters.Phrases {
		if pattern := phraseTokens(phrase); len(pattern) > 0 {
			m.patterns = append(m.patterns, pattern)
		}
	}
	for _, block := range view.Document.Blocks {
		m.total += len(block.Sentences)
	}
	for _, block := range view.Document.Blocks {
		if err := m.block(ctx, block); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (m *phraseMatcher) block(ctx context.Context, block document.Block) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	evaluated := false
	for _, sentence := range block.Sentences {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, pattern := range m.patterns {
			observed, err := m.match(sentence, pattern)
			if err != nil {
				return err
			}
			evaluated = evaluated || observed
		}
		m.index++
	}
	return m.view.Observe(phraseObservation(block, len(m.patterns), evaluated))
}

func phraseObservation(block document.Block, patterns int, evaluated bool) feature.BlockObservation {
	result := feature.BlockObservation{BlockID: block.ID, Status: "inapplicable", Reason: "no_eligible_window"}
	switch {
	case evaluated:
		result.Status, result.Reason = "evaluated", ""
	case patterns == 0:
		result.Reason = "no_patterns"
	case len(block.Sentences) == 0:
		result.Reason = "no_sentences"
	}
	return result
}

func (m *phraseMatcher) match(sentence document.Sentence, pattern []string) (bool, error) {
	evaluated := false
	for start := 0; start+len(pattern) <= len(sentence.Tokens); start++ {
		end := start + len(pattern)
		if !m.candidate(sentence, start, end) {
			continue
		}
		evaluated = true
		if !matches(sentence.Tokens[start:end], pattern) {
			continue
		}
		if err := m.emit.Emit(phraseEvidence(sentence, start, end)); err != nil {
			return evaluated, err
		}
	}
	return evaluated, nil
}

func (m *phraseMatcher) candidate(sentence document.Sentence, start, end int) bool {
	return phrasePosition(m.view.Parameters.Positions, start, m.index, m.total) &&
		!m.view.Exempts(sentence, start, end) &&
		!slices.ContainsFunc(sentence.Tokens[start:end], func(token document.Token) bool { return token.Protected })
}

func phraseEvidence(sentence document.Sentence, start, end int) rule.Evidence {
	return rule.Evidence{Kind: "exact", Occurrences: []rule.Occurrence{tokenOccurrence(sentence, start, end)}, Activation: 1000,
		Metrics: []rule.Metric{{Name: "phrase-match", Value: 1, Unit: "matches", Onset: 0, Saturation: 1}}}
}
