package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func indirectInstruction(clauses []frameClause, index int) (rhetoricalFrame, bool) {
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
	return readerPurpose(tokens) || possibilityNoun(tokens) || nominalizedMethod(tokens) ||
		(embedded && capabilityInstruction(tokens)) || (concreteMethod(tokens) && !owned)
}

func ownedMethod(clauses []frameClause, index int) bool {
	return index > 0 && adjacentInstruction(clauses[index-1], clauses[index]) &&
		(capabilityInstruction(clauses[index-1].tokens()) || readerGoal(clauses[index-1].tokens()))
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

func readerPurpose(tokens []document.Token) bool {
	if len(tokens) < 8 || !frameWord(tokens[0], "if") {
		return false
	}
	for i := 4; i < len(tokens)-2; i++ {
		end := i
		if frameWord(tokens[i], ",") {
			end++
		} else if !frameWord(tokens[i], "you", "we") {
			continue
		}
		if intentionGoal(tokens[1:i]) && directInstruction(tokens[end:]) {
			return true
		}
	}
	return false
}

func intentionGoal(tokens []document.Token) bool {
	if len(tokens) < 3 || !frameWord(tokens[0], "you", "we") || instructionCondition(tokens) {
		return false
	}
	if frameWord(tokens[1], "want", "wish", "intend", "need") {
		return (frameWord(tokens[2], "to") && instructionAction(tokens[3:])) ||
			(frameWord(tokens[1], "need") && nominalSubject(tokens[2:]))
	}
	return interestedGoal(tokens)
}

func interestedGoal(tokens []document.Token) bool {
	i := 2
	if frameWord(tokens[1], "are") && frameWord(tokens[i], "just", "only") {
		i++
	}
	return i+2 < len(tokens) && frameWord(tokens[1], "are") && frameWord(tokens[i], "interested") &&
		frameWord(tokens[i+1], "in") && nominalSubject(tokens[i+2:])
}

func directInstruction(tokens []document.Token) bool {
	i, ok := readerActionStart(tokens)
	if !ok {
		return false
	}
	if i < len(tokens) && frameWord(tokens[i], "simply", "just") {
		i++
	}
	return !instructionCondition(tokens[i:]) &&
		(instructionAction(tokens[i:]) || (i+1 == len(tokens) && frameWord(tokens[i], "write")))
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
	case frameWord(tokens[i], "can", "could", "may"):
		return i + 1, true
	case frameWord(tokens[i], "need") && frameWord(tokens[i+1], "to"):
		return i + 2, true
	default:
		return 0, false
	}
}
