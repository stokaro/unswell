package builtin

import (
	"context"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func longSentence(ctx context.Context, view rule.View, emit rule.Emitter) error {
	for _, block := range view.Document.Blocks {
		evaluated, err := sentenceLengths(ctx, view, block, emit)
		if err != nil {
			return err
		}
		reason := ""
		if !evaluated {
			reason = "no_prose_words"
		}
		if err := observeBlock(view, block, reason); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func sentenceLengths(ctx context.Context, view rule.View, block document.Block, emit rule.Emitter) (bool, error) {
	evaluated := false
	for _, sentence := range block.Sentences {
		if err := ctx.Err(); err != nil {
			return evaluated, err
		}
		evaluated = evaluated || sentence.Words > 0
		if sentence.Words <= view.Parameters.Onset {
			continue
		}
		evidence := measured("exact", "length", "prose-words", sentence.Words,
			view.Parameters.Onset, view.Parameters.Saturation, []rule.Occurrence{sentenceOccurrence(sentence)})
		if err := emit.Emit(evidence); err != nil {
			return evaluated, err
		}
	}
	return evaluated, nil
}

func modifierCluster(ctx context.Context, view rule.View, emit rule.Emitter) error {
	lexicon := make(map[string]bool)
	for _, word := range view.Parameters.Phrases {
		lexicon[word] = true
	}
	for _, block := range view.Document.Blocks {
		evaluated, err := modifierBlock(ctx, view, block, lexicon, emit)
		if err != nil {
			return err
		}
		if err := observeBlock(view, block, tokenBlockReason(block, evaluated, len(lexicon) > 0)); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func modifierBlock(ctx context.Context, view rule.View, block document.Block, lexicon map[string]bool, emit rule.Emitter) (bool, error) {
	evaluated := false
	for _, sentence := range block.Sentences {
		if err := ctx.Err(); err != nil {
			return evaluated, err
		}
		count, occurrence, compared := modifierOccurrence(sentence, lexicon, view)
		evaluated = evaluated || compared
		if count <= view.Parameters.Onset {
			continue
		}
		evidence := measured("heuristic", "evaluative-modifiers", "tokens", count,
			view.Parameters.Onset, view.Parameters.Saturation, []rule.Occurrence{occurrence})
		if err := emit.Emit(evidence); err != nil {
			return evaluated, err
		}
	}
	return evaluated, nil
}

func modifierOccurrence(sentence document.Sentence, lexicon map[string]bool, view rule.View) (int, rule.Occurrence, bool) {
	count, evaluated := 0, false
	occurrence := rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: []document.Span{}}
	for i, token := range sentence.Tokens {
		if token.Protected || view.Exempts(sentence, i, i+1) {
			continue
		}
		evaluated = evaluated || token.Word
		if lexicon[token.Normal] && strings.HasPrefix(token.Tag, "JJ") {
			evaluated = true
			count++
			occurrence.Spans = append(occurrence.Spans, token.Spans...)
		}
	}
	return count, occurrence, evaluated
}

func connectiveOveruse(ctx context.Context, view rule.View, emit rule.Emitter) error {
	lexicon := make(map[string]bool)
	for _, phrase := range view.Parameters.Phrases {
		lexicon[phrase] = true
	}
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := connectiveBlock(view, block, lexicon, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func connectiveBlock(view rule.View, block document.Block, lexicon map[string]bool, emit rule.Emitter) error {
	if block.Words < view.Parameters.MinWords {
		return observeBlock(view, block, "insufficient_words")
	}
	occurrences, evaluated := transitionOccurrences(block, lexicon, view)
	if err := observeBlock(view, block, tokenBlockReason(block, evaluated, len(lexicon) > 0)); err != nil {
		return err
	}
	if len(occurrences) <= view.Parameters.Onset {
		return nil
	}
	return emit.Emit(measured("heuristic", "sentence-transitions", "occurrences", len(occurrences),
		view.Parameters.Onset, view.Parameters.Saturation, occurrences))
}

func transitionOccurrences(block document.Block, lexicon map[string]bool, view rule.View) ([]rule.Occurrence, bool) {
	occurrences := make([]rule.Occurrence, 0)
	evaluated := false
	for _, sentence := range block.Sentences {
		if len(sentence.Tokens) == 0 || sentence.Tokens[0].Protected || view.Exempts(sentence, 0, 1) {
			continue
		}
		evaluated = evaluated || sentence.Tokens[0].Word
		if lexicon[sentence.Tokens[0].Normal] {
			evaluated = true
			occurrences = append(occurrences, tokenOccurrence(sentence, 0, 1))
		}
	}
	return occurrences, evaluated
}
