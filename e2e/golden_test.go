package e2e_test

import (
	"encoding/json"
	"flag"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

var updateGoldens = flag.Bool("update", false, "Update reviewed e2e golden reports after checking want annotations")

type findingRecord struct {
	RuleID   string             `json:"rule_id"`
	Severity string             `json:"severity"`
	Gate     string             `json:"gate"`
	Scope    string             `json:"scope"`
	Message  string             `json:"message"`
	Primary  unswell.Location   `json:"primary"`
	Related  []unswell.Location `json:"related"`
}

type diagnosticRecord struct {
	Status   string             `json:"status"`
	Complete bool               `json:"complete"`
	Passed   bool               `json:"passed"`
	NoGate   bool               `json:"no_gate"`
	Findings []findingRecord    `json:"findings"`
	Errors   []unswell.RunError `json:"errors"`
}

func diagnostics(result unswell.RunResult) diagnosticRecord {
	record := diagnosticRecord{Status: result.Status, Complete: result.Manifest.Complete,
		Passed: result.Gate.Passed, NoGate: result.Manifest.NoGate, Findings: []findingRecord{}, Errors: result.Errors}
	for _, finding := range result.Findings {
		record.Findings = append(record.Findings, findingRecord{RuleID: finding.RuleID, Severity: finding.Severity,
			Gate: finding.Gate, Scope: finding.Scope, Message: finding.Message,
			Primary: finding.Primary, Related: finding.Related})
	}
	return record
}

func assertGoldenJSON(t *testing.T, path string, value any) {
	t.Helper()
	c := qt.New(t)
	data, err := json.MarshalIndent(value, "", "  ")
	c.Assert(err, qt.IsNil)
	assertGolden(t, path, append(data, '\n'))
}

func assertGolden(t *testing.T, path string, actual []byte) {
	t.Helper()
	c := qt.New(t)
	if *updateGoldens {
		c.Assert(os.WriteFile(path, actual, 0o600), qt.IsNil)
	}
	expected, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	c.Assert(err, qt.IsNil, qt.Commentf("missing golden %s; review the case before using -update", path))
	c.Assert(string(actual), qt.Equals, string(expected), qt.Commentf("golden: %s", path))
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
	EndColumn   int `json:"endColumn"`
	ByteOffset  int `json:"byteOffset"`
	ByteLength  int `json:"byteLength"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region sarifRegion `json:"region"`
	} `json:"physicalLocation"`
}

type sarifResult struct {
	RuleID  string `json:"ruleId"`
	Level   string `json:"level"`
	Message struct {
		Text string `json:"text"`
	} `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

func verifySARIF(t *testing.T, workspace string, result unswell.RunResult) {
	t.Helper()
	c := qt.New(t)
	data, err := fs.ReadFile(os.DirFS(workspace), "result.sarif")
	c.Assert(err, qt.IsNil)
	var report struct {
		Version string `json:"version"`
		Runs    []struct {
			ColumnKind  string        `json:"columnKind"`
			Results     []sarifResult `json:"results"`
			Invocations []struct {
				ExecutionSuccessful bool `json:"executionSuccessful"`
			} `json:"invocations"`
		} `json:"runs"`
	}
	c.Assert(json.Unmarshal(data, &report), qt.IsNil)
	c.Assert(report.Version, qt.Equals, "2.1.0")
	c.Assert(report.Runs, qt.HasLen, 1)
	run := report.Runs[0]
	c.Assert(run.ColumnKind, qt.Equals, "unicodeCodePoints")
	c.Assert(run.Invocations, qt.HasLen, 1)
	c.Assert(run.Invocations[0].ExecutionSuccessful, qt.Equals, result.Manifest.Complete)
	c.Assert(run.Results, qt.HasLen, len(result.Findings))
	for i, actual := range run.Results {
		finding := result.Findings[i]
		c.Assert(actual.RuleID, qt.Equals, finding.RuleID)
		c.Assert(actual.Level, qt.Equals, finding.Severity)
		c.Assert(actual.Message.Text, qt.Equals, finding.Message)
		c.Assert(actual.Locations, qt.HasLen, 1)
		location := actual.Locations[0].PhysicalLocation
		uri, err := url.Parse(location.ArtifactLocation.URI)
		c.Assert(err, qt.IsNil)
		c.Assert(uri.Path, qt.Equals, finding.Primary.Path)
		c.Assert(location.Region, qt.Equals, region(finding.Primary))
	}
}

func region(location unswell.Location) sarifRegion {
	start, end := location.Start, location.End
	return sarifRegion{StartLine: start.Line, StartColumn: start.Column, EndLine: end.Line, EndColumn: end.Column,
		ByteOffset: location.Span.Start, ByteLength: location.Span.End - location.Span.Start}
}
