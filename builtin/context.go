package builtin

import (
	"context"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func contextRules() []rule.Rule {
	long := descriptor(
		"syntax.long-sentence",
		"Consider splitting this long sentence without losing its conditions.",
		"syntax-load",
		"sentence",
		30,
	)
	long.Defaults.Parameters = rule.Parameters{Onset: 35, Saturation: 65}
	long.Parameters = []string{"onset", "saturation"}
	long.Examples = []rule.Example{
		{Text: strings.Repeat("word ", 40) + "ends.", Match: true},
		{Text: "The client may retry if the connection closes.", Match: false},
	}
	hype := descriptor(
		"hype.modifier-cluster",
		"Replace the cluster of evaluative modifiers with specific properties.",
		"inflation",
		"sentence",
		24,
	)
	hype.Requires = append(hype.Requires, nlp.POS, nlp.Chunks)
	hype.Defaults.Parameters = rule.Parameters{
		Phrases:    []string{"powerful", "seamless", "robust", "innovative", "transformative", "unparalleled"},
		Onset:      2,
		Saturation: 5,
	}
	hype.Parameters = []string{"phrases", "onset", "saturation"}
	hype.Examples = []rule.Example{
		{Text: "The powerful, seamless, innovative platform offers a robust, transformative experience.", Match: true},
		{Text: "The robust estimator tolerates outliers.", Match: false},
	}
	notOnly := descriptor(
		"syntax.not-only-density",
		"Repeated paired contrasts make this passage formulaic.",
		"rhetorical-patterns",
		"document",
		18,
	)
	notOnly.Defaults.Parameters = rule.Parameters{WindowSentences: 8, AllowedOccurrences: 1, SaturationOccurrences: 4}
	notOnly.Parameters = []string{"window_sentences", "allowed_occurrences", "saturation_occurrences"}
	notOnly.Examples = []rule.Example{
		{Text: "It not only reads but also writes. It not only checks but also validates.", Match: true},
		{Text: "The service not only validates requests but also checks permissions.", Match: false},
	}
	connective := descriptor(
		"density.connective-overuse",
		"Remove repeated transitions where the connection is already clear.",
		"rhetorical-patterns",
		"paragraph",
		20,
	)
	connective.Defaults.Parameters = rule.Parameters{
		Phrases:    []string{"moreover", "furthermore", "additionally"},
		MinWords:   30,
		Onset:      1,
		Saturation: 4,
	}
	connective.Parameters = []string{"phrases", "min_words", "onset", "saturation"}
	connective.Examples = []rule.Example{
		{Text: "Moreover, " + strings.Repeat("word ", 16) + "ends. Furthermore, " + strings.Repeat("word ", 16) + "ends.", Match: true},
		{Text: "Furthermore, the server may close the connection.", Match: false},
	}
	result := []rule.Rule{
		check{long, longSentence},
		check{hype, modifierCluster},
		check{notOnly, notOnlyDensity},
		check{connective, connectiveOveruse},
	}
	return append(result, repetitionRules()...)
}

func activation(metric, onset, saturation int) int {
	if saturation <= onset {
		return 0
	}
	return min(1000, max(0, (metric-onset)*1000/(saturation-onset)))
}

func measured(kind, name, unit string, value, onset, saturation int, occurrences []rule.Occurrence) rule.Evidence {
	return rule.Evidence{Kind: kind, Occurrences: occurrences, Activation: activation(value, onset, saturation),
		Metrics: []rule.Metric{{Name: name, Value: float64(value), Unit: unit, Onset: float64(onset), Saturation: float64(saturation)}}}
}

func longSentence(ctx context.Context, view rule.View, emit rule.Emitter) error {
	for _, sentence := range allSentences(view.Document) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if sentence.Words <= view.Parameters.Onset {
			continue
		}
		evidence := measured(
			"exact",
			"length",
			"prose-words",
			sentence.Words,
			view.Parameters.Onset,
			view.Parameters.Saturation,
			[]rule.Occurrence{sentenceOccurrence(sentence)},
		)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func modifierCluster(ctx context.Context, view rule.View, emit rule.Emitter) error {
	lexicon := make(map[string]bool)
	for _, word := range view.Parameters.Phrases {
		lexicon[word] = true
	}
	for _, sentence := range allSentences(view.Document) {
		if err := ctx.Err(); err != nil {
			return err
		}
		count := 0
		occurrence := rule.Occurrence{BlockID: sentence.BlockID, SentenceID: sentence.ID, Spans: []document.Span{}}
		for _, token := range sentence.Tokens {
			if lexicon[token.Normal] && strings.HasPrefix(token.Tag, "JJ") {
				count++
				occurrence.Spans = append(occurrence.Spans, token.Spans...)
			}
		}
		if count <= view.Parameters.Onset {
			continue
		}
		evidence := measured(
			"heuristic",
			"evaluative-modifiers",
			"tokens",
			count,
			view.Parameters.Onset,
			view.Parameters.Saturation,
			[]rule.Occurrence{occurrence},
		)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func notOnlyDensity(ctx context.Context, view rule.View, emit rule.Emitter) error {
	sentences := allSentences(view.Document)
	for start := 0; start < len(sentences); start++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		occurrences := make([]rule.Occurrence, 0)
		last := start
		for offset, sentence := range sentences[start:min(len(sentences), start+view.Parameters.WindowSentences)] {
			words := " " + strings.Join(normalizedWords(sentence), " ") + " "
			if (strings.Contains(words, " not only ") || strings.Contains(words, " not just ")) && strings.Contains(words, " but ") {
				occurrences = append(occurrences, sentenceOccurrence(sentence))
				last = start + offset
			}
		}
		if len(occurrences) <= view.Parameters.AllowedOccurrences {
			continue
		}
		evidence := measured(
			"heuristic",
			"paired-contrasts",
			"sentences",
			len(occurrences),
			view.Parameters.AllowedOccurrences,
			view.Parameters.SaturationOccurrences,
			occurrences,
		)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
		start = last
	}
	return nil
}

func connectiveOveruse(ctx context.Context, view rule.View, emit rule.Emitter) error {
	lexicon := make(map[string]bool)
	for _, phrase := range view.Parameters.Phrases {
		lexicon[phrase] = true
	}
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if block.Words < view.Parameters.MinWords {
			continue
		}
		occurrences := transitionOccurrences(block, lexicon)
		if len(occurrences) <= view.Parameters.Onset {
			continue
		}
		evidence := measured(
			"heuristic",
			"sentence-transitions",
			"occurrences",
			len(occurrences),
			view.Parameters.Onset,
			view.Parameters.Saturation,
			occurrences,
		)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func transitionOccurrences(block document.Block, lexicon map[string]bool) []rule.Occurrence {
	occurrences := make([]rule.Occurrence, 0)
	for _, sentence := range block.Sentences {
		if len(sentence.Tokens) > 0 && lexicon[sentence.Tokens[0].Normal] {
			occurrences = append(occurrences, tokenOccurrence(sentence, 0, 1))
		}
	}
	return occurrences
}
