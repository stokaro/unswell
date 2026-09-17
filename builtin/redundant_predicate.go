package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func redundantPredicate(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() || quotedClaim(c.sentence.Tokens) || instructionGuard(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	if metadataPredicate(tokens) {
		return localFrame(c, 0)
	}
	for i := 1; i < min(len(tokens)-3, 18); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") {
			continue
		}
		if predicateRepeats(relationSubject(tokens[:i]), tokens[i+1:]) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func metadataPredicate(tokens []document.Token) bool {
	tokens = withoutArticle(tokens)
	return len(tokens) == 5 && frameWord(tokens[0], "category", "type") &&
		frameWord(tokens[1], "label") && frameWord(tokens[2], "describes", "identifies", "names") &&
		frameWord(tokens[3], "the", "a") && frameWord(tokens[4], tokens[0].Normal)
}

func predicateRepeats(head string, rest []document.Token) bool {
	switch head {
	case "path", "location", "locations":
		return frameWord(rest[0], "located", "situated") && frameWord(rest[1], "at", "in")
	case "reason":
		return frameWord(rest[0], "because")
	default:
		return false
	}
}

// A relation must be the subject head, not a noun inside a file description or
// a control-flow condition. Modifiers after for/of may identify an opaque operand.
func relationSubject(tokens []document.Token) string {
	head := ""
	modifier := false
	for _, token := range tokens {
		if frameWord(token, "for", "of") && head != "" {
			modifier = true
			continue
		}
		if modifier {
			if !token.Protected && !relationNominal(token) {
				return ""
			}
			continue
		}
		if token.Protected || !relationNominal(token) {
			return ""
		}
		if strings.HasPrefix(token.Tag, "NN") {
			head = token.Normal
		}
	}
	return head
}

func relationNominal(token document.Token) bool {
	return strings.HasPrefix(token.Tag, "NN") || strings.HasPrefix(token.Tag, "JJ") ||
		frameWord(token, "the", "a", "an", "this", "that")
}
