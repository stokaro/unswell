package builtin

import "github.com/stokaro/unswell/document"

func scopedInformationNotice(c frameClause, i int) bool {
	return !quotedClaim(c.sentence.Tokens) && (informationNotice(c.tokens()[i:]) || cognitiveAnnouncement(c.tokens()[i:]))
}

// Notice frames describe presenting information, not its operational truth. A
// following condition remains source context and is not a reason to erase it.
func informationNotice(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "it") {
		return false
	}
	i := 1
	if !frameWord(tokens[i], "should", "can", "may", "must") {
		return false
	}
	i++
	if frameWord(tokens[i], "also") {
		i++
	}
	if i+2 >= len(tokens) || !frameWord(tokens[i], "be") || !frameWord(tokens[i+1], "noted", "mentioned", "observed") {
		return false
	}
	return frameWord(tokens[i+2], "that")
}

func cognitiveAnnouncement(tokens []document.Token) bool {
	for i := 1; i+2 < min(len(tokens), 18); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !informationSubject(tokens[:i]) {
			continue
		}
		rest := tokens[i+1:]
		if !frameWord(rest[0], "worth") || !frameWord(rest[1],
			"knowing", "noting", "remembering", "understanding", "mentioning", "stating", "saying") {
			return false
		}
		return !phraseOwnsNotice(tokens[:i], rest)
	}
	return false
}

// The existing phrase rule owns this exact announcement.
func phraseOwnsNotice(subject, rest []document.Token) bool {
	return len(subject) == 1 && frameWord(subject[0], "it") && len(rest) > 2 &&
		frameWord(rest[1], "noting") && frameWord(rest[2], "that")
}

func informationSubject(tokens []document.Token) bool {
	if len(tokens) == 0 || frameWord(tokens[0], "and", "but", "or", "so") {
		return false
	}
	if len(tokens) == 1 && frameWord(tokens[0], "this", "that", "which", "it") {
		return true
	}
	if !nominalSubject(tokens) {
		return false
	}
	for _, token := range tokens {
		if token.Protected || frameWord(token, "no", "not", "never", "without", "neither") {
			return false
		}
	}
	return true
}
