package builtin

import "github.com/stokaro/unswell/document"

func documentMapFrame(c frameClause) (rhetoricalFrame, bool) {
	for i := range c.tokens() {
		if !embeddedClauseStart(c.tokens(), i) || !documentMap(c.tokens()[i:]) {
			continue
		}
		candidate := c
		candidate.start += i
		if rhetoricAttributed(candidate) || rhetoricQuoted(candidate) {
			continue
		}
		frame, ok := localFrame(candidate, 0)
		frame.explanation = editorialExplanation{
			message:    "This clause describes the page as a metaphorical map; introduce its subject directly.",
			suggestion: "Keep the topic, scope and reference links. Review whether the page needs to describe itself as a map.",
		}
		return frame, ok
	}
	return rhetoricalFrame{}, false
}

// A relative topic identifies self-description, rather than a literal map of
// a place, address space, schema, or other concrete mapped object.
func documentMap(tokens []document.Token) bool {
	i := documentSubjectEnd(tokens)
	return i > 0 && i+5 < len(tokens) && frameWord(tokens[i], "is", "was") &&
		frameWord(tokens[i+1], "a", "the") && frameWord(tokens[i+2], "map", "roadmap") &&
		frameWord(tokens[i+3], "of", "to") && frameWord(tokens[i+4], "what", "how", "which", "where")
}
