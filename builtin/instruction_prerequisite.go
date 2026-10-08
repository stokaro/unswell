package builtin

import "github.com/stokaro/unswell/document"

// The complete reader prerequisite retains its installed state. A deadline,
// qualified state, optional requirement or architectural actor is not this frame.
func installedPrerequisite(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || c.start != 0 || rhetoricQuoted(c) || rhetoricAttributed(c) || prerequisiteGuard(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	if !installedPrerequisiteClause(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	frame, _ := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message: "Future and possession auxiliaries lengthen this installed-tools prerequisite.",
		suggestion: "State the installed-tools prerequisite directly, for example 'you need ... installed'. " +
			"Preserve the purpose, required installed state, versions and conditions, and keep recommended items optional. " +
			"Retain future wording when it expresses a separate deadline or state change.",
	}
	return frame, true
}

func installedPrerequisiteClause(tokens []document.Token) bool {
	start := prerequisiteReaderStart(tokens)
	if start < 0 || len(tokens)-start < 9 {
		return false
	}
	tokens = tokens[start:]
	return matches(tokens[:7], []string{"you", "will", "need", "to", "have", "the", "following"}) &&
		frameWord(tokens[len(tokens)-1], "installed") && len(tokens)-8 <= 8 &&
		nominalSubject(tokens[7:len(tokens)-1])
}

func prerequisiteReaderStart(tokens []document.Token) int {
	if frameWord(tokens[0], "you") {
		return 0
	}
	start := 1
	if len(tokens) > 3 && matches(tokens[:3], []string{"in", "order", "to"}) {
		start = 3
	} else if !frameWord(tokens[0], "to") {
		return -1
	}
	for end := start + 2; end < min(len(tokens), 14); end++ {
		if frameWord(tokens[end], ",") && grammaticalAction(tokens[start:end]) {
			return end + 1
		}
	}
	return -1
}

func prerequisiteGuard(tokens []document.Token) bool {
	if projectionGuard(tokens) || instructionCondition(tokens) {
		return true
	}
	for _, token := range tokens {
		if frameWord(token, "optional", "recommended", "before", "after", "by", "once", "while", "during",
			"deadline", "tomorrow", "today", "later", "next", "soon", "eventually", "already", "still") {
			return true
		}
	}
	return false
}
