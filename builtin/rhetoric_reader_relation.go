package builtin

import "github.com/stokaro/unswell/document"

// An anaphoric tail may predict a generic reader's preference without making a
// majority claim. A named actor, explicit requirement or condition is different.
func readerResult(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "which", "that", "this", "it") ||
		!frameWord(tokens[1], "is", "was") || !frameWord(tokens[2], "what") {
		return false
	}
	return genericReaderPredicate(tokens[3:])
}

func genericReaderPredicate(reader []document.Token) bool {
	if frameWord(reader[0], "a", "an", "the") {
		reader = reader[1:]
	}
	if len(reader) < 2 || len(reader) > 14 || !frameWord(reader[0], "reader", "operator", "user", "readers", "operators", "users") {
		return false
	}
	predicate := reader[len(reader)-1]
	if !frameWord(predicate, "wants", "want", "understands", "understand", "misses", "miss") {
		return false
	}
	if len(reader) == 2 {
		return true
	}
	if !frameWord(reader[1], "reading", "looking", "reviewing") {
		return false
	}
	return readerModifier(reader[2 : len(reader)-1])
}

func readerModifier(tokens []document.Token) bool {
	for _, token := range tokens {
		if token.Protected || !token.Word || !affirmativeRelation([]document.Token{token}) ||
			frameWord(token, "requires", "require", "specified", "requested", "surveyed", "reported", "that", "who") {
			return false
		}
	}
	return true
}
