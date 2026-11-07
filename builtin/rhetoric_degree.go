package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// Promotional degree is local to the qualitative modifier. A capability's
// operands and uncertainty remain source context, never replacement facts.
func promotionalDegree(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if promotionalDegreeExcluded(clauses, index) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		end, explanation := degreeAt(c.tokens(), i)
		if promotionalDegreeAllowed(c, i, end) {
			c.end, c.start = c.start+end, c.start+i
			frame, ok := localFrame(c, 0)
			frame.explanation = explanation
			return frame, ok
		}
	}
	return rhetoricalFrame{}, false
}

func promotionalDegreeExcluded(clauses []frameClause, index int) bool {
	c := clauses[index]
	return c.end-c.start < 2 || c.end-c.start > 48 || len(c.sentence.Tokens) > 96 ||
		question(c.sentence) || degreeGrounded(clauses, index)
}

func promotionalDegreeAllowed(c frameClause, start, end int) bool {
	return end > start && !degreeNegated(c.tokens()[:end]) &&
		(!contextualRhetoricScoped(c) || readerIntentOnly(c.sentence.Tokens, c.start+start))
}

func degreeAt(tokens []document.Token, at int) (int, editorialExplanation) {
	if end := degreeModifierEnd(tokens, at); end > at {
		return end, degreeExplanation()
	}
	if embeddedClauseStart(tokens, at) && superlativeBenefit(tokens[at:]) {
		return len(tokens), editorialExplanation{
			message: "This benefit is ranked above the alternatives without a stated comparison criterion.",
			suggestion: "State the benefit and the criterion supporting its rank. " +
				"Verify the comparison before changing or removing the ranking; preserve the named alternatives, conditions and uncertainty.",
			fallback: degreeFallback(),
		}
	}
	return 0, editorialExplanation{}
}

func degreeExplanation() editorialExplanation {
	return editorialExplanation{
		message: "This modifier evaluates a capability without stating its degree or the criterion for that judgment.",
		suggestion: "State the concrete capability and the criterion or measured extent behind the modifier. " +
			"Verify that basis before removing or weakening a claim; preserve the operands, conditions, uncertainty and established commitments.",
		fallback: degreeFallback(),
	}
}

func degreeFallback() string {
	return "Keep the concrete behavior, checks, links, measured results, conditions, uncertainty and established commitments. " +
		"State the criteria behind quality or benefit claims. Verify their basis before removing or weakening any claim."
}

// A reader's intent to perform an operation does not substantiate an ease or
// quality claim about the tool used for that operation. Other conditions and
// attribution retain the conservative scope boundary.
func readerIntentOnly(tokens []document.Token, candidate int) bool {
	if quotedClaim(tokens) {
		return false
	}
	found := false
	for i, token := range tokens {
		if token.Tag == "CD" && !indefiniteOne(tokens, i) {
			return false
		}
		if readerIntentAt(tokens, i) {
			if !readerIntentMainClause(tokens, i, candidate) {
				return false
			}
			found = true
			continue
		}
		if contextualScopeWord(token) {
			return false
		}
	}
	return found
}

func readerIntentMainClause(tokens []document.Token, condition, candidate int) bool {
	for i := condition + 4; i < candidate; i++ {
		if frameWord(tokens[i], ",") ||
			i+1 < candidate && frameWord(tokens[i], "we") &&
				frameWord(tokens[i+1], "have", "provide", "offer") {
			return true
		}
	}
	return false
}

func readerIntentAt(tokens []document.Token, index int) bool {
	return index+3 < len(tokens) && frameWord(tokens[index], "if") &&
		frameWord(tokens[index+1], "you") && frameWord(tokens[index+2], "want", "wish") &&
		frameWord(tokens[index+3], "to")
}

func degreeModifierEnd(tokens []document.Token, at int) int {
	if at+1 < len(tokens) && emphaticQuality(tokens[at], tokens[at+1]) {
		return at + 2
	}
	return easeModifierEnd(tokens, at)
}

func emphaticQuality(modifier, quality document.Token) bool {
	if frameWord(modifier, "blazingly", "blindingly", "phenomenally", "astonishingly", "remarkably",
		"incredibly", "amazingly", "unbelievably", "ridiculously") {
		return frameWord(quality, "fast", "quick", "efficient", "powerful", "flexible", "reliable",
			"robust", "portable", "intuitive", "scalable", "responsive", "easy", "usable", "adoptable", "extensible")
	}
	return frameWord(modifier, "heavily", "highly", "extensively") && frameWord(quality, "optimized") ||
		frameWord(modifier, "perfectly") && frameWord(quality, "adoptable", "usable", "intuitive", "straightforward")
}

// Ease-to-use must modify a software artifact. An infinitive in an ordinary
// comparison or a statement about a specific operation is not this modifier.
func easeModifierEnd(tokens []document.Token, at int) int {
	end := at + 1
	if !frameWord(tokens[at], "easy-to-use", "simple-to-use", "effortless-to-use") {
		if at+3 >= len(tokens) || !frameWord(tokens[at], "easy", "simple", "effortless") ||
			!frameWord(tokens[at+1], "to") || !frameWord(tokens[at+2], "use") {
			return 0
		}
		end = at + 3
	}
	for i := end; i < min(end+3, len(tokens)); i++ {
		if softwareArtifact(tokens[i]) {
			return i + 1
		}
		if tokens[i].Protected || !strings.HasPrefix(tokens[i].Tag, "JJ") {
			return 0
		}
	}
	return 0
}

func softwareArtifact(token document.Token) bool {
	return frameWord(token, "client", "library", "package", "interface", "api", "tool", "service", "framework")
}

func superlativeBenefit(tokens []document.Token) bool {
	if len(tokens) < 5 || !frameWord(tokens[0], "this", "that", "it") {
		return false
	}
	i := 1
	if frameWord(tokens[i], "will", "may", "might", "could", "would", "can") {
		i++
	}
	if i+3 >= len(tokens) || !frameWord(tokens[i], "is", "was", "be") || !frameWord(tokens[i+1], "the") ||
		!frameWord(tokens[i+2], "biggest", "greatest", "largest", "best") ||
		!frameWord(tokens[i+3], "win", "benefit", "advantage", "improvement", "gain") {
		return false
	}
	return i+4 == len(tokens)
}

func degreeNegated(tokens []document.Token) bool {
	for _, token := range tokens {
		if frameWord(token, "not", "n't", "no", "never", "cannot", "can't") {
			return true
		}
	}
	return false
}

// A neighboring measurement can establish the local claim's degree. This is a
// conservative applicability boundary, not a test that the evidence is true.
func degreeGrounded(clauses []frameClause, index int) bool {
	c := clauses[index]
	if c.end < len(c.sentence.Tokens) && frameWord(c.sentence.Tokens[c.end], ":") {
		return true
	}
	for _, neighbor := range clauses[max(0, index-1):min(len(clauses), index+2)] {
		if neighbor.sentence.BlockID != c.sentence.BlockID {
			continue
		}
		sameSentence := neighbor.sentence.ID == c.sentence.ID
		for _, token := range neighbor.sentence.Tokens {
			if degreeMeasurement(token, sameSentence) || sameSentence && frameWord(token, "because", "by", "through") {
				return true
			}
		}
	}
	return false
}

func degreeMeasurement(token document.Token, sameSentence bool) bool {
	// A bare count in another sentence need not support this sentence's degree.
	// Named measurements can establish a nearby criterion.
	return sameSentence && token.Tag == "CD" || frameWord(token, "benchmark", "benchmarks", "measured", "measurement",
		"median", "percentile", "compared", "comparison", "criterion", "criteria")
}
