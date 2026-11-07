package builtin

import "github.com/stokaro/unswell/document"

// qualityClaim identifies a local subjective descriptor without treating the
// described capability or its technical operands as disposable prose.
func qualityClaim(c frameClause) (rhetoricalFrame, bool) {
	if !c.eligible() || len(c.sentence.Tokens) > 96 || quotedClaim(c.sentence.Tokens) ||
		contextualRhetoricScoped(c) || qualityMechanism(c.tokens()) {
		return rhetoricalFrame{}, false
	}
	for i := range c.tokens() {
		end := qualityClaimEnd(c.tokens(), i)
		if end <= i || degreeNegated(c.tokens()[:end]) {
			continue
		}
		local := c
		local.end, local.start = c.start+end, c.start+i
		frame, ok := localFrame(local, 0)
		frame.explanation = editorialExplanation{
			message: "This quality claim has no stated criterion or measurable scope.",
			suggestion: "State the concrete capability and the criterion for this judgment. " +
				"Verify the claim before weakening or removing it; preserve technical conditions and identifiers.",
			fallback: degreeFallback(),
		}
		return frame, ok
	}
	return rhetoricalFrame{}, false
}

func qualityClaimEnd(tokens []document.Token, at int) int {
	if end := superlativeFeatureEnd(tokens, at); end > at {
		return end
	}
	if end := pairedQualityArtifactEnd(tokens, at); end > at {
		return end
	}
	if at+1 < len(tokens) && frameWord(tokens[at], "reasonable") &&
		frameWord(tokens[at+1], "defaults") {
		return at + 2
	}
	if at+1 < len(tokens) && frameWord(tokens[at], "important") &&
		frameWord(tokens[at+1], "effect", "impact") && qualityEffect(tokens[at+2:]) {
		return at + 2
	}
	return 0
}

func superlativeFeatureEnd(tokens []document.Token, at int) int {
	if at+3 >= len(tokens) || !frameWord(tokens[at], "most") ||
		!frameWord(tokens[at+1], "comprehensive") ||
		!frameWord(tokens[at+2], "feature", "capability") ||
		!frameWord(tokens[at+3], "set", "coverage") {
		return 0
	}
	for _, token := range tokens[at+4:] {
		if frameWord(token, "among", "compared", "measured", "benchmark", "benchmarks", "table") ||
			token.Tag == "CD" {
			return 0
		}
	}
	return at + 4
}

func pairedQualityArtifactEnd(tokens []document.Token, at int) int {
	if at+3 >= len(tokens) || !frameWord(tokens[at], "powerful", "flexible", "intuitive", "scalable") ||
		!frameWord(tokens[at+1], "and") ||
		!frameWord(tokens[at+2], "powerful", "flexible", "intuitive", "scalable") ||
		frameWord(tokens[at], tokens[at+2].Normal) {
		return 0
	}
	if frameWord(tokens[at+3], "query") && at+4 < len(tokens) && frameWord(tokens[at+4], "language") {
		return at + 5
	}
	if softwareArtifact(tokens[at+3]) {
		return at + 4
	}
	return 0
}

func qualityEffect(tokens []document.Token) bool {
	for _, token := range tokens[:min(len(tokens), 8)] {
		if frameWord(token, "quality") {
			return true
		}
	}
	return false
}
