package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

func definitionEcho(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 1; i < min(len(tokens), 9); i++ {
		if frameWord(tokens[i], "is", "are", "was", "were") && !quotedClaim(c.sentence.Tokens) {
			if end := persistentIdentity(tokens[:i], tokens[i+1:]); end > 0 {
				c.end = c.start + i + 1 + end
				return localFrame(c, 0)
			}
		}
		if frameWord(tokens[i], "is", "are", "was", "were") && echoedDefinition(tokens[:i], tokens[i+1:]) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func persistentIdentity(subject, rest []document.Token) int {
	if len(rest) < 3 || !frameWord(rest[0], "still") {
		return 0
	}
	end := len(rest)
	for i := 1; i < len(rest); i++ {
		if frameWord(rest[i], ",") {
			end = i
			break
		}
	}
	if nominalEcho(withoutArticle(subject), withoutArticle(rest[1:end])) {
		return end
	}
	return 0
}

func echoedDefinition(subject, rest []document.Token) bool {
	for i := 1; i < min(len(rest)-2, 9); i++ {
		if !frameWord(rest[i], ",") || !frameWord(rest[i+1], "not") || !frameWord(rest[i+2], "a", "an", "the") {
			continue
		}
		a, b := withoutArticle(subject), withoutArticle(rest[:i])
		return nominalEcho(a, b) && len(rest[i+3:]) > 0
	}
	return false
}

func nominalEcho(a, b []document.Token) bool {
	return len(a) > 0 && strings.HasPrefix(a[len(a)-1].Tag, "NN") && nominalSubject(a) && proseTokens(a) &&
		proseTokens(b) && slices.EqualFunc(a, b, func(x, y document.Token) bool { return x.Normal == y.Normal })
}

func withoutArticle(tokens []document.Token) []document.Token {
	if len(tokens) > 0 && frameWord(tokens[0], "a", "an", "the") {
		return tokens[1:]
	}
	return tokens
}

func sloganContrast(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := 2; i+2 < len(tokens); i++ {
		if sloganAlternatives(tokens, i) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func sloganAlternatives(tokens []document.Token, i int) bool {
	pair := (frameWord(tokens[i], "rather") && frameWord(tokens[i+1], "than")) ||
		(frameWord(tokens[i], "instead") && frameWord(tokens[i+1], "of"))
	if !pair || i+3 != len(tokens) {
		return false
	}
	end := i
	if frameWord(tokens[end-1], ",") {
		end--
	}
	if end < 2 || !abstractVerification(tokens[end-1], tokens[i+2]) {
		return false
	}
	return sloganSubject(tokens[:end-1])
}

func sloganSubject(subject []document.Token) bool {
	if frameWord(subject[0], "so", "thus") {
		subject = subject[1:]
	}
	return len(subject) > 0 && len(subject) <= 5 && nominalSubject(subject) &&
		!slices.ContainsFunc(subject, func(t document.Token) bool {
			return !t.Protected && (strings.HasPrefix(t.Tag, "RB") || frameWord(t, "not", "no", "never", "only", "always"))
		})
}

func abstractVerification(left, right document.Token) bool {
	return (frameWord(left, "checks", "check", "verifies", "verify", "verified", "checked") &&
		frameWord(right, "trusts", "trust", "trusted", "assumes", "assume", "assumed")) ||
		(frameWord(left, "proves", "prove", "proved") && frameWord(right, "asserts", "assert", "asserted"))
}
