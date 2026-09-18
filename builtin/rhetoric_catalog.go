package builtin

import (
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func localRhetoricRules() []rule.Rule {
	return []rule.Rule{
		localRhetoricRule("filler.instruction-scaffolding", instructionScaffolding,
			"State the supported action directly; preserve its conditions and optionality.",
			"Remove narrated setup or nested support layers around the action and its method. "+
				"Keep prerequisites, actors, operands, alternatives and limits; a capability is not an obligation.",
			"An affirmative impersonal capability with a transitive configuration or presentation action, "+
				"or a reader-intention clause introducing an instruction. Anaphoric gerund methods, nominalized "+
				"methods and reader-purpose clauses are also recognized. Nested enables/allows plus an action nominalization "+
				"and passive infinitive, and nominal actions performed through a gerund method, also qualify. "+
				"Actor-to-action projection recognizes ability nouns, modal used-to/used-for actions, named method predicates, "+
				"and nested intended-to-enable actions. Narrated prerequisites and repeated same-actor obligations can introduce an action. "+
				"Standalone generic reader enablement requires an adjacent method; concrete capability explanations remain controls. "+
				"Infinitive reader goals admit coordinated actions. Relative operation or actor antecedents can carry nested support; "+
				"capable-of gerunds, imperative assurance and purpose-to-steps announcements also qualify. "+
				"A second method may be anchored by a repeated two-noun object and explicit reader action. "+
				"Preserve modal meaning, named actors and conditions. "+
				"Named actor permissions and simple passive predicates do not. Adjacent method announcements remain related evidence. "+
				"Failure, permission, negation and quoted constructions do not match.",
			[]rule.Example{{Text: "It is possible to display a label. This is done by using the label option.", Match: true},
				{Text: "It is possible to lose data after a failed write."}}),
		localRhetoricRule("repetition.redundant-predicate", redundantPredicate,
			"The predicate repeats the relation already named by its subject.",
			"State the relation once. Preserve the path, actor, conditions and technical distinction.",
			"A category/type label describing its own unqualified category/type, or a path/location subject "+
				"followed by is located/situated at/in, or a reason subject followed by is because. "+
				"The grammatical subject must name the repeated relation; files, identifiers and control-flow conditions do not qualify.",
			[]rule.Example{{Text: "The default path for the configuration file is located at `/etc/example.conf`.", Match: true},
				{Text: "The configuration file is located at `/etc/example.conf`."}}),
		localRhetoricRule("filler.document-justification", documentJustification,
			"State the scope or destination directly instead of explaining why this page exists.",
			"Keep the link, ownership, and scope; remove the explanation of the document's existence or deliberate non-repetition.",
			"A document subject followed by 'exists so/to/because', deliberate non-repetition of material, "+
				"a statement that a document part earns its place, a page explaining what its fields mean, editorial "+
				"ownership, document dedication or future-presentation clauses, or announcements "+
				"that definitions or counts are repeated or omitted here. "+
				"Ownership may use a pronoun only after an adjacent explicit document announcement in the same block. "+
				"Storage existence and ordinary scope exclusions do not match.",
			[]rule.Example{{Text: "This page exists so the operator is reachable from here.", Match: true},
				{Text: "This page does not describe authentication."}}),
		localRhetoricRule("filler.evaluative-closure", evaluativeClosure,
			"This clause announces information or endorses a result; state the information directly.",
			"Keep the behavior, reasons, conditions and instructions. Remove the announcement or judgment while "+
				"preserving the information it introduces.",
			"An anaphoric purpose tail ('which is the point of doing it', 'which is what doing it is for'), "+
				"a bare whole-point/design declaration, an anaphoric assertion of value or verification, "+
				"an information subject announcing cognitive worth, abstract feature-purpose or intentionality closures, "+
				"output claiming to prove understanding, or an impersonal modal notice followed by a that-clause. "+
				"Also recognizes cognitive notices modifying information nouns, gerund-anaphor purpose clefts, "+
				"relative feature-purpose predicates and discourse subjects declaring importance or editorial integrity. "+
				"Information judgments may join adjacent evaluative predicates and retain a quoted noun-phrase subject. "+
				"Discourse subjects admit selecting importance and attention predicates; author notices and document-outcome endorsements qualify. "+
				"Gerund action subjects can close with abstract functional clefts. Bare noun/counts compounds remain excluded. "+
				"Attribution is scoped through the candidate clause; later operational reporting does not erase a preceding judgment. "+
				"Concrete component purposes and measured evaluations stay excluded; "+
				"notice and cognitive-worth frames retain their attached conditions.",
			[]rule.Example{{Text: "The approval stops applying, which is what binding it to a digest is for.", Match: true},
				{Text: "The command starts a new batch, which is what the wrapper is for."}}),
		localRhetoricRule("filler.unscoped-assurance", unscopedAssurance,
			"State the behavior, scope or evidence behind this quality claim.",
			"Keep the technical conditions. Replace a bare quality announcement, nearly-always-right judgment, "+
				"or universal learning promise with specific behavior.",
			"A positive copular clause announces intentional quality without a mechanism, calls a choice nearly always right, "+
				"or promises the fastest/best/easiest way to understand everything. "+
				"Also recognizes affirmative extreme quality predicates and unrestricted claims about "+
				"what everybody or nobody knows, wants or can verify, unsupported majority claims about reader preferences or practice, "+
				"and anaphoric result tails predicting generic readers' preferences or understanding. "+
				"Ordinary copular quality or difficulty judgments require a main subject and a complete predicate without a stated local mechanism. "+
				"Technical adjective-noun phrases, conditional, quantified and protected constructions are excluded.",
			[]rule.Example{{Text: "The configuration is intentionally explicit.", Match: true},
				{Text: "The configuration is explicit about unknown keys."}}),
		localRhetoricRule("repetition.definition-echo", definitionEcho,
			"The definition repeats its subject; state the distinguishing information directly.",
			"State the definition or distinction once. Preserve modifiers, negation and limits.",
			"A positive copula repeats the same nominal phrase, ignoring articles, immediately before ', not' and a nominal alternative. "+
				"Also recognizes a still-qualified repetition of the same nominal or gerund subject, "+
				"including a bounded both-object ellipsis. Added modifiers, "+
				"conditions and code identifiers are excluded; bare uncontrasted identities stay outside the rule.",
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
	d.Version = localRhetoricVersion(id)
	d.Requires = append(d.Requires, nlp.POS)
	d.Description = description + " Reports complete matched clauses, once per clause."
	d.Limitations = "Bounded token constructions, not a semantic or authorship classifier. " +
		"Candidates are limited to 48 tokens; local tails may occur in sentences of up to 96 tokens. " +
		"Protected text is never construction vocabulary. " +
		"Warnings request an editorial review and do not prove that a sentence can be deleted. " +
		"Defaults are experimental policy proposals; review results describe only their declared samples."
	d.Parameters = []string{"window_sentences", "allowed_occurrences", "saturation_occurrences"}
	d.Defaults.Parameters = rule.Parameters{WindowSentences: 8, SaturationOccurrences: 1}
	d.Examples = examples
	if id == "filler.instruction-scaffolding" {
		d.Description += " An immediate anaphoric method may be related across adjacent prose paragraphs, without crossing structural boundaries."
		return check{d, instructionRhetoric(find, id, suggestion)}
	}
	return check{d, localRhetoric(find, id, suggestion)}
}

func localRhetoricVersion(id string) string {
	switch id {
	case "filler.document-justification":
		return "5"
	case "filler.instruction-scaffolding":
		return "10"
	case "filler.evaluative-closure":
		return "9"
	case "repetition.redundant-predicate", "repetition.explanatory-restart":
		return "2"
	case "repetition.definition-echo":
		return "3"
	case "filler.unscoped-assurance":
		return "7"
	default:
		return "1"
	}
}
