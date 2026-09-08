package e2e_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

type trustedRecord struct {
	Stage       string                    `json:"stage"`
	Diagnostics diagnosticRecord          `json:"diagnostics"`
	Policy      *unswell.PolicyComparison `json:"policy_comparison"`
	Selection   *unswell.ChangeSelection  `json:"selection"`
}

func TestTrustedWorkflow(t *testing.T) {
	c := qt.New(t)
	binary, workspace := buildCLI(t), t.TempDir()
	policy, err := os.ReadFile("testdata/policy.yaml")
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "policy.yaml", policy)
	writeCommittedStage(t, workspace, "before")
	fixtureGit(t, workspace, "init", "-b", "main")
	fixtureGit(t, workspace, "add", "guide.md", "policy.yaml")
	fixtureGit(t, workspace, "commit", "-m", "Establish required rules")
	base := fixtureGit(t, workspace, "rev-parse", "HEAD")
	writeFixture(c, workspace, "policy.yaml", []byte("version: 1\nextends: [builtin:custom]\n"))
	fixtureGit(t, workspace, "commit", "-am", "Disable candidate rules")
	var records []trustedRecord
	for _, stage := range []struct {
		name string
		exit int
	}{{"policy", 1}, {"permission", 1}, {"edited", 0}} {
		source, wants := trustedStage(t, workspace, stage.name)
		fixtureGit(t, workspace, "commit", "-am", stage.name, "--allow-empty")
		stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "guide.md", "--config", "policy.yaml",
			"--changed-from", base, "--policy-from-base", "--report", "json:result.json", "--report", "sarif:result.sarif"}, nil)
		c.Assert(code, qt.Equals, stage.exit, qt.Commentf("%s %s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(workspace, "result.json"), &result)
		c.Assert(result.Manifest.Git.BaseCommit, qt.Equals, base)
		c.Assert(result.Manifest.Git.Clean, qt.IsTrue)
		c.Assert(result.PolicyComparison.Complete, qt.IsTrue)
		c.Assert(result.PolicyComparison.FullScan, qt.IsTrue)
		c.Assert(matchWants(result.Findings, wants), qt.IsNil)
		verifyLocations(t, result, map[string][]byte{"guide.md": source})
		verifySARIF(t, workspace, result)
		assertInputsUnchanged(t, workspace, map[string][]byte{"guide.md": source})
		records = append(records, trustedRecord{Stage: stage.name, Diagnostics: diagnostics(result),
			Policy: result.PolicyComparison, Selection: result.Changes})
	}
	assertGoldenJSON(t, "trusteddata/workflow.golden.json", records)
}

func trustedStage(t *testing.T, workspace, stage string) ([]byte, []expectation) {
	t.Helper()
	if stage == "policy" {
		return writeCommittedStage(t, workspace, "before")
	}
	c := qt.New(t)
	data, err := fs.ReadFile(os.DirFS("trusteddata"), stage+".md.txt")
	c.Assert(err, qt.IsNil)
	source, wants, err := annotatedSource("guide.md", string(data))
	c.Assert(err, qt.IsNil)
	writeFixture(c, workspace, "guide.md", []byte(source))
	return []byte(source), wants
}
