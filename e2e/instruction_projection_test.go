package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestInstructionProjectionRevision(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md": "The adapter has the ability to query a local index.\n\n" +
			"The function may be used to configure a label.\n\n" +
			"The map is intended to allow extra values to be passed to the template.\n\n" +
			"It is possible to configure the queue.\n\nThis is done by using the console.",
		"revision.md": "The adapter can query a local index.\n\n" +
			"You may use the function to configure a label.\n\n" +
			"The map supports passing extra values to the template.\n\nYou can configure the queue through the console.",
		"control.md": "The console allows administrators to configure the service.\n\n" +
			"The adapter has the ability to query a local index only after approval.",
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
		c.Assert(result.Findings, qt.HasLen, 4)
		related := 0
		for _, f := range result.Findings {
			c.Assert(f.RuleID, qt.Equals, "filler.instruction-scaffolding")
			c.Assert(f.RuleVersion, qt.Equals, "9")
			c.Assert(f.Primary.Path, qt.Equals, "draft.md")
			c.Assert(files["draft.md"][f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
			for _, loc := range f.Related {
				related++
				c.Assert(files["draft.md"][loc.Span.Start:loc.Span.End], qt.Equals, loc.Snippet)
			}
		}
		c.Assert(related, qt.Equals, 1)
	}
}
