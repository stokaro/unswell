package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNounStackStopsAtRecordedPredicates(t *testing.T) {
	for _, row := range []struct{ name, text string }{
		{"existence before coordination", "An unofficial Chinese language documentation translation exists but may be outdated " +
			"and should not be relied upon as the primary reference."},
		{"predicate before colon", "For example, a typical build output path resembles:"},
		{"predicate before period", "Preview bindings can provide an additional preview command, " +
			"even when a default preview command exists."},
		{"predicate after version subject", "Versions 0.13.3, 0.13.2, and 0.13.1 fix duplicate preview lines, " +
			"clearing races, and large ANSI output."},
		{"leading imperative", "fix duplicate preview lines."},
		{"configured predicate before modal", "The analysis completion state applies but may change."},
		{"configured predicate before verb", "The analysis completion state applies and remains relevant."},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "syntax.noun-stack", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}

func TestNounPredicateGuardsKeepNounUsesAndSourceSpans(t *testing.T) {
	for _, row := range []struct{ name, text, want string }{
		{"interior noun", "The bug fix review process is recorded.", "bug fix review process"},
		{"plural head", "The security patch deployment fixes are recorded.", "security patch deployment fixes"},
		{"coordinated noun", "The security patch deployment fixes and improvements are recorded.", "security patch deployment fixes"},
		{"configured noun before coordination", "The service request response stores and caches are recorded.",
			"service request response stores"},
		{"stack before predicate", "The service request response status code exists but may change.", "service request response status code"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "syntax.noun-stack", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Equals, row.want)
			c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.want))
			c.Assert(finding.Primary.Span.End, qt.Equals, strings.Index(row.text, row.want)+len(row.want))
			c.Assert(finding.RuleVersion, qt.Equals, "5")
		})
	}
}

func TestNounPredicateDictionaryCanDeclareDualUseForms(t *testing.T) {
	c := qt.New(t)
	text := "The bug fix review process is recorded."
	result := singleRuleResult(t, "syntax.noun-stack", text, "{verbs: [review], nouns: [review]}", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	result = singleRuleResult(t, "syntax.noun-stack", text, "{verbs: [review], nouns: []}", "")
	c.Assert(result.Findings, qt.HasLen, 0)
	result = singleRuleResult(t, "syntax.noun-stack", text, "", "vocabulary:\n  terms: [bug fix]\n"+
		"  term_exemptions: [syntax.noun-stack]\n")
	c.Assert(result.Findings, qt.HasLen, 0)
	text = "The review process ownership document is recorded."
	result = singleRuleResult(t, "syntax.noun-stack", text, "{verbs: [review], nouns: [review]}", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "review process ownership document")
}
