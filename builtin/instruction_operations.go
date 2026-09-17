package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func supportAdverb(tokens []document.Token, at int) int {
	if at < len(tokens) && frameWord(tokens[at], "also") {
		return at + 1
	}
	return at
}

// A method/function head establishes an operation role even when the name is
// tagged as a verb. The role is not inferred for an arbitrary component.
func operationSubject(tokens []document.Token) bool {
	if len(tokens) < 2 || len(tokens) > 6 || !frameWord(tokens[len(tokens)-1], "method", "function") {
		return false
	}
	start := 0
	if frameWord(tokens[0], "the", "a", "this", "that") {
		start++
	}
	if start+1 >= len(tokens) {
		return false
	}
	for _, token := range tokens[start : len(tokens)-1] {
		if !token.Protected && (!token.Word || frameWord(token, "not", "no", "only", "which", "that", "when", "if")) {
			return false
		}
	}
	return true
}

func actionOperand(tokens []document.Token) bool {
	if projectionOperand(tokens) {
		return true
	}
	// Whether/if can be a queried value, not a condition on execution. Keep
	// the complement intact; a read-if-ready condition is not this role.
	if len(tokens) < 5 || !frameWord(tokens[0], "determine", "check", "verify", "test", "establish") ||
		!frameWord(tokens[1], "whether", "if") {
		return false
	}
	for _, token := range tokens[2:] {
		if strings.HasPrefix(token.Tag, "NN") || token.Protected {
			return true
		}
	}
	return false
}

func enabledComplement(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	if at+2 < len(tokens) && frameWord(tokens[at], "be") && projectedParticiple(tokens[at+1]) {
		return instructionProjection{action: at + 1, layers: layers}, true
	}
	return projectAction(tokens, at, layers)
}
