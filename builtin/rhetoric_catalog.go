package builtin

import (
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func localRhetoricRules() []rule.Rule {
	return []rule.Rule{
		localRhetoricRule("filler.document-justification", documentJustification,
			"State the scope or destination directly instead of explaining why this page exists.",
			"Keep the link, ownership, and scope; remove the explanation of the document's existence or deliberate non-repetition.",
			"A document subject followed by 'exists so/to/because', deliberate non-repetition of material, "+
				"a statement that a document part earns its place, a page explaining what its fields mean, or editorial ownership. "+
				"Ownership may use a pronoun only after an adjacent explicit document announcement in the same block. "+
				"Storage existence and ordinary scope exclusions do not match.",
			[]rule.Example{{Text: "This page exists so the operator is reachable from here.", Match: true},
				{Text: "This page does not describe authentication."}}),
		localRhetoricRule("filler.evaluative-closure", evaluativeClosure,
			"This evaluative clause announces a purpose or endorses a result; state the consequence directly.",
			"Keep the preceding behavior and its conditions. Remove the closing judgment if it adds no separate reason or instruction.",
			"An anaphoric purpose tail ('which is the point of doing it', 'which is what doing it is for'), "+
				"a bare whole-point/design declaration, an anaphoric assertion of value or verification, "+
				"or a nominal subject announcing that information is worth knowing. "+
				"Named purposes, conditions and measurements are excluded from the added constructions.",
			[]rule.Example{{Text: "The approval stops applying, which is what binding it to a digest is for.", Match: true},
				{Text: "The command starts a new batch, which is what the wrapper is for."}}),
		localRhetoricRule("filler.unscoped-assurance", unscopedAssurance,
			"State the behavior, scope or evidence behind this quality claim.",
			"Keep the technical conditions. Replace a bare quality announcement, nearly-always-right judgment, "+
				"or universal learning promise with specific behavior.",
			"A positive copular clause announces intentional quality without a mechanism, calls a choice nearly always right, "+
				"or promises the fastest/best/easiest way to understand everything. "+
				"Also recognizes affirmative extreme quality predicates and unrestricted claims about "+
				"what everybody or nobody knows, wants or can verify. "+
				"Conditional, quantified and protected constructions are excluded.",
			[]rule.Example{{Text: "The configuration is intentionally explicit.", Match: true},
				{Text: "The configuration is explicit about unknown keys."}}),
		localRhetoricRule("repetition.definition-echo", definitionEcho,
			"The definition repeats its subject before the contrast; explain the distinction directly.",
			"Replace the repeated definition with the behavior that distinguishes the alternatives. Preserve the negation and limitation.",
			"A positive copula repeats the same nominal phrase, ignoring articles, immediately before ', not' and a nominal alternative. "+
				"Added modifiers, conditions, code identifiers, and uncontrasted identity statements are excluded.",
			[]rule.Example{{Text: "The default is a default, not a fallback.", Match: true},
				{Text: "The default is a safe fallback, not a mandatory value."}}),
		localRhetoricRule("syntax.slogan-contrast", sloganContrast,
			"The contrast names abstract actions without their objects; name the check and its failure condition.",
			"Use the concrete verification step or retain the explanation that follows. Do not remove a real alternative or safety condition.",
			"A short nominal subject and a bare check/verify/prove verb opposed to a bare trust/assume/assert verb. "+
				"Both verbs must lack objects and qualifiers; concrete alternatives and refusal/guessing contrasts are excluded.",
			[]rule.Example{{Text: "So Ptah checks, rather than trusts.", Match: true},
				{Text: "The store checks a fencing token rather than trusting a lease it read."}}),
	}
}

func localRhetoricRule(id string, find frameFinder, summary, suggestion, description string, examples []rule.Example) rule.Rule {
	d := descriptor(id, summary, "rhetorical-patterns", "block", 12)
	d.Contexts = []string{"paragraph", "comment", "string", "list-item"}
	d.BlockObservations = true
	d.TermExemptions = true
	if id == "filler.document-justification" || id == "filler.evaluative-closure" {
		d.Version = "3"
	}
	if id == "filler.unscoped-assurance" {
		d.Version = "2"
	}
	d.Requires = append(d.Requires, nlp.POS)
	d.Description = description + " Reports complete matched clauses, once per clause, within a single block."
	d.Limitations = "Bounded token constructions, not a semantic or authorship classifier. " +
		"Clauses are limited to 48 tokens; protected text is never construction vocabulary. " +
		"Warnings request an editorial review and do not prove that a sentence can be deleted. " +
		"Defaults are experimental policy proposals; review results describe only their declared samples."
	d.Parameters = []string{"window_sentences", "allowed_occurrences", "saturation_occurrences"}
	d.Defaults.Parameters = rule.Parameters{WindowSentences: 8, SaturationOccurrences: 1}
	d.Examples = examples
	return check{d, localRhetoric(find, id, suggestion)}
}
