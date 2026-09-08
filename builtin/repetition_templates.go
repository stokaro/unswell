package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func syntaxTemplates(ctx context.Context, view rule.View, emit rule.Emitter) error {
	budget := repetitionBudget{ctx, view.MaxCandidates}
	groups := make(map[string][]repetitionSentence)
	for _, item := range repetitionSentences(view) {
		if err := budget.spend(len(item.sentence.Tokens)); err != nil {
			return err
		}
		key := syntaxTemplate(view, item.sentence)
		if key != "" {
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
	return ctx.Err()
}

func syntaxTemplate(view rule.View, sentence document.Sentence) string {
	if len(sentence.Tokens) == 0 || sentence.Tokens[0].Tag == "VB" || question(sentence) {
		return ""
	}
	parts := make([]string, 0, len(sentence.Tokens)+1)
	for i, token := range sentence.Tokens {
		if token.Protected || token.Tag == "" {
			return ""
		}
		part := templateTag(token.Tag)
		if !token.Word || view.Exempts(sentence, i, i+1) {
			part = token.Normal
		}
		parts = append(parts, part)
	}
	parts = append(parts, repetitionSignature(view, sentence))
	return strings.Join(parts, "|")
}

func templateTag(tag string) string {
	for _, prefix := range []string{"NN", "VB", "JJ", "RB"} {
		if strings.HasPrefix(tag, prefix) {
			return prefix
		}
	}
	return tag
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
