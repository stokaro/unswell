package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func numericalHookRule() rule.Rule {
	d := descriptor("filler.unnamed-numerical-choice",
		"Name the choice instead of introducing a count of unnamed answers or options.", "rhetorical-patterns", "block", 12)
	d.Contexts = []string{"paragraph", "comment", "string"}
	d.BlockObservations = true
	d.TermExemptions = true
	d.Defaults.Parameters = rule.Parameters{WindowSentences: 1, SaturationOccurrences: 1}
	d.Parameters = []string{"allowed_occurrences", "saturation_occurrences"}
	d.Description = "A prose block opens with a count and answers/options/choices, followed by ', and/but the right/best one depends on' " +
		"and a stated selection condition. Reports that clause; the actual task is left unnamed by the opening."
	d.Limitations = "Experimental editorial policy, not authorship attribution, enumeration validation, " +
		"or established population precision. " +
		"Recognizes two through twenty in words and 2 through 99 in digits, only in the first sentence of a prose block. " +
		"Clauses are limited to 48 tokens. Named alternatives, task headings, lists, requirements, quotations and questions do not qualify. " +
		"Other numerical hooks remain unsupported. Keep the selection condition when naming the actual choice."
	d.Examples = []rule.Example{
		{Text: "Four answers, and the right one depends on whether you control the writes.", Match: true},
		{Text: "Two schemes: env and file."},
		{Text: "Choose a consistency mode based on whether you control source writes."},
	}
	return check{d, numericalHooks}
}

func numericalHooks(ctx context.Context, view rule.View, emit rule.Emitter) error {
	m := newEditorialMatcher(ctx, view)
	observations := newCandidateObservations(view, proseBlock)
	for _, block := range view.Document.Blocks {
		if err := m.spend(); err != nil {
			return err
		}
		if !proseBlock(block) || len(block.Sentences) == 0 {
			continue
		}
		observations.advance(block.ID, candidateEvaluated)
		advice := rhetoricAdvice{emit, "Name the actual configuration or decision. Preserve the stated condition and any necessary quantity; " +
			"this warning does not prove that the count is wrong."}
		if err := evaluateFrames(m, block.Sentences[:1], numericalChoice, "unnamed-numerical-choice", advice); err != nil {
			return err
		}
	}
	return observations.finish(ctx, view)
}

func numericalChoice(clauses []frameClause, index int) (rhetoricalFrame, bool) {
	c := clauses[index]
	if len(clauses) != 1 || c.start != 0 || !c.eligible() || rhetoricQuoted(c) || rhetoricAttributed(c) {
		return rhetoricalFrame{}, false
	}
	if !numericalChoiceLead(c.tokens()) || numericalRequirement(c.sentence.Tokens) {
		return rhetoricalFrame{}, false
	}
	return rhetoricalFrame{parts: []frameClause{c}}, true
}

func numericalChoiceLead(tokens []document.Token) bool {
	if len(tokens) < 11 || tokens[0].Protected || !frameWord(tokens[1], "answers", "choices", "options") {
		return false
	}
	count := framingCount(tokens[0].Normal)
	return count >= 2 && count <= 99 && unnamedChoiceTail(tokens[2:])
}

func unnamedChoiceTail(tokens []document.Token) bool {
	return frameWord(tokens[0], ",") && frameWord(tokens[1], "and", "but") &&
		frameWord(tokens[2], "the") && frameWord(tokens[3], "right", "best") &&
		matches(tokens[4:7], []string{"one", "depends", "on"})
}

func numericalRequirement(tokens []document.Token) bool {
	// An explicit requirement elsewhere in the sentence can justify the
	// count. Negation inside the selection condition remains source context.
	for _, token := range tokens {
		if frameWord(token, "must", "shall", "required", "requires", "require", "exactly", "fixed", "mandatory") {
			return true
		}
	}
	return false
}
