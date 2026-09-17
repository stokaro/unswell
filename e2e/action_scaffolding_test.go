package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestActionScaffoldingRevision(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md": "The clause allows specifying labels to be attached to the alert.\n\n" +
			"User management can be done by using the connection API.",
		"revision.md": "The clause adds labels to the alert.\n\nYou can manage users through the connection API.",
		"control.md":  "The interface allows configuration files to be uploaded.\n\nThe console allows administrators to configure the service.",
	}
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(work, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		stdout, stderr, code := invoke(t, binary, work, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "draft.md", "revision.md", "control.md"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(work, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		c.Assert(result.Findings, qt.HasLen, 2)
		for _, f := range result.Findings {
			c.Assert(f.RuleID, qt.Equals, "filler.instruction-scaffolding")
			c.Assert(f.RuleVersion, qt.Equals, "9")
			c.Assert(f.Primary.Path, qt.Equals, "draft.md")
			c.Assert(files["draft.md"][f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
		}
	}
}
