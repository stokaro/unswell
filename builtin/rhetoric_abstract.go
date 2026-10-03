package builtin

import "github.com/stokaro/unswell/document"

// A bare abstract identity turns information into a conclusion without naming
// the relation it establishes. Preserve an explicit mode prefix in the span.
func abstractInformationFrame(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	if len(tokens) > 2 && frameWord(tokens[0], "offline", "online") && frameWord(tokens[1], ",") {
		tokens = tokens[2:]
	}
	if !abstractInformationIdentity(tokens) {
		return rhetoricalFrame{}, false
	}
	frame, ok := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message: "This sentence presents an abstract identity; state the concrete behavior it explains.",
		suggestion: "Keep the stated conditions, identifiers and technical limits. " +
			"Replace the bare abstract identity with the concrete behavior explained nearby.",
	}
	return frame, ok
}

func abstractInformationIdentity(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "the", "this", "that") ||
		!frameWord(tokens[1], "declaration", "description", "specification", "configuration", "document") ||
		!frameWord(tokens[2], "is", "was") || !frameWord(tokens[3], "the") {
		return false
	}
	return len(tokens) == 5 && frameWord(tokens[4], "evidence", "proof", "answer", "truth") ||
		len(tokens) == 6 && frameWord(tokens[4], "whole") && frameWord(tokens[5], "truth")
}

// An anaphoric abstract contrast comments on a behavior rather than naming
// its operational alternatives. Concrete subjects and qualified predicates
// cannot establish this frame.
func abstractRecastFrame(c frameClause) (rhetoricalFrame, bool) {
	if rhetoricQuoted(c) || rhetoricAttributed(c) || rhetoricCondition(c) || !abstractRecast(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	frame, ok := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message: "This contrast recasts the behavior as an abstract judgment; state the concrete constraint directly.",
		suggestion: "Keep unsupported operations, conditions and alternatives. " +
			"Replace the rhetorical judgment with the concrete scope of the behavior.",
	}
	return frame, ok
}

func abstractRecast(tokens []document.Token) bool {
	if len(tokens) < 8 || !frameWord(tokens[0], "this", "that") || !frameWord(tokens[1], "is", "was") ||
		!frameWord(tokens[2], "a", "an", "the") {
		return false
	}
	i := 3
	if frameWord(tokens[i], "design", "safety", "architectural") {
		i++
	}
	if !frameWord(tokens[i], "boundary", "constraint", "choice", "decision", "policy", "tradeoff", "compromise") {
		return false
	}
	i = abstractAlternativeStart(tokens, i+1)
	return abstractAlternative(tokens, i)
}

func abstractAlternative(tokens []document.Token, i int) bool {
	if i == 0 || i+1 >= len(tokens) || !frameWord(tokens[i], "a", "an", "the") {
		return false
	}
	i++
	if frameWord(tokens[i], "missing", "absent") {
		i++
	}
	return i+1 == len(tokens) && frameWord(tokens[i], "feature", "preference", "accident", "defect", "bug", "mistake", "failure")
}

func abstractAlternativeStart(tokens []document.Token, i int) int {
	if i+2 >= len(tokens) {
		return 0
	}
	if frameWord(tokens[i], "rather") && frameWord(tokens[i+1], "than") ||
		frameWord(tokens[i], "instead") && frameWord(tokens[i+1], "of") ||
		frameWord(tokens[i], ",") && frameWord(tokens[i+1], "not") {
		return i + 2
	}
	return 0
}
