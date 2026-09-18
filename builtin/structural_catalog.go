package builtin

import "github.com/stokaro/unswell/rule"

func structuralWordingRules() []rule.Rule {
	return []rule.Rule{
		localRhetoricRule("filler.nominal-support", nominalSupport,
			"Consider expressing the action or frequency without this support noun.",
			"Use a direct verb or modifier where it preserves the meaning. Keep the actor, instrument, "+
				"frequency, qualifications and conditions; do not delete the operation or its operands.",
			"An inflected make followed by use of an operand, an instrumental by/through/with the use of "+
				"construction, or a frequency or execution-mode modifier carried by basis or manner. "+
				"Modified use nouns, legal use restrictions, quantities and unrelated nominal subjects do not match.",
			[]rule.Example{{Text: "The reader makes use of the version marker.", Match: true},
				{Text: "The reader makes careful use of the version marker."},
				{Text: "The license restricts use of the renderer."}}),
		localRhetoricRule("repetition.redundant-support", redundantSupport,
			"This clause marks the same relationship twice; check whether one marker is enough.",
			"Keep one marker for the additive, example or sequencing relationship. Preserve distinct coordinated "+
				"actions and any literal comparison, quantity or negation.",
			"An additive also with a clause-final as well, an in-addition-to introduction with also in the "+
				"same main clause, a like-introduced example immediately restated by for example, or adjacent "+
				"then finally sequencing. Separate finite or to-infinitive predicates and literal as-well-as comparisons do not qualify.",
			[]rule.Example{{Text: "The renderer also supports custom labels as well.", Match: true},
				{Text: "The renderer also supports custom labels as well as colors."},
				{Text: "The renderer supports labels, and the reader supports them as well."}}),
	}
}
