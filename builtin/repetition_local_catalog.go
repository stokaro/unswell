package builtin

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func localRepetitionRules() []rule.Rule {
	word := descriptor("repetition.adjacent-word", "This word occurs twice in succession; check for an accidental duplicate.",
		"repetition", "sentence", 12)
	word.Requires = append(word.Requires, nlp.POS)
	word.Version = "2"
	word.BlockObservations, word.TermExemptions = true, true
	word.Description = "Locates adjacent equal prose words separated only by whitespace. " +
		"Quoted examples, protected text, repeated adjectives/adverbs, and ambiguous grammatical repetitions are excluded. " +
		"Quotation protection follows the repeated tokens, including quotes spanning sentences in one block."
	word.Limitations += " Deliberate repetition and technical compounds can still match; this is not an authorship signal."
	word.Limitations += " Unclosed quotations and ambiguous plural possessives inside single quotes protect the remaining passage."
	word.Examples = []rule.Example{{Text: "The scanner requires requires an external decoder.", Match: true},
		{Text: "The client had had enough time to close the connection."}}
	claim := descriptor("repetition.repeated-claim", "These passages repeat an assertion; check whether the explanation can be shared.",
		"repetition", "document", 15)
	claim.Version = "2"
	claim.Contexts = []string{"paragraph", "comment", "string"}
	claim.Requires = append(claim.Requires, nlp.POS)
	claim.RequiresStructure, claim.BlockObservations, claim.TermExemptions = true, true, true
	claim.Defaults.Parameters = rule.Parameters{MinWords: 5, WindowSentences: 16}
	claim.Parameters = []string{"min_words", "window_sentences"}
	claim.Description = "Compares short complete assertions, assertions containing opaque inline-code atoms, " +
		"and prefixes before a nonrestrictive ', which' clause, within one section and sentence window. " +
		"Case, operands, punctuation, negation and conditions remain part of the exact identity. " +
		"Also compares adjacent explicit reformulations of an only restriction as rejection of its negated property. " +
		"Actors, predicates, tense, properties and objects must agree. A bounded possessive dependency-role conversion " +
		"compares an owner's license restriction with its refusal of code not licensed under that property."
	claim.Limitations += " Requires a surface finite-verb cue; it does not resolve pronouns or prove general semantic equivalence. " +
		"Questions, quoted text, lists, cells, arbitrary code, and claims over 64 tokens are excluded. " +
		"Full unprotected sentences of twelve or more words remain with repetition.exact-sentence. " +
		"A relative-clause prefix must match a complete assertion, not another partial prefix. " +
		"Restriction reformulations require That is or In other words, stay inside one block, and reject " +
		"quantities, conditions, quotations, protected operands and unresolved scope or modality."
	claim.Examples = []rule.Example{
		{Text: "The command exits `1` because drift was found, which confirms the check worked.\n\n" +
			"The command exits `1` because drift was found.", Format: document.Markdown, Match: true},
		{Text: "The command exits `1` because drift was found.\n\n" +
			"The command exits `0` because drift was found.", Format: document.Markdown},
	}
	restart := localRhetoricRule("repetition.explanatory-restart", explanatoryRestart,
		"The explanation repeats a judgment; state it once and name the concrete cause.",
		"State the judgment once. Retain an actual cause and its qualifications; replace a repeated bare judgment with its concrete mechanism.",
		"A nominal subject and a copular adjective are repeated as 'and it/they is/are ADJECTIVE because'. "+
			"The repeated copula and adjective must agree; negation, quantities, conditions and ambiguous noun antecedents are excluded. "+
			"A reason-subject setup also includes a final abstract optimization or design judgment repeating the same adjective.",
		[]rule.Example{{Text: "The process is complex, and it is complex because several callbacks share state.", Match: true},
			{Text: "The process is complex because several callbacks share state."}})
	return []rule.Rule{check{word, adjacentWords}, check{claim, repeatedClaims}, restart}
}
