package builtin

import "github.com/stokaro/unswell/document"

// Discourse appraisals are separate from the operational explanation that
// follows. Construction words remain unprotected and literal uses stay open.
func discourseStance(c frameClause) (rhetoricalFrame, bool) {
	end := stanceEnd(c.tokens())
	if end == 0 {
		return rhetoricalFrame{}, false
	}
	c.end = c.start + end
	if rhetoricCondition(c) || rhetoricQuoted(c) || rhetoricAttributed(c) || stanceCriteria(c) {
		return rhetoricalFrame{}, false
	}
	frame, ok := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message:    "This clause appraises the explanation or outcome; state the information directly.",
		suggestion: stanceGuidance,
		fallback:   stanceGuidance,
	}
	return frame, ok
}

const stanceGuidance = "Keep the actions, numeric operands, absent operations, qualifications and established commitments. " +
	"State the information or operational consequence directly. Verify the basis before removing or weakening any claim."

func stanceEnd(tokens []document.Token) int {
	if len(tokens) < 5 {
		return 0
	}
	if end := continuedAnaphoricEvaluation(tokens); end > 0 {
		return end
	}
	if conversionAppraisal(tokens) || abstractCompletion(tokens) || bareValueIdentity(tokens) {
		return len(tokens)
	}
	return 0
}

func continuedAnaphoricEvaluation(tokens []document.Token) int {
	if !frameWord(tokens[0], "which", "this", "that", "it") || !frameWord(tokens[1], "is", "was") {
		return 0
	}
	rest := tokens[2:]
	if len(rest) > 3 && bareQuestion(rest[:3]) && stanceContinuation(rest[3:]) {
		return 5
	}
	if len(rest) > 4 && bareWorth(rest[:4]) && stanceContinuation(rest[4:]) {
		return 6
	}
	return 0
}

func stanceContinuation(tokens []document.Token) bool {
	if len(tokens) > 1 && frameWord(tokens[0], ",") {
		tokens = tokens[1:]
	}
	return len(tokens) > 1 && frameWord(tokens[0], "and", "but", "because")
}

func conversionAppraisal(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "this", "that", "it", "which") ||
		!frameWord(tokens[1], "is", "was") || !frameWord(tokens[2], "the", "an") ||
		!frameWord(tokens[3], "honest") || !frameWord(tokens[4], "conversion", "translation", "interpretation") {
		return false
	}
	return len(tokens) == 5 || len(tokens) > 6 && frameWord(tokens[5], "of")
}

// A demonstrative plan or step can carry a completion metaphor. An actual
// feedback/control loop, or an ordinary component function, is not this frame.
func abstractCompletion(tokens []document.Token) bool {
	for i := 2; i+4 < min(len(tokens), 20); i++ {
		if !frameWord(tokens[i], "is", "was") || !discourseAction(tokens[:i]) || !plannedAction(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		return len(rest) == 4 && frameWord(rest[0], "what") && frameWord(rest[1], "closes", "completes") &&
			frameWord(rest[2], "the", "this", "that") && frameWord(rest[3], "loop", "circle", "cycle") &&
			!literalLoop(tokens)
	}
	return false
}

func plannedAction(tokens []document.Token) bool {
	return len(tokens) >= 3 && frameWord(tokens[1], "this", "that", "these", "those") &&
		frameWord(tokens[len(tokens)-1], "plan", "plans", "step", "steps", "sequence")
}

func literalLoop(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "feedback", "controller", "circuit", "voltage", "current", "signal",
			"iteration", "iterates", "graph", "edge", "motor", "servo") {
			return true
		}
	}
	return false
}

func bareValueIdentity(tokens []document.Token) bool {
	for i := 1; i+3 < min(len(tokens), 18); i++ {
		if !frameWord(tokens[i], "is", "was") || !nominalSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		return len(rest) == 3 && frameWord(rest[0], "the", "a") &&
			frameWord(rest[1], "useful", "interesting", "important", "noteworthy") &&
			frameWord(rest[2], "part", "point", "detail", "aspect")
	}
	return false
}

func stanceCriteria(c frameClause) bool {
	for _, token := range c.sentence.Tokens {
		if frameWord(token, "measured", "survey", "surveyed", "benchmark", "benchmarks",
			"median", "percentile", "respondents", "criterion", "criteria", "compared") {
			return true
		}
	}
	return literalLoop(c.sentence.Tokens)
}
