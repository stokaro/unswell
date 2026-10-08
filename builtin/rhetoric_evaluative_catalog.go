package builtin

import "github.com/stokaro/unswell/rule"

func evaluativeClosureRule() rule.Rule {
	return localRhetoricRule("filler.evaluative-closure", evaluativeClosure,
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
			"Document endorsements admit helped/helps with a reader object or a bare/to-infinitive action. "+
			"Gerund action subjects can close with abstract functional clefts. Bare noun/counts compounds remain excluded. "+
			"Discourse subjects contrasting verification with assertion receive a local evidence-commentary explanation; "+
			"a bare it requires an information antecedent in its sentence. "+
			"Concrete quantities, checking actors and explicit methods remain excluded from that construction. "+
			"Bare information-subject/abstract-value identities retain their written offline/online prefix. "+
			"Anaphoric abstract recasting contrasts receive a constraint-focused explanation; "+
			"concrete alternatives and conditions before or after the candidate remain excluded. "+
			"Attribution is scoped through the candidate clause; later operational reporting does not erase a preceding judgment. "+
			"Bare anaphoric question and worth predicates retain local spans before coordinated or causal explanations. "+
			"Deictic conversion appraisals, demonstrative-plan completion metaphors and unqualified abstract value labels "+
			"receive guidance that preserves operational facts and requires verification before weakening claims. "+
			"Literal control loops, explicit criteria and operational consequences remain excluded from these constructions. "+
			"Concrete component purposes and measured evaluations stay excluded; "+
			"notice and cognitive-worth frames retain their attached conditions.",
		[]rule.Example{{Text: "The approval stops applying, which is what binding it to a digest is for.", Match: true},
			{Text: "The command starts a new batch, which is what the wrapper is for."}})
}
