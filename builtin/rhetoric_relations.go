package builtin

import "github.com/stokaro/unswell/document"

// Relations retain the written subject and predicate. These bounded surface
// checks do not infer dependencies or equality between different propositions.
func rhetoricalRelationEnd(candidate frameClause) int {
	if rhetoricQuoted(candidate) || evaluationScoped(candidate) {
		return 0
	}
	tokens := candidate.tokens()
	if rhetoricalRelation(tokens) {
		return len(tokens)
	}
	for i, token := range tokens {
		if frameWord(token, ",", "and", "but") && rhetoricalRelation(tokens[:i]) {
			return i
		}
	}
	return 0
}

func rhetoricalRelation(tokens []document.Token) bool {
	if !affirmativeRelation(tokens) {
		return false
	}
	for i := 1; i < min(len(tokens)-2, 18); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") {
			continue
		}
		subject, predicate := tokens[:i], tokens[i+1:]
		if purposeRelationSubject(subject) && (purposeComplement(predicate) || relativePurpose(predicate)) {
			return true
		}
		if discourseSubject(subject) && abstractEvaluation(predicate) {
			return true
		}
	}
	return false
}

func purposeRelationSubject(tokens []document.Token) bool {
	return purposeSubject(tokens) || gerundAnaphor(tokens) && affirmativeRelation(tokens)
}

func affirmativeRelation(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "not", "n't", "never", "no", "neither", "nor", "only", "must", "should", "could", "would") {
			return false
		}
	}
	return true
}

func discourseSubject(tokens []document.Token) bool {
	if len(tokens) == 0 || frameWord(tokens[0], "and", "but", "or", "so") {
		return false
	}
	if deicticEvaluationSubject(tokens) {
		return true
	}
	if ordinalAnaphor(tokens) {
		return true
	}
	return len(tokens) >= 2 && positiveRhetoricSubject(tokens) && discourseNoun(tokens[len(tokens)-1])
}

func ordinalAnaphor(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "the") &&
		frameWord(tokens[1], "first", "last", "second", "third") && frameWord(tokens[2], "of") &&
		frameWord(tokens[3], "these", "those", "them")
}

func discourseNoun(token document.Token) bool {
	return frameWord(token, "distinction", "difference", "choice", "decision", "entry", "example", "explanation",
		"paragraph", "argument", "point", "row", "case", "result", "outcome")
}

func relativePurpose(tokens []document.Token) bool {
	if len(tokens) > 0 && frameWord(tokens[0], "exactly", "precisely") {
		tokens = tokens[1:]
	}
	if len(tokens) < 6 || !frameWord(tokens[0], "the") ||
		!frameWord(tokens[1], "question", "review", "incompatibility", "failure", "problem", "reason", "purpose") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "that", "which") {
		i++
	}
	return featureExistence(tokens[i:])
}

func featureExistence(tokens []document.Token) bool {
	i := 0
	if len(tokens) < 4 || !frameWord(tokens[i], "the", "this", "that") ||
		!frameWord(tokens[i+1], "feature", "rule", "verb", "flag", "baseline", "check") ||
		!frameWord(tokens[i+2], "exists", "existed") {
		return false
	}
	return existenceAction(tokens[i+3:])
}

func existenceAction(rest []document.Token) bool {
	return len(rest) == 1 && frameWord(rest[0], "for") || len(rest) == 2 && frameWord(rest[0], "to") &&
		frameWord(rest[1], "express", "prevent", "answer", "address", "capture", "perform")
}

func abstractEvaluation(tokens []document.Token) bool {
	if wholeAbstractValue(tokens) {
		return true
	}
	if len(tokens) >= 3 && frameWord(tokens[0], "the") && frameWord(tokens[1], "one") {
		tokens = tokens[2:]
		if frameWord(tokens[0], "that", "which") {
			tokens = tokens[1:]
		}
	} else if len(tokens) >= 2 && frameWord(tokens[0], "what") {
		tokens = tokens[1:]
	}
	return importancePredicate(tokens) || editorialIntegrity(tokens)
}

func importancePredicate(tokens []document.Token) bool {
	return len(tokens) >= 1 && frameWord(tokens[0], "matters", "counts") &&
		(len(tokens) == 1 || len(tokens) == 2 && frameWord(tokens[1], "most", "here"))
}

func editorialIntegrity(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "keeps", "kept") && frameWord(tokens[1], "the", "this", "that") &&
		discourseNoun(tokens[2]) && frameWord(tokens[3], "honest", "coherent")
}

// A noun modified by worth + cognitive gerund announces how information should
// be received. Numerical context before that phrase is not part of the notice.
func nominalNotice(c frameClause) (rhetoricalFrame, bool) {
	if quotedClaim(c.sentence.Tokens) || !affirmativeRelation(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 1; i+1 < len(tokens); i++ {
		if !informationWorth(tokens[i-1 : i+2]) {
			continue
		}
		start := i - 1
		if start > 0 && frameWord(tokens[start-1], "a", "an", "the", "this", "that") {
			start--
		}
		candidate := c
		candidate.start += start
		candidate.end = c.start + i + 2
		if candidate.eligible() && !candidateRhetoricScoped(candidate) {
			return localFrame(candidate, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func informationWorth(tokens []document.Token) bool {
	return informationNoun(tokens[0]) && frameWord(tokens[1], "worth") && cognitiveAct(tokens[2])
}

func informationNoun(token document.Token) bool {
	return frameWord(token, "decision", "distinction", "difference", "fact", "detail", "observation", "point", "information")
}

func cognitiveAct(token document.Token) bool {
	return frameWord(token, "knowing", "noting", "remembering", "understanding", "mentioning", "stating", "saying")
}
