package e2e_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
)

type baselineRecord struct {
	Stage       string               `json:"stage"`
	Diagnostics diagnosticRecord     `json:"diagnostics"`
	Comparison  *baseline.Comparison `json:"comparison"`
	States      []string             `json:"finding_states"`
	Assessments []unswell.Assessment `json:"assessments"`
	Accepted    int                  `json:"accepted_gate_reasons"`
}

func TestBaselineWorkflow(t *testing.T) {
	c := qt.New(t)
	binary, workspace := buildCLI(t), t.TempDir()
	policy, err := os.ReadFile("testdata/policy.yaml")
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "policy.yaml", policy)
	writeBaselineStage(t, workspace, "before")
	stdout, stderr, code := invoke(t, binary, workspace,
		[]string{"baseline", "create", "guide.md", "--config", "policy.yaml", "--output", "accepted.json"}, nil)
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s %s", stdout, stderr))
	c.Assert(stdout, qt.Equals, "Baseline create: 3 accepted entries across 1 documents.\n")
	var records []baselineRecord
	for _, stage := range []struct {
		name, mode string
		exit       int
	}{{"moved", "new", 0}, {"moved", "all", 1}, {"changed", "new", 1}} {
		records = append(records, checkBaselineStage(t, binary, workspace, stage.name, stage.mode, stage.exit))
	}
	stdout, stderr, code = invoke(t, binary, workspace,
		[]string{"baseline", "update", "guide.md", "--config", "policy.yaml", "--baseline", "accepted.json"}, nil)
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s %s", stdout, stderr))
	c.Assert(stdout, qt.Equals, "Baseline update: 6 accepted entries across 1 documents.\n")
	records = append(records, checkBaselineStage(t, binary, workspace, "changed", "new", 0))
	records = append(records, checkBaselineStage(t, binary, workspace, "clean", "new", 0))
	assertGoldenJSON(t, "baselinedata/workflow.golden.json", records)
	assertInputsUnchanged(t, workspace, map[string][]byte{"policy.yaml": policy})
}

func writeBaselineStage(t *testing.T, workspace, stage string) ([]byte, []expectation) {
	t.Helper()
	c := qt.New(t)
	data, err := fs.ReadFile(os.DirFS("baselinedata"), stage+".md.txt")
	c.Assert(err, qt.IsNil)
	text, wants, err := annotatedSource("guide.md", string(data))
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "guide.md", []byte(text))
	return []byte(text), wants
}

func checkBaselineStage(t *testing.T, binary, workspace, stage, mode string, exit int) baselineRecord {
	t.Helper()
	c := qt.New(t)
	source, wants := writeBaselineStage(t, workspace, stage)
	accepted, err := fs.ReadFile(os.DirFS(workspace), "accepted.json")
	c.Assert(err, qt.IsNil)
	stdout, stderr, code := invoke(t, binary, workspace, []string{"baseline", "check", "guide.md", "--config", "policy.yaml",
		"--baseline", "accepted.json", "--gate-mode", mode, "--report", "json:result.json", "--report", "sarif:result.sarif"}, nil)
	c.Assert(code, qt.Equals, exit, qt.Commentf("%s %s", stdout, stderr))
	assertInputsUnchanged(t, workspace, map[string][]byte{"guide.md": source, "accepted.json": accepted})
	var result unswell.RunResult
	decodeFile(t, filepath.Join(workspace, "result.json"), &result)
	c.Assert(matchWants(result.Findings, wants), qt.IsNil)
	verifyLocations(t, result, map[string][]byte{"guide.md": source})
	verifySARIF(t, workspace, result)
	record := baselineRecord{Stage: stage + "/" + mode, Diagnostics: diagnostics(result), Comparison: result.Baseline,
		Assessments: result.Assessments, Accepted: len(result.Gate.Accepted)}
	for _, finding := range result.Findings {
		record.States = append(record.States, finding.BaselineState)
	}
	return record
}
