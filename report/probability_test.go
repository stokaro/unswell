package report_test

import (
	"bytes"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func modelResult(estimate *float64) unswell.RunResult {
	return unswell.RunResult{
		SchemaVersion: unswell.SchemaVersion, Status: "complete",
		Documents: []unswell.DocumentResult{{Name: "guide.md", Format: document.Markdown}},
		Findings:  []unswell.Finding{},
		Assessments: []unswell.Assessment{{Path: "guide.md", Scope: "sentence", Words: 9, Status: "available",
			SlopProbability: estimate, ProbabilityStatus: "available"},
			{Path: "guide.md", Scope: "sentence", Words: 3, Status: "available", ProbabilityStatus: "insufficient_evidence"},
			{Path: "guide.md", Scope: "paragraph", Words: 12, Status: "available", ProbabilityStatus: "unsupported_unit"}},
		Errors: []unswell.RunError{},
		Gate:   unswell.GateDecision{Passed: true, Reasons: []unswell.GateReason{}},
		Manifest: unswell.Manifest{ToolVersion: unswell.Version, SelectionMode: "explicit", Complete: true,
			Probability: &unswell.ProbabilityModel{PackID: "fixture-pack", SHA256: "0", Version: "unswell-probability-pack-v1",
				Task: "editorial_needs_revision", Rubric: "fixture-rubric-v1", Kind: "sentence",
				DeclaredStatus: "experimental", HumanCorpus: "not_qualified", MinWords: 5}},
	}
}

// Declared pack values are data, so each format escapes them its own way.
func TestHumanReportsDescribeAConfiguredProbabilityModel(t *testing.T) {
	value := 0.5
	for _, row := range []struct{ format, identity string }{
		{"text", "fixture-pack (experimental, corpus not_qualified)"},
		{"markdown", "fixture&#45;pack (experimental, corpus not&#95;qualified)"},
		{"html", "fixture-pack (experimental, corpus not_qualified)"},
	} {
		t.Run(row.format, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(report.Write(&output, row.format, modelResult(&value), report.Options{}), qt.IsNil)
			c.Assert(output.String(), qt.Contains, row.identity)
			c.Assert(output.String(), qt.Contains, "estimated for 1 of 2 sentence units")
			c.Assert(output.String(), qt.Not(qt.Contains), "no calibrated")
		})
	}
}

func TestHumanReportsKeepModelFreeWording(t *testing.T) {
	c := qt.New(t)
	result := modelResult(nil)
	result.Manifest.Probability = nil
	for _, format := range []string{"text", "markdown", "html"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, result, report.Options{}), qt.IsNil)
		c.Assert(output.String(), qt.Contains, "no calibrated")
		c.Assert(output.String(), qt.Not(qt.Contains), "fixture-pack")
	}
}
