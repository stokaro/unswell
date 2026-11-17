package builtin

import (
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// habitRules are rules of a page as a whole. Each names a habit that shows
// only when a page keeps to it from its first paragraph to its last, and
// reports the page once.
func habitRules() []rule.Rule {
	denial := surfacePattern("syntax.closing-denial-density",
		"Several paragraphs end by saying what something is not; end most of them on what it is or does.",
		"rhetorical-patterns", "document", 0,
		rule.Example{Text: "# Runner\n\nThe runner reads the old history. It does not rewrite a file.\n\n" +
			"The check compares two schemas. The comparison runs on a copy; it does not call the service.\n\n" +
			"We ran the plan on a copy. This is a rehearsal, not a migration.\n", Format: document.Markdown, Match: true},
		rule.Example{Text: "# Runner\n\nThe runner does not rewrite a file. It reads the old history.\n\n" +
			"The check compares two schemas, and it runs in a minute.\n\n" +
			"We ran the plan on a copy. The copy has the same tables.\n", Format: document.Markdown})
	denial.Contexts, denial.RequiresStructure = []string{"paragraph"}, true
	denial.TermExemptions = false
	denial.Defaults.Enabled = true
	denial.Defaults.Parameters = rule.Parameters{AllowedOccurrences: 2, Onset: 8}
	denial.Parameters = []string{"allowed_occurrences", "onset"}
	denial.Description = "Counts the paragraphs of two sentences or more whose last sentence denies: a last clause after a " +
		"semicolon or a dash that opens with a negation, a tail of the form \", not X\", or a main clause with its negation " +
		"among the first six words. Reports the page once when more than allowed_occurrences paragraphs end that way and " +
		"they are at least onset percent of those paragraphs."
	denial.Limitations += " One denial at the end of a paragraph is an honest limit; the rule reports a habit, not a sentence." +
		" A condition, a question, an instruction, a sentence that addresses the reader, and \"not only\", \"not yet\"," +
		" and the like are not counted. The forms are English, and they are read from the words, not from a parse."
	register := surfacePattern("syntax.uniform-register",
		"This page never leaves one formal register: no contraction, no aside in parentheses, no question, and no "+
			"sentence that opens with And, But, or So. Say one or two things the way you would to a colleague.",
		"rhetorical-patterns", "document", 0,
		rule.Example{Text: "# Server\n\n" + strings.Repeat("The server reads the request and sends the reply to the client. ", 26) + "\n",
			Format: document.Markdown, Match: true},
		rule.Example{Text: "# Server\n\n" + strings.Repeat("The server reads the request and sends the reply to the client. ", 26) +
			"It doesn't keep the request.\n", Format: document.Markdown})
	register.Contexts, register.RequiresStructure = []string{"paragraph"}, true
	register.TermExemptions = false
	register.Defaults.Enabled = true
	register.Defaults.Parameters = rule.Parameters{MinSentences: 25}
	register.Parameters = []string{"min_sentences", "allowed_occurrences"}
	register.Description = "Counts the sentences of running paragraphs and those of them with a casual mark: a question, an " +
		"exclamation, a first word And, But, So, Or, or Yet, a contraction, or parentheses that hold two words or more. " +
		"Reports the page once when it has at least min_sentences sentences and at most allowed_occurrences of them have a mark."
	register.Limitations += " Reference documentation and a specification are often written this way by people, on purpose." +
		" The marks are English. A possessive is not a contraction, and one word in parentheses is not an aside."
	return []rule.Rule{check{denial, closingDenials}, check{register, uniformRegister}}
}
