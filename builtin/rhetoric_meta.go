package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func documentJustification(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if c.eligible() {
		for i := range c.tokens() {
			if embeddedClauseStart(c.tokens(), i) && documentEarnsPlace(c.tokens()[i:]) {
				return localFrame(c, i)
			}
			if embeddedClauseStart(c.tokens(), i) && documentJustifies(c.tokens()[i:]) {
				return localFrame(c, i)
			}
		}
	}
	return rhetoricalFrame{}, false
}

func documentJustifies(tokens []document.Token) bool {
	i := documentSubjectEnd(tokens)
	if i == 0 || i+2 >= len(tokens) {
		return false
	}
	if frameWord(tokens[i], "exists", "exist", "existed") {
		return frameWord(tokens[i+1], "so", "to", "because")
	}
	if !frameWord(tokens[i], "deliberately", "intentionally", "purposely") {
		return false
	}
	i++
	switch {
	case frameWord(tokens[i], "does", "do", "did") && frameWord(tokens[i+1], "not", "n't"):
		i += 2
	case frameWord(tokens[i], "doesn't", "don't", "didn't"):
		i++
	default:
		return false
	}
	return i+1 < len(tokens) && frameWord(tokens[i], "restate", "repeat", "duplicate", "reproduce", "retell")
}

func evaluativeClosure(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		if !embeddedClauseStart(c.tokens(), i) {
			continue
		}
		if end := evaluationEnd(c.tokens()[i:]); end > 0 {
			c.end = c.start + i + end
			return localFrame(c, i)
		}
		if evaluativeTail(c.tokens()[i:]) {
			return localFrame(c, i)
		}
	}
	return rhetoricalFrame{}, false
}

func evaluativeTail(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "which", "that", "this", "it") || !frameWord(tokens[1], "is", "was") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "exactly", "precisely") {
		i++
	}
	if honestAnswer(tokens[i:]) {
		return true
	}
	return tokens[0].Normal == "which" && anaphoricPurpose(tokens[i:])
}

func honestAnswer(tokens []document.Token) bool {
	return len(tokens) >= 3 && frameWord(tokens[0], "the", "an") && frameWord(tokens[1], "honest") &&
		frameWord(tokens[2], "answer", "response", "outcome", "result")
}

func anaphoricPurpose(tokens []document.Token) bool {
	if len(tokens) >= 5 && frameWord(tokens[0], "the") && frameWord(tokens[1], "point", "purpose") &&
		frameWord(tokens[2], "of") {
		return gerundAnaphor(tokens[3:])
	}
	end := len(tokens)
	return end >= 5 && frameWord(tokens[0], "what") && gerundAnaphor(tokens[1:]) &&
		frameWord(tokens[end-2], "is", "was") && frameWord(tokens[end-1], "for")
}

func gerundAnaphor(tokens []document.Token) bool {
	return len(tokens) >= 2 && !tokens[0].Protected && tokens[0].Word && strings.HasSuffix(tokens[0].Normal, "ing") &&
		frameWord(tokens[1], "it", "them", "this", "that", "these", "those")
}
