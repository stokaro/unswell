package claimreview_test

import (
	"math"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
	"github.com/stokaro/unswell/rule"
)

func measuredRule() unswell.Finding {
	return unswell.Finding{ID: "original-finding", RuleID: "syntax.long-sentence", RuleVersion: "1",
		Message: "Consider splitting this long sentence without losing its conditions.",
		Evidence: rule.Evidence{Kind: "exact", Activation: 100,
			Occurrences: []rule.Occurrence{{BlockID: 1, SentenceID: 2, Spans: []document.Span{{Start: 0, End: 20}}}},
			Metrics:     []rule.Metric{{Name: "length", Value: 38, Unit: "prose-words", Onset: 35, Saturation: 65}}}}
}

func TestRuleExplanationPreservesMeasurementsWithoutInventingAnEdit(t *testing.T) {
	c := qt.New(t)
	finding := measuredRule()
	reason, action, err := claimreview.ExplainRule(finding)
	c.Assert(err, qt.IsNil)
	c.Assert(reason, qt.Contains, "syntax.long-sentence v1")
	c.Assert(reason, qt.Contains, "length=38 prose-words; onset=35, saturation=65")
	c.Assert(reason, qt.Contains, "do not prove a wording defect")
	c.Assert(action, qt.Equals, finding.Message)
	c.Assert(finding.Evidence.Message, qt.Equals, "")
	c.Assert(finding.Evidence.Suggestion, qt.Equals, "")
}

func TestRuleExplanationKeepsExplicitEvidenceAndAction(t *testing.T) {
	c := qt.New(t)
	finding := measuredRule()
	finding.Evidence.Message = "The same lead-in occurs twice in the café description."
	finding.Evidence.Suggestion = "Keep the earlier condition while removing the later repeated lead-in."
	reason, action, err := claimreview.ExplainRule(finding)
	c.Assert(err, qt.IsNil)
	c.Assert(reason, qt.Contains, finding.Evidence.Message)
	c.Assert(action, qt.Equals, finding.Evidence.Suggestion)
}

func TestRuleExplanationRejectsMissingOrNonfiniteEvidence(t *testing.T) {
	tests := []struct {
		name string
		edit func(*unswell.Finding)
	}{
		{"missing version", func(f *unswell.Finding) { f.RuleVersion = "" }},
		{"missing diagnostic", func(f *unswell.Finding) { f.Message = "" }},
		{"derived threshold", func(f *unswell.Finding) { f.Derived = true }},
		{"missing metrics", func(f *unswell.Finding) { f.Evidence.Metrics = nil }},
		{"missing occurrences", func(f *unswell.Finding) { f.Evidence.Occurrences = nil }},
		{"bad kind", func(f *unswell.Finding) { f.Evidence.Kind = "authorship" }},
		{"bad activation", func(f *unswell.Finding) { f.Evidence.Activation = 1001 }},
		{"nan value", func(f *unswell.Finding) { f.Evidence.Metrics[0].Value = math.NaN() }},
		{"infinite onset", func(f *unswell.Finding) { f.Evidence.Metrics[0].Onset = math.Inf(1) }},
		{"infinite saturation", func(f *unswell.Finding) { f.Evidence.Metrics[0].Saturation = math.Inf(-1) }},
		{"missing unit", func(f *unswell.Finding) { f.Evidence.Metrics[0].Unit = "" }},
		{"oversized explanation", func(f *unswell.Finding) { f.Evidence.Message = strings.Repeat("x", 2401) }},
		{"oversized action", func(f *unswell.Finding) { f.Evidence.Suggestion = strings.Repeat("x", 2401) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			finding := measuredRule()
			test.edit(&finding)
			reason, action, err := claimreview.ExplainRule(finding)
			c.Assert(err, qt.IsNotNil)
			c.Assert(reason, qt.Equals, "")
			c.Assert(action, qt.Equals, "")
		})
	}
}
