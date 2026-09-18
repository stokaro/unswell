package e2e_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestAdjacentQuotationRepair(t *testing.T) {
	c := qt.New(t)
	workspace := t.TempDir()
	const policy = "version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  repetition.adjacent-word: {enabled: true, gate: forbid}\n"
	c.Assert(os.WriteFile(filepath.Join(workspace, ".unswell.yaml"), []byte(policy), 0o600), qt.IsNil)
	binary := buildCLI(t)
	for _, row := range []struct {
		name, text string
		code       int
	}{
		{"draft", "\ufeff# Café\r\n\r\nThe library was named “ImGui” when\r\nwhen the maintainer released it.", 1},
		{"revision", `The library was named "ImGui" when the maintainer released it.`, 0},
		{"quotation", `The manual says "First sentence. The client must must retry. Last sentence."`, 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(os.WriteFile(filepath.Join(workspace, "source.md"), []byte(row.text), 0o600), qt.IsNil)
			stdout, stderr, code := invoke(t, binary, workspace, []string{"check", "--include-source",
				"--report", "json:result.json", "source.md"}, nil)
			c.Assert(code, qt.Equals, row.code, qt.Commentf("%s\n%s", stdout, stderr))
			var result unswell.RunResult
			decodeFile(t, filepath.Join(workspace, "result.json"), &result)
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, row.code)
			if row.code == 1 {
				finding := result.Findings[0]
				c.Assert(finding.RuleVersion, qt.Equals, "2")
				c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(row.text, "when"))
				c.Assert(finding.Related[0].Span.Start, qt.Equals, strings.LastIndex(row.text, "when"))
			}
		})
	}
}
