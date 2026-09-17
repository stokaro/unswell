package builtin

import "github.com/stokaro/unswell/document"

func documentEarnsPlace(tokens []document.Token) bool {
	for i := 1; i < min(len(tokens)-6, 8); i++ {
		if !frameWord(tokens[i], "is", "was") || !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		return len(rest) == 6 && frameWord(rest[0], "where") && frameWord(rest[1], "the", "this") &&
			frameWord(rest[2], "column", "table", "diagram", "example", "section", "page") &&
			frameWord(rest[3], "earns", "earned") && frameWord(rest[4], "its") && frameWord(rest[5], "place")
	}
	return false
}

// evaluationEnd keeps a following operational clause outside the diagnostic.
// Complements naming a purpose, condition or measurement prevent the match.
func evaluationEnd(tokens []document.Token) int {
	for i := 1; i < min(len(tokens)-2, 8); i++ {
		if !frameWord(tokens[i], "is", "was") || !positiveRhetoricSubject(tokens[:i]) {
			continue
		}
		anaphoric := i == 1 && frameWord(tokens[0], "this", "that", "it", "which")
		if end := evaluationComplementEnd(tokens[i+1:], anaphoric); end > 0 {
			return i + 1 + end
		}
	}
	return 0
}

func evaluationComplementEnd(tokens []document.Token, anaphoric bool) int {
	if len(tokens) >= 3 && frameWord(tokens[0], "the") && frameWord(tokens[1], "whole", "entire") &&
		frameWord(tokens[2], "point", "design", "idea") && rhetoricEnd(tokens, 3) {
		return 3
	}
	if anaphoric {
		return anaphoricEvaluationEnd(tokens)
	}
	return 0
}

func anaphoricEvaluationEnd(tokens []document.Token) int {
	if bareVerification(tokens) || bareWorth(tokens) || bareQuestion(tokens) || bareWorthSeeing(tokens) {
		return len(tokens)
	}
	return 0
}

func bareQuestion(tokens []document.Token) bool {
	return len(tokens) == 3 && frameWord(tokens[0], "exactly", "precisely") &&
		frameWord(tokens[1], "the") && frameWord(tokens[2], "point", "question", "idea")
}

func bareWorthSeeing(tokens []document.Token) bool {
	return len(tokens) == 3 && frameWord(tokens[0], "worth") && frameWord(tokens[1], "seeing", "noting", "asking") &&
		frameWord(tokens[2], "which", "why", "how")
}

func bareVerification(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "measured", "verified", "tested", "proved", "proven") &&
		frameWord(tokens[1], "rather") && frameWord(tokens[2], "than") && frameWord(tokens[3], "assumed", "asserted")
}

func bareWorth(tokens []document.Token) bool {
	return len(tokens) == 4 && frameWord(tokens[0], "the") && frameWord(tokens[1], "case", "part") &&
		frameWord(tokens[2], "worth") && frameWord(tokens[3], "having", "keeping")
}

func rhetoricEnd(tokens []document.Token, end int) bool {
	return end == len(tokens) || (end+1 < len(tokens) && frameWord(tokens[end], ",") && frameWord(tokens[end+1], "and"))
}

func positiveRhetoricSubject(tokens []document.Token) bool {
	if len(tokens) == 1 && frameWord(tokens[0], "which", "this", "that", "it") {
		return true
	}
	if len(tokens) == 0 || !nominalSubject(tokens) {
		return false
	}
	for _, token := range tokens {
		if frameWord(token, "not", "no", "never", "neither", "without") || token.Tag == "CD" {
			return false
		}
	}
	return true
}
