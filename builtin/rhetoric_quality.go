package builtin

import "github.com/stokaro/unswell/document"

// qualitativeAssurance requires a copular quality predicate, not an adjective
// inside a technical noun phrase. It requests scope, not proof of falsehood.
func qualitativeAssurance(c frameClause) (rhetoricalFrame, bool) {
	if contextualRhetoricScoped(c) || instructionGuard(c.tokens()) || question(c.sentence) || qualityMechanism(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		candidate, ok := localRhetoricCandidate(c, i)
		if ok && embeddedClauseStart(c.tokens(), i) && qualityJudgment(candidate.tokens()) {
			return localFrame(candidate, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func qualityJudgment(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-1, 16); i++ {
		if !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		if end := qualityCopula(tokens[i:]); end > 0 && qualityComplement(tokens[i+end:]) {
			return true
		}
	}
	return false
}

func qualityCopula(tokens []document.Token) int {
	if verb, negated := frameCopula(tokens[0]); verb != "" && !negated {
		if len(tokens) > 3 && frameWord(tokens[1], "supposed") && frameWord(tokens[2], "to") && frameWord(tokens[3], "be") {
			return 4
		}
		return 1
	}
	if len(tokens) > 1 && frameWord(tokens[0], "should", "would") && frameWord(tokens[1], "be") {
		return 2
	}
	return 0
}

func qualityComplement(tokens []document.Token) bool {
	i := 0
	for i < min(len(tokens), 3) && frameWord(tokens[i], "also", "now", "very", "much", "fairly", "quite", "rather", "particularly") {
		i++
	}
	if i >= len(tokens) || !qualitativeAdjective(tokens[i]) {
		return false
	}
	return qualityContinuation(tokens[i+1:])
}

func qualityContinuation(tokens []document.Token) bool {
	if len(tokens) == 0 {
		return true
	}
	if frameWord(tokens[0], "and") {
		return qualityComplement(tokens[1:])
	}
	if len(tokens) < 2 {
		return false
	}
	if frameWord(tokens[0], "to") {
		return grammaticalAction(tokens[1:]) || len(tokens) == 2 && instructionVerb(tokens[1:])
	}
	return frameWord(tokens[0], "than") && (nominalSubject(tokens[1:]) || methodAction(tokens[1:]))
}

func qualitativeAdjective(token document.Token) bool {
	return frameWord(token, "easy", "easier", "simple", "simpler", "straightforward", "difficult", "harder",
		"complex", "complicated", "powerful", "flexible", "fast", "faster", "slow", "slower", "smart", "smarter",
		"efficient", "inefficient", "expensive", "cheap", "intuitive", "effortless")
}

func qualityMechanism(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "because", "since", "by", "through", "due", "thanks", "as") {
			return true
		}
	}
	return false
}
