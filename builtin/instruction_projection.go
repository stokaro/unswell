package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// instructionProjection tracks a support chain ending in an action. These are
// bounded surface relations, not dependency edges or a license to rewrite text.
type instructionProjection struct {
	action, layers int
}

func projectedInstruction(c frameClause) bool {
	if !c.eligible() || len(c.sentence.Tokens) > 96 || quotedClaim(c.sentence.Tokens) ||
		projectionGuard(c.sentence.Tokens) {
		return false
	}
	tokens := c.tokens()
	for verb := 1; verb+3 < len(tokens); verb++ {
		if !projectionSubject(tokens[:verb]) {
			continue
		}
		p, ok := projectSupport(tokens, verb, 0)
		if ok && p.layers > 0 && projectionOperand(tokens[p.action:]) {
			return true
		}
	}
	return false
}

func projectionSubject(tokens []document.Token) bool {
	if len(tokens) > 2 && frameWord(tokens[0], "optionally", "also") && frameWord(tokens[1], ",") {
		tokens = tokens[2:]
	}
	if len(tokens) == 0 || len(tokens) > 12 || !nominalSubject(tokens) {
		return false
	}
	for _, token := range tokens {
		if projectionRelativeOrNumber(token, len(tokens)) {
			return false
		}
	}
	return projectionActorHead(tokens)
}

func projectionRelativeOrNumber(token document.Token, length int) bool {
	return !token.Protected && (token.Tag == "CD" || length > 1 && frameWord(token, "that", "which", "who", "whom"))
}

func projectionActorHead(tokens []document.Token) bool {
	last := tokens[len(tokens)-1]
	if len(tokens) > 1 && frameWord(last, "also") {
		last = tokens[len(tokens)-2]
	}
	// The tagger sometimes marks a noun such as "library" as an adjective.
	// A determiner anchors that nominal role; an issue number alone cannot.
	return last.Protected || strings.HasPrefix(last.Tag, "NN") ||
		last.Tag == "JJ" && frameWord(tokens[0], "the", "a", "an", "this", "that") ||
		frameWord(last, "this", "that", "it", "they", "these", "those")
}

func projectionGuard(tokens []document.Token) bool {
	if instructionGuard(tokens) {
		return true
	}
	for _, token := range tokens {
		if frameWord(token, "no", "says", "said", "reported", "according", "authorization", "authorized",
			"access", "credentials", "grant", "grants", "deny", "denies", "restricted", "hazard", "damage", "leak") ||
			(!token.Protected && strings.HasSuffix(token.Normal, "n't")) {
			return true
		}
	}
	return false
}

func projectSupport(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	if at+2 >= len(tokens) || layers >= 3 {
		return instructionProjection{}, false
	}
	switch {
	case frameWord(tokens[at], "has", "have"):
		return projectAbility(tokens, at+1, layers)
	case frameWord(tokens[at], "allows", "allow", "enables", "enable"):
		return projectEnabling(tokens, at+1, layers)
	case frameWord(tokens[at], "can", "may", "could"):
		if frameWord(tokens[at+1], "be") {
			return projectPassive(tokens, at+2, layers+1)
		}
	case frameWord(tokens[at], "is", "are", "be"):
		return projectPassive(tokens, at+1, layers)
	}
	return instructionProjection{}, false
}

func projectAbility(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	if at < len(tokens) && frameWord(tokens[at], "also") {
		at++
	}
	if at+3 >= len(tokens) || !frameWord(tokens[at], "the", "an", "a") ||
		!frameWord(tokens[at+1], "ability", "capability") || !frameWord(tokens[at+2], "to") {
		return instructionProjection{}, false
	}
	return projectAction(tokens, at+3, layers+1)
}

func projectPassive(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	if at+2 >= len(tokens) || !frameWord(tokens[at+1], "to") {
		return instructionProjection{}, false
	}
	if frameWord(tokens[at], "intended", "designed") {
		// A single purpose relation carries information; require another layer.
		return projectSupport(tokens, at+2, layers+1)
	}
	if layers > 0 && frameWord(tokens[at], "used") {
		return projectAction(tokens, at+2, layers+1)
	}
	return instructionProjection{}, false
}

func projectEnabling(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	readerEnd := genericInstructionReader(tokens, at)
	if readerEnd > at && readerEnd+1 < len(tokens) && frameWord(tokens[readerEnd], "to") {
		return projectAction(tokens, readerEnd+1, layers+1)
	}
	// Intended-to-allow plus a passive action has two support layers. A bare
	// allows + object + passive is a capability statement and stays a control.
	if layers == 0 {
		return instructionProjection{}, false
	}
	for end := at + 1; end+3 < min(len(tokens), at+13); end++ {
		if frameWord(tokens[end], "to") && frameWord(tokens[end+1], "be") &&
			nominalSubject(tokens[at:end]) && projectedParticiple(tokens[end+2]) {
			return instructionProjection{end + 2, layers + 1}, true
		}
	}
	return instructionProjection{}, false
}

func genericInstructionReader(tokens []document.Token, at int) int {
	if at < len(tokens) && frameWord(tokens[at], "the", "a") {
		at++
	}
	if at < len(tokens) && frameWord(tokens[at], "you", "user", "users", "reader", "readers") {
		return at + 1
	}
	return 0
}

func projectAction(tokens []document.Token, at, layers int) (instructionProjection, bool) {
	if at >= len(tokens) {
		return instructionProjection{}, false
	}
	if p, ok := projectSupport(tokens, at, layers); ok {
		return p, true
	}
	if projectionVerb(tokens[at]) && projectionOperand(tokens[at:]) {
		return instructionProjection{at, layers}, true
	}
	return instructionProjection{}, false
}

func projectionVerb(token document.Token) bool {
	if token.Protected || !token.Word || frameWord(token, "be", "have", "allow", "enable") {
		return false
	}
	// Infinitive syntax fixes the role; POS supplies open vocabulary. The
	// fallback covers common base forms the tagger sometimes treats as nouns.
	return token.Tag == "VB" || token.Tag == "VBP" || frameWord(token,
		"run", "query", "fetch", "read", "send", "pass", "reuse", "process", "sort", "register", "load", "save",
		"parse", "compute", "convert", "compare", "validate", "verify", "configure", "set", "specify", "provide")
}

func projectionOperand(tokens []document.Token) bool {
	for _, token := range tokens[1:] {
		if frameWord(token, "if", "when", "whenever", "until", "because", ",", "and", "or") {
			return false
		}
		if token.Protected || strings.HasPrefix(token.Tag, "NN") || frameWord(token, "it", "them", "this", "these") {
			return true
		}
	}
	return false
}

func projectedParticiple(token document.Token) bool {
	return !token.Protected && (token.Tag == "VBN" || frameWord(token,
		"passed", "specified", "configured", "provided", "stored", "read", "written", "selected", "created", "defined"))
}
