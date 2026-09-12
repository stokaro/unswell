package unswell_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// assertBudgetAbstention checks that exactly one rule abstained on one document
// with an exhausted budget, that the run stayed complete, that the manifest
// names the rule, and that the rule left no findings.
func assertBudgetAbstention(t *testing.T, result unswell.RunResult, path, ruleID, detail string) {
	t.Helper()
	c := qt.New(t)
	c.Assert(result.Status, qt.Equals, "complete")
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Errors, qt.HasLen, 0)
	c.Assert(result.Abstentions, qt.HasLen, 1)
	abstention := result.Abstentions[0]
	c.Assert(abstention.Path, qt.Equals, path)
	c.Assert(abstention.RuleID, qt.Equals, ruleID)
	c.Assert(abstention.Reason, qt.Equals, rule.ReasonBudgetExhausted)
	c.Assert(abstention.Detail, qt.Contains, detail)
	index := slices.IndexFunc(result.Manifest.Rules, func(d rule.Descriptor) bool { return d.ID == ruleID })
	c.Assert(index >= 0, qt.IsTrue)
	c.Assert(abstention.RuleVersion, qt.Equals, result.Manifest.Rules[index].Version)
	c.Assert(result.Manifest.AbstainedRules, qt.DeepEquals, []string{ruleID})
	for _, finding := range result.Findings {
		c.Assert(finding.RuleID, qt.Not(qt.Equals), ruleID)
	}
}

// budgetLimited wraps one catalog rule so that only its own candidate budget
// is exhausted; the engine-level budgets keep their configured value.
func budgetLimited(rules []rule.Rule, id string) []rule.Rule {
	limited := slices.Clone(rules)
	for i, implementation := range limited {
		if implementation.Descriptor().ID == id {
			limited[i] = candidateBudgetRule{implementation}
		}
	}
	return limited
}

func catalogExample(t *testing.T, id string) string {
	t.Helper()
	for _, implementation := range builtin.Rules() {
		if d := implementation.Descriptor(); d.ID == id {
			return d.Examples[0].Text
		}
	}
	t.Fatalf("no catalog example for %s", id)
	return ""
}

func findingsOf(result unswell.RunResult, ruleID string) []unswell.Finding {
	var selected []unswell.Finding
	for _, finding := range result.Findings {
		if finding.RuleID == ruleID {
			selected = append(selected, finding)
		}
	}
	return selected
}

const abstentionConfig = "version: 1\nextends: [builtin:custom]\nrules:\n" +
	"  format.em-dash-density: {enabled: true, gate: forbid}\n" +
	"  policy.banned-phrases: {enabled: true, gate: forbid, parameters: {phrases: [alpha beta]}}\n"

// One rule running out of budget abstains on that document only; every other
// rule's findings, the gate, and the completion state are those of a normal run.
func TestBudgetAbstentionKeepsOtherRuleFindings(t *testing.T) {
	c := qt.New(t)
	text := catalogExample(t, "format.em-dash-density") + "\n\nAlpha beta describes the setup."
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	normal, err := unswell.New(unswell.Options{Config: []byte(abstentionConfig)})
	c.Assert(err, qt.IsNil)
	want, err := normal.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(want.Abstentions, qt.IsNil)
	c.Assert(want.Manifest.AbstainedRules, qt.IsNil)
	c.Assert(findingsOf(want, "format.em-dash-density"), qt.HasLen, 1)
	c.Assert(findingsOf(want, "policy.banned-phrases"), qt.HasLen, 1)
	c.Assert(want.Gate.Reasons, qt.HasLen, 2)
	limited, err := unswell.New(unswell.Options{Config: []byte(abstentionConfig),
		Rules: budgetLimited(builtin.Rules(), "format.em-dash-density")})
	c.Assert(err, qt.IsNil)
	got, err := limited.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	assertBudgetAbstention(t, got, "guide.md", "format.em-dash-density", "editorial pattern checks exceed max_candidates")
	c.Assert(got.Findings, qt.DeepEquals, findingsOf(want, "policy.banned-phrases"))
	c.Assert(got.Gate.Passed, qt.IsFalse)
	c.Assert(got.Gate.Reasons, qt.HasLen, 1)
	c.Assert(got.Gate.Reasons[0].FindingID, qt.Equals, got.Findings[0].ID)
	c.Assert(got.Documents[0].ProseWords, qt.Equals, want.Documents[0].ProseWords)
}

// Abstentions are merged in source order whatever the worker count, so a
// parallel run equals a sequential one.
func TestAbstentionsStayOrderedAcrossWorkers(t *testing.T) {
	c := qt.New(t)
	text := catalogExample(t, "format.em-dash-density")
	sources := []document.Source{
		{Name: "z.md", Format: document.Markdown, Bytes: []byte(text)},
		{Name: "a.md", Format: document.Markdown, Bytes: []byte(text)},
	}
	var results []unswell.RunResult
	for _, jobs := range []int{1, 2} {
		engine, err := unswell.New(unswell.Options{Config: []byte(abstentionConfig), Jobs: jobs,
			Rules: budgetLimited(builtin.Rules(), "format.em-dash-density")})
		c.Assert(err, qt.IsNil)
		result, err := engine.AnalyzeAll(t.Context(), sources)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Status, qt.Equals, "complete")
		c.Assert(result.Abstentions, qt.HasLen, 2)
		c.Assert(result.Abstentions[0].Path, qt.Equals, "a.md")
		c.Assert(result.Abstentions[1].Path, qt.Equals, "z.md")
		c.Assert(result.Manifest.AbstainedRules, qt.DeepEquals, []string{"format.em-dash-density"})
		results = append(results, result)
	}
	c.Assert(results[1], qt.DeepEquals, results[0])
}

// A permission for the abstaining rule had nothing to cover; it stays unused
// in the audit record without failing the document.
func TestAbstentionExcusesUnusedPermission(t *testing.T) {
	c := qt.New(t)
	text := "<!-- unswell-disable-next-block format.em-dash-density -- Intentional punctuation. -->\n\n" +
		catalogExample(t, "format.em-dash-density")
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	config := []byte(abstentionConfig + "suppressions: {reject_unused: true}\n")
	normal, err := unswell.New(unswell.Options{Config: config})
	c.Assert(err, qt.IsNil)
	want, err := normal.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(want.Suppressions, qt.HasLen, 1)
	c.Assert(want.Suppressions[0].Status, qt.Equals, "used")
	limited, err := unswell.New(unswell.Options{Config: config, Rules: budgetLimited(builtin.Rules(), "format.em-dash-density")})
	c.Assert(err, qt.IsNil)
	got, err := limited.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	assertBudgetAbstention(t, got, "guide.md", "format.em-dash-density", "editorial pattern checks exceed max_candidates")
	c.Assert(got.Suppressions, qt.HasLen, 1)
	c.Assert(got.Suppressions[0].Status, qt.Equals, "unused")
	c.Assert(got.Gate.Passed, qt.IsTrue)
}

type abstainingRule struct {
	reason  string
	invalid bool
}

func (abstainingRule) Descriptor() rule.Descriptor {
	return rule.Descriptor{ID: "test.abstain", Version: "1", Group: "test", Scope: "paragraph",
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "none"}}
}

func (r abstainingRule) Evaluate(_ context.Context, _ rule.View, emit rule.Emitter) error {
	if r.invalid {
		_ = emit.Emit(rule.Evidence{Kind: "bogus"})
	}
	return rule.Abstain(r.reason, nil)
}

// Only a declared, valid reason is an abstention. An invalid reason or a
// latched emitter failure keeps the operational-error path.
func TestUndeclaredAbstentionsStayFatal(t *testing.T) {
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("The client opens a connection.")}
	for _, tc := range []struct {
		name    string
		rule    abstainingRule
		message string
	}{
		{"declared reason", abstainingRule{reason: rule.ReasonBudgetExhausted}, ""},
		{"invalid reason", abstainingRule{reason: "Not Valid"}, "test.abstain: rule abstained: Not Valid"},
		{"emitter failure", abstainingRule{reason: rule.ReasonBudgetExhausted, invalid: true},
			"test.abstain emitted invalid evidence: invalid evidence kind"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{tc.rule},
				Config: []byte("version: 1\nrules:\n  test.abstain: {enabled: true}\n")})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), source)
			if tc.message != "" {
				c.Assert(err, qt.ErrorMatches, tc.message)
				c.Assert(result.Status, qt.Equals, "incomplete")
				c.Assert(result.Abstentions, qt.IsNil)
				return
			}
			c.Assert(err, qt.IsNil)
			assertBudgetAbstention(t, result, "guide.md", "test.abstain", "")
			c.Assert(result.Abstentions[0].Detail, qt.Equals, "")
		})
	}
}
