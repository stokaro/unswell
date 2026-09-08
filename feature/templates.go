package feature

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

// POSPattern owns a surface template or an explicit absence reason. The zero
// value is unavailable. Methods permit concurrent reads and Key exposes text.
type POSPattern struct {
	key, reason string
	available   bool
}

// Available reports whether a surface template was prepared.
func (p POSPattern) Available() bool { return p.available }

// Key returns the owned surface key, without a rule's protected signature.
func (p POSPattern) Key() string { return p.key }

// Reason returns an absence reason, or an empty string for an available key.
func (p POSPattern) Reason() string {
	if !p.available && p.reason == "" {
		return "not_computed"
	}
	return p.reason
}

// PreparePOSPattern folds the existing tag families without inferring missing
// POS. A nil literal mask selects no word literals; otherwise it must cover every
// input token. Punctuation is always literal. The caller supplies term selection
// and later adds protected signatures. Tokens remain unchanged during the call;
// the result retains no source buffers. This is not a dependency or meaning model.
func PreparePOSPattern(ctx context.Context, sentence document.Sentence, literal []bool, limits SequenceLimits) (POSPattern, error) {
	if err := validateSequence(ctx, sentence.Tokens, len(sentence.Text), limits); err != nil {
		return POSPattern{}, err
	}
	if !utf8.ValidString(sentence.Text) || (literal != nil && len(literal) != len(sentence.Tokens)) {
		return POSPattern{}, fmt.Errorf("invalid sentence text or literal mask")
	}
	if reason := templateExclusion(sentence); reason != "" {
		return POSPattern{reason: reason}, nil
	}
	return preparePOSParts(ctx, sentence.Tokens, literal)
}

func templateExclusion(sentence document.Sentence) string {
	if len(sentence.Tokens) == 0 {
		return "empty"
	}
	if sentence.Tokens[0].Tag == "VB" {
		return "imperative"
	}
	if strings.HasSuffix(strings.TrimSpace(sentence.Text), "?") {
		return "question"
	}
	return ""
}

func preparePOSParts(ctx context.Context, tokens []document.Token, literal []bool) (POSPattern, error) {
	parts := make([]string, 0, len(tokens))
	for i, token := range tokens {
		if err := ctx.Err(); err != nil {
			return POSPattern{}, err
		}
		if token.Protected {
			return POSPattern{reason: "protected_token"}, nil
		}
		if token.Tag == "" {
			return POSPattern{reason: "missing_pos"}, nil
		}
		part := templateTag(token.Tag)
		if !token.Word || (literal != nil && literal[i]) {
			part = token.Normal
		}
		parts = append(parts, part)
	}
	key := strings.Clone(strings.Join(parts, "|"))
	if err := ctx.Err(); err != nil {
		return POSPattern{}, err
	}
	return POSPattern{key: key, available: true}, nil
}

func templateTag(tag string) string {
	for _, prefix := range []string{"NN", "VB", "JJ", "RB"} {
		if strings.HasPrefix(tag, prefix) {
			return prefix
		}
	}
	return tag
}
