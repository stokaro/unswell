package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

func documentJustification(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if !c.eligible() {
		return rhetoricalFrame{}, false
	}
	if frame, ok := contextualDocumentFrame(clauses, index); ok {
		return frame, true
	}
	for i := range c.tokens() {
		if documentMaintenanceStart(c, i) {
			return localFrame(c, i)
		}
		if embeddedClauseStart(c.tokens(), i) &&
			(documentEarnsPlace(c.tokens()[i:]) || documentJustifies(c.tokens()[i:])) {
			return localFrame(c, i)
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
	if len(c.sentence.Tokens) > 96 || question(c.sentence) {
		return rhetoricalFrame{}, false
	}
	if discourseEvaluation(c) {
		return localFrame(c, 0)
	}
	for i := range c.tokens() {
		candidate, ok := localRhetoricCandidate(c, i)
		if !ok || !embeddedClauseStart(c.tokens(), i) {
			continue
		}
		if frame, found := evaluationCandidateFrame(c, candidate, i); found {
			return frame, true
		}
	}
	return nominalNotice(c)
}

func evaluationCandidateFrame(c, candidate frameClause, start int) (rhetoricalFrame, bool) {
	if rhetoricAttributed(candidate) || rhetoricQuoteOpen(candidate) {
		return rhetoricalFrame{}, false
	}
	if frame, found := informationEvaluation(candidate); found {
		return frame, true
	}
	if rhetoricQuoted(candidate) {
		return rhetoricalFrame{}, false
	}
	if scopedInformationNotice(c, start) {
		return localFrame(candidate, 0)
	}
	if end := evaluationEnd(candidate.tokens()); end > 0 {
		candidate.end = candidate.start + end
		return localFrame(candidate, 0)
	}
	if evaluativeTail(candidate.tokens()) {
		return localFrame(candidate, 0)
	}
	if end := rhetoricalRelationEnd(candidate); end > 0 {
		candidate.end = candidate.start + end
		return localFrame(candidate, 0)
	}
	if candidateEvaluation(candidate) {
		return localFrame(candidate, 0)
	}
	return rhetoricalFrame{}, false
}

func candidateEvaluation(candidate frameClause) bool {
	if evaluationScoped(candidate) {
		return false
	}
	tokens := candidate.tokens()
	return contextualEvaluation(tokens) || purposeEvaluation(tokens) || functionalCleft(tokens) ||
		intentionalityAnnouncement(tokens) || outputUnderstanding(tokens)
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
