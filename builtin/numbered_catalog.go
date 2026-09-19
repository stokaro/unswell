package builtin

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func numberedFramingRule() rule.Rule {
	d := descriptor("filler.numbered-section-framing",
		"The introduction and heading emphasize a count instead of naming the section's subject.", "rhetorical-patterns", "document", 12)
	d.Contexts = []string{"paragraph", "heading"}
	d.RequiresStructure = true
	d.BlockObservations = true
	d.TermExemptions = true
	d.Defaults.Parameters = rule.Parameters{SaturationOccurrences: 1}
	d.Parameters = []string{"allowed_occurrences", "saturation_occurrences"}
	d.Description = "A paragraph opens with a count and a generic content noun; the next heading repeats only that count and noun. " +
		"Reports the opening clause and the related heading. Code, lists, intervening paragraphs, and excluded content break the pair."
	d.Limitations = "An experimental editorial warning, not an authorship detector or a count consistency check. " +
		"Recognizes two through twenty in words and 2 through 99 in digits. Limited to lines, steps, rules, points, things, ways, and tips. " +
		"Explicit requirement and negation cues are excluded. The rule cannot establish whether an unstated count is technically necessary. " +
		"Opening sentences are limited to 96 tokens, matched clauses to 48, and headings to four. " +
		"A named task in the heading, a standalone numerical hook, and nonadjacent headings are outside its scope."
	d.Examples = []rule.Example{
		{Format: document.Markdown, Text: "Four lines of configuration control where requests go.\n\n## The four lines", Match: true},
		{Format: document.Markdown, Text: "Two schemes are supported: env and file.\n\n## Credential references"},
		{Format: document.Markdown, Text: "Four steps initialize the database.\n\n## Four steps to initialize a database"},
	}
	return check{d, numberedFraming}
}
