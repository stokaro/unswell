package builtin

import (
	"bytes"
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type editorialEvent struct {
	first, last int
	occurrences []rule.Occurrence
}

type eventFinder func(*editorialMatcher, []document.Sentence, int) ([]editorialEvent, error)

func editorialWindow(find eventFinder, metric string) func(context.Context, rule.View, rule.Emitter) error {
	return func(ctx context.Context, view rule.View, emit rule.Emitter) error {
		matcher := newEditorialMatcher(ctx, view)
		runs, err := proseRuns(ctx, view.Document)
		if err != nil {
			return err
		}
		for _, run := range runs {
			events, err := matcher.events(run, find)
			if err != nil {
				return err
			}
			if err := emitWindows(ctx, view.Parameters, metric, events, emit); err != nil {
				return err
			}
		}
		return ctx.Err()
	}
}

func (m *editorialMatcher) events(run []document.Sentence, find eventFinder) ([]editorialEvent, error) {
	var events []editorialEvent
	for i := range run {
		if err := m.ctx.Err(); err != nil {
			return nil, err
		}
		found, err := find(m, run, i)
		if err != nil {
			return nil, err
		}
		events = append(events, found...)
	}
	return events, nil
}

func proseRuns(ctx context.Context, doc *document.Document) ([][]document.Sentence, error) {
	var runs [][]document.Sentence
	var current []document.Sentence
	var previous *document.Block
	flush := func() {
		if len(current) > 0 {
			runs = append(runs, current)
			current = nil
		}
	}
	for i := range doc.Blocks {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		block := &doc.Blocks[i]
		if !adjacentParagraphs(doc, previous, block) {
			flush()
		}
		previous = block
		if !proseBlock(*block) {
			continue
		}
		for _, sentence := range block.Sentences {
			if protectedSentence(sentence) {
				flush()
				continue
			}
			current = append(current, sentence)
		}
	}
	flush()
	return runs, ctx.Err()
}

func adjacentParagraphs(doc *document.Document, before, after *document.Block) bool {
	if before == nil || before.Kind != "paragraph" || after.Kind != "paragraph" {
		return false
	}
	start, end := before.Span.End, after.Span.Start
	return start <= end && end <= len(doc.Source) && len(bytes.TrimSpace(doc.Source[start:end])) == 0
}

func emitWindows(ctx context.Context, p rule.Parameters, metric string, events []editorialEvent, emit rule.Emitter) error {
	end := 0
	for start := 0; start < len(events); {
		if err := ctx.Err(); err != nil {
			return err
		}
		end = max(start, end)
		for end < len(events) && events[end].last-events[start].first < p.WindowSentences {
			end++
		}
		if end-start <= p.AllowedOccurrences {
			start++
			continue
		}
		var occurrences []rule.Occurrence
		for _, event := range events[start:end] {
			occurrences = append(occurrences, event.occurrences...)
		}
		evidence := measured("heuristic", metric, "patterns", end-start, p.AllowedOccurrences, p.SaturationOccurrences, occurrences)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
		start = end
	}
	return nil
}

func phraseEvents(opening bool) eventFinder {
	return func(m *editorialMatcher, sentences []document.Sentence, index int) ([]editorialEvent, error) {
		matches, err := m.phrases(sentences[index], opening)
		if err != nil {
			return nil, err
		}
		var events []editorialEvent
		for _, match := range matches {
			events = append(events, editorialEvent{index, index,
				[]rule.Occurrence{tokenOccurrence(sentences[index], match.start, match.end)}})
		}
		return events, nil
	}
}
