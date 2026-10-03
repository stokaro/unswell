package builtin

import "github.com/stokaro/unswell/document"

// evidencePosture recognizes commentary contrasting verification with an
// assertion. The written discourse subject is required: a concrete measured
// quantity or an actor performing a check is a different construction.
func evidencePosture(c frameClause) (rhetoricalFrame, bool) {
	end, alternative := evidencePostureEnd(c.tokens())
	if end == 0 {
		return rhetoricalFrame{}, false
	}
	c.end = c.start + end
	attribution := c
	attribution.end = c.start + alternative
	if rhetoricAttributed(attribution) || rhetoricQuoted(c) || rhetoricCondition(c) || !evidenceAnaphor(c) {
		return rhetoricalFrame{}, false
	}
	frame, ok := localFrame(c, 0)
	frame.explanation = editorialExplanation{
		message: "This clause comments on verification; state the check and its result directly.",
		suggestion: "Keep the method, technical conditions, links and measured results. " +
			"Shorten the commentary contrasting verification with assertion while preserving its factual claims.",
	}
	return frame, ok
}

func evidencePostureEnd(tokens []document.Token) (int, int) {
	for i := 1; i < min(14, len(tokens)-3); i++ {
		if !frameWord(tokens[i], "is", "are", "was", "were") || !evidenceSubject(tokens[:i]) {
			continue
		}
		end, alternative := evidenceComparison(tokens[i+1:])
		if end > 0 {
			return i + 1 + end, i + 1 + alternative
		}
	}
	return 0, 0
}

func evidenceSubject(tokens []document.Token) bool {
	if len(tokens) == 1 {
		return frameWord(tokens[0], "this", "that", "which", "it")
	}
	if len(tokens) == 2 && frameWord(tokens[0], "the", "this", "that", "these", "those") {
		return frameWord(tokens[1], "claim", "statement", "statements", "conclusion", "conclusions", "coverage")
	}
	return evidenceRelativeSubject(tokens)
}

func evidenceRelativeSubject(tokens []document.Token) bool {
	if len(tokens) < 3 || !frameWord(tokens[0], "what") ||
		!frameWord(tokens[len(tokens)-1], "covers", "shows", "describes") {
		return false
	}
	return len(tokens) == 3 && frameWord(tokens[1], "it", "this", "that") ||
		len(tokens) == 4 && frameWord(tokens[1], "the", "this", "that") &&
			frameWord(tokens[2], "page", "guide", "section", "report", "table")
}

func evidenceComparison(tokens []document.Token) (int, int) {
	i := 0
	if len(tokens) > 0 && frameWord(tokens[0], "actually", "empirically", "explicitly") {
		i++
	}
	if i >= len(tokens) || !frameWord(tokens[i], "measured", "verified", "checked", "tested",
		"validated", "demonstrated", "proved", "proven") {
		return 0, 0
	}
	i++
	i = evidenceContrastStart(tokens, i)
	if i == 0 {
		return 0, 0
	}
	return evidenceAlternative(tokens, i)
}

func evidenceAlternative(tokens []document.Token, i int) (int, int) {
	if i < len(tokens) && frameWord(tokens[i], "simply", "merely", "just") {
		i++
	}
	if i >= len(tokens) || !frameWord(tokens[i], "assumed", "asserted", "claimed", "promised", "guessed") ||
		!evidenceContrastComplete(tokens[i+1:]) {
		return 0, 0
	}
	return i + 1, i
}

func evidenceContrastStart(tokens []document.Token, i int) int {
	if i+2 >= len(tokens) {
		return 0
	}
	if frameWord(tokens[i], "rather") && frameWord(tokens[i+1], "than") ||
		frameWord(tokens[i], ",", "and") && frameWord(tokens[i+1], "not") {
		return i + 2
	}
	return 0
}

func evidenceContrastComplete(rest []document.Token) bool {
	if len(rest) == 0 {
		return true
	}
	return len(rest) > 1 && frameWord(rest[0], ",", "and", "but") &&
		!frameWord(rest[1], "by", "using", "against", "under", "with", "because", "according", "if", "when", "unless")
}

// A bare it needs an explicit information antecedent in its own sentence.
// Numeric quantities and concrete entities cannot establish this relation.
func evidenceAnaphor(c frameClause) bool {
	if !frameWord(c.tokens()[0], "it") {
		return true
	}
	information := false
	for _, token := range c.sentence.Tokens[:c.start] {
		if token.Tag == "CD" {
			return false
		}
		information = information || frameWord(token, "claim", "statement", "statements", "conclusion", "conclusions",
			"result", "coverage", "page", "guide", "section", "report", "table")
	}
	return information
}
