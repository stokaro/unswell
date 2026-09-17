package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestRhetoricalRelationsRevision(t *testing.T) {
	c := qt.New(t)
	workspace := t.TempDir()
	files := map[string]string{
		"draft.md": "Importing it is what this verb is for.\n\n" +
			"This is what readers understand.\n\nSeveral rows carry a decision worth stating.",
		"revision.md": "Import replays the selected migration directory.\n\n" +
			"The diagnostic names the conflicting table.\n\nThe setting selects the storage backend.",
	}
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(workspace, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "draft.md", "revision.md"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(workspace, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		found := make(map[string][]string)
		for _, finding := range result.Findings {
			c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
			span := finding.Primary.Span
			c.Assert(files["draft.md"][span.Start:span.End], qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Evidence.Suggestion, qt.Not(qt.Equals), "")
			found[finding.RuleID] = append(found[finding.RuleID], finding.Primary.Snippet)
		}
		c.Assert(found, qt.DeepEquals, map[string][]string{
			"filler.evaluative-closure": {"Importing it is what this verb is for", "a decision worth stating"},
			"filler.unscoped-assurance": {"This is what readers understand"},
		})
	}
}
