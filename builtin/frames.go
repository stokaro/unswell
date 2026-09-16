package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// frameClause retains token indexes into the existing NLP result. Protected
// operands remain opaque; their contents never become construction vocabulary.
type frameClause struct {
	sentence            document.Sentence
	start, end, ordinal int
}

type rhetoricalFrame struct {
	parts []frameClause
}

type frameFinder func([]frameClause, int) (rhetoricalFrame, bool)

func rhetoricalFrames(find frameFinder, construction string) func(context.Context, rule.View, rule.Emitter) error {
	return func(ctx context.Context, view rule.View, emit rule.Emitter) error {
		m := newEditorialMatcher(ctx, view)
		runs, err := proseRuns(ctx, view.Document, false)
		if err != nil {
			return err
		}
		for _, run := range runs {
			if err := evaluateFrames(m, run, find, construction, emit); err != nil {
				return err
			}
		}
		for _, block := range view.Document.Blocks {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := observeBlock(view, block, proseMinimumReason(block, 1)); err != nil {
				return err
			}
		}
		return ctx.Err()
	}
}

func evaluateFrames(m *editorialMatcher, run []document.Sentence, find frameFinder, construction string, emit rule.Emitter) error {
	var clauses []frameClause
	for ordinal, sentence := range run {
		if err := m.spend(); err != nil {
			return err
		}
		clauses = append(clauses, frameClauses(sentence, ordinal)...)
	}
	var events []editorialEvent
	for i := range clauses {
		if err := m.spend(); err != nil {
			return err
		}
		frame, ok := find(clauses, i)
		if !ok || frameExempt(m.view, frame) {
			continue
		}
		event := editorialEvent{first: frame.parts[0].ordinal, last: frame.parts[len(frame.parts)-1].ordinal}
		for _, part := range frame.parts {
			event.occurrences = append(event.occurrences, frameOccurrence(part))
		}
		events = append(events, event)
	}
	return emitWindows(m.ctx, m.view.Parameters, construction, events, emit)
}

func frameClauses(sentence document.Sentence, ordinal int) []frameClause {
	var result []frameClause
	start := 0
	for i, token := range sentence.Tokens {
		if !token.Protected && slices.Contains([]string{";", ":", "—", "–", "--", ".", "!", "?"}, token.Normal) {
			result = append(result, frameClause{sentence, start, i, ordinal})
			start = i + 1
		}
	}
	if start < len(sentence.Tokens) {
		result = append(result, frameClause{sentence, start, len(sentence.Tokens), ordinal})
	}
	return result
}

func frameExempt(view rule.View, frame rhetoricalFrame) bool {
	return slices.ContainsFunc(frame.parts, func(part frameClause) bool {
		return view.Exempts(part.sentence, part.start, part.end)
	})
}

// Include opaque operands as source context, as sentence diagnostics do. Only
// the unprotected operators and the declared subject relation establish a frame.
func frameOccurrence(part frameClause) rule.Occurrence {
	return tokenOccurrence(part.sentence, part.start, part.end)
}

func (c frameClause) tokens() []document.Token { return c.sentence.Tokens[c.start:c.end] }

func (c frameClause) eligible() bool {
	return c.end-c.start >= 3 && c.end-c.start <= 48 && !question(c.sentence)
}

func frameWord(token document.Token, words ...string) bool {
	return !token.Protected && slices.Contains(words, token.Normal)
}
