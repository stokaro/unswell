package builtin

import "github.com/stokaro/unswell/document"

func documentMaintenanceStart(c frameClause, i int) bool {
	return !quotedClaim(c.sentence.Tokens) && documentMaintenance(c.tokens()[i:]) &&
		(embeddedClauseStart(c.tokens(), i) || frameWord(c.tokens()[i], "rather", "instead"))
}

func documentMaintenance(tokens []document.Token) bool {
	if len(tokens) < 5 {
		return false
	}
	if frameWord(tokens[0], "instead") && frameWord(tokens[1], "of") && frameWord(tokens[2], "being") {
		return repeatedHere(tokens[3:])
	}
	if frameWord(tokens[0], "rather") && frameWord(tokens[1], "than") {
		end := documentSubjectEnd(tokens[2:])
		return end > 0 && 2+end+1 < len(tokens) && frameWord(tokens[2+end], "restating", "repeating")
	}
	return listedInstead(tokens) || countRepeatedHere(tokens)
}

func listedInstead(tokens []document.Token) bool {
	return len(tokens) > 6 && frameWord(tokens[0], "they", "these") && frameWord(tokens[1], "are") &&
		frameWord(tokens[2], "listed", "documented") && frameWord(tokens[3], "here") &&
		frameWord(tokens[4], "rather") && frameWord(tokens[5], "than") && frameWord(tokens[6], "left")
}

func countRepeatedHere(tokens []document.Token) bool {
	if !frameWord(tokens[0], "the", "this", "that", "neither", "these", "those") ||
		!frameWord(tokens[1], "count", "counts", "definition", "definitions", "explanation", "explanations", "list") ||
		!frameWord(tokens[2], "is", "are") {
		return false
	}
	i := 3
	if frameWord(tokens[i], "not") {
		i++
	}
	return repeatedHere(tokens[i:])
}

func repeatedHere(tokens []document.Token) bool {
	return len(tokens) == 2 && frameWord(tokens[0], "repeated", "restated", "duplicated") && frameWord(tokens[1], "here")
}
