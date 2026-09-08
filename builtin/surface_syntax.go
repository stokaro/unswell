package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func nominalizationChains(ctx context.Context, view rule.View, emit rule.Emitter) error {
	verbs, nouns := wordDictionary(view.Parameters.Verbs), wordDictionary(view.Parameters.Nouns)
	matcher := newEditorialMatcher(ctx, view)
	return surfaceSentences(matcher, func(sentence document.Sentence) error {
		return nominalizationSentence(matcher, sentence, verbs, nouns, emit)
	})
}

func surfaceSentences(m *editorialMatcher, visit func(document.Sentence) error) error {
	for _, block := range m.view.Document.Blocks {
		if !proseBlock(block) {
			continue
		}
		for _, sentence := range block.Sentences {
			if err := visit(sentence); err != nil {
				return err
			}
		}
	}
	return m.ctx.Err()
}

func nominalizationSentence(m *editorialMatcher, sentence document.Sentence, verbs, nouns map[string]bool, emit rule.Emitter) error {
	for start, token := range sentence.Tokens {
		if err := m.spend(); err != nil {
			return err
		}
		if !verbs[token.Normal] || !strings.HasPrefix(token.Tag, "VB") || token.Protected {
			continue
		}
		end := nominalizationEnd(m.view, sentence, start, nouns)
		if end <= start {
			continue
		}
		if err := emit.Emit(measured("heuristic", "nominalization-chains", "chains", 1, 0, 1,
			[]rule.Occurrence{tokenOccurrence(sentence, start, end)})); err != nil {
			return err
		}
	}
	return nil
}

func wordDictionary(words []string) map[string]bool {
	result := make(map[string]bool, len(words))
	for _, word := range words {
		result[document.Normalize(word)] = true
	}
	return result
}

func nominalizationEnd(view rule.View, sentence document.Sentence, start int, nouns map[string]bool) int {
	i := afterNominalModifiers(sentence.Tokens, start+1, 4)
	if i+2 >= len(sentence.Tokens) {
		return start
	}
	if !nominalHead(view, sentence, i, nouns) || !ofConnector(sentence.Tokens[i+1]) {
		return start
	}
	end := afterNominalModifiers(sentence.Tokens, i+2, 6)
	if end >= len(sentence.Tokens) || !commonNoun(sentence.Tokens[end]) {
		return start
	}
	if view.Exempts(sentence, start, end+1) {
		return start
	}
	return end + 1
}

func nominalHead(view rule.View, sentence document.Sentence, index int, nouns map[string]bool) bool {
	noun := sentence.Tokens[index]
	return nouns[noun.Normal] && commonNoun(noun) && !view.Exempts(sentence, index, index+1)
}

func afterNominalModifiers(tokens []document.Token, start, limit int) int {
	end := start
	for end < min(start+limit, len(tokens)) && nominalModifier(tokens[end]) {
		end++
	}
	return end
}

func ofConnector(token document.Token) bool {
	return !token.Protected && token.Normal == "of" && token.Tag == "IN"
}

func nominalModifier(token document.Token) bool {
	return !token.Protected && (token.Tag == "DT" || token.Tag == "PRP$" || strings.HasPrefix(token.Tag, "JJ"))
}

func commonNoun(token document.Token) bool {
	return !token.Protected && (token.Tag == "NN" || token.Tag == "NNS") && !protectedIdentifier(token, 1)
}

func nounStacks(ctx context.Context, view rule.View, emit rule.Emitter) error {
	matcher := newEditorialMatcher(ctx, view)
	return surfaceSentences(matcher, func(sentence document.Sentence) error {
		for _, chunk := range sentence.Chunks {
			if chunk.Kind != "NP" {
				continue
			}
			if err := nounChunk(matcher, sentence, chunk, emit); err != nil {
				return err
			}
		}
		return nil
	})
}

func nounChunk(m *editorialMatcher, sentence document.Sentence, chunk document.Chunk, emit rule.Emitter) error {
	if chunk.FirstToken < 0 || chunk.EndToken <= chunk.FirstToken || chunk.EndToken > len(sentence.Tokens) {
		return fmt.Errorf("invalid NP chunk token range")
	}
	start := chunk.FirstToken
	for i := chunk.FirstToken; i <= chunk.EndToken; i++ {
		if err := m.spend(); err != nil {
			return err
		}
		if i < chunk.EndToken && eligibleCommonNoun(m.view, sentence, i) {
			continue
		}
		if i-start > m.view.Parameters.Onset {
			if err := emit.Emit(measured("heuristic", "consecutive-common-nouns", "nouns", i-start,
				m.view.Parameters.Onset, m.view.Parameters.Saturation, []rule.Occurrence{tokenOccurrence(sentence, start, i)})); err != nil {
				return err
			}
		}
		start = i + 1
	}
	return nil
}

func eligibleCommonNoun(view rule.View, sentence document.Sentence, index int) bool {
	return commonNoun(sentence.Tokens[index]) && !view.Exempts(sentence, index, index+1)
}

func passiveEvents(m *editorialMatcher, sentences []document.Sentence, index int) ([]editorialEvent, error) {
	sentence := sentences[index]
	if sentence.Words < m.view.Parameters.MinWords {
		return nil, nil
	}
	var occurrences []rule.Occurrence
	for i, token := range sentence.Tokens {
		if err := m.spend(); err != nil {
			return nil, err
		}
		if token.Protected || !strings.HasPrefix(token.Tag, "VB") ||
			!slices.Contains([]string{"am", "is", "are", "was", "were", "be", "been", "being"}, token.Normal) {
			continue
		}
		end := passiveEnd(sentence, i)
		if end > i && !m.view.Exempts(sentence, i, end) {
			occurrences = append(occurrences, tokenOccurrence(sentence, i, end))
		}
	}
	if len(occurrences) == 0 {
		return nil, nil
	}
	return []editorialEvent{{index, index, occurrences}}, nil
}

func passiveEnd(sentence document.Sentence, start int) int {
	i := start + 1
	for i < min(start+5, len(sentence.Tokens)) && !sentence.Tokens[i].Protected && strings.HasPrefix(sentence.Tokens[i].Tag, "RB") {
		i++
	}
	if i < len(sentence.Tokens) && !sentence.Tokens[i].Protected && sentence.Tokens[i].Tag == "VBN" {
		return i + 1
	}
	return start
}
