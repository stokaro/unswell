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
	word.BlockObservations, word.TermExemptions = true, true
	word.Description = "Locates adjacent equal prose words separated only by whitespace. " +
		"Quoted examples, protected text, repeated adjectives/adverbs, and ambiguous grammatical repetitions are excluded."
	word.Limitations += " Deliberate repetition and technical compounds can still match; this is not an authorship signal."
	word.Examples = []rule.Example{{Text: "The scanner requires requires an external decoder.", Match: true},
		{Text: "The client had had enough time to close the connection."}}
	claim := descriptor("repetition.repeated-claim", "These passages repeat an assertion; check whether the explanation can be shared.",
		"repetition", "document", 15)
	claim.Contexts = []string{"paragraph", "comment", "string"}
	claim.Requires = append(claim.Requires, nlp.POS)
	claim.RequiresStructure, claim.BlockObservations, claim.TermExemptions = true, true, true
	claim.Defaults.Parameters = rule.Parameters{MinWords: 5, WindowSentences: 16}
	claim.Parameters = []string{"min_words", "window_sentences"}
	claim.Description = "Compares short complete assertions, assertions containing opaque inline-code atoms, " +
		"and prefixes before a nonrestrictive ', which' clause, within one section and sentence window. " +
		"Case, operands, punctuation, negation and conditions remain part of the identity."
	claim.Limitations += " Requires a surface finite-verb cue; it does not resolve pronouns or prove semantic equivalence. " +
		"Questions, quoted text, lists, cells, arbitrary code, and claims over 64 tokens are excluded. " +
		"Full unprotected sentences of twelve or more words remain with repetition.exact-sentence. " +
		"A relative-clause prefix must match a complete assertion, not another partial prefix."
	claim.Examples = []rule.Example{
		{Text: "The command exits `1` because drift was found, which confirms the check worked.\n\n" +
			"The command exits `1` because drift was found.", Format: document.Markdown, Match: true},
		{Text: "The command exits `1` because drift was found.\n\n" +
			"The command exits `0` because drift was found.", Format: document.Markdown},
	}
	restart := localRhetoricRule("repetition.explanatory-restart", explanatoryRestart,
		"The explanation repeats its initial judgment before giving the reason.",
		"State the judgment once and retain the actual cause and its qualifications.",
		"A nominal subject and a copular adjective are repeated as 'and it/they is/are ADJECTIVE because'. "+
			"The repeated copula and adjective must agree; negation, quantities, conditions and ambiguous noun antecedents are excluded.",
		[]rule.Example{{Text: "The process is complex, and it is complex because several callbacks share state.", Match: true},
			{Text: "The process is complex because several callbacks share state."}})
	return []rule.Rule{check{word, adjacentWords}, check{claim, repeatedClaims}, restart}
}
