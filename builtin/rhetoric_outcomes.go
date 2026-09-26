package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// outcomeAnnouncement links a generic possibility announcement to two adjacent
// conditional branches. It does not treat an isolated list introduction as empty.
func outcomeAnnouncement(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if index+2 >= len(clauses) || !outcomeOpening(c) || !genericOutcomes(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	first, second := clauses[index+1], clauses[index+2]
	if first.ordinal != c.ordinal+1 || second.ordinal != first.ordinal+1 ||
		c.sentence.BlockID != first.sentence.BlockID || first.sentence.BlockID != second.sentence.BlockID ||
		!conditionalOutcome(first) || !conditionalOutcome(second) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{c, first, second}}, true
}

func genericOutcomes(tokens []document.Token) bool {
	if len(tokens) < 4 || !matches(tokens[:2], []string{"there", "are"}) {
		return false
	}
	rest := outcomeQuantity(tokens[2:])
	if len(rest) > 1 && frameWord(rest[0], "possible", "different") {
		rest = rest[1:]
	}
	if len(rest) == 0 || !frameWord(rest[0], "outcomes", "possibilities", "cases", "ways") {
		return false
	}
	return outcomeReference(rest[1:])
}

func outcomeQuantity(tokens []document.Token) []document.Token {
	switch {
	case len(tokens) >= 4 && frameWord(tokens[0], "a") && frameWord(tokens[1], "couple", "number") && frameWord(tokens[2], "of"):
		return tokens[3:]
	case frameWord(tokens[0], "several", "various", "multiple", "some"):
		return tokens[1:]
	default:
		return nil
	}
}

func outcomeReference(tokens []document.Token) bool {
	if len(tokens) == 0 {
		return true
	}
	if frameWord(tokens[0], "how") {
		tokens = tokens[1:]
	}
	if len(tokens) < 4 || !frameWord(tokens[0], "this", "that") {
		return false
	}
	at := outcomeSubjectEnd(tokens)
	return at != 1 && outcomeEnd(tokens[at:])
}

func outcomeEnd(tokens []document.Token) bool {
	if len(tokens) < 2 || !frameWord(tokens[0], "may", "can") || !frameWord(tokens[1], "end") {
		return false
	}
	return len(tokens) == 2 || len(tokens) == 3 && frameWord(tokens[2], "up")
}

func outcomeSubjectEnd(tokens []document.Token) int {
	at := 1
	for at < len(tokens) && at <= 2 && !tokens[at].Protected && strings.HasPrefix(tokens[at].Tag, "NN") {
		at++
	}
	return at
}

func conditionalOutcome(c frameClause) bool {
	if !outcomeOpening(c) || !frameWord(c.tokens()[0], "if") {
		return false
	}
	tokens := c.tokens()
	for at := 3; at+2 < len(tokens); at++ {
		if frameWord(tokens[at], ",") {
			return outcomeFinite(tokens[1:at]) &&
				(outcomeFinite(tokens[at+1:]) || grammaticalAction(tokens[at+1:]))
		}
	}
	return unpunctuatedOutcome(tokens)
}

func outcomeFinite(tokens []document.Token) bool {
	for at, token := range tokens {
		if !token.Protected && (outcomePredicate(tokens, at) || outcomeMatchPredicate(tokens, at)) {
			return true
		}
	}
	return false
}

// The POS backend can tag "cache matches/misses" as two nouns. These bounded
// written predicates follow a noun subject; this does not reinterpret
// arbitrary plural nouns or change the shared NLP representation.
func outcomeMatchPredicate(tokens []document.Token, at int) bool {
	return at > 0 && frameWord(tokens[at], "matches", "misses") &&
		!tokens[at-1].Protected && strings.HasPrefix(tokens[at-1].Tag, "NN")
}

// In an unpunctuated conditional, require two finite predicates without a
// coordinator joining the final pair. Modal complements are one predicate.
// This is a surface safeguard, not a dependency parse.
func unpunctuatedOutcome(tokens []document.Token) bool {
	previous, last := -1, -1
	for at, token := range tokens {
		if token.Protected || !outcomePredicate(tokens, at) {
			continue
		}
		previous, last = last, at
	}
	if previous < 0 {
		return false
	}
	for _, token := range tokens[previous+1 : last] {
		if frameWord(token, "and", "or", "but", "if", "unless", "while", "when", "because", "that", "which") {
			return false
		}
	}
	return true
}

func outcomePredicate(tokens []document.Token, at int) bool {
	switch tokens[at].Tag {
	case "MD", "VBP", "VBZ", "VBD":
		return true
	case "VB":
		return at > 0 && (strings.HasPrefix(tokens[at-1].Tag, "NN") || tokens[at-1].Tag == "PRP")
	default:
		return false
	}
}

func outcomeOpening(c frameClause) bool {
	return c.eligible() && c.start == 0 && len(frameClauses(c.sentence, 0)) == 1 &&
		!rhetoricQuoted(c) && !rhetoricAttributed(c)
}
