package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func editorialClaims(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	m.collectPhraseObservations()
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) {
			continue
		}
		for _, sentence := range block.Sentences {
			if err := m.claims(sentence, emit); err != nil {
				return err
			}
		}
	}
	return m.observePhraseBlocks()
}

func (m *editorialMatcher) claims(sentence document.Sentence, emit rule.Emitter) error {
	if err := m.ctx.Err(); err != nil {
		return err
	}
	if qualifiedClaim(sentence) {
		return nil
	}
	matches, err := m.phrases(sentence, false)
	if err != nil {
		return err
	}
	for _, occurrence := range matchOccurrences(sentence, matches) {
		if err := emit.Emit(measured("heuristic", "unqualified-phrase", "matches", 1, 0, 1, []rule.Occurrence{occurrence})); err != nil {
			return err
		}
	}
	return nil
}

func stackedHedging(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if err := m.hedgeBlock(block, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (m *editorialMatcher) hedgeBlock(block document.Block, emit rule.Emitter) error {
	if err := m.ctx.Err(); err != nil {
		return err
	}
	if !proseBlock(block) {
		return observeBlock(m.view, block, "unsupported_unit")
	}
	evaluated := false
	for _, sentence := range block.Sentences {
		compared, err := m.hedges(sentence, emit)
		if err != nil {
			return err
		}
		evaluated = evaluated || compared
	}
	return observeBlock(m.view, block, tokenBlockReason(block, evaluated, len(m.words) > 0))
}

func (m *editorialMatcher) hedges(sentence document.Sentence, emit rule.Emitter) (bool, error) {
	var occurrences []rule.Occurrence
	evaluated := false
	seen := make(map[string]bool)
	flush := func() error {
		p := m.view.Parameters
		err := emitHedges(p, occurrences, emit)
		occurrences = nil
		clear(seen)
		return err
	}
	for i, token := range sentence.Tokens {
		if err := m.ctx.Err(); err != nil {
			return evaluated, err
		}
		if clauseBoundary(token) {
			if err := flush(); err != nil {
				return evaluated, err
			}
			continue
		}
		evaluated = evaluated || m.view.Observer != nil && !m.view.Exempts(sentence, i, i+1)
		if seen[token.Normal] || !m.isHedge(sentence, i) {
			continue
		}
		if err := m.spend(); err != nil {
			return evaluated, err
		}
		seen[token.Normal] = true
		occurrences = append(occurrences, tokenOccurrence(sentence, i, i+1))
	}
	return evaluated, flush()
}

func (m *editorialMatcher) isHedge(sentence document.Sentence, i int) bool {
	token := sentence.Tokens[i]
	return m.words[token.Normal] && !m.view.Exempts(sentence, i, i+1) && (token.Tag == "MD" || strings.HasPrefix(token.Tag, "RB"))
}

func emitHedges(p rule.Parameters, occurrences []rule.Occurrence, emit rule.Emitter) error {
	if len(occurrences) <= p.Onset {
		return nil
	}
	return emit.Emit(measured("heuristic", "distinct-hedges", "cues", len(occurrences), p.Onset, p.Saturation, occurrences))
}

func clauseBoundary(token document.Token) bool {
	return token.Protected || !token.Word || token.Tag == "CC" ||
		slices.Contains([]string{"if", "unless", "because", "although"}, token.Normal)
}
