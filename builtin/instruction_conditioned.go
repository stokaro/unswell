package builtin

import "github.com/stokaro/unswell/document"

// A temporal safety or possibility statement keeps its condition. Only an
// immediate anaphoric invocation of a named operation supplies the extra setup
// layer; neither a safety statement nor an ordinary passive is a finding alone.
func conditionedMethodInstruction(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	if index+1 >= len(clauses) || !conditionedMethodLink(clauses[index], clauses[index+1]) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{
		parts: []frameClause{clauses[index], clauses[index+1]},
		explanation: editorialExplanation{
			message: "The prerequisite and action are introduced before a separate method-invocation announcement.",
			suggestion: "Combine the action and its named method under the existing prerequisite. " +
				"Preserve the safety or possibility qualifier, prior state, actors, operands and limits; " +
				"do not turn permission or capability into an obligation.",
		},
	}, true
}

func conditionedMethodLink(a, b frameClause) bool {
	return wholeInstructionClause(a) && wholeInstructionClause(b) && a.ordinal+1 == b.ordinal &&
		conditionedAction(a.tokens()) && anaphoricInvocation(b.tokens())
}

// Do not infer a complete instruction from a clause before a semicolon,
// quotation, attribution or restriction. The original sentence stays intact.
func wholeInstructionClause(c frameClause) bool {
	if !c.eligible() || c.start != 0 || rhetoricQuoted(c) || rhetoricAttributed(c) || projectionGuard(c.tokens()) {
		return false
	}
	return c.end == len(c.sentence.Tokens) ||
		c.end+1 == len(c.sentence.Tokens) && frameWord(c.sentence.Tokens[c.end], ".")
}

func conditionedAction(tokens []document.Token) bool {
	if len(tokens) < 10 || !frameWord(tokens[0], "after", "once", "when") {
		return false
	}
	for end := 4; end+6 < len(tokens); end++ {
		if !frameWord(tokens[end], ",") {
			continue
		}
		action := tokens[end+1:]
		return prerequisiteState(tokens[1:end]) && matches(action[:2], []string{"it", "is"}) &&
			frameWord(action[2], "safe", "possible") && frameWord(action[3], "to") &&
			grammaticalAction(action[4:])
	}
	return false
}

func prerequisiteState(tokens []document.Token) bool {
	for verb := 1; verb+1 < len(tokens); verb++ {
		if !nominalSubject(tokens[:verb]) {
			continue
		}
		if frameWord(tokens[verb], "is", "are") &&
			(projectedParticiple(tokens[verb+1]) || frameWord(tokens[verb+1], "ready", "complete")) {
			return true
		}
		if verb+2 < len(tokens) && frameWord(tokens[verb], "has", "have") &&
			frameWord(tokens[verb+1], "been") && projectedParticiple(tokens[verb+2]) {
			return true
		}
	}
	return false
}

func anaphoricInvocation(tokens []document.Token) bool {
	if len(tokens) < 8 || !matches(tokens[:2], []string{"to", "do"}) || !frameWord(tokens[2], "this", "that") {
		return false
	}
	start := 3
	if frameWord(tokens[start], ",") {
		start++
	}
	for verb := start + 2; verb+1 < min(len(tokens), start+7); verb++ {
		if operationSubject(tokens[start:verb]) && passiveInvocation(tokens[verb:]) {
			return true
		}
	}
	return false
}

func passiveInvocation(tokens []document.Token) bool {
	end := 1
	if frameWord(tokens[0], "can", "may") && len(tokens) > 2 && frameWord(tokens[1], "be") {
		end++
	} else if !frameWord(tokens[0], "is") {
		return false
	}
	if !frameWord(tokens[end], "called", "invoked") {
		return false
	}
	end++
	if end == len(tokens) {
		return true
	}
	return invocationOperand(tokens[end:])
}

func invocationOperand(tokens []document.Token) bool {
	if !frameWord(tokens[0], "on", "with", "from") || !projectionOperand(tokens) {
		return false
	}
	for _, token := range tokens {
		if finiteWord(token) || frameWord(token, "if", "when", "once", "after", "while", "until", "because",
			"and", "or", "by", "automatically") {
			return false
		}
	}
	return true
}
