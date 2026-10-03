package e2e_test

import (
	"html"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestVerificationExplanationReports(t *testing.T) {
	c := qt.New(t)
	work := t.TempDir()
	root, err := os.OpenRoot(work)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Check(root.Close(), qt.IsNil) })
	sources := map[string]string{
		"draft.md":    "Reliability is ensured by [**rigorous testing**](https://example.test/tests).\n\nNothing is taken on trust.",
		"revision.md": "Tests cover the parser's input validation.\n\nThe client compares the server's reported version before continuing.",
		"control.md":  "These tests verify that the parser rejects missing fields.\n\nCorrectness is guaranteed under the stated assumptions.",
	}
	for name, text := range sources {
		c.Assert(os.WriteFile(filepath.Join(work, name), []byte(text), 0o600), qt.IsNil)
	}
	binary := buildCLI(t)
	for _, profile := range []string{"technical", "strict"} {
		stdout, stderr, code := invoke(t, binary, work, []string{"check", "--profile", profile,
			"--include-source", "--report", "json:result.json", "--report", "text:result.txt",
			"--report", "sarif:result.sarif", "--report", "html:result.html", "--report", "markdown:result.md",
			"draft.md", "revision.md", "control.md"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		var result unswell.RunResult
		decodeFile(t, filepath.Join(work, "result.json"), &result)
		c.Assert(result.Manifest.Complete, qt.IsTrue)
		c.Assert(result.Findings, qt.HasLen, 2)
		for _, finding := range result.Findings {
			c.Assert(finding.RuleID, qt.Equals, "filler.unscoped-assurance")
			c.Assert(finding.RuleVersion, qt.Equals, "9")
			c.Assert(finding.Primary.Path, qt.Equals, "draft.md")
			c.Assert(sources["draft.md"][finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Evidence.Message, qt.Equals, finding.Message)
			for _, name := range []string{"result.txt", "result.sarif", "result.html", "result.md"} {
				data, err := root.ReadFile(name)
				c.Assert(err, qt.IsNil)
				c.Assert(html.UnescapeString(string(data)), qt.Contains, finding.Message, qt.Commentf("%s", name))
			}
		}
		stdout, stderr, code = invoke(t, binary, work,
			[]string{"report", "result.json", "--format", "html", "--output", "saved.html"}, nil)
		c.Assert(code, qt.Equals, 0, qt.Commentf("%s\n%s", stdout, stderr))
		data, err := root.ReadFile("saved.html")
		c.Assert(err, qt.IsNil)
		for _, finding := range result.Findings {
			c.Assert(string(data), qt.Contains, finding.Message)
		}
	}
}
