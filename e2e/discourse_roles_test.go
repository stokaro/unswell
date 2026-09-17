package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestDiscourseRolesRevision(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md": "The source column is the part worth reading.\n\n" +
			"I want to stress that the command resets the cache.\n\nThe interface is very easy to use.",
		"revision.md": "The source column names the registry.\n\n" +
			"The command resets the cache.\n\nThe interface accepts a JSON request.",
		"control.md": "The report contains the verdict, the row counts, and the findings.\n\n" +
			"The query is faster because the index covers the key.\n\nThe response includes a token that can be used to fetch records.",
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
			if f.RuleID != "filler.evaluative-closure" && f.RuleID != "filler.unscoped-assurance" && f.RuleID != "filler.instruction-scaffolding" {
				continue
			}
			counts[f.RuleID]++
			c.Assert(f.Primary.Path, qt.Equals, "draft.md")
			c.Assert(files["draft.md"][f.Primary.Span.Start:f.Primary.Span.End], qt.Equals, f.Primary.Snippet)
		}
		c.Assert(counts, qt.DeepEquals, map[string]int{"filler.evaluative-closure": 2, "filler.unscoped-assurance": 1})
	}
}
