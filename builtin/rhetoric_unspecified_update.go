package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// unspecifiedUpdate identifies a complete update announcement that names only
// an indefinite asset set and an unspecified latest version. A second clause
// or sentence can carry the missing scope, so neither is treated as a match.
func unspecifiedUpdate(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	if len(clauses) != 1 || index != 0 {
		return rhetoricalFrame{}, false
	}
	c := clauses[index]
	if !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) || rhetoricCondition(c) {
		return rhetoricalFrame{}, false
	}
	if !unspecifiedUpdateShape(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	frame, ok := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message: "This update announcement leaves the affected assets and destination version unspecified.",
		suggestion: "Name the affected assets and destination version, or link the exact update record. " +
			"Preserve any known scope and avoid inventing a version.",
	}
	return frame, ok
}

func unspecifiedUpdateShape(tokens []document.Token) bool {
	verb, ok := unspecifiedUpdateTail(tokens)
	if !ok {
		return false
	}
	if verb > 0 && frameWord(tokens[verb-1], "been") {
		verb--
	}
	if verb < 3 || !frameWord(tokens[verb-1], "have", "has", "are", "were") {
		return false
	}
	subject := tokens[1 : verb-1]
	return len(subject) > 0 && len(subject) <= 8 && unspecifiedUpdateSubject(subject)
}

func unspecifiedUpdateTail(tokens []document.Token) (int, bool) {
	if len(tokens) < 8 || len(tokens) > 18 {
		return 0, false
	}
	if !frameWord(tokens[0], "some", "several", "various") {
		return 0, false
	}
	last := len(tokens) - 1
	if !frameWord(tokens[last-3], "to") || !frameWord(tokens[last-2], "the") ||
		!frameWord(tokens[last-1], "latest") || !frameWord(tokens[last], "version", "versions") {
		return 0, false
	}
	verb := last - 4
	if !frameWord(tokens[verb], "updated", "upgraded") {
		return 0, false
	}
	return verb, true
}

func unspecifiedUpdateSubject(tokens []document.Token) bool {
	nouns := 0
	for _, token := range tokens {
		if token.Protected || !token.Word || token.Tag == "CD" {
			return false
		}
		if strings.HasPrefix(token.Tag, "NN") {
			nouns++
		}
	}
	return nouns > 0
}
