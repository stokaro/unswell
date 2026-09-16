package builtin

import "github.com/stokaro/unswell/document"

func metadiscourseFrames(clauses []frameClause, i int) (rhetoricalFrame, bool) {
	clause := clauses[i]
	if !clause.eligible() || !documentAnnouncement(clause.tokens()) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{clause}}, true
}

func documentAnnouncement(tokens []document.Token) bool {
	start := documentSubjectEnd(tokens)
	if start == 0 {
		start = documentPrefaceEnd(tokens)
	}
	if start == 0 || start >= len(tokens) {
		return false
	}
	if frameWord(tokens[start], "will", "shall") {
		start++
	}
	return start+1 < len(tokens) && documentVerb(tokens[start])
}

func documentSubjectEnd(tokens []document.Token) int {
	// Limit the subject to a determiner, one optional spatial modifier, and a
	// document noun. Arbitrary occurrences of "page" do not establish a frame.
	if len(tokens) < 3 || !frameWord(tokens[0], "this", "these", "the", "our") {
		return 0
	}
	i := 1
	if frameWord(tokens[i], "following", "next", "present", "current") {
		i++
	}
	if i < len(tokens) && documentNoun(tokens[i]) {
		return i + 1
	}
	return 0
}

func documentPrefaceEnd(tokens []document.Token) int {
	if len(tokens) < 6 || !frameWord(tokens[0], "in", "throughout") {
		return 0
	}
	end := documentSubjectEnd(tokens[1:])
	if end == 0 {
		return 0
	}
	end++
	if end < len(tokens) && frameWord(tokens[end], ",") {
		end++
	}
	if end < len(tokens) && frameWord(tokens[end], "we", "i") {
		return end + 1
	}
	return 0
}

func documentNoun(token document.Token) bool {
	return frameWord(token, "page", "pages", "section", "sections", "guide", "guides", "chapter", "chapters",
		"document", "documents", "article", "articles", "tutorial", "tutorials", "overview", "reference")
}

func documentVerb(token document.Token) bool {
	return frameWord(token, "define", "defines", "defined", "describe", "describes", "described", "explain", "explains", "explained",
		"cover", "covers", "covered", "discuss", "discusses", "discussed", "outline", "outlines", "outlined",
		"introduce", "introduces", "introduced", "summarize", "summarizes", "summarized", "document", "documents", "documented",
		"show", "shows", "showed", "demonstrate", "demonstrates", "demonstrated", "examine", "examines", "examined")
}
