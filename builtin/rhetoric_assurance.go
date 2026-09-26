package builtin

import "github.com/stokaro/unswell/document"

func unscopedAssurance(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if frame, ok := verificationAssurance(c); ok {
		return frame, true
	}
	if frame, ok := readerAssurance(c); ok {
		return frame, true
	}
	return clauseAssurance(c)
}

func clauseAssurance(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() {
		return rhetoricalFrame{}, false
	}
	// A colon introduces the scope or mechanism in the same sentence.
	if c.end < len(c.sentence.Tokens) && frameWord(c.sentence.Tokens[c.end], ":") {
		return rhetoricalFrame{}, false
	}
	if frame, ok := qualitativeAssurance(c); ok {
		return frame, true
	}
	if frame, ok := contextualAssuranceFrame(c); ok {
		return frame, true
	}
	for i := range c.tokens() {
		if !embeddedClauseStart(c.tokens(), i) {
			continue
		}
		if end := assuranceEnd(c.tokens()[i:]); end > 0 {
			c.end = c.start + i + end
			return localFrame(c, i)
		}
	}
	return rhetoricalFrame{}, false
}

func assuranceEnd(tokens []document.Token) int {
	for i := 1; i < min(len(tokens)-2, 12); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		if len(rest) == 2 && frameWord(rest[0], "intentionally", "deliberately", "purposely") &&
			frameWord(rest[1], "explicit", "simple", "clear", "safe", "robust", "reliable", "straightforward") {
			return len(tokens)
		}
		if end := nearlyAlwaysRight(rest); end > 0 {
			return i + 1 + end
		}
		if comprehensionPromise(rest) {
			return len(tokens)
		}
	}
	return 0
}

func nearlyAlwaysRight(tokens []document.Token) int {
	if len(tokens) < 5 || !frameWord(tokens[0], "the") || !frameWord(tokens[1], "right", "best", "correct") ||
		!frameWord(tokens[2], "answer", "choice", "solution", "approach", "default") ||
		!frameWord(tokens[3], "nearly", "almost") || !frameWord(tokens[4], "always") || !rhetoricEnd(tokens, 5) {
		return 0
	}
	return 5
}

func comprehensionPromise(tokens []document.Token) bool {
	if !universalComprehension(tokens) {
		return false
	}
	i := 6
	if i < len(tokens) && frameWord(tokens[i], "else") {
		i++
	}
	if i < len(tokens) && frameWord(tokens[i], "here") {
		i++
	}
	return i == len(tokens)
}

func universalComprehension(tokens []document.Token) bool {
	return len(tokens) >= 6 && frameWord(tokens[0], "the") && frameWord(tokens[1], "fastest", "easiest", "best") &&
		frameWord(tokens[2], "way") && frameWord(tokens[3], "to") && frameWord(tokens[4], "understand", "learn", "grasp") &&
		frameWord(tokens[5], "everything")
}
