package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestInformationEvaluationRevision(t *testing.T) {
	c := qt.New(t)
	workspace := t.TempDir()
	files := map[string]string{
		"draft.md": "That distinction is the whole value of the verb: the probe says the key failed.\n\n" +
			"That is the whole guarantee, and it is worth stating precisely: both records commit together.",
		"revision.md":   "The probe says the key failed.\n\nBoth records commit together.",
		"quotation.md":  "The author says: that distinction is the whole value of the verb.",
		"definition.md": "A form that allows a user to upload a file could be written like this in HTML.",
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
		var snippets []string
		for _, finding := range result.Findings {
			if finding.RuleID != "filler.evaluative-closure" && finding.RuleID != "filler.instruction-scaffolding" {
				continue
			}
			c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
			c.Assert(finding.RuleID, qt.Equals, "filler.evaluative-closure")
			span := finding.Primary.Span
			c.Assert(files["draft.md"][span.Start:span.End], qt.Equals, finding.Primary.Snippet)
			snippets = append(snippets, finding.Primary.Snippet)
		}
		c.Assert(snippets, qt.DeepEquals, []string{
			"That distinction is the whole value of the verb",
			"That is the whole guarantee, and it is worth stating precisely",
		})
	}
}
