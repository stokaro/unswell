package builtin

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func exactRepetitionRule(sample string) rule.Rule {
	exact := descriptor(
		"repetition.exact-sentence",
		"This sentence is repeated; consider whether every occurrence is needed.",
		"repetition",
		"document",
		30,
	)
	exact.Defaults.Parameters = rule.Parameters{MinWords: 12, Window: "document"}
	exact.Version = "3"
	exact.RequiresStructure = true
	exact.Description = "Compares sentences within one structural section, with separate table-cell and leading-condition scopes."
	exact.Description += " Protected operands are compared by literal source identity without interpreting their contents."
	exact.Limitations += " Protected operands do not contribute to the prose word minimum. Identity comparison shares a bounded work budget."
	exact.BlockObservations = true
	exact.Parameters = []string{"min_words", "window"}
	exact.Examples = []rule.Example{{Text: sample + " " + sample, Match: true}, {Text: sample, Match: false}}
	opaque := "The client reads the value from `LEFT` before sending the request to the configured server."
	exact.Examples = append(exact.Examples, rule.Example{Text: opaque + "\n\n" + opaque, Format: document.Markdown, Match: true})
	return check{exact, exactRepetition}
}

// Protected operands participate only as literal source identities. Their
// contents do not become prose tokens or contribute to the word minimum.
func exactSentenceKey(view rule.View, sentence document.Sentence, budget *repetitionBudget) (string, error) {
	hasProtected, hasProse := exactSentenceFlags(sentence)
	if !hasProtected {
		return sentenceKey(sentence), nil
	}
	if !hasProse {
		return "", nil
	}
	var identity []byte
	for _, token := range sentence.Tokens {
		if err := budget.spend(len(token.Normal)); err != nil {
			return "", err
		}
		if !token.Protected {
			identity = appendExactPart(identity, 'n', []byte(token.Normal))
			continue
		}
		var err error
		identity, err = appendExactOperand(identity, token, view.Document.Source, budget)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("protected:%x", sha256.Sum256(identity)), nil
}

func exactSentenceFlags(sentence document.Sentence) (hasProtected, hasProse bool) {
	for _, token := range sentence.Tokens {
		hasProtected = hasProtected || token.Protected
		hasProse = hasProse || token.Word && !token.Protected
	}
	return hasProtected, hasProse
}

func appendExactOperand(identity []byte, token document.Token, source []byte, budget *repetitionBudget) ([]byte, error) {
	if len(token.Spans) == 0 {
		return nil, fmt.Errorf("protected repetition operand has no original source ranges")
	}
	identity = append(identity, 'o')
	identity = binary.AppendUvarint(identity, uint64(len(token.Spans)))
	for _, span := range token.Spans {
		if !span.Valid(len(source)) {
			return nil, fmt.Errorf("protected repetition operand has an invalid source range")
		}
		if err := budget.spend(span.End - span.Start); err != nil {
			return nil, err
		}
		identity = appendExactPart(identity, 'p', source[span.Start:span.End])
	}
	return identity, nil
}

func appendExactPart(identity []byte, kind byte, value []byte) []byte {
	identity = append(identity, kind)
	identity = binary.AppendUvarint(identity, uint64(len(value)))
	return append(identity, value...)
}
