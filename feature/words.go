package feature

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

// LexicalContract identifies set overlap and the candidate content-word filter.
const LexicalContract = "unswell-lexical-features-v1"

// WordLimits bounds an already selected normalized word sequence.
type WordLimits struct {
	MaxWords, MaxUniqueWords, MaxBytes int
}

// WordSet owns distinct normalized words for repeated comparisons. Its zero value
// is unavailable; a constructed empty set is available. All methods permit
// concurrent reads. Words and ContentKeys explicitly return source-derived text.
type WordSet struct {
	words map[string]struct{}
	keys  []string
}

// NewWordSet copies an already selected normalized word sequence. It does not
// tokenize, normalize, or decide whether protected or exempt words belong in it.
// The caller must preserve those source policies and record preprocessing identity.
func NewWordSet(ctx context.Context, words []string, limits WordLimits) (WordSet, error) {
	if err := ctx.Err(); err != nil {
		return WordSet{}, err
	}
	if err := validateWordLimits(limits); err != nil {
		return WordSet{}, err
	}
	if len(words) > limits.MaxWords {
		return WordSet{}, fmt.Errorf("lexical input exceeds max_words")
	}
	wordsByKey, err := copyWords(ctx, words, limits)
	if err != nil {
		return WordSet{}, err
	}
	set := WordSet{words: wordsByKey}
	set.keys = make([]string, 0, len(set.words))
	for word := range set.words {
		set.keys = append(set.keys, word)
	}
	slices.Sort(set.keys)
	if err := ctx.Err(); err != nil {
		return WordSet{}, err
	}
	return set, nil
}

func copyWords(ctx context.Context, words []string, limits WordLimits) (map[string]struct{}, error) {
	set := make(map[string]struct{})
	bytes := 0
	for _, word := range words {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		bytes += len(word)
		if bytes > limits.MaxBytes {
			return nil, fmt.Errorf("lexical input exceeds max_bytes")
		}
		if word == "" || !utf8.ValidString(word) || strings.ContainsRune(word, 0) {
			return nil, fmt.Errorf("lexical input contains an empty or invalid normalized word")
		}
		if _, exists := set[word]; !exists {
			set[strings.Clone(word)] = struct{}{}
		}
		if len(set) > limits.MaxUniqueWords {
			return nil, fmt.Errorf("lexical input exceeds max_unique_words")
		}
	}
	return set, nil
}

func validateWordLimits(limits WordLimits) error {
	if limits.MaxWords < 1 || limits.MaxWords > 1<<30 || limits.MaxUniqueWords < 1 ||
		limits.MaxUniqueWords > limits.MaxWords || limits.MaxBytes < 1 || limits.MaxBytes > 1<<30 {
		return fmt.Errorf("invalid lexical limits")
	}
	return nil
}

// Available distinguishes a constructed set from an unavailable zero value.
func (s WordSet) Available() bool { return s.words != nil }

// Len returns the distinct word count. Check Available before interpreting zero.
func (s WordSet) Len() int { return len(s.words) }

// Words returns owned, sorted normalized words, including the candidate stopwords.
func (s WordSet) Words() []string { return slices.Clone(s.keys) }

// ContentKeys returns sorted candidate-index keys using the versioned filter.
// The filter affects candidate retrieval, not the set's overlap denominator.
func (s WordSet) ContentKeys() []string {
	var keys []string
	for _, word := range s.keys {
		if InformativeWord(word) {
			keys = append(keys, word)
		}
	}
	return keys
}

// InformativeWord implements the existing English repetition candidate filter.
// Length is measured in bytes for compatibility. It does not determine whether a
// word is meaningful, whether a term is appropriate, or whether prose needs editing.
func InformativeWord(word string) bool {
	return len(word) > 2 && !slices.Contains([]string{
		"the", "and", "for", "that", "this", "with", "from", "into", "are", "was", "were", "has", "have", "had",
		"its", "their", "they", "them", "you", "your", "our", "can", "will", "would", "could", "should", "may",
		"must", "not", "but", "also", "than", "then", "when", "which", "each", "all", "any", "one", "more", "most",
		"been", "being", "these", "those", "there", "here", "such", "some", "other", "through", "about", "over",
	}, word)
}
