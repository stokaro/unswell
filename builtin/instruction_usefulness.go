package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
)

// Keep both parts of the chain. The relative clause can qualify an operand,
// so its purpose cannot be transferred to the main action or deleted by a rule.
func supportedActionUsefulness(c frameClause) (rhetoricalFrame, bool) {
	if !wholeInstructionClause(c) {
		return rhetoricalFrame{}, false
	}
	tokens := c.tokens()
	for relative := 6; relative+6 < len(tokens); relative++ {
		if !usefulnessPurpose(tokens[relative:]) {
			continue
		}
		prefix := c
		prefix.end = c.start + relative
		if frameWord(tokens[relative-1], ",") {
			prefix.end--
		}
		// Helpfulness alone is not a support chain. The main clause must
		// independently contain an existing supported-action construction.
		if !projectedInstructionWithMethod(prefix, false) {
			continue
		}
		frame, _ := localFrame(c, 0)
		frame.explanation = editorialExplanation{
			message: "The action and possible purpose are wrapped in separate support and helpfulness clauses.",
			suggestion: "State the action and possible purpose directly. " +
				"Preserve the relative clause's antecedent, actors, operands, methods and conditions, and the modality of both clauses; " +
				"a possible benefit is not a guarantee and a capability is not an obligation.",
		}
		return frame, true
	}
	return rhetoricalFrame{}, false
}

func usefulnessPurpose(tokens []document.Token) bool {
	if !usefulnessRelative(tokens) {
		return false
	}
	purpose := tokens[5:]
	if len(purpose) > 7 || !frameWord(purpose[len(purpose)-1], "purpose", "purposes") ||
		!nominalPurposeHead(purpose[len(purpose)-2]) {
		return false
	}
	// The bounded nominal phrase excludes numbers, finite clauses, conditions
	// and protected vocabulary. Its content stays in the original evidence.
	for _, token := range purpose[:len(purpose)-1] {
		if !nominalPurposeWord(token) {
			return false
		}
	}
	return true
}

func usefulnessRelative(tokens []document.Token) bool {
	return len(tokens) >= 7 && frameWord(tokens[0], "that", "which") &&
		frameWord(tokens[1], "can", "may", "might", "could") && frameWord(tokens[2], "be") &&
		frameWord(tokens[3], "helpful", "useful") && frameWord(tokens[4], "for")
}

func nominalPurposeWord(token document.Token) bool {
	return !token.Protected && token.Word && (nominalPurposeHead(token) || strings.HasPrefix(token.Tag, "JJ"))
}

func nominalPurposeHead(token document.Token) bool {
	return token.Tag == "NN" || token.Tag == "NNS" || token.Tag == "VBG"
}
