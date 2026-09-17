package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

func positiveRestriction(tokens []document.Token) (restriction, bool) {
	for i := 1; i+3 < len(tokens) && i < 10; i++ {
		if !frameWord(tokens[i+1], "only") || !restrictionPositiveVerb(tokens[i]) {
			continue
		}
		actor, tense := restrictionActorTense(tokens[:i], tokens[i])
		if !restrictionActor(actor) {
			continue
		}
		property, object, ok := qualifiedRestrictionObject(tokens[i+2:])
		if ok {
			return restriction{actor, tokens[i], property, object, tense}, true
		}
	}
	return restriction{}, false
}

// The matching negative clause supplies the verbal cue when the tagger labels
// a positive predicate such as process as a preposition. Lexical identity and
// the explicit only-object construction are still required.
func restrictionPositiveVerb(token document.Token) bool {
	return token.Word && !frameWord(token, "be", "is", "are", "have", "has", "do", "does", "not", "never", "only")
}

func restrictionVerb(token document.Token) bool {
	return slices.Contains([]string{"VB", "VBZ", "VBP"}, token.Tag) &&
		!frameWord(token, "be", "is", "are", "have", "has", "do", "does")
}

func restrictionActorTense(tokens []document.Token, verb document.Token) ([]document.Token, string) {
	if len(tokens) > 1 && frameWord(tokens[len(tokens)-1], "will") {
		return tokens[:len(tokens)-1], "will"
	}
	if verb.Tag == "VB" {
		return nil, ""
	}
	return tokens, "present"
}

func qualifiedRestrictionObject(tokens []document.Token) ([]document.Token, string, bool) {
	if len(tokens) < 2 || len(tokens) > 5 || !strings.HasPrefix(tokens[len(tokens)-1].Tag, "NN") {
		return nil, "", false
	}
	property := tokens[:len(tokens)-1]
	if !restrictionProperty(property) {
		return nil, "", false
	}
	return property, tokens[len(tokens)-1].Normal, true
}

func restrictionProperty(tokens []document.Token) bool {
	return len(tokens) > 0 && len(tokens) <= 4 && !slices.ContainsFunc(tokens, func(t document.Token) bool {
		return !slices.Contains([]string{"JJ", "VBN", "RB"}, t.Tag) || frameWord(t, "not", "never", "only", "also", "always")
	})
}

func negativeRestriction(tokens []document.Token) (restriction, bool) {
	for i := 1; i+4 < len(tokens) && i < 11; i++ {
		actor, verb, tense, start := negativeRestrictionHead(tokens, i)
		if start == 0 || !restrictionActor(actor) || !restrictionVerb(verb) {
			continue
		}
		property, object, ok := negativeRestrictionObject(tokens[start:])
		if ok {
			return restriction{actor, verb, property, object, tense}, true
		}
	}
	return restriction{}, false
}

func negativeRestrictionHead(tokens []document.Token, i int) ([]document.Token, document.Token, string, int) {
	if frameWord(tokens[i], "never") {
		actor, tense := restrictionActorTense(tokens[:i], tokens[i+1])
		return actor, tokens[i+1], tense, i + 2
	}
	if frameWord(tokens[i], "do", "does", "will") && frameWord(tokens[i+1], "not") {
		tense := "present"
		if frameWord(tokens[i], "will") {
			tense = "will"
		}
		return tokens[:i], tokens[i+2], tense, i + 3
	}
	return nil, document.Token{}, "", 0
}

func negativeRestrictionObject(tokens []document.Token) ([]document.Token, string, bool) {
	if len(tokens) >= 6 && matches(tokens[:2], []string{"on", "code"}) &&
		matches(tokens[2:5], []string{"that", "is", "not"}) {
		return licenseProperty(tokens[5:])
	}
	if len(tokens) >= 5 && strings.HasPrefix(tokens[0].Tag, "NN") && frameWord(tokens[1], "that", "which") &&
		frameWord(tokens[2], "is", "are") && frameWord(tokens[3], "not") && restrictionProperty(tokens[4:]) {
		return tokens[4:], tokens[0].Normal, true
	}
	return nil, "", false
}
