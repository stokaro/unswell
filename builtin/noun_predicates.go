package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

type nounPredicates struct {
	verbs map[string]bool
	nouns map[string]bool
}

// matches limits configured verb forms without changing the provider's tags.
// A dual-use form such as "fix" remains a noun after a nominal modifier or
// another common noun. A leading form after a numeric subject or clause break
// can be a predicate. These are surface guards, not dependency analysis.
func (p nounPredicates) matches(sentence document.Sentence, chunk document.Chunk, start, i int) bool {
	if !p.verbs[sentence.Tokens[i].Normal] {
		return false
	}
	if p.nouns[sentence.Tokens[i].Normal] &&
		(i != start || i > 0 && nominalModifier(sentence.Tokens[i-1])) {
		return false
	}
	if i+1 < chunk.EndToken {
		return true
	}
	return i+1 >= len(sentence.Tokens) ||
		slices.Contains([]string{"TO", "IN", ".", ",", ":"}, sentence.Tokens[i+1].Tag) ||
		coordinatedPredicate(sentence.Tokens, i+1)
}

func coordinatedPredicate(tokens []document.Token, next int) bool {
	if next+1 >= len(tokens) || tokens[next].Protected || tokens[next].Tag != "CC" {
		return false
	}
	token := tokens[next+1]
	return !token.Protected && (strings.HasPrefix(token.Tag, "VB") || token.Tag == "MD")
}
