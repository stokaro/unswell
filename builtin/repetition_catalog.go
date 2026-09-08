package builtin

import (
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func repetitionPattern(id, summary string, weight int, examples ...rule.Example) rule.Descriptor {
	d := experimentalPattern(id, summary, "repetition", examples...)
	d.Defaults.Score = rule.Score{Weight: weight, Cap: weight * 2}
	d.Defaults.Parameters.MinWords = 12
	d.Defaults.Parameters.ProtectNegation = true
	d.Defaults.Parameters.ProtectNumbers = true
	d.Defaults.Parameters.ProtectIdentifiers = true
	d.Requires = append(d.Requires, nlp.POS)
	d.Parameters = append(d.Parameters, "min_words", "protect_negation", "protect_numbers", "protect_identifiers")
	d.Limitations += " Lexical and POS signals do not establish repeated meaning; technical exceptions require a declared policy."
	return d
}

func extendedRepetitionRules() []rule.Rule {
	ngram := repetitionPattern("repetition.ngram-density", "This phrase recurs in a short prose window; check whether each use is needed.", 12,
		rule.Example{Text: "We describe the delivery process and its approach to dependable service for the team. " +
			"They review the delivery process and its approach to dependable service for the reader. " +
			"You explain the delivery process and its approach to dependable service for the group.", Match: true},
		rule.Example{Text: "The timeout is thirty seconds for a connection attempt, and the client waits for its response."})
	ngram.Defaults.Parameters.MinNgramWords, ngram.Defaults.Parameters.MaxNgramWords = 3, 8
	ngram.Defaults.Parameters.AllowedOccurrences, ngram.Defaults.Parameters.SaturationOccurrences = 2, 5
	ngram.Parameters = append(ngram.Parameters, "min_ngram_words", "max_ngram_words")
	ngram.Description = "Groups repeated consecutive prose n-grams in sentence windows; longer covering phrases take precedence."
	ngram.Limitations += " Activation counts occurrences; reported coverage uses the prose words between the first and last matching sentence."
	template := repetitionPattern("repetition.syntax-template",
		"Several sentences repeat a POS shape; check whether the structure serves the text.", 6,
		rule.Example{Text: "The careful writer describes the simple process for the entire local team. " +
			"The curious reader reviews the clear procedure for the entire small group. " +
			"The skilled editor explains the useful approach for the entire new audience.", Match: true},
		rule.Example{Text: "Open the connection to the server and wait for the complete response. " +
			"Close the connection to the server and wait for the complete response. " +
			"Check the connection to the server and wait for the complete response."})
	template.Defaults.Parameters.AllowedOccurrences, template.Defaults.Parameters.SaturationOccurrences = 2, 5
	template.Description = "Groups complete normalized POS shapes with at least two lexical realizations; " +
		"imperative openings and lists are excluded."
	template.Limitations += " POS grouping is a surface heuristic, not a dependency analysis or a claim that the facts are duplicated."
	return append([]rule.Rule{check{ngram, ngramDensity}, check{template, syntaxTemplates}}, overlapRules()...)
}

func overlapRules() []rule.Rule {
	paragraph := repetitionPattern("repetition.paragraph-overlap",
		"These prose blocks have high lexical overlap; check for repeated information.", 18,
		rule.Example{Text: "The client opens a connection to the server and sends the request with its credentials.\n\n" +
			"The client creates a connection to the server and sends the request with its credentials.", Match: true},
		rule.Example{Text: "The client opens a connection to the server and sends the request with its credentials.\n\n" +
			"The client never opens a connection to the server and sends the request with its credentials."})
	paragraph.Defaults.Parameters.WindowSentences = 0
	paragraph.Defaults.Parameters.WindowBlocks, paragraph.Defaults.Parameters.Similarity = 8, 0.85
	paragraph.Parameters = []string{"min_words", "window_blocks", "similarity", "allowed_occurrences", "saturation_occurrences",
		"protect_negation", "protect_numbers", "protect_identifiers"}
	paragraph.Description = "Builds connected clusters of Jaccard-qualified block pairs through an expiring inverted index."
	paragraph.Limitations += " The block window bounds each pair, not the complete cluster; " +
		"not every pair in a connected cluster must qualify."
	heading := headingEchoDescriptor()
	summary := repetitionPattern("repetition.summary-echo",
		"This summary has high lexical overlap with preceding prose; check what it adds.", 12,
		rule.Example{Text: "The client opens a connection to the server and sends the request with its credentials.\n\n" +
			"## Summary\n\nThe client creates a connection to the server and sends the request with its credentials.",
			Format: document.Markdown, Match: true},
		rule.Example{Text: "The client opens a connection to the server and sends the request with its credentials.\n\n" +
			"## Details\n\nThe client creates a connection to the server and sends the request with its credentials.", Format: document.Markdown})
	summary.Scope, summary.Contexts, summary.RequiresStructure = "section", []string{"heading", "paragraph"}, true
	summary.Defaults.Parameters = paragraph.Defaults.Parameters
	summary.Defaults.Parameters.Phrases = []string{"summary", "conclusion", "conclusions"}
	summary.Parameters = append(slices.Clone(paragraph.Parameters), "phrases")
	summary.Description = "Compares paragraphs under explicitly configured selected summary headings with preceding nonsummary paragraphs."
	summary.Limitations += " Summary scope follows grammar-derived heading ancestry. " +
		"Other titles require configuration; lexical overlap can be intentional."
	return []rule.Rule{check{paragraph, paragraphOverlap}, check{heading, headingEcho}, check{summary, summaryEcho}}
}

func headingEchoDescriptor() rule.Descriptor {
	heading := repetitionPattern("repetition.heading-echo",
		"The first paragraph closely echoes its heading; check whether it adds information.", 6,
		rule.Example{Text: "# A practical approach to the delivery process\n\nA practical approach to the delivery process.",
			Format: document.Markdown, Match: true},
		rule.Example{Text: "# Retry budget\n\nThe retry budget limits the total time spent reconnecting after a transport failure.",
			Format: document.Markdown})
	heading.Scope, heading.Contexts, heading.RequiresStructure = "section", []string{"heading", "paragraph"}, true
	heading.Defaults.Parameters.WindowSentences = 0
	heading.Defaults.Parameters.MinWords, heading.Defaults.Parameters.Similarity = 6, 0.85
	heading.Parameters = []string{"min_words", "similarity", "protect_negation", "protect_numbers", "protect_identifiers"}
	heading.Description = "Compares a selected heading with its immediately following paragraph using exact set Jaccard."
	heading.Limitations += " Short definitions may be useful; code and intervening blocks break adjacency."
	return heading
}
