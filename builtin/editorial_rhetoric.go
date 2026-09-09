package builtin

import (
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type editorialPrefix struct {
	length   int
	eligible bool
}

func (m *editorialMatcher) fixedPattern(phrase string) []string {
	parts, ok := m.fixed[phrase]
	if !ok {
		parts = phraseTokens(phrase)
		m.fixed[phrase] = parts
	}
	return parts
}

func (m *editorialMatcher) prefixTokens(sentence document.Sentence, phrases ...string) editorialPrefix {
	result := editorialPrefix{}
	for _, phrase := range phrases {
		parts := m.fixedPattern(phrase)
		if len(parts) > len(sentence.Tokens) {
			continue
		}
		if m.windowObservations != nil && !m.view.Exempts(sentence, 0, len(parts)) {
			result.eligible = true
		}
		if matches(sentence.Tokens[:len(parts)], parts) {
			result.length = len(parts)
			return result
		}
	}
	return result
}

func notOnlyEvents(m *editorialMatcher, sentences []document.Sentence, index int) ([]editorialEvent, error) {
	sentence := sentences[index]
	start := -1
	for i, token := range sentence.Tokens {
		m.observeNotOnlyStart(sentence, i)
		if token.Normal == ";" {
			start = -1
		}
		if token.Normal == "but" && start >= 0 {
			return []editorialEvent{{index, index, []rule.Occurrence{tokenOccurrence(sentence, start, i+1)}}}, nil
		}
		if token.Normal == "not" && i+1 < len(sentence.Tokens) && slices.Contains([]string{"only", "just"}, sentence.Tokens[i+1].Normal) {
			if err := m.spend(); err != nil {
				return nil, err
			}
			if start < 0 {
				start = i
			}
		}
	}
	return nil, nil
}

func pairedContrastEvents(m *editorialMatcher, sentences []document.Sentence, i int) ([]editorialEvent, error) {
	if err := m.spend(); err != nil {
		return nil, err
	}
	first := sentences[i]
	if i+1 >= len(sentences) || first.BlockID != sentences[i+1].BlockID || question(first) {
		return nil, nil
	}
	second := sentences[i+1]
	a := m.prefixTokens(first, "it is not about", "it's not about")
	b := m.prefixTokens(second, "it is about", "it's about")
	m.observeContrastPair(first, second, a, b)
	if a.length == 0 || b.length == 0 || question(second) ||
		m.view.Exempts(first, 0, a.length) || m.view.Exempts(second, 0, b.length) {
		return nil, nil
	}
	return []editorialEvent{{i, i + 1, []rule.Occurrence{
		tokenOccurrence(first, 0, a.length), tokenOccurrence(second, 0, b.length),
	}}}, nil
}

func whetherEvents(m *editorialMatcher, sentences []document.Sentence, i int) ([]editorialEvent, error) {
	if err := m.spend(); err != nil {
		return nil, err
	}
	sentence := sentences[i]
	start := m.prefixTokens(sentence, "whether you are", "whether you're")
	m.observeWhetherStart(sentence, "whether you are", "whether you're")
	if start.length == 0 {
		return nil, nil
	}
	or := false
	for end := start.length; end < min(len(sentence.Tokens), 32); end++ {
		token := sentence.Tokens[end]
		or = or || token.Normal == "or"
		if token.Normal == "," {
			if or && !m.view.Exempts(sentence, 0, end+1) {
				return []editorialEvent{{i, i, []rule.Occurrence{tokenOccurrence(sentence, 0, end+1)}}}, nil
			}
			break
		}
	}
	return nil, nil
}

func questionEvents(m *editorialMatcher, sentences []document.Sentence, i int) ([]editorialEvent, error) {
	first := sentences[i]
	if !question(first) || i+1 >= len(sentences) || first.BlockID != sentences[i+1].BlockID {
		return nil, nil
	}
	answer := sentences[i+1]
	if !shortProseAnswer(answer, m.view.Parameters.MaxAnswerWords) {
		return nil, nil
	}
	m.observeQuestionPair(first, answer)
	matches, err := m.phrases(first, true)
	if err != nil || len(matches) == 0 {
		return nil, err
	}
	match := matches[0]
	if match.end != len(first.Tokens) || m.view.Exempts(answer, 0, len(answer.Tokens)) {
		return nil, nil
	}
	return []editorialEvent{{i, i + 1, []rule.Occurrence{
		tokenOccurrence(first, match.start, match.end), sentenceOccurrence(answer),
	}}}, nil
}

func shortProseAnswer(sentence document.Sentence, maxWords int) bool {
	if sentence.Words == 0 || sentence.Words > maxWords || question(sentence) {
		return false
	}
	return !slices.ContainsFunc(sentence.Tokens, func(token document.Token) bool {
		return token.Protected || strings.ContainsFunc(token.Text, unicode.IsDigit)
	})
}

func triadEvents(m *editorialMatcher, sentences []document.Sentence, index int) ([]editorialEvent, error) {
	sentence := sentences[index]
	var events []editorialEvent
	for i := 0; i < len(sentence.Tokens); i++ {
		m.observeWindowWord(sentence, i)
		if !m.adjective(sentence, i) {
			continue
		}
		if err := m.spend(); err != nil {
			return nil, err
		}
		indices, end := m.adjectiveList(sentence, i)
		if adjectiveTriad(sentence, indices) {
			occurrence := rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID}
			for _, at := range indices {
				occurrence.Spans = append(occurrence.Spans, sentence.Tokens[at].Spans...)
			}
			events = append(events, editorialEvent{index, index, []rule.Occurrence{occurrence}})
		}
		i = end - 1
	}
	return events, nil
}

func (m *editorialMatcher) adjective(sentence document.Sentence, i int) bool {
	token := sentence.Tokens[i]
	return m.words[token.Normal] && adjectiveCandidate(token.Tag) && !token.Protected && !m.view.Exempts(sentence, i, i+1)
}

func adjectiveCandidate(tag string) bool {
	return strings.HasPrefix(tag, "JJ") || strings.HasPrefix(tag, "NN")
}

func adjectiveTriad(sentence document.Sentence, indices []int) bool {
	if len(indices) != 3 {
		return false
	}
	adjectives := 0
	for _, i := range indices {
		if strings.HasPrefix(sentence.Tokens[i].Tag, "JJ") {
			adjectives++
		}
	}
	return adjectives >= 2
}

func (m *editorialMatcher) adjectiveList(sentence document.Sentence, start int) ([]int, int) {
	indices := []int{start}
	end := start + 1
	for end < len(sentence.Tokens) {
		next := end
		for next < len(sentence.Tokens) && slices.Contains([]string{",", "and"}, sentence.Tokens[next].Normal) {
			next++
		}
		if next == end || next >= len(sentence.Tokens) || !m.adjective(sentence, next) {
			break
		}
		indices = append(indices, next)
		end = next + 1
	}
	return indices, end
}
