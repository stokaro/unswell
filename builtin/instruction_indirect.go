package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func indirectInstruction(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	if frame, ok := actionScaffolding(clauses[index]); ok {
		return frame, true
	}
	c := clauses[index]
	if !c.eligible() || quotedClaim(c.sentence.Tokens) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := range tokens {
		if !embeddedClauseStart(tokens, i) || instructionGuard(tokens[i:]) {
			continue
		}
		if indirectForm(tokens[i:], i > 0 || c.start > 0, ownedMethod(clauses, index)) {
			return localFrame(c, i)
		}
	}
	return rhetoricalFrame{}, false
}

func indirectForm(tokens []document.Token, embedded, owned bool) bool {
	return possibilityNoun(tokens) || nominalizedMethod(tokens) ||
		procedureAnnouncement(tokens) ||
		(embedded && capabilityInstruction(tokens)) || (concreteMethod(tokens) && !owned)
}

func ownedMethod(clauses []frameClause, index int) bool {
	return index > 0 && adjacentInstruction(clauses[index-1], clauses[index]) &&
		(instructionAnnouncement(clauses[index-1]) || readerGoal(clauses[index-1].tokens()))
}

func possibilityNoun(tokens []document.Token) bool {
	return len(tokens) >= 7 && frameWord(tokens[0], "there") && frameWord(tokens[1], "is") &&
		frameWord(tokens[2], "the", "a") && frameWord(tokens[3], "possibility") &&
		frameWord(tokens[4], "to") && !instructionCondition(tokens) && instructionAction(tokens[5:])
}

func concreteMethod(tokens []document.Token) bool {
	if !methodAnnouncement(tokens) {
		return false
	}
	for i := 3; i < min(len(tokens)-1, 7); i++ {
		if frameWord(tokens[i-1], "by", "through") && methodAction(tokens[i:]) {
			return true
		}
	}
	return false
}

func methodAction(tokens []document.Token) bool {
	if len(tokens) < 2 || tokens[0].Protected || !tokens[0].Word ||
		!strings.HasSuffix(tokens[0].Normal, "ing") || frameWord(tokens[0], "being", "having") {
		return false
	}
	return tokens[0].Tag == "VBG" || frameWord(tokens[0], "using", "defining", "providing", "configuring")
}

func nominalizedMethod(tokens []document.Token) bool {
	start := 0
	if frameWord(tokens[0], "the") {
		start++
	}
	if start >= len(tokens) || !frameWord(tokens[start], "attachment", "configuration", "selection",
		"installation", "creation", "declaration", "definition", "activation", "deactivation") {
		return false
	}
	for i := start + 1; i+4 < min(len(tokens), 22); i++ {
		if !frameWord(tokens[i], "is") || !nominalSubject(tokens[start:i]) || !frameWord(tokens[i+1], "done", "performed") {
			continue
		}
		return methodComplement(tokens[i+2:])
	}
	return false
}

func methodComplement(tokens []document.Token) bool {
	return (frameWord(tokens[0], "by") && methodAction(tokens[1:])) ||
		(frameWord(tokens[0], "as") && frameWord(tokens[1], "per", "shown") && frameWord(tokens[2], "below", "above"))
}

func readerActionStart(tokens []document.Token) (int, bool) {
	if !frameWord(tokens[0], "you", "we") {
		return 0, true
	}
	i := 1
	if i < len(tokens) && frameWord(tokens[i], "just") {
		i++
	}
	if i+1 >= len(tokens) {
		return 0, false
	}
	switch {
	case frameWord(tokens[i], "can", "could", "may", "would", "should"):
		return i + 1, true
	case frameWord(tokens[i], "need", "have") && frameWord(tokens[i+1], "to"):
		return i + 2, true
	default:
		return 0, false
	}
}
