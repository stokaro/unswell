package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// actionScaffolding requires a nominal action or gerund on both sides of the
// support verb. An ordinary ability, permission, or passive actor is insufficient.
func actionScaffolding(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || quotedClaim(c.sentence.Tokens) || instructionGuard(c.tokens()) || instructionCondition(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 1; i+4 < len(tokens); i++ {
		if nestedEnabling(tokens, i) || passiveActionMethod(tokens, i) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func nestedEnabling(tokens []document.Token, verb int) bool {
	if !frameWord(tokens[verb], "allows", "enables") || !nominalSubject(tokens[:verb]) {
		return false
	}
	for end := verb + 2; end+2 < min(len(tokens), verb+23); end++ {
		if !frameWord(tokens[end], "to") || !frameWord(tokens[end+1], "be") {
			continue
		}
		return nominalAction(tokens[verb+1:end]) && passiveParticiple(tokens[end+2])
	}
	return false
}

func passiveActionMethod(tokens []document.Token, verb int) bool {
	if !nominalAction(tokens[:verb]) {
		return false
	}
	end := verb
	if frameWord(tokens[end], "can", "could", "may") {
		end++
		if !frameWord(tokens[end], "be") {
			return false
		}
	} else if !frameWord(tokens[end], "is", "are") {
		return false
	}
	return performedMethod(tokens[end+1:])
}

func performedMethod(tokens []document.Token) bool {
	if len(tokens) < 3 {
		return false
	}
	end := 1
	if frameWord(tokens[0], "carried") && frameWord(tokens[1], "out") {
		end++
	} else if !frameWord(tokens[0], "done", "performed", "accomplished", "achieved") {
		return false
	}
	if frameWord(tokens[end], "using") {
		return projectionOperand(tokens[end:])
	}
	if !frameWord(tokens[end], "by", "through") {
		return false
	}
	end++
	if end < len(tokens) && frameWord(tokens[end], "directly", "simply") {
		end++
	}
	return methodAction(tokens[end:])
}

func nominalAction(tokens []document.Token) bool {
	if len(tokens) == 0 || !nominalSubject(tokens) {
		return false
	}
	if methodAction(tokens) {
		return true
	}
	// Use the nominal head, not an action word modifying a concrete object:
	// configuration files and deployment artifacts are objects, not actions.
	end := len(tokens)
	for i, token := range tokens {
		if frameWord(token, "of", "for", "on", "in", "within") {
			end = i
			break
		}
	}
	if end == 0 {
		return false
	}
	return frameWord(tokens[end-1], "configuration", "installation", "creation", "declaration",
		"definition", "activation", "deactivation", "selection", "attachment", "management",
		"deployment", "setup", "cleanup", "delivery", "assembly", "registration", "modification",
		"initialization", "execution", "validation", "verification", "generation", "transformation",
		"conversion", "migration", "authentication", "synchronization", "serialization")
}

func passiveParticiple(token document.Token) bool {
	return !token.Protected && token.Word && (token.Tag == "VBN" || strings.HasSuffix(token.Normal, "ed"))
}
