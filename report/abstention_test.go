package report_test

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/rule"
)

// exhaustedRule limits one catalog rule to a single candidate after the engine
// has prepared the document, so only that rule's own budget runs out.
type exhaustedRule struct{ rule.Rule }

func (r exhaustedRule) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	view.MaxCandidates = 1
	return r.Rule.Evaluate(ctx, view, emit)
}

const abstentionDetail = "editorial pattern checks exceed max_candidates"

func abstainedResult(t *testing.T) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	rules := slices.Clone(builtin.Rules())
	for i, implementation := range rules {
		if implementation.Descriptor().ID == "format.em-dash-density" {
			rules[i] = exhaustedRule{implementation}
		}
	}
	engine, err := unswell.New(unswell.Options{Rules: rules, Features: []string{"activation/format.em-dash-density"},
		Config: []byte("version: 1\nextends: [builtin:custom]\nrules:\n  format.em-dash-density: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	text := strings.Repeat("The client opens the connection — then checks the reply from the server. ", 4)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	c.Assert(result.Abstentions, qt.DeepEquals, []unswell.RuleAbstention{{Path: "guide.md", RuleID: "format.em-dash-density",
		RuleVersion: result.Abstentions[0].RuleVersion, Reason: rule.ReasonBudgetExhausted, Detail: abstentionDetail}})
	return result
}

func TestWritersSurfaceRuleAbstentions(t *testing.T) {
	result := abstainedResult(t)
	for _, tc := range []struct{ format, want string }{
		{"text", "  abstained: guide.md: format.em-dash-density abstained (budget_exhausted): " + abstentionDetail + "\n"},
		{"markdown", "- Abstained: guide&#46;md — format&#46;em&#45;dash&#45;density abstained &#40;budget&#95;exhausted&#41;: " +
			"editorial pattern checks exceed max&#95;candidates"},
		{"html", "Rule abstained: guide.md — format.em-dash-density (budget_exhausted): " + abstentionDetail},
	} {
		t.Run(tc.format, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(report.Write(&output, tc.format, result, report.Options{}), qt.IsNil)
			c.Assert(output.String(), qt.Contains, tc.want)
			c.Assert(output.String(), qt.Not(qt.Contains), "error:")
		})
	}
}

func TestSARIFReportsAbstentionsAsWarnings(t *testing.T) {
	c := qt.New(t)
	result := abstainedResult(t)
	var output bytes.Buffer
	c.Assert(report.Write(&output, "sarif", result, report.Options{}), qt.IsNil)
	c.Assert(report.ValidateSARIF(output.Bytes()), qt.IsNil)
	var log struct {
		Runs []struct {
			Invocations []struct {
				ExecutionSuccessful bool `json:"executionSuccessful"`
				Notifications       []struct {
					Level      string            `json:"level"`
					Message    map[string]string `json:"message"`
					Properties map[string]string `json:"properties"`
				} `json:"toolExecutionNotifications"`
			} `json:"invocations"`
			Properties struct {
				Abstentions []unswell.RuleAbstention `json:"abstentions"`
			} `json:"properties"`
		} `json:"runs"`
	}
	c.Assert(json.Unmarshal(output.Bytes(), &log), qt.IsNil)
	invocation := log.Runs[0].Invocations[0]
	c.Assert(invocation.ExecutionSuccessful, qt.IsTrue)
	c.Assert(invocation.Notifications, qt.HasLen, 1)
	c.Assert(invocation.Notifications[0].Level, qt.Equals, "warning")
	c.Assert(invocation.Notifications[0].Message["text"], qt.Equals,
		"guide.md: format.em-dash-density abstained (budget_exhausted): "+abstentionDetail)
	c.Assert(invocation.Notifications[0].Properties["rule_id"], qt.Equals, "format.em-dash-density")
	c.Assert(invocation.Notifications[0].Properties["reason"], qt.Equals, rule.ReasonBudgetExhausted)
	c.Assert(log.Runs[0].Properties.Abstentions, qt.DeepEquals, result.Abstentions)
}

// A saved result keeps its abstentions and the absent activation values, and
// the reader rejects an abstention the manifest does not name or that carries
// an invalid reason.
func TestSavedResultsKeepAndValidateAbstentions(t *testing.T) {
	c := qt.New(t)
	result := abstainedResult(t)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	restored, err := report.Read(bytes.NewReader(saved.Bytes()))
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, result)
	c.Assert(restored.Manifest.AbstainedRules, qt.DeepEquals, []string{"format.em-dash-density"})
	c.Assert(len(restored.Features.Sources[0].Units) > 0, qt.IsTrue)
	for _, unit := range restored.Features.Sources[0].Units {
		c.Assert(unit.Values[0].Number, qt.IsNil)
		c.Assert(unit.Values[0].Reason, qt.Equals, "inapplicable/budget_exhausted")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*unswell.RunResult)
		want   string
	}{
		{"unnamed rule", func(r *unswell.RunResult) { r.Manifest.AbstainedRules = nil }, "manifest does not name the abstaining rules"},
		{"invalid reason", func(r *unswell.RunResult) { r.Abstentions[0].Reason = "Not Valid" }, "invalid rule abstention"},
		{"missing rule", func(r *unswell.RunResult) { r.Abstentions[0].RuleID = "" }, "invalid rule abstention"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			var edited unswell.RunResult
			c.Assert(json.Unmarshal(saved.Bytes(), &edited), qt.IsNil)
			tc.mutate(&edited)
			data, err := json.Marshal(edited)
			c.Assert(err, qt.IsNil)
			_, err = report.Read(bytes.NewReader(data))
			c.Assert(err, qt.ErrorMatches, tc.want)
		})
	}
}
