package builtin

import "github.com/stokaro/unswell/document"

// A local candidate stays within 48 tokens even when its sentence has a longer
// technical premise. Quotation and reported-claim boundaries still apply to the
// sentence; numeric operands in a separate premise do not scope the candidate.
func localRhetoricCandidate(c frameClause, start int) (frameClause, bool) {
	c.start += start
	return c, len(c.sentence.Tokens) <= 96 && c.eligible()
}

func candidateRhetoricScoped(c frameClause) bool {
	if quotedClaim(c.sentence.Tokens) {
		return true
	}
	for _, token := range c.sentence.Tokens {
		if frameWord(token, "according", "claims", "claimed", "says", "said", "argues", "argued", "rejects", "refutes", "disputes") {
			return true
		}
	}
	for i, token := range c.tokens() {
		if (token.Tag == "CD" && !indefiniteOne(c.tokens(), i)) ||
			frameWord(token, "if", "when", "unless", "whenever", "except", "without", "measured", "survey", "surveyed",
				"benchmark", "benchmarks", "median", "percentile", "respondents", "according") {
			return true
		}
	}
	return false
}

func purposeEvaluation(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-2, 8); i++ {
		if frameWord(tokens[i], "is", "was") && purposeSubject(tokens[:i]) {
			return purposeComplement(tokens[i+1:])
		}
	}
	return false
}

func purposeSubject(tokens []document.Token) bool {
	if deicticEvaluationSubject(tokens) {
		return true
	}
	return len(tokens) == 3 && frameWord(tokens[0], "this", "that") &&
		frameWord(tokens[1], "first", "second", "last", "other") && frameWord(tokens[2], "case", "result", "outcome")
}

func purposeComplement(tokens []document.Token) bool {
	if len(tokens) > 0 && frameWord(tokens[0], "exactly", "precisely") {
		tokens = tokens[1:]
	}
	return namedPurpose(tokens) || relevantPart(tokens) || nominalWorth(tokens) || existencePurpose(tokens)
}

func namedPurpose(tokens []document.Token) bool {
	// Abstract feature/finding purposes add no named action. A concrete wrapper,
	// lock, buffer or database purpose is not enough to establish this frame.
	if len(tokens) < 5 || !frameWord(tokens[0], "what") || !frameWord(tokens[1], "the", "this", "that") ||
		!frameWord(tokens[2], "feature", "finding", "refusal", "verb", "distinction", "difference", "rule") ||
		!frameWord(tokens[3], "is", "was") {
		return false
	}
	return purposeEnd(tokens[4:])
}

func purposeEnd(tokens []document.Token) bool {
	return len(tokens) == 1 && frameWord(tokens[0], "for", "about") ||
		len(tokens) == 2 && frameWord(tokens[0], "there") && frameWord(tokens[1], "for")
}

func relevantPart(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "the") &&
		frameWord(tokens[1], "part", "question", "point", "thing") && frameWord(tokens[2], "that", "which") &&
		frameWord(tokens[3], "matters", "counts")
}

func nominalWorth(tokens []document.Token) bool {
	for i := 2; i < min(len(tokens)-1, 6); i++ {
		if frameWord(tokens[i], "worth") && nominalSubject(tokens[:i]) &&
			(cognitiveWorth(tokens[i:]) || cognitiveWorthAbout(tokens[i:])) {
			return true
		}
	}
	return false
}

func cognitiveWorthAbout(tokens []document.Token) bool {
	return len(tokens) >= 4 && frameWord(tokens[0], "worth") && frameWord(tokens[1], "knowing", "hearing") &&
		frameWord(tokens[2], "about") && frameWord(tokens[3], "before", "now", "here")
}

func existencePurpose(tokens []document.Token) bool {
	if len(tokens) < 6 || !frameWord(tokens[0], "the") || !frameWord(tokens[1], "fact", "case", "reason") ||
		!frameWord(tokens[2], "the", "this", "that") {
		return false
	}
	i := 3
	if frameWord(tokens[i], "whole", "entire") {
		i++
	}
	return i+3 == len(tokens) && frameWord(tokens[i], "feature", "rule", "verb", "distinction") &&
		frameWord(tokens[i+1], "exists", "existed") && frameWord(tokens[i+2], "for")
}

func intentionalityAnnouncement(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "this", "that") &&
		frameWord(tokens[1], "direction", "distinction", "difference", "choice", "separation") &&
		frameWord(tokens[2], "is", "was") && frameWord(tokens[3], "deliberate", "intentional")
}

func outputUnderstanding(tokens []document.Token) bool {
	for i := 2; i < min(len(tokens)-4, 6); i++ {
		if !frameWord(tokens[i], "proves", "proved") || !positiveRhetoricSubject(tokens[:i]) ||
			!frameWord(tokens[i-1], "output", "sql", "result", "response") {
			continue
		}
		end := i + 1
		if frameWord(tokens[end], "that") {
			end++
		}
		if understoodInput(tokens[end:]) {
			return true
		}
	}
	return false
}

func understoodInput(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-2, 4); i++ {
		if frameWord(tokens[i], "understood", "understands") && nominalSubject(tokens[:i]) &&
			frameWord(tokens[i+1], "the", "your") && frameWord(tokens[i+2], "desired", "intended") {
			return true
		}
	}
	return false
}
