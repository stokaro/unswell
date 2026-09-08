package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type tokenMatch struct{ start, end int }

type editorialMatcher struct {
	ctx          context.Context
	view         rule.View
	patterns     map[string][][]string
	fixed        map[string][]string
	words        map[string]bool
	checks       int
	observations *editorialPhraseObservations
}

func newEditorialMatcher(ctx context.Context, view rule.View) *editorialMatcher {
	m := &editorialMatcher{ctx: ctx, view: view, patterns: make(map[string][][]string),
		fixed: make(map[string][]string), words: make(map[string]bool)}
	for _, phrase := range view.Parameters.Phrases {
		parts := phraseTokens(phrase)
		if len(parts) > 0 {
			m.patterns[parts[0]] = append(m.patterns[parts[0]], parts)
			m.words[document.Normalize(phrase)] = true
		}
	}
	for _, patterns := range m.patterns {
		slices.SortStableFunc(patterns, func(a, b []string) int { return len(b) - len(a) })
	}
	return m
}

func (m *editorialMatcher) spend() error {
	if err := m.ctx.Err(); err != nil {
		return err
	}
	m.checks++
	if m.checks > m.view.MaxCandidates {
		return fmt.Errorf("editorial pattern checks exceed max_candidates")
	}
	return nil
}

func (m *editorialMatcher) phrases(sentence document.Sentence, opening bool) ([]tokenMatch, error) {
	var result []tokenMatch
	for i := 0; i < len(sentence.Tokens); i++ {
		if err := m.ctx.Err(); err != nil {
			return nil, err
		}
		if opening && i > 0 {
			break
		}
		if err := m.observePhraseStart(sentence, i); err != nil {
			return nil, err
		}
		end, err := m.phraseAt(sentence, i)
		if err != nil {
			return nil, err
		}
		if end > i {
			result = append(result, tokenMatch{i, end})
			i = end - 1
		}
	}
	return result, nil
}

func (m *editorialMatcher) phraseAt(sentence document.Sentence, start int) (int, error) {
	for _, pattern := range m.patterns[sentence.Tokens[start].Normal] {
		if err := m.spend(); err != nil {
			return 0, err
		}
		end := start + len(pattern)
		if end <= len(sentence.Tokens) && matches(sentence.Tokens[start:end], pattern) && !m.view.Exempts(sentence, start, end) {
			return end, nil
		}
	}
	return start, nil
}

func proseBlock(block document.Block) bool {
	return block.Kind == "paragraph" || block.Kind == "comment" || block.Kind == "string"
}

func protectedSentence(sentence document.Sentence) bool {
	return slices.ContainsFunc(sentence.Tokens, func(token document.Token) bool { return token.Protected })
}

func question(sentence document.Sentence) bool {
	return strings.HasSuffix(strings.TrimSpace(sentence.Text), "?")
}

func hasWords(sentence document.Sentence, words []string) bool {
	return slices.ContainsFunc(sentence.Tokens, func(token document.Token) bool { return slices.Contains(words, token.Normal) })
}

func qualifiedClaim(sentence document.Sentence) bool {
	return question(sentence) || hasWords(sentence, []string{
		"not", "n't", "no", "cannot", "if", "unless", "except", "when", "provided", "assuming", "within", "under",
		"may", "might", "could", "can", "would", "should",
	})
}

func matchOccurrences(sentence document.Sentence, matches []tokenMatch) []rule.Occurrence {
	result := make([]rule.Occurrence, 0, len(matches))
	for _, match := range matches {
		result = append(result, tokenOccurrence(sentence, match.start, match.end))
	}
	return result
}
