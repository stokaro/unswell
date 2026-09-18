package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func nominalSupport(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := range tokens {
		if !licensedUsage(tokens[:i]) && (nominalUse(tokens[i:]) || instrumentalUse(tokens[i:]) || supportModifier(tokens[i:])) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func nominalUse(tokens []document.Token) bool {
	return len(tokens) >= 4 && frameWord(tokens[0], "make", "makes", "made", "making") &&
		frameWord(tokens[1], "use") && frameWord(tokens[2], "of") && supportOperand(tokens[3:])
}

func instrumentalUse(tokens []document.Token) bool {
	return len(tokens) >= 5 && frameWord(tokens[0], "by", "through", "with") &&
		matches(tokens[1:4], []string{"the", "use", "of"}) && supportOperand(tokens[4:])
}

func supportOperand(tokens []document.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	i := 0
	if frameWord(tokens[i], "a", "an", "the", "this", "that", "these", "those", "its", "your", "their") {
		i++
	}
	for ; i < min(len(tokens), 5); i++ {
		if tokens[i].Protected || strings.HasPrefix(tokens[i].Tag, "NN") {
			return true
		}
		if !strings.HasPrefix(tokens[i].Tag, "JJ") {
			return false
		}
	}
	return false
}

func supportModifier(tokens []document.Token) bool {
	if len(tokens) < 4 || !frameWord(tokens[1], "a", "an") {
		return false
	}
	if frameWord(tokens[0], "on") && frameWord(tokens[3], "basis") {
		return frameWord(tokens[2], "periodic", "regular", "daily", "weekly", "monthly", "annual", "yearly", "hourly")
	}
	return frameWord(tokens[0], "in") && frameWord(tokens[3], "manner") &&
		frameWord(tokens[2], "asynchronous", "synchronous", "sequential", "parallel", "deterministic", "consistent")
}

func licensedUsage(tokens []document.Token) bool {
	for i := 0; i+1 < len(tokens); i++ {
		if frameWord(tokens[i], "right", "rights", "permission", "authorization") && frameWord(tokens[i+1], "to") {
			return true
		}
	}
	return false
}
