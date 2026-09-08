package mapping

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/stokaro/unswell/document"
)

type canonicalPart struct {
	Text          string `json:"text,omitempty"`
	ProtectedHash string `json:"protected_hash,omitempty"`
}

// Canonical collapses prose whitespace while retaining case, punctuation, and
// protected source hashes. It encodes explicit boundaries rather than joining
// prose across code or URLs. Callers may hash this representation for identity;
// it is not input for editorial rules or a measure of semantic equivalence.
func Canonical(mapped document.MappedText, source []byte) (string, error) {
	if len(mapped.Text) != len(mapped.Map) {
		return "", fmt.Errorf("canonical text requires one source map entry per byte")
	}
	parts := make([]canonicalPart, 0)
	start := 0
	for i := 0; i < len(mapped.Text); i++ {
		if mapped.Text[i] != 0 {
			continue
		}
		parts = appendCanonicalText(parts, mapped.Text[start:i])
		span := mapped.Map[i]
		if !span.Valid(len(source)) {
			return "", fmt.Errorf("canonical protected boundary has an invalid source span")
		}
		hash := fmt.Sprintf("%x", sha256.Sum256(source[span.Start:span.End]))
		parts = append(parts, canonicalPart{ProtectedHash: hash})
		start = i + 1
	}
	parts = appendCanonicalText(parts, mapped.Text[start:])
	data, err := json.Marshal(parts)
	return string(data), err
}

func appendCanonicalText(parts []canonicalPart, text string) []canonicalPart {
	if text = strings.Join(strings.Fields(text), " "); text != "" {
		parts = append(parts, canonicalPart{Text: text})
	}
	return parts
}
