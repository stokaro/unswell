package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestPromotionalDegreeReportsPreserveSourceAndVerification(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	files := map[string]string{
		"draft.md": "## Blazingly fast\n\nThe engine is heavily optimized for request batching.\n\n" +
			"That might be the greatest benefit.",
		"criteria.md": "The client is blazingly fast: the benchmark reports its throughput.\n\n" +
			"The engine is heavily optimized by coalescing queued requests.\n\n" +
			"That might be the greatest benefit in the measured workload.",
		"control.md": "The service is highly available.\n\nThe module is strongly typed.",
	}
	for name, text := range files {
		c.Assert(os.WriteFile(filepath.Join(work, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		stdout, stderr, code := invoke(t, binary, work, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "draft.md", "criteria.md", "control.md"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(work, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		c.Assert(result.Findings, qt.HasLen, 3)
		for _, finding := range result.Findings {
			c.Assert(finding.RuleID, qt.Equals, "filler.unscoped-assurance")
			c.Assert(finding.RuleVersion, qt.Equals, "10")
			c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
			c.Assert(files["draft.md"][finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Evidence.Message, qt.Equals, finding.Message)
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "Verify")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "uncertainty")
		}
	}
}
