package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestRestrictionReformulationRevision(t *testing.T) {
	c := qt.New(t)
	workspace := t.TempDir()
	files := map[string]string{
		"draft.md": "In particular, ripgrep's dependencies (direct and transitive) will always be limited to permissive licenses. " +
			"That is, ripgrep will never depend on code that is not permissively licensed.",
		"revision.md": "Ripgrep's dependencies (direct and transitive) will always be limited to permissive licenses.",
		"distinct.md": "The client accepts only signed requests. That is, the server never accepts requests that are not signed.",
	}
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(workspace, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "."}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(workspace, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		count := 0
		for _, finding := range result.Findings {
			if finding.RuleID != "repetition.repeated-claim" {
				continue
			}
			count++
			c.Assert(finding.RuleVersion, qt.Equals, "2")
			c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
			c.Assert(finding.Related, qt.HasLen, 1)
			span := finding.Primary.Span
			c.Assert(files["draft.md"][span.Start:span.End], qt.Equals, finding.Primary.Snippet)
		}
		c.Assert(count, qt.Equals, 1)
	}
}
