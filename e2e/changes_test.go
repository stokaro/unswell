package e2e_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

type changeRecord struct {
	Stage       string                   `json:"stage"`
	Diagnostics diagnosticRecord         `json:"diagnostics"`
	Selection   *unswell.ChangeSelection `json:"selection"`
	States      []string                 `json:"finding_states"`
	Assessments []unswell.Assessment     `json:"assessments"`
	Unchanged   int                      `json:"unchanged_gate_reasons"`
}

func TestCommittedWorkflow(t *testing.T) {
	c := qt.New(t)
	binary, workspace := buildCLI(t), t.TempDir()
	policy, err := os.ReadFile("testdata/policy.yaml")
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "policy.yaml", policy)
	writeCommittedStage(t, workspace, "before")
	fixtureGit(t, workspace, "init", "-b", "main")
	fixtureGit(t, workspace, "add", "guide.md", "policy.yaml")
	fixtureGit(t, workspace, "commit", "-m", "Initial documentation")
	base := fixtureGit(t, workspace, "rev-parse", "HEAD")
	var records []changeRecord
	for _, stage := range []struct {
		name string
		exit int
	}{{"moved", 0}, {"joined", 1}, {"edited", 1}} {
		records = append(records, checkCommittedStage(t, binary, workspace, base, stage.name, stage.exit))
	}
	assertGoldenJSON(t, "changesdata/workflow.golden.json", records)
	writeCommittedStage(t, workspace, "moved")
	stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "guide.md", "--config", "policy.yaml",
		"--changed-from", base, "--no-gate"}, nil)
	c.Assert(code, qt.Equals, 2, qt.Commentf("%s %s", stdout, stderr))
	c.Assert(stderr, qt.Contains, "bytes do not match HEAD")
	assertInputsUnchanged(t, workspace, map[string][]byte{"policy.yaml": policy})
}

func fixtureGit(t *testing.T, workspace string, args ...string) string {
	t.Helper()
	argv := append([]string{"-C", workspace, "-c", "commit.gpgsign=false", "-c", "core.autocrlf=false",
		"-c", "user.name=Unswell E2E", "-c", "user.email=e2e@example.invalid"}, args...)
	// #nosec G204 -- Fixed fixture operations use argv in a temporary repository without a remote.
	command := exec.CommandContext(t.Context(), "git", argv...)
	output, err := command.CombinedOutput()
	qt.New(t).Assert(err, qt.IsNil, qt.Commentf("fixture Git: %s", output))
	return strings.TrimSpace(string(output))
}

func writeCommittedStage(t *testing.T, workspace, stage string) ([]byte, []expectation) {
	t.Helper()
	c := qt.New(t)
	data, err := fs.ReadFile(os.DirFS("changesdata"), stage+".md.txt")
	c.Assert(err, qt.IsNil)
	text, wants, err := annotatedSource("guide.md", string(data))
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "guide.md", []byte(text))
	return []byte(text), wants
}

func checkCommittedStage(t *testing.T, binary, workspace, base, stage string, exit int) changeRecord {
	t.Helper()
	c := qt.New(t)
	source, wants := writeCommittedStage(t, workspace, stage)
	fixtureGit(t, workspace, "commit", "-am", stage)
	head := fixtureGit(t, workspace, "rev-parse", "HEAD")
	stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "guide.md", "--config", "policy.yaml",
		"--changed-from", base, "--report", "json:result.json", "--report", "sarif:result.sarif"}, nil)
	c.Assert(code, qt.Equals, exit, qt.Commentf("%s %s", stdout, stderr))
	var result unswell.RunResult
	decodeFile(t, filepath.Join(workspace, "result.json"), &result)
	c.Assert(result.Manifest.Git, qt.DeepEquals, &unswell.GitSelection{
		RequestedRef: base, BaseCommit: base, HeadCommit: head, Clean: true,
	})
	c.Assert(matchWants(result.Findings, wants), qt.IsNil)
	verifyLocations(t, result, map[string][]byte{"guide.md": source})
	verifySARIF(t, workspace, result)
	assertInputsUnchanged(t, workspace, map[string][]byte{"guide.md": source})
	record := changeRecord{Stage: stage, Diagnostics: diagnostics(result), Selection: result.Changes,
		Assessments: result.Assessments, Unchanged: len(result.Gate.Unchanged)}
	for _, finding := range result.Findings {
		record.States = append(record.States, finding.ChangeState)
	}
	return record
}
