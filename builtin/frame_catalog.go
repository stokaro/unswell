package builtin

import (
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func frameRules() []rule.Rule {
	reframe := frameDescriptor("syntax.repeated-reframing", "denial-redefinition",
		"Repeated denial and redefinition: review whether both clauses add information.")
	reframe.Requires = append(reframe.Requires, nlp.POS)
	reframe.Defaults.Parameters.AllowedOccurrences = 1
	reframe.Description = "Groups adjacent copular denial/redefinition clauses with a repeated subject or anaphoric pronoun. " +
		"Each pair stays in one block; repetition is counted within eight sentences of uninterrupted prose."
	reframe.Examples = []rule.Example{
		{Text: "Backups are not a checkbox. They are your last defense. Testing is not a phase. It is a commitment.", Match: true},
		{Text: "A dev database is not the target. It is a disposable replay target."},
	}
	meta := frameDescriptor("filler.document-metadiscourse", "document-metadiscourse",
		"Document self-description: check whether this announcement or navigation helps the reader.")
	meta.Description = "Recognizes document subjects followed by communicative verbs, and 'in this document, we' announcements. " +
		"Groups their complete clauses within eight sentences of uninterrupted prose; one clause is sufficient."
	meta.Examples = []rule.Example{
		{Text: "This page defines all four roles; other pages link here instead of redefining them.", Match: true},
		{Text: "See the configuration reference for the complete list of environment variables."},
	}
	return []rule.Rule{
		check{reframe, rhetoricalFrames(denialFrames, "denial-redefinition")},
		check{meta, rhetoricalFrames(metadiscourseFrames, "document-metadiscourse")},
	}
}

func frameDescriptor(id, construction, summary string) rule.Descriptor {
	d := descriptor(id, summary, "rhetorical-patterns", "document", 0)
	d.Contexts = []string{"paragraph", "comment", "string"}
	d.BlockObservations = true
	d.TermExemptions = true
	d.Defaults.Severity = "note"
	d.Defaults.Parameters = rule.Parameters{WindowSentences: 8, SaturationOccurrences: 4}
	d.Parameters = []string{"window_sentences", "allowed_occurrences", "saturation_occurrences"}
	d.Limitations = "The " + construction + " frame is a bounded token/POS heuristic, not a dependency or semantic parse. " +
		"Clauses are limited to 48 tokens and subjects to eight. Protected text is opaque, never a lexical cue. " +
		"Recognition does not establish redundancy, authorship, or a need for revision. " +
		"Technical definitions and useful navigation can match; defaults carry no score or gate. Precision is unqualified."
	return d
}
