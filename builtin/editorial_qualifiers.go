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
	return ctx.Err()
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

func weakIntensifiers(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) || block.Words < max(1, view.Parameters.MinWords) {
			continue
		}
		occurrences, err := m.intensifiers(block)
		if err != nil {
			return err
		}
		rate := float64(len(occurrences)) * 100 / float64(block.Words)
		p := view.Parameters
		if rate <= float64(p.Onset) {
			continue
		}
		level := min(1000, int((rate-float64(p.Onset))*1000/float64(p.Saturation-p.Onset)))
		evidence := rule.Evidence{Kind: "heuristic", Activation: level, Occurrences: occurrences,
			Metrics: []rule.Metric{
				{Name: "intensifier-density", Value: rate, Unit: "matches/100-prose-words", Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
				{Name: "intensifiers", Value: float64(len(occurrences)), Unit: "tokens"},
				{Name: "prose-length", Value: float64(block.Words), Unit: "prose-words"},
			}}
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (m *editorialMatcher) intensifiers(block document.Block) ([]rule.Occurrence, error) {
	var occurrences []rule.Occurrence
	for _, sentence := range block.Sentences {
		for i := 0; i+1 < len(sentence.Tokens); i++ {
			if err := m.ctx.Err(); err != nil {
				return nil, err
			}
			if !m.isIntensifier(sentence, i) {
				continue
			}
			if err := m.spend(); err != nil {
				return nil, err
			}
			occurrences = append(occurrences, tokenOccurrence(sentence, i, i+1))
		}
	}
	return occurrences, nil
}

func (m *editorialMatcher) isIntensifier(sentence document.Sentence, i int) bool {
	token, next := sentence.Tokens[i], sentence.Tokens[i+1]
	if !m.words[token.Normal] || token.Protected || next.Protected || m.view.Exempts(sentence, i, i+1) {
		return false
	}
	if slices.Contains([]string{"first", "last", "same", "next", "previous"}, next.Normal) {
		return false
	}
	return strings.HasPrefix(token.Tag, "RB") && (strings.HasPrefix(next.Tag, "JJ") || strings.HasPrefix(next.Tag, "RB"))
}

func stackedHedging(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) {
			continue
		}
		for _, sentence := range block.Sentences {
			if err := m.hedges(sentence, emit); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

func (m *editorialMatcher) hedges(sentence document.Sentence, emit rule.Emitter) error {
	var occurrences []rule.Occurrence
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
			return err
		}
		if clauseBoundary(token) {
			if err := flush(); err != nil {
				return err
			}
			continue
		}
		if !seen[token.Normal] && m.isHedge(sentence, i) {
			if err := m.spend(); err != nil {
				return err
			}
			seen[token.Normal] = true
			occurrences = append(occurrences, tokenOccurrence(sentence, i, i+1))
		}
	}
	return flush()
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
