package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestClauseBoundaryRepairs(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md": "The library includes a function `Decode` that can be used to parse records.\n\n" +
			"To configure the client, the process is straightforward.\n\n" +
			"The library provides the ability to parse records.\n\nSetting up a configuration file is simple.",
		"revision.md": "The `Decode` function parses records.\n\nSet the client address in its configuration file.",
		"control.md": "Raise `--connect-timeout` for databases that are slow to accept connections.\n\n" +
			"The response includes a checksum `function` that can be used to cache results.",
	}
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
		counts := make(map[string]int)
		for _, f := range result.Findings {
			if f.RuleID != "filler.instruction-scaffolding" && f.RuleID != "filler.unscoped-assurance" {
				continue
			}
			counts[f.RuleID]++
			c.Assert(f.Primary.Path, qt.Equals, "draft.md")
			c.Assert(files["draft.md"][f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
		}
		c.Assert(counts, qt.DeepEquals, map[string]int{"filler.instruction-scaffolding": 2, "filler.unscoped-assurance": 2})
	}
}
