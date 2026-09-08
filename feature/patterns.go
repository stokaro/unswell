package feature

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

// PatternContract identifies n-gram selection and surface POS normalization.
const PatternContract = "unswell-pattern-features-v1"

// SequenceLimits bounds input tokens and the total bytes of normalized words,
// POS tags, and any sentence text consumed by a preprocessing call.
type SequenceLimits struct {
	MaxTokens, MaxBytes int
}

func validateSequence(ctx context.Context, tokens []document.Token, textBytes int, limits SequenceLimits) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateSequenceLimits(limits); err != nil {
		return err
	}
	if len(tokens) > limits.MaxTokens {
		return ErrTokenLimit
	}
	bytes := int64(textBytes)
	for _, token := range tokens {
		if err := ctx.Err(); err != nil {
			return err
		}
		bytes += int64(len(token.Normal)) + int64(len(token.Tag))
		if bytes > int64(limits.MaxBytes) {
			return fmt.Errorf("sequence exceeds max_bytes")
		}
		if err := validateSequenceToken(token); err != nil {
			return err
		}
	}
	if bytes > int64(limits.MaxBytes) {
		return fmt.Errorf("sequence exceeds max_bytes")
	}
	return nil
}

func validateSequenceLimits(limits SequenceLimits) error {
	if limits.MaxTokens < 1 || limits.MaxTokens > 1<<30 || limits.MaxBytes < 1 || limits.MaxBytes > 1<<30 {
		return fmt.Errorf("invalid sequence limits")
	}
	return nil
}

func validateSequenceToken(token document.Token) error {
	if !utf8.ValidString(token.Normal) || !utf8.ValidString(token.Tag) || len(token.Tag) > 128 ||
		strings.ContainsRune(token.Normal, 0) || strings.ContainsRune(token.Tag, 0) {
		return fmt.Errorf("sequence contains an invalid normalized token or POS tag")
	}
	if token.Word && !token.Protected && token.Normal == "" {
		return fmt.Errorf("sequence word has empty normalization")
	}
	return nil
}

// PatternCatalog describes source-derived preprocessing keys. These are not
// numeric Value measurements and must not be included in saved reports by default.
// Callers must identify source selection, NLP, terms, and normalization policies.
func PatternCatalog() []Descriptor {
	return []Descriptor{
		{ID: "prose-ngram", Version: "1", Family: "repetition", Type: "string", Unit: "normalized-key", Scope: "candidate",
			Formula:       "Consecutive 3-8 Word tokens without Protected tokens; require two distinct candidate content words.",
			Normalization: "Join provider Normal with spaces unchanged; use the versioned lexical content-word filter.",
			Requires:      []nlp.Capability{nlp.Tokens}, MinWords: 3,
			Missing:     "No candidate means no eligible sequence; callback errors and resource exhaustion are incomplete computation.",
			Limitations: "Candidates require caller term exclusions, protected signatures, and occurrence policy before qualification."},
		{ID: "surface-pos-template", Version: "1", Family: "repetition", Type: "string", Unit: "normalized-key", Scope: "sentence",
			Formula:       "Join token tags; fold NN/VB/JJ/RB prefixes; retain punctuation and explicitly selected term tokens literally.",
			Normalization: "Preserve provider Normal for literal tokens; join parts with the existing vertical-bar delimiter.",
			Requires:      []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS},
			Missing:       "Absent for empty input, questions, imperative openings, protected tokens, or missing POS; tags are never inferred.",
			Limitations:   "Surface shape only; caller adds protected signatures. Delimited keys do not establish semantic equivalence."},
	}
}
