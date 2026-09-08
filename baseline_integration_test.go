package unswell_test

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
)

const debtPolicy = `version: 1
extends: [builtin:custom]
rules:
  filler.announced-importance:
    enabled: true
    gate: forbid
gate:
  paragraph_score: {fail_at: 10, min_words: 1}
  sentence_score: {fail_at: 10, min_words: 1}
`

const debtProse = "It is important to note that the client may retry `Upload` 3 times."

func TestBaselineAcceptsMovedDebtWithoutChangingEvidence(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("# Retries\n\n" + debtProse)}
	data, before := captureDebt(t, debtPolicy, source)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	source.Bytes = []byte("\ufeff# **Retries**\r\n\r\nAn unrelated introduction.\r\n\r\n" +
		"It is **important** to note that the client may retry `Upload`\r\n3 times.")
	after, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(before.Gate.Passed, qt.IsFalse)
	c.Assert(after.Gate.Passed, qt.IsTrue)
	c.Assert(after.Gate.Accepted, qt.HasLen, 3)
	c.Assert(after.Baseline.Stale, qt.HasLen, 0)
	c.Assert(after.Findings, qt.HasLen, 3)
	for i, finding := range after.Findings {
		c.Assert(finding.BaselineState, qt.Equals, "existing")
		c.Assert(finding.Suppressed, qt.IsFalse)
		c.Assert(finding.BaselineFingerprint, qt.Equals, before.Findings[i].BaselineFingerprint)
	}
	for _, assessment := range after.Assessments {
		if assessment.SlopScore != 0 {
			c.Assert(assessment.SlopScore, qt.Equals, float64(15))
			c.Assert(assessment.EffectiveSlopScore, qt.Equals, float64(15))
			c.Assert(assessment.SlopProbability, qt.IsNil)
		}
	}
	c.Assert(bytes.Contains(data, []byte("Upload")), qt.IsFalse)
}

func TestBaselineRechecksSemanticAndStructuralChanges(t *testing.T) {
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("# Retries\n\n" + debtProse)}
	data, _ := captureDebt(t, debtPolicy, source)
	for _, text := range []string{
		"# Retries\n\n" + strings.ReplaceAll(debtProse, "may retry", "may not retry"),
		"# Retries\n\n" + strings.ReplaceAll(debtProse, "3 times", "4 times"),
		"# Retries\n\n" + strings.ReplaceAll(debtProse, "Upload", "upload"),
		"# Uploads\n\n" + debtProse,
		"# Retries\n\n" + debtProse + " This sentence changes the paragraph.",
	} {
		c := qt.New(t)
		engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
		c.Assert(err, qt.IsNil)
		source.Bytes = []byte(text)
		result, err := engine.Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Gate.Passed, qt.IsFalse, qt.Commentf("%s", text))
		c.Assert(result.Baseline.Stale, qt.HasLen, 3)
		for _, finding := range result.Findings {
			c.Assert(finding.BaselineState, qt.Equals, "new")
		}
	}
}

func TestBaselineDistinguishesDeclarationsAndRejectsAmbiguity(t *testing.T) {
	c := qt.New(t)
	prose := strings.ReplaceAll(debtProse, "`", "")
	source := document.Source{Name: "sample.go", Format: document.Go,
		Bytes: []byte("package sample\nvar first = \"" + prose + "\"\nvar second = \"" + prose + "\"\n")}
	data, _ := captureDebt(t, debtPolicy, source)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	source.Bytes = []byte(strings.ReplaceAll(string(source.Bytes), "first =", "third ="))
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "new")
	c.Assert(result.Findings[1].BaselineState, qt.Equals, "existing")
	c.Assert(result.Findings[0].ID, qt.Not(qt.Equals), result.Findings[1].ID)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	ambiguous, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	result, err = ambiguous.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte(debtProse + "\n\n" + debtProse)})
	c.Assert(err, qt.ErrorMatches, ".*ambiguous.*")
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.BaselineSnapshot.Complete, qt.IsFalse)
}

func TestBaselineGateModesAndCompatibility(t *testing.T) {
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(debtProse)}
	data, _ := captureDebt(t, debtPolicy, source)
	for _, mode := range []string{"all", "new"} {
		c := qt.New(t)
		policy := strings.ReplaceAll(debtPolicy, "gate: forbid", "gate: forbid\n    severity: error")
		engine, err := unswell.New(unswell.Options{Config: []byte(policy), Baseline: data, GateMode: mode})
		c.Assert(err, qt.IsNil)
		result, err := engine.Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Gate.Passed, qt.Equals, mode == "new")
		c.Assert(result.Findings[0].BaselineState, qt.Equals, "existing")
	}
	c := qt.New(t)
	policy := strings.ReplaceAll(debtPolicy, "gate: forbid", "gate: forbid\n    score: {weight: 30, cap: 30}")
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), Baseline: data, GateMode: "new", NoGate: true})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.ErrorMatches, ".*compatibility changed.*")
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Status, qt.Equals, "incomplete")
	_, err = unswell.New(unswell.Options{GateMode: "new"})
	c.Assert(err, qt.ErrorMatches, ".*requires a baseline")
}

func TestBaselineConcurrentCallsOwnTheirResults(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(debtProse)}
	data, _ := captureDebt(t, debtPolicy, source)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new", Jobs: 4})
	c.Assert(err, qt.IsNil)
	clear(data)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			result, analyzeErr := engine.Analyze(t.Context(), source)
			c.Check(analyzeErr, qt.IsNil)
			c.Check(result.Gate.Passed, qt.IsTrue)
			result.Baseline.Matches[0].State = "corrupted"
			result.BaselineSnapshot.Candidates[0].Path = "other.md"
		})
	}
	workers.Wait()
}

func captureDebt(t *testing.T, policy string, source document.Source) ([]byte, unswell.RunResult) {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), CollectBaseline: true, GateMode: "all"})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	file, err := baseline.Create(t.Context(), *result.BaselineSnapshot)
	c.Assert(err, qt.IsNil)
	data, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	return data, result
}
