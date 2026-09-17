package builtin

import "github.com/stokaro/unswell/document"

// informationEvaluation keeps a complete judgment and can join two adjacent
// evaluative predicates. Operational continuations remain outside the span.
func informationEvaluation(c frameClause) (rhetoricalFrame, bool) {
	tokens := c.tokens()
	end := informationUnitEnd(tokens)
	if informationScoped(c, end) {
		return rhetoricalFrame{}, false
	}
	if !informationJudgment(tokens[:end]) {
		return rhetoricalFrame{}, false
	}
	next := end
	if next < len(tokens) && frameWord(tokens[next], ",") {
		next++
	}
	if next+1 < len(tokens) && frameWord(tokens[next], "and", "but") && informationJudgment(tokens[next+1:]) {
		end = len(tokens)
	}
	c.end = c.start + end
	return localFrame(c, 0)
}

func informationUnitEnd(tokens []document.Token) int {
	for i, token := range tokens {
		if frameWord(token, ",", "and", "but") {
			return i
		}
	}
	return len(tokens)
}

func informationJudgment(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens), 18); i++ {
		subject, rest := tokens[:i], tokens[i+1:]
		if !evaluationSubject(subject) {
			continue
		}
		if importanceVerb(tokens[i], subject, rest) {
			return true
		}
		if frameWord(tokens[i], "is", "are", "was", "were") && informationComplement(rest) {
			return true
		}
	}
	return false
}

func evaluationSubject(tokens []document.Token) bool {
	return deicticEvaluationSubject(tokens) || discourseSubject(tokens) || quotedInformationSubject(tokens)
}

// A quoted noun phrase can be the information being evaluated. A quotation of
// the judgment itself does not have this shape and remains excluded.
func quotedInformationSubject(tokens []document.Token) bool {
	if len(tokens) < 3 || !quotedClaim(tokens[:1]) || !quotedClaim(tokens[len(tokens)-1:]) {
		return false
	}
	inside := tokens[1 : len(tokens)-1]
	return len(inside) <= 10 && !quotedClaim(inside) && nominalSubject(inside)
}

func informationComplement(tokens []document.Token) bool {
	if wholeAbstractValue(tokens) || bareAbstractAnnouncement(tokens) || discourseSelection(tokens) {
		return true
	}
	if len(tokens) > 2 && frameWord(tokens[0], "the", "a", "one") &&
		frameWord(tokens[1], "part", "point", "detail", "fact", "distinction", "difference", "thing", "reason") {
		return cognitiveInformation(tokens[2:])
	}
	return false
}

func bareAbstractAnnouncement(tokens []document.Token) bool {
	return len(tokens) == 3 && frameWord(tokens[0], "the") && frameWord(tokens[1], "whole", "entire") &&
		frameWord(tokens[2], "point", "purpose", "value", "benefit", "idea", "design", "guarantee", "criterion")
}

func cognitiveInformation(tokens []document.Token) bool {
	if len(tokens) < 2 || !frameWord(tokens[0], "worth") ||
		!cognitiveAct(tokens[1]) && !frameWord(tokens[1], "reading") {
		return false
	}
	if len(tokens) == 2 {
		return true
	}
	return len(tokens) == 3 && frameWord(tokens[2], "precisely", "carefully", "closely", "twice", "again", "here") ||
		len(tokens) == 4 && frameWord(tokens[2], "in") && frameWord(tokens[3], "full")
}

func importanceAudience(tokens []document.Token) bool {
	return len(tokens) == 0 || len(tokens) == 1 && frameWord(tokens[0], "here", "most") ||
		len(tokens) == 2 && frameWord(tokens[0], "for", "to") &&
			frameWord(tokens[1], "readers", "users", "operators", "maintainers", "embedders")
}

// Bare counts after a noun is also a plural nominal (row counts). Require an
// anaphoric subject or an explicit audience to resolve that ambiguous reading.
func importanceVerb(verb document.Token, subject, rest []document.Token) bool {
	return importanceAudience(rest) && (frameWord(verb, "matters") ||
		frameWord(verb, "counts") && (len(subject) == 1 || len(rest) > 0))
}

func informationScoped(c frameClause, end int) bool {
	tokens := c.tokens()
	candidate := c
	candidate.end = c.start + end
	if end+1 < len(tokens) && frameWord(tokens[end], ",") && !frameWord(tokens[end+1], "and", "but") {
		candidate.end = c.end
	}
	return evaluationScoped(candidate)
}
