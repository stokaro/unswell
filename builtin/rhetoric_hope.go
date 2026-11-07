package builtin

import "github.com/stokaro/unswell/document"

// Hope framing is an editorial relation, not a match on the word "hope" alone.
// Preserve the source clause so the proposed edit can retain its condition or
// current technical status.
func hopeFraming(candidate frameClause) (rhetoricalFrame, bool) {
	if rhetoricQuoted(candidate) || rhetoricAttributed(candidate) {
		return rhetoricalFrame{}, false
	}
	tokens := candidate.tokens()
	var explanation editorialExplanation
	switch {
	case authorDocumentHope(tokens):
		explanation = editorialExplanation{
			message:    "The hope about this document does not state a supported action.",
			suggestion: "Keep the stated goal and document scope; remove the hope about the document.",
		}
	case impersonalHopeRequirement(tokens):
		explanation = editorialExplanation{
			message:    "The hoped-if-would construction buries the design goal.",
			suggestion: "State the intended behavior directly; preserve the condition, actor and modification limit.",
		}
	case speculativeHopeRemoval(tokens):
		explanation = editorialExplanation{
			message:    "An undated hope about a feature's future adds no current instruction.",
			suggestion: "Keep the current feature behavior and warnings; omit the hope unless there is a concrete plan to report.",
		}
	default:
		return rhetoricalFrame{}, false
	}
	frame, _ := localFrame(candidate, 0)
	frame.explanation = explanation
	return frame, true
}

func authorDocumentHope(tokens []document.Token) bool {
	if len(tokens) < 12 || !frameWord(tokens[0], "i", "we") || !frameWord(tokens[1], "hope") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "that") {
		i++
	}
	end := documentSubjectEnd(tokens[i:])
	if end == 0 {
		return false
	}
	i = conjoinedProcessEnd(tokens, i+end)
	return i > 0 && anaphoricHopeGoal(tokens[i:])
}

func conjoinedProcessEnd(tokens []document.Token, i int) int {
	if i+5 >= len(tokens) || !frameWord(tokens[i], "and") || !frameWord(tokens[i+1], "the", "its") ||
		!frameWord(tokens[i+2], "process", "processes", "procedure", "procedures", "guidance", "steps") {
		return 0
	}
	i += 3
	if i+1 < len(tokens) && frameWord(tokens[i], "it", "they") && frameWord(tokens[i+1], "entails", "contain", "contains") {
		i += 2
	}
	return i
}

func anaphoricHopeGoal(tokens []document.Token) bool {
	if len(tokens) < 4 || !frameWord(tokens[0], "will", "would", "may") || !frameWord(tokens[1], "help") {
		return false
	}
	i := 2
	if frameWord(tokens[i], "to") {
		i++
	}
	return i+2 == len(tokens) && frameWord(tokens[i], "achieve", "accomplish", "realize") &&
		frameWord(tokens[i+1], "that", "this")
}

func impersonalHopeRequirement(tokens []document.Token) bool {
	if len(tokens) < 12 {
		return false
	}
	i := impersonalHopePrefixEnd(tokens)
	return i > 0 && conditionalHopeTail(tokens[i:])
}

func impersonalHopePrefixEnd(tokens []document.Token) int {
	if !frameWord(tokens[0], "it") || !frameWord(tokens[1], "is", "was") {
		return 0
	}
	i := 2
	if frameWord(tokens[i], "also") {
		i++
	}
	if i+2 >= len(tokens) || !frameWord(tokens[i], "hoped") ||
		!frameWord(tokens[i+1], "that") || !frameWord(tokens[i+2], "if") {
		return 0
	}
	return i + 3
}

func conditionalHopeTail(tokens []document.Token) bool {
	modal := -1
	for j, token := range tokens {
		if token.Tag == "CD" || frameWord(token, "not", "never", "unless", "because", "measured") {
			return false
		}
		if frameWord(token, "would", "could") {
			modal = j
		}
	}
	return modal >= 3 && modal+2 < len(tokens)
}

func speculativeHopeRemoval(tokens []document.Token) bool {
	if len(tokens) < 9 || !frameWord(tokens[0], "hopefully") {
		return false
	}
	i := somedayPrefixEnd(tokens[1:])
	return i > 0 && futureRemovalTail(tokens[1+i:])
}

func somedayPrefixEnd(tokens []document.Token) int {
	i := 0
	if frameWord(tokens[i], ",") {
		i++
	}
	switch {
	case i+1 < len(tokens) && frameWord(tokens[i], "some") && frameWord(tokens[i+1], "day"):
		i += 2
	case frameWord(tokens[i], "someday"):
		i++
	default:
		return 0
	}
	if i < len(tokens) && frameWord(tokens[i], ",") {
		i++
	}
	return i
}

func futureRemovalTail(tokens []document.Token) bool {
	modal := -1
	for j, token := range tokens {
		if token.Tag == "CD" || frameWord(token, "if", "when", "unless", "because", "not", "never") {
			return false
		}
		if frameWord(token, "will") {
			modal = j
		}
	}
	return futureRemovalPredicate(tokens, modal)
}

func futureRemovalPredicate(tokens []document.Token, modal int) bool {
	if modal < 1 || modal+2 >= len(tokens) || modal > len(tokens)-3 {
		return false
	}
	if frameWord(tokens[modal+1], "similarly", "eventually") {
		modal++
	}
	return modal+3 == len(tokens) && frameWord(tokens[modal+1], "become") &&
		frameWord(tokens[modal+2], "unnecessary", "obsolete", "redundant")
}
