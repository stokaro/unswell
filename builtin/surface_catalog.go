package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func surfacePattern(id, summary, group, scope string, weight int, examples ...rule.Example) rule.Descriptor {
	d := experimentalPattern(id, summary, group, examples...)
	d.Scope = scope
	d.Defaults.Score = rule.Score{Weight: weight, Cap: weight * 2}
	d.Defaults.Parameters = rule.Parameters{}
	d.Parameters = nil
	d.Limitations += " This surface measurement does not establish a grammatical error, a need for revision, or authorship."
	return d
}

func surfaceRules() []rule.Rule {
	rules := surfaceSyntaxRules()
	rules = append(rules, surfaceReadabilityRules()...)
	return append(rules, surfaceFormatRules()...)
}

func surfaceSyntaxRules() []rule.Rule {
	nominal := surfacePattern("syntax.nominalization-chain",
		"Consider a direct verb while preserving the subject, conditions, and technical meaning.", "syntax-load", "sentence", 12,
		rule.Example{Text: "We perform an evaluation of the implementation before the release.", Match: true},
		rule.Example{Text: "The evaluation of the implementation found a defect. The client performs an operation."})
	nominal.BlockObservations = true
	nominal.Requires = append(nominal.Requires, nlp.POS)
	nominal.Defaults.Parameters.Verbs = []string{"perform", "performs", "performed", "performing", "conduct", "conducts", "conducted",
		"conducting", "undertake", "undertakes", "undertook", "undertaken", "make", "makes", "made", "making"}
	nominal.Defaults.Parameters.Nouns = []string{"analysis", "assessment", "evaluation", "examination", "inspection", "investigation",
		"measurement", "review", "verification", "validation", "consideration"}
	nominal.Parameters = []string{"verbs", "nouns"}
	nominal.Description = "Matches a configured weak verb, a configured common noun, and an of-complement with a common-noun head."
	nominal.Limitations += " Inflected verb forms are explicit dictionary entries. Suffixes alone never activate this rule."
	noun := surfacePattern("syntax.noun-stack", "Review this common-noun sequence; keep established technical terms.",
		"syntax-load", "sentence", 8,
		rule.Example{Text: "The service request response status code changes.", Match: true},
		rule.Example{Text: "The access control policy requirements apply.", Match: true},
		rule.Example{Text: "Package document defines source coordinates and the neutral prose model."},
		rule.Example{Text: "Emit latches validation failures even when a custom rule ignores the error."},
		rule.Example{Text: "A binary origin probability cannot measure the fraction of words written by AI."},
		rule.Example{Text: "The request has a status code. The client uses TransportCacheEntry."})
	noun.BlockObservations = true
	noun.Version = "3"
	noun.Requires = append(noun.Requires, nlp.POS, nlp.Chunks)
	noun.Defaults.Parameters = rule.Parameters{Onset: 3, Saturation: 7}
	noun.Parameters = []string{"onset", "saturation"}
	noun.Description = "Counts NN modifiers followed by an NN/NNS head within a shallow NP chunk, stopping at terms and identifiers."
	noun.Limitations += " A shallow NP is not a dependency tree, and a noun sequence can be an appropriate domain term."
	noun.Limitations += " Interior NNS tags cause abstention: they can be plural modifiers or misclassified finite verbs."
	noun.Limitations += " The negative modal \"cannot\" breaks a noun run even when tagged NN; other tagging ambiguity remains possible."
	passive := passiveDescriptor()
	insertion := insertionDescriptor()
	return []rule.Rule{check{nominal, nominalizationChains}, check{noun, nounStacks},
		check{passive, editorialWindow(passiveEvents, "passive-candidate-sentences", false)}, check{insertion, parentheticalLoad}}
}

func passiveDescriptor() rule.Descriptor {
	d := surfacePattern("syntax.passive-candidate-density",
		"Several sentences contain be-plus-participle candidates; review voice only where the meaning allows it.",
		"syntax-load", "document", 6,
		rule.Example{Text: "The request is carefully validated by the server before execution. " +
			"The response is securely recorded by the client after completion.", Match: true},
		rule.Example{Text: "The request must be validated before execution. The client is ready for the next request."})
	d.BlockObservations = true
	d.Requires = append(d.Requires, nlp.POS)
	d.Defaults.Parameters = rule.Parameters{MinWords: 8, WindowSentences: 8, AllowedOccurrences: 1, SaturationOccurrences: 4}
	d.Parameters = []string{"min_words", "window_sentences", "allowed_occurrences", "saturation_occurrences"}
	d.Description = "Counts sentences containing a be verb, up to four adverbs, and VBN in bounded contiguous prose windows."
	d.Limitations += " The same POS shape can describe a state." +
		" This rule is not syntax.passive-density and does not establish grammatical voice."
	return d
}

func insertionDescriptor() rule.Descriptor {
	d := surfacePattern("syntax.parenthetical-load",
		"Review the load of balanced insertions; retain conditions, qualifications, and necessary definitions.",
		"syntax-load", "paragraph", 8,
		rule.Example{Text: "The client (which may retry after a temporary failure when the configured budget still permits another attempt) " +
			"opens a connection to the server.", Match: true},
		rule.Example{Text: "The application programming interface (API) lets the client send requests to the server and receive responses."})
	d.BlockObservations = true
	d.Defaults.Parameters = rule.Parameters{MinWords: 20, Onset: 12, Saturation: 36,
		AllowedOccurrences: 2, SaturationOccurrences: 5, AllowedDepth: 1, SaturationDepth: 4}
	d.Parameters = []string{"min_words", "onset", "saturation", "allowed_occurrences",
		"saturation_occurrences", "allowed_depth", "saturation_depth"}
	d.Description = "Measures nonexempt words, pair count, and nesting of balanced parentheses and square brackets within a prose block."
	d.Limitations += " Unbalanced or protected-crossing frames are not matched. Depth is capped at 256 with an explicit error."
	return d
}

func surfaceReadabilityRules() []rule.Rule {
	long := surfacePattern("readability.long-paragraph",
		"Consider paragraph boundaries while keeping each condition with the statement it qualifies.", "readability", "paragraph", 8,
		rule.Example{Text: strings.Repeat("The client "+strings.Repeat("carefully ", 29)+"waits. ", 4), Match: true},
		rule.Example{Text: strings.Repeat("The client opens connections. ", 40)})
	long.Requires = append(long.Requires, nlp.POS)
	long.TermExemptions = false
	long.SharedFeatures = true
	long.BlockObservations = true
	long.Defaults.Parameters = rule.Parameters{Onset: 120, Saturation: 240, SentenceWords: 25, MinLongSentences: 2}
	long.Parameters = []string{"onset", "saturation", "sentence_words", "min_long_sentences"}
	long.Description = "Requires both paragraph length and a configured count of long sentences; reports local descriptive measurements."
	long.Limitations += " Length, lexical diversity, and POS ratios are descriptive; reference material can justify them."
	grade := surfacePattern("readability.grade-metric",
		"The ARI formula exceeds the selected level; consider the audience and keep necessary technical terms.",
		"readability", "paragraph", 0,
		rule.Example{Text: strings.Repeat("The implementation requires comprehensive configuration, systematic verification, "+
			"and consistent documentation of operational prerequisites before production deployment. ", 4), Match: true},
		rule.Example{Text: strings.Repeat("The client opens a link. The server sends a reply. ", 6)})
	grade.Requires = append(grade.Requires, nlp.POS)
	grade.TermExemptions = false
	grade.SharedFeatures = true
	grade.BlockObservations = true
	grade.Defaults.Parameters = rule.Parameters{MinWords: 50, MinSentences: 2, Onset: 12, Saturation: 20}
	grade.Parameters = []string{"min_words", "min_sentences", "onset", "saturation"}
	grade.Description = "Applies 4.71*characters/words + 0.5*words/sentences - 21.43 to local extracted prose using explicit token counts."
	grade.Limitations += " Characters are Unicode letters and digits within prose tokens." +
		" No age, comprehension, quality, or authorship is predicted."
	return []rule.Rule{check{long, longParagraph}, check{grade, readabilityMetric}}
}

func surfaceFormatRules() []rule.Rule {
	dash := surfacePattern("format.em-dash-density", "Em dashes exceed the selected local density; punctuation can be intentional.",
		"readability", "paragraph", 0,
		rule.Example{Text: strings.Repeat("The client opens the connection — then checks the reply from the server. ", 4), Match: true},
		rule.Example{Text: "The client opens a connection — then waits. " + strings.Repeat("The server sends the response. ", 8)})
	dash.BlockObservations = true
	dash.TermExemptions = false
	dash.Defaults.Parameters = rule.Parameters{MinWords: 40, Onset: 2, Saturation: 6, AllowedOccurrences: 1}
	dash.Parameters = []string{"min_words", "onset", "saturation", "allowed_occurrences"}
	dash.Description = "Counts U+2014 in selected prose per 100 prose words, with mapped source ranges including encoded Markdown entities."
	list := surfacePattern("format.list-fragmentation", "Several short lists fragment this section; check whether the grouping helps readers.",
		"readability", "section", 0,
		rule.Example{Text: "The release offers these benefits.\n\n- Clear reports\n- Useful summaries\n\n" +
			"Its output has these traits.\n\n- Simple navigation\n- Consistent terminology\n\n" +
			"The workflow has these qualities.\n\n- Quick checks\n- Direct feedback", Format: document.Markdown, Match: true},
		rule.Example{Text: "Follow these steps.\n\n1. Open the connection\n2. Send the request\n3. Check the response",
			Format: document.Markdown})
	list.Contexts, list.RequiresStructure = []string{"heading", "paragraph", "list-item"}, true
	list.Requires = append(list.Requires, nlp.POS)
	list.Defaults.Parameters = rule.Parameters{WindowBlocks: 32, MaxItemWords: 10, MaxListItems: 3,
		AllowedOccurrences: 2, SaturationOccurrences: 5}
	list.Parameters = []string{"window_blocks", "max_item_words", "max_list_items", "allowed_occurrences", "saturation_occurrences"}
	list.Description = "Counts complete short flat unordered lists in a bounded section window using grammar-derived list ownership."
	list.Limitations += " Ordered/task/nested lists, sentence items, procedure openings, identifiers," +
		" and selected reference headings are excluded."
	return []rule.Rule{check{dash, emDashDensity}, check{list, listFragmentation}}
}
