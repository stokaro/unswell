package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// An unnamed, discourse-marked part of logic introduces an action through a
// support predicate. A concrete component or a numbered architectural part does
// not establish this frame. Keep the complete action and its conditions.
func abstractActionCarrier(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) || projectionGuard(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for verb := 5; verb+2 < min(len(tokens), 14); verb++ {
		if !frameWord(tokens[verb], "allows", "enables", "supports") ||
			!abstractActionSubject(tokens[:verb]) || !methodAction(tokens[verb+1:]) ||
			!actionOperand(tokens[verb+1:]) {
			continue
		}
		frame, _ := localFrame(c, 0)
		frame.explanation = editorialExplanation{
			message: "An unnamed part of logic introduces the supported action indirectly.",
			suggestion: "Name the component that can perform the action instead of announcing an unnamed part of logic. " +
				"Keep the capability, quantities, operands, conditions and limits; a capability is not an obligation.",
		}
		return frame, true
	}
	return rhetoricalFrame{}, false
}

func abstractActionSubject(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[len(tokens)-1], "logic", "code") {
		return false
	}
	start := abstractCarrierStart(tokens)
	if start == 0 {
		return false
	}
	// Only an unqualified nominal modifier can occur before the generic head.
	// Protected names, proper names, quantities and relational clauses would
	// give the part a distinct identity and are not construction vocabulary.
	for _, token := range tokens[start : len(tokens)-1] {
		if !abstractCarrierModifier(token) {
			return false
		}
	}
	return true
}

// The caller supplies at least five tokens, so the optional modifier cannot
// move the piece/of prefix past the generic head.
func abstractCarrierStart(tokens []document.Token) int {
	if !frameWord(tokens[0], "one", "a", "an", "the", "another") {
		return 0
	}
	at := 1
	if frameWord(tokens[at], "final", "additional", "further") {
		at++
	} else if !frameWord(tokens[0], "another") {
		return 0
	}
	if frameWord(tokens[at], "piece", "bit", "part") && frameWord(tokens[at+1], "of") {
		return at + 2
	}
	return 0
}

func abstractCarrierModifier(token document.Token) bool {
	return !token.Protected && token.Word &&
		(token.Tag == "NN" || token.Tag == "NNS" || strings.HasPrefix(token.Tag, "JJ"))
}
