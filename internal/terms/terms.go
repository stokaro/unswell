// Package terms matches explicit vocabulary at complete token boundaries.
package terms

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/jdkato/prose/v3/tokenize"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// Matcher is immutable after Compile and can be shared by concurrent analyses.
type Matcher struct {
	patterns  map[string][][]string
	sensitive bool
}

// Compile validates complete-token terms and rejects normalized duplicates.
func Compile(terms []string, sensitive bool) (*Matcher, error) {
	if len(terms) > 10000 {
		return nil, fmt.Errorf("vocabulary exceeds 10000 terms")
	}
	m := &Matcher{patterns: make(map[string][][]string), sensitive: sensitive}
	seen := make(map[string]bool)
	for _, term := range terms {
		words, err := m.termWords(term)
		if err != nil {
			return nil, err
		}
		key := strings.Join(words, "\x00")
		if seen[key] {
			return nil, fmt.Errorf("duplicate vocabulary term %q", term)
		}
		seen[key] = true
		m.patterns[words[0]] = append(m.patterns[words[0]], words)
	}
	return m, nil
}

func (m *Matcher) termWords(term string) ([]string, error) {
	if len(term) == 0 || len(term) > 1000 || !utf8.ValidString(term) || strings.ContainsRune(term, '\x00') {
		return nil, fmt.Errorf("terms require 1 to 1000 valid UTF-8 bytes without protected markers")
	}
	if !strings.ContainsFunc(term, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsNumber(r) }) {
		return nil, fmt.Errorf("term %q has no letters or numbers", term)
	}
	var words []string
	for _, token := range tokenize.New().Tokenize(term) {
		words = append(words, m.normal(token.Text))
	}
	if len(words) == 0 || len(words) > 32 {
		return nil, fmt.Errorf("terms require 1 to 32 tokens")
	}
	return words, nil
}

func (m *Matcher) normal(text string) string {
	if m.sensitive {
		return text
	}
	return document.Normalize(text)
}

// Find returns exact token ranges without changing documents or NLP annotations.
func (m *Matcher) Find(ctx context.Context, doc *document.Document, budget int) ([]rule.TokenRange, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var result []rule.TokenRange
	for _, block := range doc.Blocks {
		for _, sentence := range block.Sentences {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			found, err := m.sentence(ctx, sentence, &budget)
			if err != nil {
				return nil, err
			}
			result = append(result, found...)
			if len(result) > 100000 {
				return nil, fmt.Errorf("term matches exceed 100000 ranges")
			}
		}
	}
	return result, nil
}

func (m *Matcher) sentence(ctx context.Context, sentence document.Sentence, budget *int) ([]rule.TokenRange, error) {
	var result []rule.TokenRange
	for start, token := range sentence.Tokens {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if token.Protected {
			continue
		}
		for _, pattern := range m.patterns[m.normal(token.Text)] {
			match, err := m.matches(sentence.Tokens[start:], pattern, budget)
			if err != nil {
				return nil, err
			}
			if match {
				result = append(result, rule.TokenRange{BlockID: sentence.BlockID, SentenceID: sentence.ID, Start: start, End: start + len(pattern)})
			}
		}
	}
	return result, nil
}

func (m *Matcher) matches(tokens []document.Token, pattern []string, budget *int) (bool, error) {
	*budget--
	if *budget < 0 {
		return false, fmt.Errorf("term matching exceeds max_candidates")
	}
	if len(pattern) > len(tokens) {
		return false, nil
	}
	for i, word := range pattern {
		*budget--
		if *budget < 0 {
			return false, fmt.Errorf("term matching exceeds max_candidates")
		}
		if tokens[i].Protected || m.normal(tokens[i].Text) != word {
			return false, nil
		}
	}
	return true, nil
}
