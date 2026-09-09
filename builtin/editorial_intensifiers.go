package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func weakIntensifiers(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if err := m.intensifierBlock(block, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (m *editorialMatcher) intensifierBlock(block document.Block, emit rule.Emitter) error {
	if err := m.ctx.Err(); err != nil {
		return err
	}
	if reason := proseMinimumReason(block, m.view.Parameters.MinWords); reason != "" {
		return observeBlock(m.view, block, reason)
	}
	if len(m.words) == 0 {
		return observeBlock(m.view, block, "no_patterns")
	}
	occurrences, evaluated, err := m.intensifiers(block)
	if err != nil {
		return err
	}
	if err := emitIntensifiers(m.view.Parameters, block.Words, occurrences, emit); err != nil {
		return err
	}
	return m.view.Observe(phraseObservation(block, len(m.words), evaluated))
}

func emitIntensifiers(p rule.Parameters, words int, occurrences []rule.Occurrence, emit rule.Emitter) error {
	rate := float64(len(occurrences)) * 100 / float64(words)
	if rate <= float64(p.Onset) {
		return nil
	}
	level := min(1000, int((rate-float64(p.Onset))*1000/float64(p.Saturation-p.Onset)))
	evidence := rule.Evidence{Kind: "heuristic", Activation: level, Occurrences: occurrences,
		Metrics: []rule.Metric{
			{Name: "intensifier-density", Value: rate, Unit: "matches/100-prose-words", Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
			{Name: "intensifiers", Value: float64(len(occurrences)), Unit: "tokens"},
			{Name: "prose-length", Value: float64(words), Unit: "prose-words"},
		}}
	return emit.Emit(evidence)
}

func (m *editorialMatcher) intensifiers(block document.Block) ([]rule.Occurrence, bool, error) {
	var occurrences []rule.Occurrence
	evaluated := false
	for _, sentence := range block.Sentences {
		for i := 0; i+1 < len(sentence.Tokens); i++ {
			if err := m.ctx.Err(); err != nil {
				return nil, evaluated, err
			}
			evaluated = evaluated || m.view.Observer != nil && m.intensifierCandidate(sentence, i)
			if !m.isIntensifier(sentence, i) {
				continue
			}
			if err := m.spend(); err != nil {
				return nil, evaluated, err
			}
			occurrences = append(occurrences, tokenOccurrence(sentence, i, i+1))
		}
	}
	return occurrences, evaluated, nil
}

func (m *editorialMatcher) intensifierCandidate(sentence document.Sentence, i int) bool {
	token, next := sentence.Tokens[i], sentence.Tokens[i+1]
	return !token.Protected && !next.Protected && !m.view.Exempts(sentence, i, i+1) &&
		!slices.Contains([]string{"first", "last", "same", "next", "previous"}, next.Normal)
}

func (m *editorialMatcher) isIntensifier(sentence document.Sentence, i int) bool {
	token, next := sentence.Tokens[i], sentence.Tokens[i+1]
	return m.words[token.Normal] && m.intensifierCandidate(sentence, i) &&
		strings.HasPrefix(token.Tag, "RB") && (strings.HasPrefix(next.Tag, "JJ") || strings.HasPrefix(next.Tag, "RB"))
}
