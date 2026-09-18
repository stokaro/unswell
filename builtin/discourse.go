package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// discourseEvaluation requires a detached cue and the proposition it introduces.
// The cue is not construction vocabulary when protected or quoted. Conditions
// belong to the proposition and stay in the finding and editing advice.
func discourseEvaluation(c frameClause) bool {
	if !c.eligible() || rhetoricAttributed(c) || rhetoricQuoted(c) {
		return false
	}
	tokens := c.tokens()
	if end := detachedEvaluationEnd(tokens); end > 0 {
		return discourseProposition(tokens[end:])
	}
	if end := readerInferenceEnd(tokens); end > 0 {
		return discourseProposition(tokens[end:])
	}
	return false
}

func detachedEvaluationEnd(tokens []document.Token) int {
	if len(tokens) > 3 && frameWord(tokens[0], "interestingly", "surprisingly", "thankfully",
		"fortunately", "unfortunately", "obviously", "unsurprisingly", "remarkably") && frameWord(tokens[1], ",") {
		return 2
	}
	if len(tokens) > 4 && matches(tokens[:3], []string{"of", "course", ","}) {
		return 3
	}
	return 0
}

func readerInferenceEnd(tokens []document.Token) int {
	if len(tokens) < 6 || !matches(tokens[:2], []string{"as", "you"}) {
		return 0
	}
	i := 2
	if frameWord(tokens[i], "can") {
		i++
	}
	if !frameWord(tokens[i], "see", "infer", "gather", "know", "remember") {
		return 0
	}
	i++
	if frameWord(tokens[i], ",") {
		return i + 1
	}
	if !frameWord(tokens[i], "from") {
		return 0
	}
	for end := i + 2; end < min(len(tokens)-2, i+18); end++ {
		if frameWord(tokens[end], ",") {
			if nominalSubject(tokens[i+1 : end]) {
				return end + 1
			}
			return 0
		}
	}
	return 0
}

// A nominal subject followed by a finite predicate distinguishes a detached
// evaluation from a heading, a standalone reply, or an adverb of manner.
func discourseProposition(tokens []document.Token) bool {
	if len(tokens) < 3 || !qualitySubjectStart(tokens[0]) {
		return false
	}
	for verb := 1; verb < min(len(tokens)-1, 16); verb++ {
		token := tokens[verb]
		if token.Protected || !nominalSubject(tokens[:verb]) {
			continue
		}
		if discourseFinite(token, tokens[verb+1]) {
			return true
		}
	}
	return false
}

func discourseFinite(token, next document.Token) bool {
	if token.Tag == "VBP" || token.Tag == "VBZ" || token.Tag == "VBD" {
		return true
	}
	return token.Tag == "MD" && !next.Protected && strings.HasPrefix(next.Tag, "VB")
}
