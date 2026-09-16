package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

type copularClause struct {
	subject  []document.Token
	copula   string
	negative bool
}

func denialFrames(clauses []frameClause, i int) (rhetoricalFrame, bool) {
	if i+1 >= len(clauses) {
		return rhetoricalFrame{}, false
	}
	a, b := clauses[i], clauses[i+1]
	if a.sentence.BlockID != b.sentence.BlockID || !a.eligible() || !b.eligible() {
		return rhetoricalFrame{}, false
	}
	left, lok := parseCopularClause(a.tokens())
	right, rok := parseCopularClause(b.tokens())
	if !lok || !rok || !denialRedefines(left, right) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{a, b}}, true
}

func denialRedefines(left, right copularClause) bool {
	return left.negative && !right.negative && left.copula == right.copula && linkedSubjects(left.subject, right.subject)
}

func parseCopularClause(tokens []document.Token) (copularClause, bool) {
	for i := 1; i < min(len(tokens), 9); i++ {
		copula, negative := frameCopula(tokens[i])
		if copula == "" {
			continue
		}
		result := copularClause{subject: tokens[:i], copula: copula, negative: negative}
		end := i + 1
		if end < len(tokens) && frameWord(tokens[end], "not", "n't") {
			result.negative = true
			end++
		}
		if end < len(tokens) && frameWord(tokens[end], "only") {
			return copularClause{}, false
		}
		if end < len(tokens) && frameWord(tokens[end], "just", "merely") {
			end++
		}
		return result, nominalSubject(result.subject) && nominalComplement(tokens[end:])
	}
	return copularClause{}, false
}

func frameCopula(token document.Token) (string, bool) {
	if token.Protected {
		return "", false
	}
	switch token.Normal {
	case "is", "'s":
		return "is", false
	case "are", "'re":
		return "are", false
	case "was", "were":
		return token.Normal, false
	case "isn't", "aren't", "wasn't", "weren't":
		return strings.TrimSuffix(token.Normal, "n't"), true
	default:
		return "", false
	}
}

func nominalSubject(tokens []document.Token) bool {
	return !slices.ContainsFunc(tokens, func(token document.Token) bool {
		if token.Protected {
			return false
		}
		return !token.Word || frameWord(token, "if", "when", "unless", "although", "because", "which", "who", "there") ||
			(strings.HasPrefix(token.Tag, "VB") && token.Tag != "VBG") || token.Tag == "MD"
	})
}

func nominalComplement(tokens []document.Token) bool {
	if len(tokens) == 0 {
		return false
	}
	first := tokens[0]
	if frameWord(first, "a", "an", "the", "your", "our", "my", "their", "its", "about") {
		return len(tokens) > 1
	}
	return first.Protected ||
		strings.HasPrefix(first.Tag, "NN")
}

func linkedSubjects(a, b []document.Token) bool {
	if len(b) == 1 && frameWord(b[0], "it", "they", "this", "these", "that", "those") {
		return true
	}
	return slices.EqualFunc(a, b, func(left, right document.Token) bool {
		return !left.Protected && !right.Protected && left.Normal == right.Normal
	})
}
