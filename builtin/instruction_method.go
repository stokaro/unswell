package builtin

import "github.com/stokaro/unswell/document"

// A nominal instrument can link an adjacent action to its anaphoric method.
// It cannot independently establish a narrated instruction: a passive agent
// such as "the next worker" still describes who performs the operation.
func nominalInstructionMethod(tokens []document.Token) bool {
	if !methodAnnouncement(tokens) {
		return false
	}
	for i := 3; i < min(len(tokens), 7); i++ {
		if frameWord(tokens[i-1], "by", "through", "with", "using") && methodInstrument(tokens[i:]) {
			return true
		}
	}
	return false
}

func methodInstrument(tokens []document.Token) bool {
	end := len(tokens)
	for i, token := range tokens {
		if frameWord(token, "below", "above", "shown", "described") {
			end = i
			break
		}
	}
	if end == 0 || end > 8 || !nominalSubject(tokens[:end]) {
		return false
	}
	return frameWord(tokens[end-1], "notation", "syntax", "option", "flag", "attribute",
		"directive", "command", "expression")
}
