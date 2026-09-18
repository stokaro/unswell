package builtin

import "github.com/stokaro/unswell/document"

func redundantSupport(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for i := range tokens {
		if duplicateSequence(tokens, i) || duplicateExample(tokens, i) || additiveSupport(tokens, i) {
			return localFrame(c, 0)
		}
	}
	return rhetoricalFrame{}, false
}

func duplicateSequence(tokens []document.Token, i int) bool {
	return i+2 < len(tokens) && frameWord(tokens[i], "then") &&
		frameWord(tokens[i+1], "finally") && grammaticalAction(tokens[i+2:])
}

func duplicateExample(tokens []document.Token, i int) bool {
	if i == 0 || i+3 >= len(tokens) || !frameWord(tokens[i], "like") || !methodNoun(tokens[i-1]) {
		return false
	}
	i++
	if frameWord(tokens[i], ",") {
		i++
	}
	return i+2 < len(tokens) && frameWord(tokens[i], "for") && frameWord(tokens[i+1], "example") &&
		!frameWord(tokens[i+2], "when", "if", "how")
}

func additiveSupport(tokens []document.Token, i int) bool {
	if !frameWord(tokens[i], "also") || supportBoundary(tokens) {
		return false
	}
	end := len(tokens)
	if i+3 < end && frameWord(tokens[end-2], "as") && frameWord(tokens[end-1], "well") {
		return true
	}
	return i > 4 && matches(tokens[:3], []string{"in", "addition", "to"})
}

// Additional predicates can give the two markers different scopes. Keep them
// outside this bounded rule even when a more detailed parse could relate them.
func supportBoundary(tokens []document.Token) bool {
	verbs := 0
	for i, token := range tokens {
		if separateSupportAction(tokens[i:]) {
			return true
		}
		if frameWord(token, "and", "but", "or", "while", "whereas", "if", "when", "unless",
			"although", "because", "not", "n't", "never", "which", "that") {
			return true
		}
		if finiteWord(token) {
			verbs++
		}
	}
	return verbs != 1
}

func separateSupportAction(tokens []document.Token) bool {
	return len(tokens) > 1 && frameWord(tokens[0], "to") && projectionVerb(tokens[1])
}

func finiteWord(token document.Token) bool {
	return !token.Protected && token.Word &&
		(token.Tag == "VBZ" || token.Tag == "VBP" || token.Tag == "VBD" || token.Tag == "MD")
}
