package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestEmDashFrequencyReportsADocumentOnce(t *testing.T) {
	engine, err := unswell.New(unswell.Options{})
	qt.New(t).Assert(err, qt.IsNil)
	plain := strings.Repeat("The server sends the response to the client. ", 60)
	cases := []struct {
		name, file, text string
		dashes           int
	}{
		{"three in a short page", "guide.md",
			"# Client\n\nIt connects — then waits. It keeps the socket — for a while.\n\n- It closes — on a timeout.\n", 3},
		{"two are allowed", "guide.md", "# Client\n\nIt connects — then waits. It keeps the socket — for a while.\n", 0},
		{"three in a long page", "guide.md", "# Client\n\nIt connects — then waits — and closes — on a timeout.\n\n" +
			strings.Repeat(plain+"\n\n", 3), 0},
		{"heading and list item count", "guide.md", "# Client — overview\n\n- It connects — then waits.\n- It closes — on a timeout.\n", 3},
		{"code and tables do not count", "guide.md", "# Client\n\nIt connects — then waits.\n\n```text\na — b — c — d\n```\n\n" +
			"| Step — one | State — two |\n| --- | --- |\n| Open — close | Done — now |\n", 0},
		{"hyphens and en dashes", "guide.md", "# Client\n\nIt connects - then waits – and closes -- on a timeout - or an error.\n", 0},
		{"comments in code", "client.go", "package client\n\n// Start opens — then waits — and closes — on a timeout.\nfunc Start() {}\n", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			format := document.Markdown
			if tc.file == "client.go" {
				format = document.Go
			}
			result, err := engine.Analyze(t.Context(), document.Source{Name: tc.file, Format: format, Bytes: []byte(tc.text)})
			c.Assert(err, qt.IsNil)
			findings, dashes := 0, 0
			for _, finding := range result.Findings {
				if finding.RuleID != "format.em-dash-frequency" {
					continue
				}
				findings++
				c.Assert(finding.Severity, qt.Equals, "warning")
				for _, location := range append([]unswell.Location{finding.Primary}, finding.Related...) {
					c.Assert(tc.text[location.Span.Start:location.Span.End], qt.Equals, "—")
					dashes++
				}
			}
			c.Assert(dashes, qt.Equals, tc.dashes)
			c.Assert(findings <= 1, qt.IsTrue)
			c.Assert(result.Gate.Passed, qt.IsTrue)
		})
	}
}
