package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func localRhetoric(find frameFinder, metric, suggestion string) func(context.Context, rule.View, rule.Emitter) error {
	return func(ctx context.Context, view rule.View, emit rule.Emitter) error {
		matcher := newEditorialMatcher(ctx, view)
		for _, block := range view.Document.Blocks {
			if err := ctx.Err(); err != nil {
				return err
			}
			reason := rhetoricBlockReason(block)
			if reason == "" {
				advice := rhetoricAdvice{emit, suggestion}
				if err := evaluateFrames(matcher, block.Sentences, find, metric, advice); err != nil {
					return err
				}
			}
			if err := observeBlock(view, block, reason); err != nil {
				return err
			}
		}
		return ctx.Err()
	}
}

type rhetoricAdvice struct {
	emit       rule.Emitter
	suggestion string
}

// Emit attaches the rule's editing guidance without changing its evidence.
func (r rhetoricAdvice) Emit(evidence rule.Evidence) error {
	evidence.Suggestion = r.suggestion
	return r.emit.Emit(evidence)
}

func rhetoricBlockReason(block document.Block) string {
	if block.Excluded || !slices.Contains([]string{"paragraph", "comment", "string", "list-item"}, block.Kind) {
		return "unsupported_unit"
	}
	if block.Words == 0 {
		return "no_prose_words"
	}
	return ""
}

func localFrame(clause frameClause, start int) (rhetoricalFrame, bool) {
	clause.start += start
	return rhetoricalFrame{parts: []frameClause{clause}}, true
}

func embeddedClauseStart(tokens []document.Token, i int) bool {
	return i == 0 || frameWord(tokens[i-1], ",", "and", "but")
}
