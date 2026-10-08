package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestDiscourseStanceReportsPreserveOperationalInformation(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"question.md": "Café 🙂.\r\n\r\nIt creates `row_2`, which is exactly the question, and the row is removed.\r\n",
		"worth.md":    "It publishes text as an enum — which is the case worth having, because the dialect supports it.",
		"plan.md":     "Applying that plan is what closes the loop.",
		"status.md":   "The non-zero status is the useful part.",
		"rollback.md": "That is the honest conversion of a migration that never had a rollback.",
		"control.md": "Applying that plan is what closes the feedback loop.\n\n" +
			"The non-zero status is the useful part because it stops CI.\n\n" +
			"The example preserves `That is the honest conversion of a migration.`",
	}
	var paths []string
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(work, name), []byte(text), 0o600), qt.IsNil)
		paths = append(paths, name)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		args := append([]string{"check", "--profile", profile, "--include-source", "--report", "json:result.json"}, paths...)
		stdout, stderr, code := invoke(t, binary, work, args, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(work, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		c.Assert(result.Findings, qt.HasLen, 5)
		for _, finding := range result.Findings {
			c.Assert(finding.RuleID, qt.Equals, "filler.evaluative-closure")
			c.Assert(finding.RuleVersion, qt.Equals, "12")
			c.Assert(finding.Primary.Path, qt.Not(qt.Equals), "control.md")
			c.Assert(files[finding.Primary.Path][finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Primary.Snippet, qt.Not(qt.Contains), "because the dialect")
			c.Assert(finding.Primary.Snippet, qt.Not(qt.Contains), "row_2")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "absent operations")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "Verify the basis")
			c.Assert(strings.Contains(finding.Primary.Snippet, "rollback"), qt.Equals, finding.Primary.Path == "rollback.md")
		}
	}
}
