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

	// Rules that produced nothing on any measured corpus. Each names a phrase
	// family the measured generators do not write. They stay in the catalog
	// because a project that does hit one wants it named, and stay off
	// because turning them on adds no finding.
	"filler.empty-transition":            silentRule + "Its list holds openings such as \"with that being said\".",
	"filler.section-announcement":        silentRule + "Its list holds openings such as \"in this section, we will\".",
	"filler.stacked-hedging":             silentRule + "It counts hedges stacked inside one sentence.",
	"hype.absolute-claim":                silentRule + "It looks for a guarantee stated without conditions.",
	"hype.metaphor-cluster":              silentRule + "Its list holds stock metaphors such as \"rich tapestry of possibilities\".",
	"hype.vague-praise":                  silentRule + "Its list holds superlatives offered in place of a measured result.",
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
	"repetition.ngram-density": "Fires an order of magnitude more on human prose, 0.603 findings per thousand " +
		"words against 0.056 in the controlled arm. Parallel technical wording is often deliberate.",
	"repetition.paragraph-overlap": "0.283 per thousand words in human prose and none in the controlled arm.",
	"repetition.syntax-template":   "0.010 per thousand words in human prose and none in the controlled arm.",
	"readability.long-paragraph":   "0.177 per thousand words in human prose against 0.040 in the controlled arm.",
	"filler.weak-intensifiers":     "0.019 per thousand words in human prose against 0.007 in the controlled arm.",
	"syntax.parenthetical-load":    "Fires more on human prose, 0.499 per thousand words against 0.300 in the controlled arm.",

	// Rules measured against a frozen hypothesis and not supported by it.
	"syntax.noun-stack": "Frozen hypothesis H1 of the pattern protocol. Its development estimate of +2.96 points " +
		"fell to +0.41 on the confirmation partition at p 0.727, so the association did not carry to unseen text.",
	"syntax.passive-candidate-density": "Names candidates without a dependency parse, and says so in its own " +
		"limitations. It runs at 1.953 per thousand words in the controlled arm against 1.112 in human prose.",

	// One generator's habit rather than a general property.
	"syntax.paired-contrast-density": "A house tic rather than a signal: none in the controlled arm and 0.007 per " +
		"thousand words in human prose, against 0.161 in one documentation tree. The per-project path is " +
		"\"corpus propose\", which learns a tree's own over-used constructions.",
}

// withOptInReasons attaches the recorded reason to every rule that ships
// disabled. A disabled rule with no recorded reason is a programming error
// and the catalog test reports it.
func withOptInReasons(rules []rule.Rule) []rule.Rule {
	result := make([]rule.Rule, 0, len(rules))
	for _, r := range rules {
		d := r.Descriptor()
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
