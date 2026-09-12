package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
)

const nearContrastSentence = "The client may retry the failed request after the server closes the connection and releases all resources."

func TestNearSentenceTechnicalContrasts(t *testing.T) {
	for _, row := range []struct{ name, source, before, after string }{
		{"permission and obligation", nearContrastSentence, "may", "must"},
		{"recommendation and obligation", nearContrastSentence, "may", "should"},
		{"missing modal", nearContrastSentence, "may ", ""},
		{"condition", nearContrastSentence, "after", "before"},
		{"conditional exception", nearContrastSentence, "after", "unless"},
		{"negation", nearContrastSentence, "may retry", "may never retry"},
		{"state", strings.Replace(nearContrastSentence, "may retry", "retries with enabled settings", 1), "enabled", "disabled"},
		{"unit", strings.Replace(nearContrastSentence, "may retry", "waits 30 seconds to retry", 1), "seconds", "minutes"},
		{"number", strings.Replace(nearContrastSentence, "may retry", "waits 30 seconds to retry", 1), "30", "60"},
		{"version", strings.Replace(nearContrastSentence, "may retry", "uses version 1.2 to retry", 1), "1.2", "1.3"},
		{"identifier", strings.Replace(nearContrastSentence, "may retry", "uses RetryFast to retry", 1), "RetryFast", "RetrySlow"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			left := row.source
			right := strings.Replace(left, row.before, row.after, 1)
			result := singleRuleResult(t, "repetition.near-sentence", left+"\n\n"+right, "", "")
			c.Assert(result.Findings, qt.HasLen, 0)
			c.Assert(result.Gate.Passed, qt.IsTrue)
			control := strings.Replace(left, "failed request", "failed operation", 1)
			positive := singleRuleResult(t, "repetition.near-sentence", left+"\n\n"+control, "", "")
			c.Assert(positive.Findings, qt.HasLen, 1)
			c.Assert(positive.Findings[0].Related, qt.HasLen, 1)
			c.Assert(positive.Gate.Passed, qt.IsFalse)
		})
	}
}

func TestNearSentenceProtectionOverrides(t *testing.T) {
	for _, row := range []struct {
		name, source, before, after, parameters string
		findings                                int
	}{
		{"negation", nearContrastSentence, "may retry", "may never retry", "protect_negation: false", 1},
		{"numbers", strings.Replace(nearContrastSentence, "may retry", "waits 30 seconds to retry", 1),
			"30", "60", "protect_numbers: false", 1},
		{"units", strings.Replace(nearContrastSentence, "may retry", "waits 30 seconds to retry", 1),
			"seconds", "minutes", "protect_numbers: false", 1},
		{"identifiers", strings.Replace(nearContrastSentence, "may retry", "uses RetryFast to retry", 1),
			"RetryFast", "RetrySlow", "protect_identifiers: false", 1},
		{"modality remains protected", nearContrastSentence, "may", "must",
			"protect_negation: false, protect_numbers: false, protect_identifiers: false", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := row.source + "\n\n" + strings.Replace(row.source, row.before, row.after, 1)
			result := singleRuleResult(t, "repetition.near-sentence", text, "{"+row.parameters+"}", "")
			c.Assert(result.Findings, qt.HasLen, row.findings)
		})
	}
}

func TestNearSentenceClusterVersionAndBaseline(t *testing.T) {
	c := qt.New(t)
	policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  repetition.near-sentence: {enabled: true, gate: forbid}\n")
	engine, err := unswell.New(unswell.Options{Config: policy, CollectBaseline: true, IncludeSource: true})
	c.Assert(err, qt.IsNil)
	first := strings.Replace(nearContrastSentence, "failed request", "failed **request**", 1)
	text := "\ufeff" + first + "\r\n\r\n" + strings.Replace(nearContrastSentence, "request", "operation", 1) +
		"\r\n\r\n" + strings.Replace(nearContrastSentence, "resources", "handles", 1) +
		"\r\n\r\n" + strings.Replace(nearContrastSentence, "may", "must", 1)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.RuleVersion, qt.Equals, "2")
	c.Assert(finding.Primary.Snippet, qt.Equals, first)
	c.Assert(finding.Primary.Start.Line, qt.Equals, 1)
	c.Assert(finding.Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(finding.Related, qt.HasLen, 2)
	c.Assert(finding.Related[0].Start.Line, qt.Equals, 3)
	c.Assert(finding.Related[1].Start.Line, qt.Equals, 5)
	c.Assert(result.BaselineSnapshot, qt.IsNotNil)
	candidates := result.BaselineSnapshot.Candidates
	c.Assert(len(candidates) > 0, qt.IsTrue)
	c.Assert(candidates[0].Kind, qt.Equals, "finding")
	c.Assert(candidates[0].RuleVersion, qt.Equals, "2")
	previous := candidates[0]
	previous.RuleVersion = "1"
	oldFingerprint, err := baseline.Fingerprint(previous)
	c.Assert(err, qt.IsNil)
	c.Assert(finding.BaselineFingerprint, qt.Not(qt.Equals), oldFingerprint)
}

func TestNearSentenceCandidateBudget(t *testing.T) {
	c := qt.New(t)
	engine := singleRuleEngine(t, "repetition.near-sentence", "", "analysis: {max_candidates: 1}\n")
	text := nearContrastSentence + "\n\n" + strings.Replace(nearContrastSentence, "may", "must", 1)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	assertBudgetAbstention(t, result, "guide.md", "repetition.near-sentence", "near-sentence candidate index budget exceeded")
	c.Assert(result.Gate.Passed, qt.IsTrue)
}
