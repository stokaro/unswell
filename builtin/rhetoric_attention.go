package builtin

import "github.com/stokaro/unswell/document"

// Attention frames evaluate receiving an explanation. Checking an operational
// condition and reading data as a prerequisite are different constructions.
func impersonalAttention(tokens []document.Token, sentenceStart bool) bool {
	i := attentionInfinitiveEnd(tokens)
	if i == 0 || sentenceStart && phraseOwnsAttention(tokens, i) {
		return false
	}
	if frameWord(tokens[i], "here") {
		i++
	}
	return i+1 < len(tokens) && frameWord(tokens[i], "that")
}

func attentionInfinitiveEnd(tokens []document.Token) int {
	if len(tokens) < 7 || !frameWord(tokens[0], "it") || !frameWord(tokens[1], "is", "was") {
		return 0
	}
	i := 2
	if frameWord(tokens[i], "also", "particularly", "especially") {
		i++
	}
	if i+3 >= len(tokens) || !frameWord(tokens[i], "important", "crucial", "essential", "useful", "interesting", "noteworthy") ||
		!frameWord(tokens[i+1], "to") || !attentionVerb(tokens[i+2]) {
		return 0
	}
	return i + 3
}

// The existing phrase rule owns these exact sentence openings.
func phraseOwnsAttention(tokens []document.Token, end int) bool {
	return end == 5 && frameWord(tokens[2], "important", "crucial") &&
		frameWord(tokens[4], "note") && frameWord(tokens[5], "that")
}

func attentionVerb(token document.Token) bool {
	return frameWord(token, "know", "note", "remember", "understand", "mention", "recognize")
}

func readerAttention(tokens []document.Token) bool {
	i := genericInstructionReader(tokens, 0)
	if i == 0 || i+2 >= len(tokens) || !frameWord(tokens[i], "has", "have", "needs", "need") ||
		!frameWord(tokens[i+1], "to") {
		return false
	}
	rest := tokens[i+2:]
	return len(rest) == 1 && (attentionVerb(rest[0]) || frameWord(rest[0], "read")) ||
		len(rest) == 2 && frameWord(rest[0], "look") && frameWord(rest[1], "at")
}

func attentionPart(tokens []document.Token) bool {
	if len(tokens) < 6 || !frameWord(tokens[0], "of") {
		return false
	}
	for at := 2; at < min(len(tokens)-3, 8); at++ {
		if !nominalSubject(tokens[1:at]) || !goalObjectEnd(tokens[at-1]) {
			continue
		}
		if readerAttention(tokens[at:]) {
			return true
		}
	}
	return false
}

func cognitiveReading(tokens []document.Token) bool {
	if len(tokens) != 3 || !frameWord(tokens[0], "reading", "knowing", "understanding") ||
		!frameWord(tokens[1], "the", "this", "that") {
		return false
	}
	return frameWord(tokens[2], "list", "explanation", "section", "paragraph", "page", "documentation", "guide")
}
