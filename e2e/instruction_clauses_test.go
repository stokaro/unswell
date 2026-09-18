package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestInstructionClausesRevision(t *testing.T) {
	files := map[string]string{
		"draft.md": "Now that we have a task in the created state we need to make sure that we wait on the task to exit.\n\n" +
			"The library is designed to allow other applications to use it.",
		"revision.md": "Once the task is created, wait on it to exit.\n\nThe library is intended for use by other applications.",
		"control.md":  "The proxy allows you to run a local server to handle API requests, while developing the UI separately.",
	}
	checkInstructionRevision(t, files, 2)
}

func checkInstructionRevision(t *testing.T, files map[string]string, want int) {
	t.Helper()
	c := qt.New(t)
	work := t.TempDir()
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(work, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		out, stderr, code := invoke(t, binary, work, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "draft.md", "revision.md", "control.md"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", out, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(work, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		count := 0
		for _, f := range result.Findings {
			if f.RuleID == "filler.instruction-scaffolding" {
				count++
				c.Assert(f.RuleVersion, qt.Equals, "10")
				c.Assert(f.Primary.Path, qt.Equals, "draft.md")
				c.Assert(files["draft.md"][f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
			}
		}
		c.Assert(count, qt.Equals, want)
	}
}
