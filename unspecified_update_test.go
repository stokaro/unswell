package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestUnspecifiedUpdateIdentifiesStandaloneVagueChange(t *testing.T) {
	const id = "filler.unspecified-update"
	for _, text := range []string{
		"Some syntaxes and themes have been updated to the latest version.",
		"Several packages were upgraded to the latest versions.",
		"Various themes are updated to the latest version.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, id, text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(result.Findings[0].Evidence.Suggestion, qt.Contains, "avoid inventing a version")
		})
	}
}

func TestUnspecifiedUpdatePreservesNamedAndExplainedChanges(t *testing.T) {
	const id = "filler.unspecified-update"
	for _, text := range []string{
		"The Rust syntax was updated to version 2.4.",
		"Some themes were updated to version 2.4.",
		"Some themes were updated to the latest version: Solarized 2.0.",
		"Some themes have been updated to the latest version. The changes are listed below.",
		"Some themes have been updated to the latest version; see the release notes.",
		"Some themes have been updated to the latest version when their packages support it.",
		"Some themes have not been updated to the latest version.",
		"The author says some themes have been updated to the latest version.",
		"Some themes have been updated to the latest version of bat 0.25.",
		"Some themes support the latest version.",
		"All syntaxes and themes have been updated to the latest version.",
		"The quoted text says \"Some themes have been updated to the latest version.\"",
		"Some `themes` have been updated to the latest version.",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestUnspecifiedUpdateKeepsOriginalUTF8Span(t *testing.T) {
	const id = "filler.unspecified-update"
	source := document.Source{Name: "changes.md", Format: document.Markdown,
		Bytes: []byte("\ufeff# Café releases\r\n\r\n- Some syntaxes and themes have been updated to the latest version.\r\n")}
	result, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(string(source.Bytes[finding.Primary.Span.Start:finding.Primary.Span.End]), qt.Equals,
		finding.Primary.Snippet)
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(string(source.Bytes), "Some syntaxes"))
}
