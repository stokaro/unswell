package report_test

import (
	"bytes"
	"encoding/json"
	"html"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestSavedTrustedPolicyPreservesGateAndPermissionAudit(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	before := []document.Source{{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Certainly! The client retries.")}}
	after := []document.Source{{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("<!-- unswell-disable-next-block scaffold.chat-preamble -- External wording. -->\n\nCertainly! The client retries.")}}
	result, err := engine.AnalyzeChangedWithOptions(t.Context(), before, after, unswell.ChangeOptions{
		TrustedPolicy: true, PolicyChanges: []unswell.PolicyChange{{Path: "policy.yaml", Kind: "config", BeforeHash: strings.Repeat("a", 64)}},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Suppressions[0].TrustState, qt.Equals, "untrusted")
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	loaded, err := report.Read(bytes.NewReader(saved.Bytes()))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	for _, format := range []string{"text", "markdown", "html", "sarif"} {
		var output bytes.Buffer
		c.Assert(report.Write(&output, format, loaded, report.Options{}), qt.IsNil)
		c.Assert(output.String(), qt.Contains, "untrusted")
		c.Assert(html.UnescapeString(output.String()), qt.Contains, "policy.yaml")
		c.Assert(output.String(), qt.Contains, strings.Repeat("a", 64))
		c.Assert(output.String(), qt.Not(qt.Contains), "Certainly! The client retries.")
		if format == "sarif" {
			assertNoAcceptedSuppression(t, output.Bytes())
		}
	}
	result.PolicyComparison.Complete = false
	saved.Reset()
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	_, err = report.Read(bytes.NewReader(saved.Bytes()))
	c.Assert(err, qt.IsNotNil)
}

func assertNoAcceptedSuppression(t *testing.T, data []byte) {
	t.Helper()
	c := qt.New(t)
	var sarif struct {
		Runs []struct {
			Results []struct {
				Suppressions []any `json:"suppressions"`
			} `json:"results"`
		} `json:"runs"`
	}
	c.Assert(json.Unmarshal(data, &sarif), qt.IsNil)
	c.Assert(sarif.Runs, qt.HasLen, 1)
	for _, finding := range sarif.Runs[0].Results {
		c.Assert(finding.Suppressions, qt.HasLen, 0)
	}
}
