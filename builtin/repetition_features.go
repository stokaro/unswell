package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
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

func proseWordSet(view rule.View, sentence document.Sentence, words map[string]bool) int {
	count := 0
	for i, token := range sentence.Tokens {
		if token.Word && !token.Protected && !view.Exempts(sentence, i, i+1) {
			words[token.Normal] = true
			count++
		}
	}
	return count
}

func wordOverlap(left, right map[string]bool) (int, int) {
	shared := 0
	for word := range right {
		if left[word] {
			shared++
		}
	}
	return shared, len(left) + len(right) - shared
}

func informativeWord(word string) bool {
	return len(word) > 2 && !slices.Contains([]string{
		"the", "and", "for", "that", "this", "with", "from", "into", "are", "was", "were", "has", "have", "had",
		"its", "their", "they", "them", "you", "your", "our", "can", "will", "would", "could", "should", "may",
		"must", "not", "but", "also", "than", "then", "when", "which", "each", "all", "any", "one", "more", "most",
		"been", "being", "these", "those", "there", "here", "such", "some", "other", "through", "about", "over",
	}, word)
}

func contentKeys(words map[string]bool) []string {
	var keys []string
	for word := range words {
		if informativeWord(word) {
			keys = append(keys, word)
		}
	}
	slices.Sort(keys)
	return keys
}
