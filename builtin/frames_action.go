package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// Action reframing observes the same negative-to-positive movement as the
// copular frame. It does not infer that the two propositions are redundant.
func actionReframing(a, b frameClause) (rhetoricalFrame, bool) {
	if rhetoricQuoted(a) || rhetoricQuoted(b) || rhetoricAttributed(a) || rhetoricAttributed(b) {
		return rhetoricalFrame{}, false
	}
	left, negative := actionSubject(a.tokens(), true)
	right, positive := actionSubject(b.tokens(), false)
	if !negative || !positive || !linkedSubjects(left, right) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{a, b}}, true
}

// A negative action needs written do-support; the affirmative predicate needs
// a finite verb. Subjects stay bounded and cannot contain a relative clause.
func actionSubject(tokens []document.Token, negative bool) ([]document.Token, bool) {
	for i := 1; i < min(len(tokens)-1, 9); i++ {
		if !nominalSubject(tokens[:i]) {
			continue
		}
		if negative && negativeAction(tokens[i:]) {
			return tokens[:i], true
		}
		if !negative && finiteAction(tokens[:i], tokens[i:]) {
			return tokens[:i], true
		}
	}
	return nil, false
}

func negativeAction(tokens []document.Token) bool {
	i := 0
	switch {
	case len(tokens) >= 4 && frameWord(tokens[0], "does", "do", "did") && frameWord(tokens[1], "not", "n't"):
		i = 2
	case len(tokens) >= 3 && frameWord(tokens[0], "doesn't", "don't", "didn't"):
		i = 1
	default:
		return false
	}
	return !tokens[i].Protected && tokens[i].Tag == "VB" && !frameWord(tokens[i], "be", "have")
}

func finiteAction(subject, tokens []document.Token) bool {
	if len(tokens) < 2 || tokens[0].Protected ||
		frameWord(tokens[0], "do", "does", "did", "is", "are", "was", "were") {
		return false
	}
	verb := tokens[0]
	finite := strings.HasPrefix(verb.Tag, "VB") && verb.Tag != "VBG" && verb.Tag != "VBN"
	// The pinned tagger can label stores as NNS after it. A singular pronoun
	// and a determiner-led object resolve that particular written ambiguity.
	finite = finite || pronounObjectAction(subject, tokens)
	return finite && !frameWord(tokens[1], "not", "n't")
}

func pronounObjectAction(subject, tokens []document.Token) bool {
	return len(subject) == 1 && frameWord(subject[0], "it", "this", "that") && len(tokens) > 2 &&
		tokens[0].Tag == "NNS" && strings.HasSuffix(tokens[0].Normal, "s") &&
		frameWord(tokens[1], "a", "an", "the", "its", "their")
}
