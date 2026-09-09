package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func syntaxTemplates(ctx context.Context, view rule.View, emit rule.Emitter) error {
	observations := newCandidateObservations(view, proseBlock)
	budget := repetitionBudget{ctx, view.MaxCandidates}
	groups := make(map[string][]repetitionSentence)
	for _, item := range repetitionSentences(view, observations) {
		if err := budget.spend(len(item.sentence.Tokens)); err != nil {
			return err
		}
		key, err := syntaxTemplate(ctx, view, item.sentence)
		if err != nil {
			return err
		}
		if key != "" {
			observations.advance(item.sentence.BlockID, candidateEvaluated)
			groups[key] = append(groups[key], item)
		}
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if err := emitTemplateWindows(&budget, view.Parameters, groups[key], emit); err != nil {
			return err
		}
	}
	return observations.finish(ctx, view)
}

func syntaxTemplate(ctx context.Context, view rule.View, sentence document.Sentence) (string, error) {
	literal := make([]bool, len(sentence.Tokens))
	for i := range sentence.Tokens {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		literal[i] = view.Exempts(sentence, i, i+1)
	}
	pattern, err := feature.PreparePOSPattern(ctx, sentence, literal, sequenceLimits(sentence))
	if err != nil || !pattern.Available() {
		return "", err
	}
	return pattern.Key() + "|" + repetitionSignature(view, sentence), nil
}

func emitTemplateWindows(budget *repetitionBudget, p rule.Parameters, items []repetitionSentence, emit rule.Emitter) error {
	end := 0
	for start := 0; start < len(items); {
		if err := budget.spend(1); err != nil {
			return err
		}
		end = max(start, end)
		for end < len(items) && items[end].ordinal-items[start].ordinal < p.WindowSentences {
			end++
		}
		if end-start <= p.AllowedOccurrences {
			start++
			continue
		}
		if err := emitTemplateCluster(p, items[start:end], emit); err != nil {
			return err
		}
		start = end
	}
	return nil
}

func emitTemplateCluster(p rule.Parameters, items []repetitionSentence, emit rule.Emitter) error {
	lexicalForms := make(map[string]bool)
	occurrences := make([]rule.Occurrence, 0, len(items))
	for _, item := range items {
		lexicalForms[sentenceKey(item.sentence)] = true
		occurrences = append(occurrences, sentenceOccurrence(item.sentence))
	}
	if len(lexicalForms) < 2 {
		return nil
	}
	evidence := measured("heuristic", "template-occurrences", "sentences", len(items),
		p.AllowedOccurrences, p.SaturationOccurrences, occurrences)
	evidence.Metrics = append(evidence.Metrics,
		rule.Metric{Name: "lexical-realizations", Value: float64(len(lexicalForms)), Unit: "sentences"})
	return emit.Emit(evidence)
}
