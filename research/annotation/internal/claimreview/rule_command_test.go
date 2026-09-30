package claimreview_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
	"github.com/stokaro/unswell/rule"
)

func savedRules(t *testing.T) (unswell.RunResult, string) {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{IncludeSource: true})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "example.txt", Format: document.Plain,
		Bytes: []byte("It is important to note that the café client must not retry when retries are disabled.")})
	c.Assert(err, qt.IsNil)
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	return result, saved.String()
}

func TestRuleCommandPreservesAllObservationsAndRedactsSourceSnippets(t *testing.T) {
	c := qt.New(t)
	result, saved := savedRules(t)
	var output bytes.Buffer
	c.Assert(claimreview.RunRules(t.Context(), strings.NewReader(saved), &output), qt.IsNil)
	var projected struct {
		Version            string `json:"version"`
		EditorialQualified bool   `json:"editorial_qualified"`
		Observations       []struct {
			ID               string             `json:"id"`
			SourceHash       string             `json:"source_sha256"`
			Diagnostic       string             `json:"diagnostic"`
			Primary          unswell.Location   `json:"primary"`
			Related          []unswell.Location `json:"related"`
			Evidence         rule.Evidence      `json:"evidence"`
			EditorialVerdict string             `json:"editorial_verdict"`
		} `json:"observations"`
	}
	c.Assert(json.Unmarshal(output.Bytes(), &projected), qt.IsNil)
	c.Assert(projected.Version, qt.Equals, claimreview.RuleEvidenceVersion)
	c.Assert(projected.EditorialQualified, qt.IsFalse)
	c.Assert(projected.Observations, qt.HasLen, len(result.Findings))
	for i, observation := range projected.Observations {
		original := result.Findings[i]
		c.Assert(observation.ID, qt.Equals, original.ID)
		c.Assert(observation.SourceHash, qt.Equals, result.Documents[0].SourceHash)
		c.Assert(observation.Diagnostic, qt.Equals, original.Message)
		c.Assert(observation.Evidence, qt.DeepEquals, original.Evidence)
		c.Assert(observation.Primary.Span, qt.Equals, original.Primary.Span)
		c.Assert(observation.Primary.Segments, qt.DeepEquals, original.Primary.Segments)
		c.Assert(observation.Primary.Snippet, qt.Equals, "")
		c.Assert(observation.EditorialVerdict, qt.Equals, "unreviewed")
		for _, location := range observation.Related {
			c.Assert(location.Snippet, qt.Equals, "")
		}
	}
	c.Assert(output.String(), qt.Not(qt.Contains), result.Documents[0].Source)
}

func TestRuleCommandRejectsIncompleteMalformedAndRepeatedReports(t *testing.T) {
	_, saved := savedRules(t)
	tests := []struct {
		name  string
		input string
	}{
		{"incomplete", strings.Replace(saved, `"status": "complete"`, `"status": "incomplete"`, 1)},
		{"malformed", `{"schema_version": "unknown"}`},
		{"repeated", saved + saved},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(claimreview.RunRules(t.Context(), strings.NewReader(test.input), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

func TestRuleCommandPreservesCancellationAndWriterFailure(t *testing.T) {
	c := qt.New(t)
	_, saved := savedRules(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(claimreview.RunRules(ctx, strings.NewReader(saved), &bytes.Buffer{}), qt.Equals, context.Canceled)
	c.Assert(claimreview.RunRules(t.Context(), strings.NewReader(saved), failingWriter{}), qt.IsNotNil)
}
