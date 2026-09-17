package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func adjacentWords(ctx context.Context, view rule.View, emit rule.Emitter) error {
	budget := &repetitionBudget{ctx, view.MaxCandidates}
	for _, block := range view.Document.Blocks {
		if err := budget.spend(1); err != nil {
			return err
		}
		evaluated := false
		if !block.Excluded {
			for _, sentence := range block.Sentences {
				if err := adjacentSentence(view, block, sentence, budget, emit, &evaluated); err != nil {
					return err
				}
			}
		}
		if err := observeBlock(view, block, tokenBlockReason(block, evaluated, true)); err != nil {
			return err
		}
	}
	return nil
}

func adjacentSentence(view rule.View, block document.Block, sentence document.Sentence, budget *repetitionBudget,
	emit rule.Emitter, evaluated *bool) error {
	if err := budget.spend(len(sentence.Tokens) + len(sentence.Text) + 1); err != nil {
		return err
	}
	if quotedClaim(sentence.Tokens) {
		return nil
	}
	for i := 1; i < len(sentence.Tokens); i++ {
		left, right := sentence.Tokens[i-1], sentence.Tokens[i]
		if !adjacentProseWords(left, right) || view.Exempts(sentence, i-1, i+1) {
			continue
		}
		*evaluated = true
		if left.Normal != right.Normal || ambiguousDuplicate(left, right, sentence.Tokens) ||
			!adjacentWordGap(block, left, right) {
			continue
		}
		occurrences := []rule.Occurrence{tokenOccurrence(sentence, i-1, i), tokenOccurrence(sentence, i, i+1)}
		if err := emit.Emit(measured("exact", "adjacent-occurrences", "occurrences", 2, 1, 2, occurrences)); err != nil {
			return err
		}
		i++
	}
	return nil
}

func ambiguousDuplicate(left, right document.Token, tokens []document.Token) bool {
	if slices.Contains([]string{"had", "that", "do", "does", "did", "is", "was", "can", "will", "her", "his",
		"not", "no", "never", "yes", "bye", "ha", "so", "as", "like"}, left.Normal) {
		return true
	}
	if strings.HasPrefix(right.Tag, "NN") && ambiguousNominalRepeat(left, tokens) {
		return true
	}
	return strings.HasPrefix(left.Tag, "JJ") && strings.HasPrefix(right.Tag, "JJ") ||
		strings.HasPrefix(left.Tag, "RB") && strings.HasPrefix(right.Tag, "RB")
}

func ambiguousNominalRepeat(left document.Token, tokens []document.Token) bool {
	if strings.HasPrefix(left.Tag, "JJ") {
		return true
	}
	return strings.HasPrefix(left.Tag, "NN") && !slices.ContainsFunc(tokens, func(t document.Token) bool {
		return !t.Protected && slices.Contains([]string{"VB", "VBZ", "VBP", "VBD", "MD"}, t.Tag)
	})
}

func quotedClaim(tokens []document.Token) bool {
	return slices.ContainsFunc(tokens, func(t document.Token) bool {
		return !t.Protected && slices.Contains([]string{"\"", "'", "“", "”", "‘", "’", "``", "''"}, t.Text)
	})
}

func adjacentProseWords(left, right document.Token) bool {
	return left.Word && right.Word && !left.Protected && !right.Protected
}

func adjacentWordGap(block document.Block, left, right document.Token) bool {
	return left.End >= 0 && left.End <= right.Start && right.Start <= len(block.Text) &&
		strings.TrimSpace(block.Text[left.End:right.Start]) == ""
}
