package report_test

import (
	"bytes"
	"html"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

func TestSuppressionAuditSurvivesSavedReports(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(`version: 1
extends: [builtin:custom]
rules:
  policy.banned-phrases:
    enabled: true
    parameters: {phrases: [robust]}
    score: {weight: 20, cap: 20}
`)})
	c.Assert(err, qt.IsNil)
	reason := "Required <script>contract</script> wording."
	text := "<!-- unswell-disable-next-block policy.banned-phrases -- " + reason + " -->\n\nThe robust client starts."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Suppressions[0].Directive.Snippet, qt.Equals, "")
	var saved bytes.Buffer
	c.Assert(report.Write(&saved, "json", result, report.Options{}), qt.IsNil)
	restored, err := report.Read(&saved)
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, result)
	for _, format := range []string{"text", "json", "sarif", "html", "markdown"} {
		t.Run(format, func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(report.Write(&output, format, restored, report.Options{}), qt.IsNil)
			c.Assert(html.UnescapeString(output.String()), qt.Contains, "contract")
			c.Assert(html.UnescapeString(output.String()), qt.Contains, "policy.banned-phrases")
			if format == "html" {
				c.Assert(output.String(), qt.Not(qt.Contains), "<script>contract</script>")
			}
		})
	}
}
