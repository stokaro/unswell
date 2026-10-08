package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestSupportedActionUsefulnessRevision(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md":    "The client can be used for associating messages with vertices that may be helpful for tracing purposes.",
		"revision.md": "You can use the client to associate messages with vertices that may help with tracing.",
		"control.md":  "Message correlation may help with tracing.",
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
		c.Assert(result.Findings, qt.HasLen, 1)
		finding := result.Findings[0]
		c.Assert(finding.RuleID, qt.Equals, "filler.instruction-scaffolding")
		c.Assert(finding.RuleVersion, qt.Equals, "15")
		c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
		c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimSuffix(files["draft.md"], "."))
		c.Assert(finding.Message, qt.Contains, "support and helpfulness clauses")
		c.Assert(finding.Evidence.Suggestion, qt.Contains, "relative clause's antecedent")
		c.Assert(files["draft.md"][finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
	}
}
