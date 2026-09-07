package builtin

import (
	"context"
	"slices"
	"strings"

	"github.com/jdkato/prose/v3/tokenize"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func phraseRules() []rule.Rule {
	definitions := []struct {
		id, summary, position string
		phrases               []string
		forbid                bool
	}{
		{"policy.banned-phrases", "Remove a phrase forbidden by this editorial policy.", "any", []string{}, true},
		{
			"scaffold.chat-preamble",
			"Start with the subject instead of a conversational acknowledgment.",
			"document-start",
			[]string{"Certainly!", "Great question!", "Absolutely!"},
			true,
		},
		{
			"scaffold.ai-self-reference",
			"Remove the assistant's self-reference.",
			"sentence-start",
			[]string{"As an AI language model", "As an AI assistant"},
			true,
		},
		{
			"scaffold.follow-up-offer",
			"Remove the conversational follow-up offer from this document.",
			"document-end",
			[]string{"Let me know if you'd like", "Let me know if you need"},
			false,
		},
		{
			"scaffold.dive-in",
			"Start with the subject instead of announcing the explanation.",
			"sentence-start",
			[]string{"Let's dive in", "Let's dive into", "Let's delve into"},
			false,
		},
		{
			"filler.announced-importance",
			"State the fact without announcing its importance.",
			"sentence-start",
			[]string{"It is important to note that", "It is worth noting that", "It is crucial to note that"},
			false,
		},
		{
			"filler.modern-world-opening",
			"Replace the generic opening with the document's subject.",
			"sentence-start",
			[]string{"In today's rapidly evolving digital landscape", "In today's fast-paced digital world"},
			false,
		},
		{
			"filler.wordy-phrase",
			"Consider a shorter phrase that preserves the condition or meaning.",
			"any",
			[]string{"due to the fact that", "in order to", "at this point in time"},
			false,
		},
	}
	result := make([]rule.Rule, 0, len(definitions))
	for _, def := range definitions {
		d := descriptor(def.id, def.summary, "scaffolding", "sentence", 15)
		d.Defaults.Parameters = rule.Parameters{Phrases: def.phrases, Positions: []string{def.position}}
		d.Parameters = []string{"phrases", "positions"}
		if def.forbid {
			d.Defaults.Gate = "forbid"
			d.Defaults.Severity = "error"
		}
		if len(def.phrases) > 0 {
			d.Examples = []rule.Example{
				{Text: def.phrases[0] + " the client opens connections.", Match: true},
				{Text: "The client opens connections.", Match: false},
			}
		} else {
			policy := "version: 1\nrules:\n  policy.banned-phrases:\n    parameters:\n      phrases: [magic solution]\n"
			d.Examples = []rule.Example{
				{Text: "This is a magic solution.", Match: true, Config: policy},
				{Text: "This is a pragmatic solution.", Match: false, Config: policy},
			}
		}
		result = append(result, check{descriptor: d, evaluate: phraseEvaluate})
	}
	return result
}

func phraseEvaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	patterns := make([][]string, 0, len(view.Parameters.Phrases))
	for _, phrase := range view.Parameters.Phrases {
		patterns = append(patterns, phraseTokens(phrase))
	}
	sentences := allSentences(view.Document)
	for index, sentence := range sentences {
		if err := ctx.Err(); err != nil {
			return err
		}
		for _, pattern := range patterns {
			if err := matchPhrase(sentence, pattern, view.Parameters.Positions, index, len(sentences), emit); err != nil {
				return err
			}
		}
	}
	return nil
}

func matchPhrase(sentence document.Sentence, pattern, positions []string, index, total int, emit rule.Emitter) error {
	if len(pattern) == 0 {
		return nil
	}
	for start := 0; start+len(pattern) <= len(sentence.Tokens); start++ {
		if !phrasePosition(positions, start, index, total) {
			continue
		}
		if !matches(sentence.Tokens[start:start+len(pattern)], pattern) {
			continue
		}
		occurrence := tokenOccurrence(sentence, start, start+len(pattern))
		evidence := rule.Evidence{Kind: "exact", Occurrences: []rule.Occurrence{occurrence}, Activation: 1000,
			Metrics: []rule.Metric{{Name: "phrase-match", Value: 1, Unit: "matches", Onset: 0, Saturation: 1}}}
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func phraseTokens(text string) []string {
	result := make([]string, 0)
	for _, token := range tokenize.New().Tokenize(document.Normalize(text)) {
		result = append(result, document.Normalize(token.Text))
	}
	return result
}

func matches(tokens []document.Token, pattern []string) bool {
	for i, token := range tokens {
		if token.Protected || token.Normal != pattern[i] {
			return false
		}
	}
	return true
}

func phrasePosition(positions []string, start, index, total int) bool {
	for _, position := range positions {
		if atPosition(position, start, index, total) {
			return true
		}
	}
	return false
}

func atPosition(position string, start, index, total int) bool {
	switch position {
	case "any":
		return true
	case "sentence-start":
		return start == 0
	case "document-start":
		return index == 0 && start == 0
	case "document-end":
		return index == total-1 && start == 0
	}
	return false
}

func allSentences(doc *document.Document) []document.Sentence {
	result := make([]document.Sentence, 0)
	for _, block := range doc.Blocks {
		result = append(result, block.Sentences...)
	}
	return result
}

func tokenOccurrence(sentence document.Sentence, start, end int) rule.Occurrence {
	spans := make([]document.Span, 0)
	for _, token := range sentence.Tokens[start:end] {
		spans = append(spans, token.Spans...)
	}
	return rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: spans}
}

func sentenceOccurrence(sentence document.Sentence) rule.Occurrence {
	return rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: slices.Clone(sentence.Spans)}
}

func normalizedWords(sentence document.Sentence) []string {
	result := make([]string, 0, sentence.Words)
	for _, token := range sentence.Tokens {
		if token.Word {
			result = append(result, token.Normal)
		}
	}
	return result
}

func sentenceKey(sentence document.Sentence) string {
	parts := make([]string, 0, len(sentence.Tokens))
	for _, token := range sentence.Tokens {
		if token.Protected {
			return ""
		}
		parts = append(parts, token.Normal)
	}
	return strings.Join(parts, " ")
}
