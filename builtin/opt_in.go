package builtin

import "github.com/stokaro/unswell/rule"

// optInReasons says why each rule that ships disabled ships that way.
// Leaving a rule off is a policy decision. Recording it here puts the reason
// beside the rule, so a reader does not have to infer it from a line missing
// in a profile.
//
// The rates quoted are findings per thousand prose words under the E1 policy,
// which enables every rule. The human figure covers the 3.60 million words of
// the historical cohort; the generated figure covers the 426 thousand words of
// the controlled arm, which is model output under recorded prompts. The
// documentation tree is the 411 thousand words of stokaro/ptah at cf75c79d,
// stated by its author to be model-written and never proofread.
// docs/rule-defaults.md carries the measurement and the decision rule.
// silentRule opens the reason of every rule that produced nothing anywhere.
// Each entry continues it with what that rule looks for, so a reader learns
// the rule's subject from the same line that says why it is off.
const silentRule = "No occurrence in 5.6 million measured prose words, human or generated. "

var optInReasons = map[string]string{
	// Surface measurements. These describe a text rather than name a defect,
	// so they stay opt-in whatever they measure.
	"format.em-dash-density": "Surface measurement. It is the one measure elevated in generated prose, " +
		"0.070 per thousand words against 0.001 in human prose, so the strict profile turns it on. " +
		"Elsewhere it stays opt-in: it reports density, not an error.",
	"readability.grade-metric": "Surface measurement of sentence and word length. It runs at 2.964 per thousand " +
		"words in the controlled arm against 0.711 in human prose, but it was frozen hypothesis H2 and returned " +
		"+0.32 points on the confirmation partition at p 0.747. Length is not wording.",
	"format.list-fragmentation": "Surface measurement of list shape. " + silentRule,
	"syntax.paired-contrast-density": "Surface measurement of contrast frequency, not redundant information. " +
		"On 36 complete exposed development pages, assistant review accepted one of 30 findings; " +
		"27 were rejected and two uncertain. Necessary technical distinctions dominate its review load. " +
		"The unchanged matcher remains available through explicit configuration.",

	// Rules that produced nothing on any measured corpus. Each names a phrase
	// family the measured generators do not write. They stay in the catalog
	// because a project that does hit one wants it named, and stay off
	// because turning them on adds no finding.
	"filler.empty-transition":            silentRule + "Its list holds openings such as \"with that being said\".",
	"filler.section-announcement":        silentRule + "Its list holds openings such as \"in this section, we will\".",
	"filler.stacked-hedging":             silentRule + "It counts hedges stacked inside one sentence.",
	"hype.absolute-claim":                silentRule + "It looks for a guarantee stated without conditions.",
	"hype.metaphor-cluster":              silentRule + "Its list holds stock metaphors such as \"rich tapestry of possibilities\".",
	"repetition.heading-echo":            silentRule + "It compares a heading against the paragraph below it.",
	"repetition.summary-echo":            silentRule + "It compares a closing section against the body it summarizes.",
	"syntax.rhetorical-question-density": silentRule + "It counts questions a paragraph asks and answers itself.",
	"syntax.triad-density":               silentRule + "It counts three-item lists used as a rhythm device.",
	"syntax.whether-preface-density":     silentRule + "It counts prefaces of the form \"whether or not X\".",
	"syntax.nominalization-chain": "Effectively silent: 0.001 findings per thousand words in human prose " +
		"and none in the controlled arm.",

	// Rules whose warning load falls mainly on human technical prose. They
	// are editorially sound and stay in the catalog; a project that wants
	// that load asks for it.
	"repetition.sentence-openers": "Its findings fall mainly on deliberate parallel wording. On the assistant-panel " +
		"corpus two reviewers both rejected 56 of 77 findings and both accepted four. Repeated form need not mean redundant information.",
	"repetition.paragraph-openers": "Its findings fall mainly on deliberate parallel wording. On the assistant-panel " +
		"corpus two reviewers both rejected 32 of 39 findings and both accepted none. Repeated openings do not establish duplicate meaning.",
	"filler.unscoped-assurance": "On real pages two reviewers both accepted 14 of 48 findings and both rejected 21. " +
		"A neighboring sentence explained most of the rejected claims, and a local rule does not read it.",
	"filler.document-metadiscourse": "On real pages two reviewers both accepted 60 of 133 findings and both rejected 31. " +
		"They disagreed on the rest: a statement of a page's scope helps one reader and delays another.",
	"repetition.ngram-density": "Fires an order of magnitude more on human prose, 0.603 findings per thousand " +
		"words against 0.056 in the controlled arm. Parallel technical wording is often deliberate.",
	"repetition.paragraph-overlap": "0.283 per thousand words in human prose and none in the controlled arm.",
	"repetition.syntax-template":   "0.010 per thousand words in human prose and none in the controlled arm.",
	"readability.long-paragraph":   "0.177 per thousand words in human prose against 0.040 in the controlled arm.",
	"filler.weak-intensifiers":     "0.019 per thousand words in human prose against 0.007 in the controlled arm.",
	"syntax.parenthetical-load":    "Fires more on human prose, 0.499 per thousand words against 0.300 in the controlled arm.",

	// Rules whose findings fall on documentation that people wrote. The shares
	// are of 121 pages written before November 2022; the strict profile turns
	// these rules on.
	"filler.wordy-phrase": "Reports on 24.8% of pages that people wrote before language models, mostly for one " +
		"purpose phrase. It does not follow what people change when they remove AI-sounding wording. On in strict.",
	"syntax.long-sentence": "Reports on 18.2% of pages that people wrote before language models. " +
		"A long sentence is a matter of lean wording, not of AI-sounding wording. On in strict.",
	"filler.instruction-scaffolding": "Reports on 9.1% of pages that people wrote before language models. On in strict.",
	"hype.vague-praise": "Reports on 9.1% of pages that people wrote before language models, mostly for one adjective. " +
		"On files before and after a de-slop commit it moves with the edit in 55% of pairs. On in strict.",
	"repetition.adjacent-word": "Reports on 7.4% of pages that people wrote before language models. " +
		"A doubled word is a typing error, not AI-sounding wording. On in strict.",
	"format.decorative-emoji": "Reports on 5.8% of pages that people wrote before language models, and is not confirmed " +
		"on likely-manual de-slop edits: lower after the edit in 20 pairs, higher in 14. On in strict.",

	// Rules measured against a frozen hypothesis and not supported by it.
	"syntax.noun-stack": "Version 4 did not establish a cross-family association in the complete-document study. " +
		"Version 5 repairs noun/verb boundaries but has no independent editorial qualification.",
	"syntax.passive-candidate-density": "Names candidates without a dependency parse, and says so in its own " +
		"limitations. It runs at 1.953 per thousand words in the controlled arm against 1.112 in human prose.",
}

// humanLoad lists the rules that ship disabled because their findings fall on
// documentation that people wrote before language models, and do not follow
// what people change when they remove AI-sounding wording. Each one reports on
// more than one in twenty of 121 such pages, and together they are lower after
// a likely-manual de-slop edit in 23 pairs and higher in 26. The emoji rule
// reports on 7 of the 121 pages and is not confirmed on those edits either.
// The strict profile turns all of them on. ADR 0058 records the measurement.
var humanLoad = map[string]bool{
	"filler.wordy-phrase":            true,
	"syntax.long-sentence":           true,
	"filler.instruction-scaffolding": true,
	"hype.vague-praise":              true,
	"repetition.adjacent-word":       true,
	"format.decorative-emoji":        true,
}

// withOptInReasons turns off the rules of humanLoad and attaches the recorded
// reason to every rule that ships disabled. A disabled rule with no recorded
// reason is a programming error and the catalog test reports it.
func withOptInReasons(rules []rule.Rule) []rule.Rule {
	result := make([]rule.Rule, 0, len(rules))
	for _, r := range rules {
		d := r.Descriptor()
		if humanLoad[d.ID] {
			d.Defaults.Enabled = false
			// The examples of a rule that ships disabled turn it on, as those of every other opt-in rule do.
			examples := make([]rule.Example, len(d.Examples))
			for index, example := range d.Examples {
				if example.Config == "" {
					example.Config = "version: 1\nrules:\n  " + d.ID + ": {enabled: true}\n"
				}
				examples[index] = example
			}
			d.Examples = examples
		}
		if !d.Defaults.Enabled {
			d.OptInReason = optInReasons[d.ID]
		}
		result = append(result, described{Rule: r, descriptor: d})
	}
	return result
}

// described carries a descriptor that replaces the wrapped rule's own.
type described struct {
	rule.Rule
	descriptor rule.Descriptor
}

// Descriptor returns the descriptor carrying the recorded opt-in reason.
func (d described) Descriptor() rule.Descriptor { return d.descriptor }
