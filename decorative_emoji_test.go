package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestDecorativeEmojiIsReportedInHeadingsAndListItems(t *testing.T) {
	// The rule ships disabled outside the strict profile, so the test turns it on.
	engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nrules:\n  format.decorative-emoji: {enabled: true}\n")})
	qt.New(t).Assert(err, qt.IsNil)
	cases := []struct {
		name, file, text string
		emoji            []string
	}{
		{"heading", "guide.md", "## \U0001F680 Quick start\n\nRun the installer.\n", []string{"\U0001F680"}},
		{"list item", "guide.md", "The tool does two things.\n\n- ⚡ It starts fast.\n- It stops cleanly.\n", []string{"⚡"}},
		{"check mark in a list", "guide.md", "The release has these parts.\n\n- \u2705 The archive is signed.\n", nil},
		{"cross in a heading", "guide.md", "## \u274C Unsupported platforms\n\nThe tool does not run there.\n", nil},
		{"check mark beside an emoji", "guide.md", "The release is ready.\n\n- \u2705 Signed \U0001F389\n", []string{"\U0001F389"}},
		{"two in one heading", "guide.md", "# ✨ Features \U0001F389\n\nThe client retries.\n", []string{"✨", "\U0001F389"}},
		{"paragraph", "guide.md", "The installer finishes \U0001F680 in one minute.\n", nil},
		{"table cell", "guide.md", "| Check | State |\n| --- | --- |\n| Archive | ✅ |\n", nil},
		{"fenced code", "guide.md", "The output follows.\n\n```text\n\U0001F680 done\n```\n", nil},
		{"plain heading", "guide.md", "## Quick start\n\nRun the installer.\n", nil},
		{"shortcode", "guide.md", "## :rocket: Quick start\n\nRun the installer.\n", nil},
		{"comment in code", "client.go", "package client\n\n// \U0001F680 Start opens the connection.\nfunc Start() {}\n", nil},
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
			var found []string
			for _, finding := range result.Findings {
				if finding.RuleID != "format.decorative-emoji" {
					continue
				}
				c.Assert(finding.Severity, qt.Equals, "warning")
				for _, location := range append([]unswell.Location{finding.Primary}, finding.Related...) {
					found = append(found, tc.text[location.Span.Start:location.Span.End])
				}
			}
			c.Assert(found, qt.DeepEquals, tc.emoji)
			c.Assert(result.Gate.Passed, qt.IsTrue)
		})
	}
}
