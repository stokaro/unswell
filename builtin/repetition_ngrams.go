package builtin

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type ngramOccurrence struct {
	sentence            document.Sentence
	ordinal, start, end int
}

type ngramGroup struct {
	words int
	items []ngramOccurrence
}

type ngramPosition struct{ sentence, token int }

type ngramAnalysis struct {
	view         rule.View
	budget       repetitionBudget
	groups       map[string]*ngramGroup
	covered      map[ngramPosition]int
	wordPrefix   []int
	observations *candidateObservations
}

func ngramDensity(ctx context.Context, view rule.View, emit rule.Emitter) error {
	a := ngramAnalysis{view: view, budget: repetitionBudget{ctx, view.MaxCandidates},
		groups: make(map[string]*ngramGroup), covered: make(map[ngramPosition]int), wordPrefix: proseWordPrefix(view.Document)}
	a.observations = newCandidateObservations(view, proseBlock)
	for _, item := range repetitionSentences(view, a.observations) {
		if err := a.addSentence(item); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(a.groups))
	for key := range a.groups {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, func(left, right string) int {
		if a.groups[left].words != a.groups[right].words {
			return a.groups[right].words - a.groups[left].words
		}
		return strings.Compare(left, right)
	})
	for _, key := range keys {
		if err := a.emitGroup(a.groups[key], emit); err != nil {
			return err
		}
	}
	return a.observations.finish(ctx, view)
}

func proseWordPrefix(doc *document.Document) []int {
	prefix := []int{0}
	for _, block := range doc.Blocks {
		if !proseBlock(block) {
			continue
		}
		for _, sentence := range block.Sentences {
			words := 0
			if !protectedSentence(sentence) {
				words = sentence.Words
			}
			prefix = append(prefix, prefix[len(prefix)-1]+words)
		}
	}
	return prefix
}

func (a *ngramAnalysis) addSentence(item repetitionSentence) error {
	sentence := item.sentence
	signature := repetitionSignature(a.view, sentence)
	visits, err := feature.ScanNgrams(a.budget.ctx, sentence.Tokens, feature.NgramOptions{
		MinWords: a.view.Parameters.MinNgramWords, MaxWords: a.view.Parameters.MaxNgramWords,
		MaxVisits: a.budget.remaining,
	}, sequenceLimits(sentence), func(candidate feature.Ngram) error {
		if !a.view.Exempts(sentence, candidate.Start, candidate.End) {
			a.observations.advance(sentence.BlockID, candidateEvaluated)
			a.add(candidate.Key+"\x00"+signature, candidate.End-candidate.Start,
				ngramOccurrence{sentence, item.ordinal, candidate.Start, candidate.End})
		}
		return nil
	})
	if errors.Is(err, feature.ErrTokenLimit) {
		return fmt.Errorf("repetition work exceeds max_candidates: %w", err)
	}
	if err != nil {
		return err
	}
	return a.budget.spend(visits)
}

func (a *ngramAnalysis) add(key string, words int, item ngramOccurrence) {
	group := a.groups[key]
	if group == nil {
		group = &ngramGroup{words: words}
		a.groups[key] = group
	}
	if len(group.items) > 0 {
		last := group.items[len(group.items)-1]
		if last.sentence.ID == item.sentence.ID && last.end > item.start {
			return
		}
	}
	group.items = append(group.items, item)
}

func (a *ngramAnalysis) emitGroup(group *ngramGroup, emit rule.Emitter) error {
	p := a.view.Parameters
	end := 0
	for start := 0; start < len(group.items); {
		if err := a.budget.spend(1); err != nil {
			return err
		}
		end = max(start, end)
		for end < len(group.items) && group.items[end].ordinal-group.items[start].ordinal < p.WindowSentences {
			end++
		}
		if end-start <= p.AllowedOccurrences {
			start++
			continue
		}
		if err := a.emitCluster(group.items[start:end], group.words, emit); err != nil {
			return err
		}
		start = end
	}
	return nil
}

func (a *ngramAnalysis) emitCluster(items []ngramOccurrence, words int, emit rule.Emitter) error {
	covered := true
	for _, item := range items {
		covered = covered && a.covered[ngramPosition{item.sentence.ID, item.start}] >= item.end
	}
	if covered {
		return nil
	}
	occurrences := make([]rule.Occurrence, 0, len(items))
	for _, item := range items {
		occurrences = append(occurrences, tokenOccurrence(item.sentence, item.start, item.end))
		for token := item.start; token < item.end; token++ {
			position := ngramPosition{item.sentence.ID, token}
			a.covered[position] = max(a.covered[position], item.end)
		}
	}
	p := a.view.Parameters
	evidence := measured("heuristic", "ngram-occurrences", "occurrences", len(items),
		p.AllowedOccurrences, p.SaturationOccurrences, occurrences)
	denominator := a.wordPrefix[items[len(items)-1].ordinal+1] - a.wordPrefix[items[0].ordinal]
	evidence.Metrics = append(evidence.Metrics,
		rule.Metric{Name: "ngram-length", Value: float64(words), Unit: "words"},
		rule.Metric{Name: "window-prose-length", Value: float64(denominator), Unit: "prose-words"},
		rule.Metric{Name: "ngram-coverage", Value: float64(len(items)*words) * 100 / float64(denominator), Unit: "matched-words/100-prose-words"})
	return emit.Emit(evidence)
}
