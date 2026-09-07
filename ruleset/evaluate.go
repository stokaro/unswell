package ruleset

import (
	"context"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type matcher interface {
	evaluate(*evaluation, unit) (matchResult, error)
}

type matchResult struct {
	ok          bool
	occurrences []rule.Occurrence
	metrics     []rule.Metric
}

type sentenceView struct {
	block    *document.Block
	sentence *document.Sentence
}

type unit []sentenceView

type evaluation struct {
	ctx       context.Context
	remaining int
	ends      map[*document.Sentence]int
}

func (e *evaluation) spend(work int) error {
	if err := e.ctx.Err(); err != nil {
		return err
	}
	if work < 0 || work > e.remaining {
		return fmt.Errorf("custom matcher exceeded max_candidates")
	}
	e.remaining -= work
	return nil
}

type compiledRule struct {
	descriptor rule.Descriptor
	match      matcher
	except     []matcher
	message    string
	heuristic  bool
}

// Descriptor returns owned metadata for this immutable rule.
func (r *compiledRule) Descriptor() rule.Descriptor {
	d := r.descriptor
	d.Contexts = slices.Clone(d.Contexts)
	d.Requires = slices.Clone(d.Requires)
	d.Examples = slices.Clone(d.Examples)
	origin := *d.Origin
	d.Origin = &origin
	return d
}

// Evaluate applies the bounded matcher and exceptions to eligible prose units.
func (r *compiledRule) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	if view.Document == nil || view.MaxCandidates <= 0 {
		return fmt.Errorf("custom rule requires a document and a positive candidate budget")
	}
	e := &evaluation{ctx: ctx, remaining: view.MaxCandidates}
	for _, current := range r.units(view.Document) {
		if err := e.spend(1); err != nil {
			return err
		}
		if words(current) == 0 {
			continue
		}
		if err := r.evaluateUnit(e, current, emit); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (r *compiledRule) units(doc *document.Document) []unit {
	var result []unit
	var whole unit
	for i := range doc.Blocks {
		block := &doc.Blocks[i]
		if !slices.Contains(r.descriptor.Contexts, block.Kind) {
			continue
		}
		paragraph := blockUnit(block)
		switch r.descriptor.Scope {
		case "sentence":
			for _, view := range paragraph {
				result = append(result, unit{view})
			}
		case "paragraph":
			result = append(result, paragraph)
		case "document":
			whole = append(whole, paragraph...)
		}
	}
	if r.descriptor.Scope == "document" {
		result = append(result, whole)
	}
	return result
}

func blockUnit(block *document.Block) unit {
	var result unit
	for i := range block.Sentences {
		if block.Sentences[i].Words > 0 {
			result = append(result, sentenceView{block: block, sentence: &block.Sentences[i]})
		}
	}
	return result
}

func (r *compiledRule) evaluateUnit(e *evaluation, current unit, emit rule.Emitter) error {
	for _, exception := range r.except {
		result, err := exception.evaluate(e, current)
		if err != nil {
			return err
		}
		if result.ok {
			return nil
		}
	}
	result, err := r.match.evaluate(e, current)
	if err != nil || !result.ok {
		return err
	}
	if len(result.occurrences) == 0 {
		result.occurrences = unitOccurrences(current)
	}
	result.occurrences = uniqueOccurrences(result.occurrences)
	kind := "exact"
	if r.heuristic {
		kind = "heuristic"
	}
	return emit.Emit(rule.Evidence{Kind: kind, Message: r.message, Activation: 1000,
		Occurrences: result.occurrences, Metrics: result.metrics})
}

func words(current unit) int {
	count := 0
	for _, view := range current {
		count += view.sentence.Words
	}
	return count
}

func tokenOccurrence(sentence *document.Sentence, start, end int) rule.Occurrence {
	result := rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID}
	for _, token := range sentence.Tokens[start:end] {
		if !token.Protected {
			result.Spans = append(result.Spans, token.Spans...)
		}
	}
	return result
}

func unitOccurrences(current unit) []rule.Occurrence {
	var result []rule.Occurrence
	for _, view := range current {
		occurrence := tokenOccurrence(view.sentence, 0, len(view.sentence.Tokens))
		if len(occurrence.Spans) > 0 {
			result = append(result, occurrence)
		}
	}
	return result
}

func uniqueOccurrences(occurrences []rule.Occurrence) []rule.Occurrence {
	slices.SortFunc(occurrences, func(a, b rule.Occurrence) int {
		if a.BlockID != b.BlockID {
			return a.BlockID - b.BlockID
		}
		if a.SentenceID != b.SentenceID {
			return a.SentenceID - b.SentenceID
		}
		left, right := document.Bounds(a.Spans), document.Bounds(b.Spans)
		if left.Start != right.Start {
			return left.Start - right.Start
		}
		if left.End != right.End {
			return left.End - right.End
		}
		return compareSpans(a.Spans, b.Spans)
	})
	return slices.CompactFunc(occurrences, func(a, b rule.Occurrence) bool {
		return a.BlockID == b.BlockID && a.SentenceID == b.SentenceID && document.Bounds(a.Spans) == document.Bounds(b.Spans)
	})
}

func compareSpans(a, b []document.Span) int {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i].Start != b[i].Start {
			return a[i].Start - b[i].Start
		}
		if a[i].End != b[i].End {
			return a[i].End - b[i].End
		}
	}
	return len(a) - len(b)
}

func addOccurrence(result *matchResult, occurrence rule.Occurrence) error {
	if len(result.occurrences) >= 10000 {
		return fmt.Errorf("custom matcher exceeded 10000 occurrences in one unit")
	}
	result.ok = true
	result.occurrences = append(result.occurrences, occurrence)
	return nil
}
