package builtin

import (
	"context"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type repetitionGroupObservation struct {
	minimumReached, evaluated bool
}

func (o repetitionGroupObservation) observe(view rule.View, block document.Block) error {
	return observeBlock(view, block, o.reason(block))
}

func (o repetitionGroupObservation) reason(block document.Block) string {
	switch {
	case len(block.Sentences) == 0:
		return "no_sentences"
	case !o.minimumReached:
		return "insufficient_words"
	case !o.evaluated:
		return "no_eligible_tokens"
	default:
		return ""
	}
}

func addExactBlock(ctx context.Context, view rule.View, block document.Block, groups map[string][]rule.Occurrence) error {
	observation := repetitionGroupObservation{}
	for _, sentence := range block.Sentences {
		if err := ctx.Err(); err != nil {
			return err
		}
		if sentence.Words < view.Parameters.MinWords {
			continue
		}
		observation.minimumReached = true
		key := sentenceKey(sentence)
		if key != "" {
			groups[key] = append(groups[key], sentenceOccurrence(sentence))
			observation.evaluated = true
		}
	}
	return observation.observe(view, block)
}

func addOpenerBlock(ctx context.Context, view rule.View, block document.Block, groups map[string][]rule.Occurrence, firstOnly bool) error {
	observation := repetitionGroupObservation{}
	for i, sentence := range block.Sentences {
		if firstOnly && i > 0 {
			break
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		words := normalizedWords(sentence)
		if sentence.Words < view.Parameters.MinWords {
			continue
		}
		observation.minimumReached = true
		if len(words) < view.Parameters.OpenerWords {
			continue
		}
		key := strings.Join(words[:view.Parameters.OpenerWords], " ")
		groups[key] = append(groups[key], sentenceOccurrence(sentence))
		observation.evaluated = true
	}
	return observation.observe(view, block)
}
