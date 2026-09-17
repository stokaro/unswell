package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

func explanatoryRestart(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() || quotedClaim(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 2; i+6 < len(tokens); i++ {
		if !frameWord(tokens[i], ",") || !frameWord(tokens[i+1], "and") || !frameWord(tokens[i+5], "because") {
			continue
		}
		if !restartPredicate(tokens[:i], tokens[i+2:i+5]) {
			continue
		}
		left, right := c, c
		left.end = c.start + i
		right.start, right.end = c.start+i+2, c.start+i+5
		return rhetoricalFrame{parts: []frameClause{left, right}}, true
	}
	return rhetoricalFrame{}, false
}

func restartPredicate(left, right []document.Token) bool {
	if len(left) < 3 || len(right) != 3 || !proseTokens(left) || !proseTokens(right) {
		return false
	}
	adjective := left[len(left)-1]
	if !strings.HasPrefix(adjective.Tag, "JJ") || adjective.Normal != right[2].Normal {
		return false
	}
	copula := len(left) - 2
	if frameWord(left[copula], "quite", "very", "rather", "extremely") {
		copula--
	}
	if !sameRestartCopula(left, copula, right[1]) {
		return false
	}
	return restartAntecedent(left[:copula], right[0])
}

func sameRestartCopula(left []document.Token, index int, right document.Token) bool {
	return index > 0 && frameWord(left[index], "is", "are", "was", "were") && left[index].Normal == right.Normal
}

func restartAntecedent(subject []document.Token, pronoun document.Token) bool {
	if !eligibleRestartSubject(subject) {
		return false
	}
	if frameWord(subject[0], "the", "this", "these", "a", "an") {
		subject = subject[1:]
	}
	if len(subject) == 0 || !strings.HasPrefix(subject[0].Tag, "NN") {
		return false
	}
	if !restartPronoun(subject[0], pronoun) {
		return false
	}
	// The explanatory reason subject can contain its own copula. Other noun
	// modifiers must remain nominal and may not introduce another antecedent.
	tail := subject[1:]
	if reasonSubject(subject) {
		return true
	}
	return nominalSubject(subject) && !slices.ContainsFunc(tail, func(t document.Token) bool {
		return strings.HasPrefix(t.Tag, "NN")
	})
}

func eligibleRestartSubject(subject []document.Token) bool {
	return len(subject) > 0 && len(subject) <= 12 && !slices.ContainsFunc(subject, func(t document.Token) bool {
		return t.Tag == "CD" || frameWord(t, "not", "no", "never", "if", "when", "unless", "except", "only")
	})
}

func restartPronoun(head, pronoun document.Token) bool {
	if head.Tag == "NNS" || head.Tag == "NNPS" {
		return frameWord(pronoun, "they")
	}
	return frameWord(pronoun, "it")
}

func reasonSubject(subject []document.Token) bool {
	tail := subject[1:]
	return frameWord(subject[0], "reason", "reasons") &&
		(len(tail) == 4 && matches(tail, []string{"for", "why", "this", "is"}) ||
			len(tail) == 3 && matches(tail, []string{"why", "this", "is"}))
}
