package builtin

import "github.com/stokaro/unswell/document"

// discourseSelection binds an importance or attention predicate to the
// information it selects. Operational complements are not selection judgments.
func discourseSelection(tokens []document.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	if frameWord(tokens[0], "exactly", "precisely") {
		tokens = tokens[1:]
	}
	if len(tokens) >= 3 && frameWord(tokens[0], "the") &&
		frameWord(tokens[1], "one", "ones", "part", "point", "detail", "thing") {
		tokens = tokens[2:]
		if frameWord(tokens[0], "that", "which") {
			tokens = tokens[1:]
		}
	}
	return importancePredicate(tokens) || cognitiveInformation(tokens) || readerMisses(tokens) || importanceQuality(tokens)
}

func readerMisses(tokens []document.Token) bool {
	return len(tokens) == 2 && frameWord(tokens[0], "people", "readers", "users") && frameWord(tokens[1], "miss", "overlook")
}

func importanceQuality(tokens []document.Token) bool {
	if len(tokens) > 1 && frameWord(tokens[0], "very", "particularly", "especially", "most") {
		tokens = tokens[1:]
	}
	if len(tokens) == 0 || !frameWord(tokens[0], "important", "significant", "noteworthy") {
		return false
	}
	rest := tokens[1:]
	return importanceAudience(rest) || len(rest) == 2 && frameWord(rest[0], "to") &&
		frameWord(rest[1], "know", "note", "remember", "understand")
}

func abstractPurposeObject(token document.Token) bool {
	return frameWord(token, "feature", "finding", "refusal", "verb", "distinction", "difference", "rule",
		"signature", "surface", "check", "criterion", "requirement", "question", "command")
}

// functionalCleft requires an anaphoric or gerund action subject and an abstract
// purpose object. A component's concrete function is not a rhetorical closure.
func functionalCleft(tokens []document.Token) bool {
	for i := 1; i+5 < min(len(tokens), 24); i++ {
		if !frameWord(tokens[i], "is", "was") || !discourseAction(tokens[:i]) {
			continue
		}
		return namedPurpose(tokens[i+1:])
	}
	return false
}

func discourseAction(tokens []document.Token) bool {
	if deicticEvaluationSubject(tokens) {
		return true
	}
	return len(tokens) >= 2 && len(tokens) <= 14 && !tokens[0].Protected && tokens[0].Tag == "VBG" &&
		nominalSubject(tokens) && affirmativeRelation(tokens)
}

func authorNotice(tokens []document.Token) bool {
	i := authorIntention(tokens)
	if i == 0 || i >= len(tokens) {
		return false
	}
	if i+2 < len(tokens) && frameWord(tokens[i], "point") && frameWord(tokens[i+1], "out") {
		i++
	} else if !frameWord(tokens[i], "stress", "emphasize", "note", "mention") {
		return false
	}
	return i+1 < len(tokens) && frameWord(tokens[i+1], "that")
}

func authorIntention(tokens []document.Token) int {
	if len(tokens) < 4 || !frameWord(tokens[0], "i", "we") {
		return 0
	}
	switch {
	case frameWord(tokens[1], "want", "need", "wish") && frameWord(tokens[2], "to"):
		return 3
	case frameWord(tokens[1], "would") && frameWord(tokens[2], "like") && frameWord(tokens[3], "to"):
		return 4
	default:
		return 0
	}
}

func documentEndorsement(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "i", "we") || !frameWord(tokens[1], "hope", "trust") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "that") {
		i++
	}
	end := documentSubjectEnd(tokens[i:])
	if end == 0 {
		return false
	}
	i += end
	if i < len(tokens) && frameWord(tokens[i], "has", "have") {
		i++
	}
	return i+1 < len(tokens) && frameWord(tokens[i], "helped", "helps", "help", "clarifies", "clarified") &&
		frameWord(tokens[i+1], "you", "readers", "users")
}
