package builtin

import (
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func editorialRules() []rule.Rule {
	rules := editorialPhraseRules()
	rules = append(rules, editorialRhetoricRules()...)
	return append(rules, editorialQualifierRules()...)
}

func experimentalPattern(id, summary, group string, examples ...rule.Example) rule.Descriptor {
	d := descriptor(id, summary, group, "document", 12)
	d.Contexts = []string{"paragraph", "comment", "string"}
	d.Defaults.Enabled = false
	d.Defaults.Parameters = rule.Parameters{WindowSentences: 8, AllowedOccurrences: 1, SaturationOccurrences: 4}
	d.Parameters = []string{"window_sentences", "allowed_occurrences", "saturation_occurrences"}
	d.TermExemptions = true
	config := "version: 1\nrules:\n  " + id + ": {enabled: true}\n"
	for i := range examples {
		examples[i].Config = config
	}
	d.Examples = examples
	return d
}

func editorialPhraseRules() []rule.Rule {
	section := experimentalPattern("filler.section-announcement", "Replace repeated section announcements with their subject.", "scaffolding",
		rule.Example{Text: "In this section, we will describe setup. In this section, we will describe deployment.", Match: true},
		rule.Example{Text: "In this section, we will explain the retry protocol."})
	section.Defaults.Parameters.Phrases = []string{"in this section, we will", "in this section we will", "this section will discuss"}
	transition := experimentalPattern("filler.empty-transition", "Check whether these repeated opening transitions add information.",
		"scaffolding",
		rule.Example{Text: "With that being said, the server starts. It goes without saying that the client waits.", Match: true},
		rule.Example{Text: "With that being said, the client must still validate the response."})
	transition.Defaults.Parameters.Phrases = []string{"with that being said", "that being said", "it goes without saying"}
	metaphor := experimentalPattern("hype.metaphor-cluster", "Replace clustered stock metaphors with specific properties.", "inflation",
		rule.Example{Text: "The service offers a rich tapestry of possibilities. Start a journey of innovation with the client.", Match: true},
		rule.Example{Text: "The renderer loads a landscape texture. The travel API returns the journey duration."})
	metaphor.Defaults.Parameters.Phrases = []string{
		"rich tapestry of possibilities", "journey of innovation", "ever-evolving digital landscape", "unlock a world of possibilities",
	}
	var rules []rule.Rule
	for _, d := range []rule.Descriptor{section, transition, metaphor} {
		d.BlockObservations = true
		d.Parameters = append(d.Parameters, "phrases")
		d.Description = "Counts nonoverlapping configured phrases in bounded prose windows; one occurrence is allowed by default."
		d.Limitations += " Only configured phrases are recognized; literal meanings and local exceptions require editorial review."
		rules = append(rules, check{d, windowPhrases(d.ID != metaphor.ID, d.ID == section.ID)})
	}
	praise := experimentalPattern("hype.vague-praise", "Replace general praise with a specific property or measured result.", "inflation",
		rule.Example{Text: "This is a game-changing solution for the team.", Match: true},
		rule.Example{Text: "The robust estimator tolerates outliers."})
	praise.Defaults.Parameters = rule.Parameters{
		Phrases: []string{"game-changing solution", "unparalleled experience", "unprecedented excellence"},
	}
	absolute := experimentalPattern("hype.absolute-claim", "Check whether this guarantee has conditions or exceptions.", "inflation",
		rule.Example{Text: "The service guarantees complete safety.", Match: true},
		rule.Example{Text: "Under these conditions, the service guarantees complete safety."})
	absolute.Defaults.Score.Weight, absolute.Defaults.Score.Cap = 0, 0
	absolute.Defaults.Parameters = rule.Parameters{
		Phrases: []string{"guarantees complete safety", "eliminates all risk", "guarantees zero downtime", "works in every environment"},
	}
	for _, d := range []rule.Descriptor{praise, absolute} {
		d.Scope = "sentence"
		d.Parameters = []string{"phrases"}
		d.Defaults.Parameters = rule.Parameters{Phrases: d.Defaults.Parameters.Phrases}
		d.Description = "Matches configured phrases in affirmative prose, excluding questions and explicit qualifications conservatively."
		d.Limitations += " This lexical screen does not determine truth, semantic vagueness, or whether every condition is stated."
		rules = append(rules, check{d, editorialClaims})
	}
	return rules
}

func editorialRhetoricRules() []rule.Rule {
	contrast := experimentalPattern("syntax.paired-contrast-density", "Check whether repeated paired contrasts carry distinct information.",
		"rhetorical-patterns",
		rule.Example{Text: "It is not about speed. It is about impact. It is not about tools. It is about outcomes.", Match: true},
		rule.Example{Text: "It is not about throughput. It is about the latency bound."})
	contrast.Description = "Counts adjacent 'It is not about'/'It is about' sentence pairs, including contractions, in bounded prose windows."
	contrast.Limitations += " Only the declared opening structures are recognized; no semantic equivalence is inferred."
	triad := experimentalPattern("syntax.triad-density", "Check repeated triads of evaluative modifiers against specific properties.",
		"inflation",
		rule.Example{Text: "A powerful, seamless, innovative platform starts. " +
			"A robust, transformative, unparalleled experience follows.", Match: true},
		rule.Example{Text: "The request requires authentication, authorization, and a valid checksum."})
	triad.Requires = append(triad.Requires, nlp.POS)
	triad.Defaults.Parameters.Phrases = evaluativeWords()
	triad.Parameters = append(triad.Parameters, "phrases")
	triad.Description = "Counts three-item lists from the evaluative dictionary with at least two adjective tags; one triad is allowed."
	triad.Limitations += " One noun tag is tolerated for an ambiguous modifier. This is a POS/dictionary heuristic, not a dependency parse."
	whether := experimentalPattern("syntax.whether-preface-density", "Reduce repeated whether-you prefaces while preserving alternatives.",
		"rhetorical-patterns",
		rule.Example{Text: "Whether you are a beginner or an expert, the tool helps. " +
			"Whether you're a writer or a reader, the tool fits.", Match: true},
		rule.Example{Text: "Whether you are using HTTP or HTTPS, validate the certificate policy."})
	whether.Description = "Counts whether-you-are prefaces containing 'or' before a comma within 32 tokens; one occurrence is allowed."
	whether.Limitations += " Recognizes one opening template, not arbitrary conditional clauses."
	question := experimentalPattern("syntax.rhetorical-question-density", "Check repeated question-and-short-answer transitions.",
		"rhetorical-patterns",
		rule.Example{Text: "The result? A better experience. The benefit? A brighter future.", Match: true},
		rule.Example{Text: "Does the client retry? Yes. Does it cache errors? No."})
	question.Defaults.Parameters.Phrases = []string{"the result?", "the outcome?", "the benefit?", "why does this matter?"}
	question.Defaults.Parameters.MaxAnswerWords = 12
	question.Parameters = append(question.Parameters, "phrases", "max_answer_words")
	question.Description = "Counts configured complete questions followed in the same block by a short nonquestion answer."
	question.Limitations += " Arbitrary questions are not classified. Numeric/code answers are excluded; disable for FAQ paths."
	return []rule.Rule{
		check{contrast, editorialWindow(pairedContrastEvents, "contrast-pairs")},
		check{triad, editorialWindow(triadEvents, "evaluative-triads")},
		check{whether, editorialWindow(whetherEvents, "whether-prefaces")},
		check{question, editorialWindow(questionEvents, "question-answer-pairs")},
	}
}

func editorialQualifierRules() []rule.Rule {
	weak := experimentalPattern("filler.weak-intensifiers", "Check whether these intensifiers add a measurable distinction.", "inflation",
		rule.Example{Text: "The very powerful client provides a really seamless workflow and an incredibly innovative interface " +
			"for every person using the service.", Match: true},
		rule.Example{Text: "The client really needs the exact identifier to retry safely."})
	weak.Scope = "paragraph"
	weak.Requires = append(weak.Requires, nlp.POS)
	weak.Defaults.Parameters = rule.Parameters{Phrases: []string{"very", "really", "incredibly", "extremely", "highly"},
		MinWords: 20, Onset: 4, Saturation: 12}
	weak.Parameters = []string{"phrases", "min_words", "onset", "saturation"}
	weak.Description = "Counts configured adverbs before adjectives/adverbs per 100 prose words, after the minimum word count."
	weak.Limitations += " POS roles are heuristic. The rule does not require deleting distinctions with technical meaning."
	hedge := experimentalPattern("filler.stacked-hedging",
		"Check whether each hedge expresses a distinct uncertainty; retain needed qualifications.", "inflation",
		rule.Example{Text: "The result may possibly perhaps change.", Match: true},
		rule.Example{Text: "The request may fail, and the server might retry."})
	hedge.Scope = "sentence"
	hedge.Requires = append(hedge.Requires, nlp.POS)
	hedge.Defaults.Parameters = rule.Parameters{Phrases: []string{"may", "might", "could", "perhaps", "possibly", "potentially", "apparently"},
		Onset: 2, Saturation: 5}
	hedge.Parameters = []string{"phrases", "onset", "saturation"}
	hedge.Description = "Counts distinct configured modal/adverb cues in one clause; commas, punctuation, conjunctions, and code end a clause."
	hedge.Limitations += " A token boundary is not a dependency parse. Multiple independent uncertainties can be justified."
	return []rule.Rule{check{weak, weakIntensifiers}, check{hedge, stackedHedging}}
}

func evaluativeWords() []string {
	return []string{"powerful", "seamless", "robust", "innovative", "transformative", "unparalleled"}
}
