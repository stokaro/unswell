package e2e_test

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dlclark/regexp2"
	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
)

type problemMatchers struct {
	Matchers []problemMatcher `json:"problemMatcher"`
}

type problemMatcher struct {
	Owner    string           `json:"owner"`
	Severity string           `json:"severity"`
	Patterns []problemPattern `json:"pattern"`
}

type problemPattern struct {
	Regexp   string `json:"regexp"`
	File     int    `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Severity int    `json:"severity"`
	Code     int    `json:"code"`
	Message  int    `json:"message"`
}

func TestGitHubProblemMatcher(t *testing.T) {
	var config problemMatchers
	decodeFile(t, "../.github/unswell-problem-matcher.json", &config)
	for _, row := range []struct{ name, severity, want, path string }{
		{"warning", "warning", "warning", "document/document.go"},
		{"error", "error", "error", "analyze.go"},
		{"note", "note", "notice", "docs/example.md"},
		{"Windows path", "warning", "warning", `C:\work\sample.go`},
		{"spaces and colons", "warning", "warning", "docs/release notes:v1.md"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			finding := unswell.Finding{RuleID: "syntax.noun-stack", Severity: row.severity, Message: "Review the noun sequence.",
				Primary: unswell.Location{Path: row.path, Start: document.Position{Line: 12, Column: 4}}}
			result := unswell.RunResult{SchemaVersion: unswell.SchemaVersion, Status: "complete", Findings: []unswell.Finding{finding}}
			var output bytes.Buffer
			c.Assert(report.Write(&output, "text", result, report.Options{}), qt.IsNil)
			line, _, _ := strings.Cut(output.String(), "\n")
			assertDiagnosticMatch(t, config, line, row.want, finding)
		})
	}
}

func assertDiagnosticMatch(t *testing.T, config problemMatchers, line, want string, finding unswell.Finding) {
	t.Helper()
	c := qt.New(t)
	matches := 0
	for _, matcher := range config.Matchers {
		c.Assert(matcher.Patterns, qt.HasLen, 1)
		pattern := matcher.Patterns[0]
		groups := problemGroups(t, pattern.Regexp, line)
		if groups == nil {
			continue
		}
		matches++
		severity := matcher.Severity
		if pattern.Severity != 0 {
			severity = groups[pattern.Severity]
		}
		c.Assert(severity, qt.Equals, want)
		c.Assert(groups[pattern.File], qt.Equals, finding.Primary.Path)
		c.Assert(groups[pattern.Line], qt.Equals, "12")
		c.Assert(groups[pattern.Column], qt.Equals, "4")
		c.Assert(groups[pattern.Code], qt.Equals, finding.RuleID)
		c.Assert(groups[pattern.Message], qt.Equals, finding.Message)
	}
	c.Assert(matches, qt.Equals, 1)
}

func TestGitHubMatcherLeavesCompilerDiagnostics(t *testing.T) {
	var config problemMatchers
	decodeFile(t, "../.github/unswell-problem-matcher.json", &config)
	for _, line := range []string{
		"broken.go:12:4: undefined: missing",
		"    example_test.go:12:4: expected one finding",
		"PASS: 2 documents, 1 findings; analysis complete.",
		"  related: document/document.go:12:4",
	} {
		t.Run(line, func(t *testing.T) {
			c := qt.New(t)
			for _, matcher := range config.Matchers {
				c.Assert(problemGroups(t, matcher.Patterns[0].Regexp, line), qt.IsNil)
			}
		})
	}
}

func problemGroups(t *testing.T, pattern, line string) []string {
	t.Helper()
	c := qt.New(t)
	compiled := regexp2.MustCompile(pattern, regexp2.ECMAScript)
	compiled.MatchTimeout = time.Second
	match, err := compiled.FindStringMatch(line)
	c.Assert(err, qt.IsNil)
	if match == nil {
		return nil
	}
	var groups []string
	for _, group := range match.Groups() {
		groups = append(groups, group.String())
	}
	return groups
}

func TestGitHubMatcherUsesECMAScript(t *testing.T) {
	c := qt.New(t)
	// This original character class succeeds in Go but not in Runner's dialect.
	pattern := `^(.+):([0-9]+):([0-9]+): (error|warning) \[([^]]+)\] (.+)$`
	line := "document/document.go:1:4: warning [syntax.noun-stack] Review the noun sequence."
	c.Assert(regexp.MustCompile(pattern).MatchString(line), qt.IsTrue)
	c.Assert(problemGroups(t, pattern, line), qt.IsNil)
}
