package builtin

import "github.com/stokaro/unswell/document"

// tutorialNarration separates an invitation to inspect an example from a real
// collective operation or a promise to change the implementation. An explicit
// now/next cue can introduce a transitive tutorial action; bare future tense
// alone cannot establish this role.
func tutorialNarration(c frameClause) bool {
	if !c.eligible() || c.start != 0 || len(c.sentence.Tokens) > 96 ||
		rhetoricQuoted(c) || rhetoricAttributed(c) || projectionGuard(c.tokens()) || tutorialScoped(c.tokens()) {
		return false
	}
	tokens := c.tokens()
	start := tutorialStart(tokens)
	return tutorialAction(tokens, start)
}

func tutorialAction(tokens []document.Token, start int) bool {
	i := start
	if i+2 >= len(tokens) {
		return false
	}
	if matches(tokens[i:i+2], []string{"let", "us"}) || matches(tokens[i:i+2], []string{"let", "'s"}) {
		return tutorialInspection(tokens[i+2:])
	}
	if frameWord(tokens[i], "let's", "let’s") {
		return tutorialInspection(tokens[i+1:])
	}
	if !matches(tokens[i:i+2], []string{"we", "will"}) {
		return false
	}
	i += 2
	return start > 0 && (tutorialInspection(tokens[i:]) || instructionAction(tokens[i:]))
}

func tutorialScoped(tokens []document.Token) bool {
	if instructionCondition(tokens) {
		return true
	}
	for _, token := range tokens {
		if frameWord(token, "tomorrow", "tonight", "today", "release", "releases", "deadline", "approval",
			"before", "after", "together") {
			return true
		}
	}
	return false
}

func tutorialStart(tokens []document.Token) int {
	if len(tokens) < 2 || !frameWord(tokens[0], "now", "next") {
		return 0
	}
	if frameWord(tokens[1], ",") {
		return 2
	}
	return 1
}

func tutorialInspection(tokens []document.Token) bool {
	if len(tokens) < 2 {
		return false
	}
	if frameWord(tokens[0], "look") {
		return len(tokens) > 2 && frameWord(tokens[1], "at") && projectionOperand(tokens[1:])
	}
	return frameWord(tokens[0], "explore", "examine", "consider", "see", "illustrate", "demonstrate") &&
		projectionOperand(tokens)
}
