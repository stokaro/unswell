package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// qualitativeAssurance requires a copular quality predicate, not an adjective
// inside a technical noun phrase. It requests scope, not proof of falsehood.
func qualitativeAssurance(c frameClause) (rhetoricalFrame, bool) {
	if contextualRhetoricScoped(c) || instructionGuard(c.tokens()) || question(c.sentence) || qualityMechanism(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		candidate, ok := localRhetoricCandidate(c, i)
		if ok && qualityClauseStart(c.tokens(), i) && qualityJudgment(candidate.tokens()) {
			return localFrame(candidate, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func qualityJudgment(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-1, 16); i++ {
		if !qualitySubject(tokens[:i]) {
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
	for i < min(len(tokens), 3) && qualityModifier(tokens[i]) {
		i++
	}
	if i >= len(tokens) || !qualitativeAdjective(tokens[i]) {
		return false
	}
	return qualityContinuation(tokens[i+1:])
}

func qualityModifier(token document.Token) bool {
	return frameWord(token, "also", "now", "very", "much", "fairly", "quite", "rather", "particularly",
		"extremely", "incredibly", "exceptionally", "remarkably", "surprisingly", "amazingly", "ridiculously")
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
		"efficient", "inefficient", "expensive", "cheap", "intuitive", "effortless", "useful", "helpful", "valuable")
}

func qualityMechanism(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "because", "since", "by", "through", "due", "thanks", "as") {
			return true
		}
	}
	return false
}

// A quality predicate needs its own subject. Conjunctions and a relative
// restriction inside an imperative cannot stand in for that subject.
func qualitySubject(tokens []document.Token) bool {
	if len(tokens) == 0 || !positiveRhetoricSubject(tokens) && !qualityWaySubject(tokens) {
		return false
	}
	if !qualitySubjectStart(tokens[0]) {
		return false
	}
	for _, token := range tokens[1:] {
		if frameWord(token, "that", "which", "who", "whom", "whose") {
			return false
		}
	}
	return true
}

// An introductory goal does not scope an ease judgment. Only an explicit
// infinitive with an operand permits a later main subject; finite clauses and
// relative restrictions remain boundaries. The goal stays out of the finding.
func qualityClauseStart(tokens []document.Token, at int) bool {
	if embeddedClauseStart(tokens, at) {
		return true
	}
	if at < 3 || at > 18 || !frameWord(tokens[0], "to") {
		return false
	}
	prefix := tokens[1:at]
	if frameWord(prefix[len(prefix)-1], ",") {
		prefix = prefix[:len(prefix)-1]
	}
	return len(prefix) > 1 && goalObjectEnd(prefix[len(prefix)-1]) &&
		grammaticalAction(prefix) && nominalSubject(prefix[1:])
}

func qualitySubjectStart(token document.Token) bool {
	return token.Protected || strings.HasPrefix(token.Tag, "NN") || token.Tag == "VBG" ||
		frameWord(token, "a", "an", "the", "this", "that", "these", "those", "it", "they", "our", "your", "its")
}

// A goal must end in an object before a new main subject begins. A determiner,
// particle or unfinished gerund cannot supply that boundary.
func goalObjectEnd(token document.Token) bool {
	return token.Protected || strings.HasPrefix(token.Tag, "NN") || token.Tag == "PRP"
}

func qualityWaySubject(tokens []document.Token) bool {
	if len(tokens) < 4 || !frameWord(tokens[0], "the", "this", "that") || !frameWord(tokens[1], "way") {
		return false
	}
	for verb := 3; verb < len(tokens); verb++ {
		token := tokens[verb]
		if token.Protected || token.Tag != "VBP" && token.Tag != "VBZ" && token.Tag != "VBD" {
			continue
		}
		return positiveRhetoricSubject(tokens[2:verb]) && nominalSubject(tokens[verb+1:])
	}
	return false
}
