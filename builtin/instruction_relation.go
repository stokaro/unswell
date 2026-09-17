package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func supportedOperand(tokens []document.Token) bool {
	for i := 2; i+4 < len(tokens); i++ {
		if !matches(tokens[i:i+3], []string{"to", "be", "used"}) ||
			!frameWord(tokens[i+3], "on", "in", "with") {
			continue
		}
		return actionOperand(tokens[:i]) && projectionOperand(tokens[i+2:])
	}
	return false
}

func relatedReaderMethod(a, b frameClause) bool {
	left, right := a.tokens(), b.tokens()
	if len(right) < 8 || !matches(right[:2], []string{"by", "using"}) ||
		!projectedInstructionWithMethod(a, true) {
		return false
	}
	for i := 3; i+3 < len(right); i++ {
		if frameWord(right[i], "you") && frameWord(right[i+1], "can", "may") &&
			grammaticalAction(right[i+2:]) && sharedMethodObject(left, right[i+2:]) {
			return true
		}
	}
	return false
}

// A repeated two-noun object is a bounded surface anchor for the method. A
// shared product name or a generic reader pronoun cannot establish the relation.
func sharedMethodObject(a, b []document.Token) bool {
	for i := 0; i+1 < len(a); i++ {
		if !methodNoun(a[i]) || !methodNoun(a[i+1]) {
			continue
		}
		for j := 0; j+1 < len(b); j++ {
			if methodNoun(b[j]) && methodNoun(b[j+1]) &&
				a[i].Normal == b[j].Normal && a[i+1].Normal == b[j+1].Normal {
				return true
			}
		}
	}
	return false
}

func methodNoun(token document.Token) bool {
	return !token.Protected && (token.Tag == "NN" || token.Tag == "NNS") &&
		!strings.HasSuffix(token.Normal, "ing")
}
