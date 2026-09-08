package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type repetitionBudget struct {
	ctx       context.Context
	remaining int
}

func (b *repetitionBudget) spend(work int) error {
	if err := b.ctx.Err(); err != nil {
		return err
	}
	if work > b.remaining {
		return fmt.Errorf("repetition work exceeds max_candidates")
	}
	b.remaining -= work
	return nil
}

type repetitionSentence struct {
	sentence document.Sentence
	ordinal  int
}

func repetitionSentences(view rule.View) []repetitionSentence {
	var result []repetitionSentence
	ordinal := 0
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) {
			continue
		}
		for _, sentence := range block.Sentences {
			if sentence.Words >= view.Parameters.MinWords && !protectedSentence(sentence) {
				result = append(result, repetitionSentence{sentence, ordinal})
			}
			ordinal++
		}
	}
	return result
}

func repetitionSignature(view rule.View, sentence document.Sentence) string {
	parts := []string{protectedSignature(sentence, view.Parameters)}
	for i, token := range sentence.Tokens {
		if contrastCue(token.Normal) || view.Exempts(sentence, i, i+1) {
			parts = append(parts, token.Normal)
		}
		if view.Parameters.ProtectNumbers && strings.ContainsFunc(token.Text, unicode.IsDigit) && i+1 < len(sentence.Tokens) {
			parts = append(parts, "unit:"+sentence.Tokens[i+1].Normal)
		}
	}
	return strings.Join(parts, "|")
}

func contrastCue(word string) bool {
	return slices.Contains([]string{
		"may", "must", "can", "should", "could", "would", "might", "shall",
		"if", "unless", "until", "when", "before", "after", "except", "only",
		"enabled", "disabled", "supported", "unsupported", "required", "optional",
		"allowed", "denied", "success", "failure",
	}, word)
}

func proseWords(view rule.View, sentence document.Sentence, words []string) []string {
	for i, token := range sentence.Tokens {
		if token.Word && !token.Protected && !view.Exempts(sentence, i, i+1) {
			words = append(words, token.Normal)
		}
	}
	return words
}

func makeWordSet(ctx context.Context, words []string, textBytes int) (feature.WordSet, error) {
	count := max(1, len(words))
	return feature.NewWordSet(ctx, words, feature.WordLimits{MaxWords: count, MaxUniqueWords: count,
		MaxBytes: min(1<<30, max(1, textBytes)*4)})
}

func sequenceLimits(sentence document.Sentence) feature.SequenceLimits {
	return feature.SequenceLimits{MaxTokens: max(1, len(sentence.Tokens)),
		MaxBytes: int(min(int64(1<<30), int64(max(1, len(sentence.Text)))*16))}
}
