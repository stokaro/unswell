package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestInstalledPrerequisiteAnnouncement(t *testing.T) {
	for _, text := range []string{
		"To work with the React UI code, you will need to have the following tools installed:",
		"You will need to have the following tools installed.",
		"To build the client, you will need to have the following development packages installed:",
		"In order to run the tests, you will need to have the following dependencies installed:",
		"To build the client, you will need to have the following `Node.js` tools installed:",
	} {
		t.Run(text, func(t *testing.T) {
			result := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimRight(text, ".:"))
			c.Assert(finding.Message, qt.Contains, "Future and possession auxiliaries")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "required installed state")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "recommended items optional")
		})
	}
}

func TestInstalledPrerequisiteSourceAndPolicy(t *testing.T) {
	id := "filler.instruction-scaffolding"
	text := "To build the café client, you will need to have the following `Node.js` tools installed:"
	markdown := "\ufeff" + text + "\r\n\r\n- Node.js 24 is required.\r\n- Recommended: an editor.\r\n"
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte(markdown)},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			finding := instructionSourceFinding(t, source)
			c := qt.New(t)
			c.Assert(finding.Primary.Snippet, qt.Contains, "café client")
			c.Assert(finding.Primary.Snippet, qt.Contains, "`Node.js` tools installed")
		})
	}
	plain := strings.ReplaceAll(text, "`Node.js`", "Node.js")
	qt.New(t).Assert(singleRuleResult(t, id, plain, "", windowTerm(id, strings.TrimSuffix(plain, ":"))).Findings, qt.HasLen, 0)
	qt.New(t).Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	long := "To build " + strings.Repeat("large ", 49) + "clients, you will need to have the following tools installed."
	qt.New(t).Assert(singleRuleResult(t, id, long, "", "").Findings, qt.HasLen, 0)
}

func instructionSourceFinding(t *testing.T, source document.Source) unswell.Finding {
	t.Helper()
	result, err := singleRuleEngine(t, "filler.instruction-scaffolding", "", "").Analyze(t.Context(), source)
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(string(source.Bytes[finding.Primary.Span.Start:finding.Primary.Span.End]), qt.Equals, finding.Primary.Snippet)
	return finding
}

func TestInstalledPrerequisiteControls(t *testing.T) {
	for _, text := range []string{
		"Install Node.js before building the client.",
		"Node.js must be installed to build the client.",
		"To work with the UI, you need the following tools installed:",
		"You need to have the following tools installed.",
		"You will need the following tools installed.",
		"You will need to have the following tools.",
		"You will need to have the following tools available.",
		"The worker will need to have the following tools installed.",
		"You might need to have the following tools installed.",
		"You will not need to have the following tools installed.",
		"You will need to have only the following tools installed.",
		"You will need to have the following tools installed unless the image includes them.",
		"You will need to have the following tools installed by Friday.",
		"By Friday, you will need to have the following tools installed.",
		"After upgrading, you will need to have the following tools installed.",
		"To meet tomorrow's deadline, you will need to have the following tools installed.",
		"You will need to have the following tools installed and configured.",
		"You will need to have the following tools that the project recommends installed.",
		"You will need to have the following permissions installed.",
		"You will need to have the following optional tools installed.",
		"The manual says you will need to have the following tools installed.",
		"The manual says: \"You will need to have the following tools installed.\"",
		"You `will need to have` the following tools installed.",
		"You will need to have the following tools `installed`.",
		"Will you need to have the following tools installed?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
