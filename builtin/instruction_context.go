package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// instructionRhetoric joins only an action announcement and its immediate
// anaphoric method. Other paragraph findings retain their existing block scope.
func instructionRhetoric(find frameFinder, metric, suggestion string) func(context.Context, rule.View, rule.Emitter) error {
	return func(ctx context.Context, view rule.View, emit rule.Emitter) error {
		matcher := newEditorialMatcher(ctx, view)
		runs, err := proseRuns(ctx, view.Document, false)
		if err != nil {
			return err
		}
		advice := rhetoricAdvice{emit, suggestion}
		for _, run := range runs {
			if err := evaluateInstructionRun(matcher, run, find, metric, advice); err != nil {
				return err
			}
		}
		return observeInstructionBlocks(matcher, find, metric, advice)
	}
}

func observeInstructionBlocks(m *editorialMatcher, find frameFinder, metric string, emit rule.Emitter) error {
	for _, block := range m.view.Document.Blocks {
		if err := m.ctx.Err(); err != nil {
			return err
		}
		reason := rhetoricBlockReason(block)
		if reason == "" && block.Kind == "list-item" {
			if err := evaluateFrames(m, block.Sentences, find, metric, emit); err != nil {
				return err
			}
		}
		if err := observeBlock(m.view, block, reason); err != nil {
			return err
		}
	}
	return m.ctx.Err()
}

func evaluateInstructionRun(m *editorialMatcher, run []document.Sentence, find frameFinder, metric string, emit rule.Emitter) error {
	start := 0
	for i := 1; i < len(run); i++ {
		if err := m.spend(); err != nil {
			return err
		}
		if run[i-1].BlockID == run[i].BlockID || instructionBlockLink(run[i-1], run[i]) {
			continue
		}
		if err := evaluateFrames(m, run[start:i], find, metric, emit); err != nil {
			return err
		}
		start = i
	}
	return evaluateFrames(m, run[start:], find, metric, emit)
}

func instructionBlockLink(a, b document.Sentence) bool {
	left, right := frameClauses(a, 0), frameClauses(b, 1)
	if len(left) != 1 || len(right) != 1 || !adjacentInstruction(left[0], right[0]) {
		return false
	}
	for _, token := range left[0].tokens() {
		if frameWord(token, "and", "or", ",") {
			return false
		}
	}
	return instructionAnnouncement(left[0]) && concreteMethod(right[0].tokens())
}

func instructionAnnouncement(c frameClause) bool {
	return capabilityInstruction(c.tokens()) || projectedInstruction(c)
}
