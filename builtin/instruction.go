package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// Instruction frames keep the action and its operands in the diagnostic. The
// suggested edit preserves optionality instead of changing a capability to a duty.
func instructionScaffolding(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if projectedInstruction(c) {
		parts := []frameClause{c}
		if instructionContinuation(clauses, index) {
			parts = append(parts, clauses[index+1])
		}
		return rhetoricalFrame{parts: parts}, true
	}
	if frame, ok := indirectInstruction(clauses, index); ok {
		return frame, true
	}
	return announcedInstruction(clauses, index)
}

func announcedInstruction(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !instructionEligible(c) {
		return rhetoricalFrame{}, false
	}
	if readerInstruction(c.tokens()) {
		return localFrame(c, 0)
	}
	if !capabilityInstruction(c.tokens()) && !readerGoal(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	// The preceding capability owns its method clause. A goal announcement owns
	// its following capability, so neither becomes a second occurrence.
	if instructionOwned(clauses, index) {
		return rhetoricalFrame{}, false
	}
	parts := []frameClause{c}
	if instructionContinuation(clauses, index) {
		parts = append(parts, clauses[index+1])
	}
	if readerGoal(c.tokens()) && len(parts) == 1 {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: parts}, true
}

func instructionEligible(c frameClause) bool {
	return c.eligible() && c.start == 0 && !instructionGuard(c.sentence.Tokens) && !quotedClaim(c.sentence.Tokens)
}

func instructionOwned(clauses []frameClause, index int) bool {
	return index > 0 && adjacentInstruction(clauses[index-1], clauses[index]) &&
		readerGoal(clauses[index-1].tokens()) && capabilityInstruction(clauses[index].tokens())
}

func instructionContinuation(clauses []frameClause, index int) bool {
	if index+1 >= len(clauses) || !adjacentInstruction(clauses[index], clauses[index+1]) {
		return false
	}
	return methodAnnouncement(clauses[index+1].tokens()) ||
		(readerGoal(clauses[index].tokens()) && capabilityInstruction(clauses[index+1].tokens()))
}

func adjacentInstruction(a, b frameClause) bool {
	return a.eligible() && b.eligible() &&
		a.ordinal+1 == b.ordinal && a.start == 0 && b.start == 0 &&
		!instructionGuard(a.tokens()) && !instructionGuard(b.tokens()) &&
		!quotedClaim(a.sentence.Tokens) && !quotedClaim(b.sentence.Tokens)
}

func capabilityInstruction(tokens []document.Token) bool {
	if len(tokens) < 6 || instructionCondition(tokens) || !frameWord(tokens[0], "it") || !frameWord(tokens[1], "is") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "also") {
		i++
	}
	return i+3 < len(tokens) && frameWord(tokens[i], "possible") && frameWord(tokens[i+1], "to") &&
		instructionAction(tokens[i+2:])
}

// Failure and permission statements need their possibility qualifier. Only a
// bounded transitive configuration or presentation action establishes a candidate.
func instructionAction(tokens []document.Token) bool {
	// The infinitive/imperative construction supplies the grammatical role.
	// POS taggers can label an uncommon base-form action as a noun.
	if len(tokens) < 2 || !frameWord(tokens[0],
		"add", "set", "specify", "configure", "enable", "disable", "activate", "deactivate",
		"show", "display", "express", "highlight", "attach", "define", "select", "choose",
		"change", "adjust", "customize", "create", "declare", "render", "format", "bind",
		"include", "insert", "escape", "put", "use", "get", "apply", "predefine", "write") {
		return false
	}
	for _, token := range tokens[1:] {
		if token.Protected || strings.HasPrefix(token.Tag, "NN") || frameWord(token, "it", "them", "this", "these") {
			return true
		}
	}
	return false
}

func instructionGuard(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "not", "n't", "never", "only", "unless", "without", "except", "provided",
			"denied", "forbidden", "unauthorized", "permission", "permissions", "privileges",
			"fail", "fails", "failure", "error", "errors", "lose", "loss", "corrupt", "corruption",
			"accidentally", "unintentionally", "risk") {
			return true
		}
	}
	return false
}

func readerGoal(tokens []document.Token) bool {
	if instructionCondition(tokens) || len(tokens) < 5 {
		return false
	}
	i := readerIntentionStart(tokens)
	return i > 0 && i+3 < len(tokens) && frameWord(tokens[i], "want", "wish", "intend") &&
		frameWord(tokens[i+1], "to") && instructionAction(tokens[i+2:])
}

func readerIntentionStart(tokens []document.Token) int {
	i := 0
	if frameWord(tokens[0], "sometimes") {
		i++
	}
	if i >= len(tokens) || !frameWord(tokens[i], "you") {
		return 0
	}
	i++
	if i < len(tokens) && frameWord(tokens[i], "may", "might") {
		i++
	}
	return i
}

func instructionCondition(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "if", "when", "whenever", "until", "unless", "provided", "except", "because") {
			return true
		}
	}
	return false
}

func readerInstruction(tokens []document.Token) bool {
	if len(tokens) < 9 || !frameWord(tokens[0], "if") {
		return false
	}
	for i := 6; i < len(tokens)-2; i++ {
		if !frameWord(tokens[i], ",") || !readerGoal(tokens[1:i]) {
			continue
		}
		rest := tokens[i+1:]
		if len(rest) > 2 && frameWord(rest[0], "you") && frameWord(rest[1], "can") {
			rest = rest[2:]
		}
		return !instructionCondition(rest) && instructionAction(rest)
	}
	return false
}

func methodAnnouncement(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "this", "that") {
		return false
	}
	i := 1
	switch {
	case frameWord(tokens[i], "can", "may") && frameWord(tokens[i+1], "be"):
		i += 2
	case frameWord(tokens[i], "is"):
		i++
	default:
		return false
	}
	if i+2 >= len(tokens) || !frameWord(tokens[i], "done", "performed", "achieved", "accomplished", "configured") ||
		!frameWord(tokens[i+1], "by", "using", "through", "with") {
		return false
	}
	return true
}
