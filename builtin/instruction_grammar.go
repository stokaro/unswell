package builtin

import "github.com/stokaro/unswell/document"

// grammaticalAction uses an explicit infinitive or imperative role. The
// impersonal possibility matcher retains its narrower configuration vocabulary.
func grammaticalAction(tokens []document.Token) bool {
	if len(tokens) < 2 || !instructionVerb(tokens) {
		return false
	}
	if actionOperand(tokens) {
		return true
	}
	// Coordinated verbs may share an object: read and write records. Require
	// another action, not an unrelated clause after a conjunction.
	return len(tokens) >= 4 && frameWord(tokens[1], "and", "or") &&
		instructionVerb(tokens[2:]) && actionOperand(tokens[2:])
}

func instructionVerb(tokens []document.Token) bool {
	return len(tokens) > 0 && !frameWord(tokens[0], "know", "believe", "think", "feel", "seem",
		"want", "need", "wish", "intend", "become", "avoid", "prevent") &&
		(projectionVerb(tokens[0]) || instructionAction(tokens))
}

func imperativeAssurance(tokens []document.Token) bool {
	return len(tokens) >= 5 && matches(tokens[:3], []string{"be", "sure", "to"}) &&
		grammaticalAction(tokens[3:]) && !instructionCondition(tokens)
}

func procedureAnnouncement(tokens []document.Token) bool {
	start := 1
	if len(tokens) < 8 {
		return false
	}
	if matches(tokens[:3], []string{"in", "order", "to"}) {
		start = 3
	} else if !frameWord(tokens[0], "to") {
		return false
	}
	for end := start + 2; end+3 < len(tokens); end++ {
		if frameWord(tokens[end], ",") && grammaticalAction(tokens[start:end]) && readerSteps(tokens[end+1:]) {
			return true
		}
	}
	return false
}

func readerSteps(tokens []document.Token) bool {
	i, ok := readerActionStart(tokens)
	if !ok || i+2 >= len(tokens) || !frameWord(tokens[i], "follow", "perform", "complete", "take") {
		return false
	}
	i++
	if frameWord(tokens[i], "the", "these", "following") {
		i++
	}
	if i < len(tokens) && frameWord(tokens[i], "following") {
		i++
	}
	return i < len(tokens) && frameWord(tokens[i], "steps", "instructions") && !instructionCondition(tokens)
}

// relativeOperation resolves only an adjacent nominal antecedent. It does not
// infer a subject through another finite clause or assign dependency edges.
func relativeOperation(tokens []document.Token, verb int) (bool, bool) {
	if verb < 2 || !frameWord(tokens[verb-1], "that", "which") {
		return false, false
	}
	end := verb - 1
	if end > 0 && frameWord(tokens[end-1], ",") {
		end--
	}
	for start := max(0, end-12); start < end; start++ {
		if start > 0 && !frameWord(tokens[start], "a", "an", "the", "this", "that") {
			continue
		}
		subject := tokens[start:end]
		if relativeActionSubject(subject) {
			return true, operationSubject(subject)
		}
	}
	return false, false
}

// Relative modal usage alone does not establish agency. A checksum, label or
// other supplied operand can be used by a caller without performing the action.
func relativeActionActor(tokens []document.Token) bool {
	return len(tokens) > 0 && frameWord(tokens[len(tokens)-1], "reader", "writer", "parser", "client", "server",
		"service", "worker", "process", "library", "utility", "tool", "function", "method", "command", "handler", "callback")
}

func relativeActionSubject(tokens []document.Token) bool {
	return operationSubject(tokens) || projectionSubject(tokens) && relativeActionActor(tokens)
}
