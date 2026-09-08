package builtin

import (
	"strings"

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
	long.BlockObservations = true
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
	hype.BlockObservations = true
	hype.Requires = append(hype.Requires, nlp.POS, nlp.Chunks)
	hype.Defaults.Parameters = rule.Parameters{
		Phrases:    evaluativeWords(),
		Onset:      2,
		Saturation: 5,
	}
	hype.Parameters = []string{"phrases", "onset", "saturation"}
	hype.TermExemptions = true
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
	notOnly.Version = "2"
	notOnly.Description = "Counts ordered not-only/not-just followed by but in bounded prose windows, allowing one matching sentence."
	notOnly.Limitations += " Version 2 stops at structural/protected boundaries and requires the contrast markers in source order."
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
	connective.BlockObservations = true
	connective.Defaults.Parameters = rule.Parameters{
		Phrases:    []string{"moreover", "furthermore", "additionally"},
		MinWords:   30,
		Onset:      1,
		Saturation: 4,
	}
	connective.Parameters = []string{"phrases", "min_words", "onset", "saturation"}
	connective.TermExemptions = true
	connective.Examples = []rule.Example{
		{Text: "Moreover, " + strings.Repeat("word ", 16) + "ends. Furthermore, " + strings.Repeat("word ", 16) + "ends.", Match: true},
		{Text: "Furthermore, the server may close the connection.", Match: false},
	}
	result := []rule.Rule{
		check{long, longSentence},
		check{hype, modifierCluster},
		check{notOnly, editorialWindow(notOnlyEvents, "paired-contrasts", false)},
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
