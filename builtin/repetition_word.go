package builtin

import (
	"context"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

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
		quotes := &adjacentQuoteScope{}
		if !block.Excluded {
			for _, sentence := range block.Sentences {
				if err := adjacentSentence(view, block, sentence, budget, emit, &evaluated, quotes); err != nil {
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
	emit rule.Emitter, evaluated *bool, quotes *adjacentQuoteScope) error {
	if err := budget.spend(len(sentence.Tokens) + len(sentence.Text) + 1); err != nil {
		return err
	}
	quoted := quotes.tokens(block.Text, sentence.Tokens)
	for i := 1; i < len(sentence.Tokens); i++ {
		left, right := sentence.Tokens[i-1], sentence.Tokens[i]
		if !unquotedAdjacentWords(sentence.Tokens, quoted, i) || view.Exempts(sentence, i-1, i+1) {
			continue
		}
		*evaluated = true
		if !sameWrittenWord(left, right, sentenceOpening(sentence.Tokens, i-1)) ||
			ambiguousDuplicate(left, right, sentence.Tokens) ||
			!adjacentWordGap(block, left, right) || !adjacentInSource(view.Document, left, right) {
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

func unquotedAdjacentWords(tokens []document.Token, quoted []bool, i int) bool {
	return !quoted[i-1] && !quoted[i] && adjacentProseWords(tokens[i-1], tokens[i])
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
	return left.Word && right.Word && !left.Protected && !right.Protected &&
		duplicateProseWord(left.Text) && duplicateProseWord(right.Text)
}

// duplicateProseWord excludes data and format conversions without excluding the
// surrounding prose. The tokenizer's Word flag also includes numeric operands.
func duplicateProseWord(text string) bool {
	letters := false
	for _, r := range text {
		switch {
		case unicode.IsLetter(r):
			letters = true
		case unicode.IsMark(r), r == '\'', r == '’', r == '-':
		default:
			return false
		}
	}
	return letters
}

// sameWrittenWord accepts a word written twice the same way. A sentence may
// open with the capitalized form of the repeated word. Elsewhere a repeated
// capitalized word is read as a name or a title, and a pair that differs in
// case as two words. On real pages two reviewers accepted no finding of
// either kind.
func sameWrittenWord(left, right document.Token, opening bool) bool {
	if left.Text == right.Text {
		return opening || !titleCase(right.Text)
	}
	first, size := utf8.DecodeRuneInString(right.Text)
	return opening && unicode.IsLower(first) && left.Text == string(unicode.ToUpper(first))+right.Text[size:]
}

// sentenceOpening reports whether only punctuation separates the token from
// the start of its sentence, or from a full stop that the segmenter kept
// inside it, as it does before a closing quotation mark.
func sentenceOpening(tokens []document.Token, index int) bool {
	for i := index - 1; i >= 0; i-- {
		if tokens[i].Word {
			return false
		}
		if slices.Contains([]string{".", "!", "?"}, tokens[i].Text) {
			return true
		}
	}
	return true
}

// titleCase reports an initial capital followed by a lowercase letter.
func titleCase(text string) bool {
	first, size := utf8.DecodeRuneInString(text)
	return unicode.IsUpper(first) && strings.IndexFunc(text[size:], unicode.IsLower) >= 0
}

// maxAdjacentSourceGap bounds the markup inspected between two words.
const maxAdjacentSourceGap = 64

// adjacentInSource rejects Markdown words that only the extracted text puts
// side by side. Emphasis, code, a closing link, or a table boundary between
// them gives each word its own role. A line break and an opening link
// bracket leave the pair adjacent for the reader.
func adjacentInSource(doc *document.Document, left, right document.Token) bool {
	if (doc.Format != document.Markdown && doc.Format != document.MDX) || len(left.Spans) == 0 || len(right.Spans) == 0 {
		return true
	}
	start, end := left.Spans[len(left.Spans)-1].End, right.Spans[0].Start
	if start < 0 || end < start || end > len(doc.Source) || end-start > maxAdjacentSourceGap {
		return false
	}
	return strings.Trim(string(doc.Source[start:end]), " \t\r\n[") == ""
}

func adjacentWordGap(block document.Block, left, right document.Token) bool {
	return left.End >= 0 && left.End <= right.Start && right.Start <= len(block.Text) &&
		strings.TrimSpace(block.Text[left.End:right.Start]) == ""
}
