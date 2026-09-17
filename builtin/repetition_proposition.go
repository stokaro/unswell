package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

func gerundIdentity(subject, rest []document.Token) bool {
	if !eligibleGerundIdentity(subject) || !eligibleGerundIdentity(rest) {
		return false
	}
	if sameRestrictionWords(subject, rest) {
		return true
	}
	// Both retains its local plural object. Added adjectives, quantities or
	// different objects cannot be dropped to manufacture equal propositions.
	return len(subject) == 3 && len(rest) == 2 && subject[2].Tag == "NNS" &&
		frameWord(subject[1], "both") && sameRestrictionWords(subject[:2], rest)
}

func eligibleGerundIdentity(tokens []document.Token) bool {
	return len(tokens) >= 2 && len(tokens) <= 7 && proseTokens(tokens) &&
		tokens[0].Tag == "VBG" && nominalSubject(tokens[1:])
}

func circularReason(left, cause []document.Token) bool {
	if len(left) < 4 || len(cause) < 3 || !proseTokens(cause) || !reasonJudgment(left) || !sameCircularJudgment(left, cause) {
		return false
	}
	subject := withoutArticle(cause[:len(cause)-2])
	if len(subject) == 0 || !frameWord(subject[0], "optimization", "optimizations", "implementation", "design", "approach", "method") {
		return false
	}
	return !slices.ContainsFunc(subject, func(t document.Token) bool {
		return t.Tag == "CD" || t.Protected || frameWord(t, "not", "no", "never", "only", "if", "when", "unless", "except",
			"because", "before", "after", "until", "without", "must", "may", "might", "should")
	})
}

func sameCircularJudgment(left, cause []document.Token) bool {
	last := cause[len(cause)-1]
	return strings.HasPrefix(last.Tag, "JJ") && last.Normal == left[len(left)-1].Normal &&
		frameWord(cause[len(cause)-2], "is", "are", "was", "were")
}

func reasonJudgment(tokens []document.Token) bool {
	end := len(tokens) - 2
	if frameWord(tokens[end], "quite", "very", "rather", "extremely") {
		end--
	}
	return end > 0 && reasonSubject(withoutArticle(tokens[:end]))
}
