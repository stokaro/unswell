package builtin

import "github.com/stokaro/unswell/document"

// narratedInstruction separates the tutorial narrator from the action. It does
// not turn an operational condition or another actor's behavior into a command.
func narratedInstruction(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || quotedClaim(c.sentence.Tokens) || projectionGuard(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 0; i+5 < len(tokens); i++ {
		if narratedAction(tokens, i) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func narratedAction(tokens []document.Token, at int) bool {
	if !frameWord(tokens[at], "we") || !narrativeStart(tokens, at) {
		return false
	}
	action, layers := narratorAction(tokens, at)
	if layers == 0 || action >= len(tokens) || at == 0 && layers < 2 {
		return false
	}
	return !frameWord(tokens[action], "avoid", "prevent", "remember", "know", "believe") &&
		projectionVerb(tokens[action]) && actionOperand(tokens[action:])
}

func narrativeStart(tokens []document.Token, at int) bool {
	if at == 0 {
		return true
	}
	if at < 5 || !matches(tokens[:4], []string{"now", "that", "we", "have"}) {
		return false
	}
	// Require a stated prerequisite rather than treating "now" as proof.
	end := at
	if frameWord(tokens[end-1], ",") {
		end--
	}
	return projectionOperand(tokens[3:end])
}

func narratorAction(tokens []document.Token, at int) (int, int) {
	i := at + 1
	if i < len(tokens) && frameWord(tokens[i], "will") {
		i++
	}
	if i+2 >= len(tokens) || !frameWord(tokens[i], "need", "have") || !frameWord(tokens[i+1], "to") {
		return 0, 0
	}
	return narratorAssurance(tokens, i+2)
}

func narratorAssurance(tokens []document.Token, i int) (int, int) {
	layers := 1
	if i+4 < len(tokens) && matches(tokens[i:i+4], []string{"make", "sure", "that", "we"}) {
		i += 4
		layers++
	} else if i+2 < len(tokens) && matches(tokens[i:i+2], []string{"ensure", "that"}) && frameWord(tokens[i+2], "we") {
		i += 3
		layers++
	}
	return i, layers
}
