package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

const repeatedPhraseProse = "The clear release notes explain the release schedule for every new reader. " +
	"Our clear release notes describe the maintenance window for the new team. " +
	"These clear release notes document the support process for each new user."

const overlapParagraph = "The client opens a connection to the server and sends the request with its credentials."

func repetitionEngine(t *testing.T, id, parameters, extra string) *unswell.Engine {
	t.Helper()
	c := qt.New(t)
	if parameters == "" {
		parameters = "{}"
	}
	config := "version: 1\nextends: [builtin:custom]\nrules:\n  " + id +
		":\n    enabled: true\n    gate: forbid\n    parameters: " + parameters + "\n" + extra
	engine, err := unswell.New(unswell.Options{Config: []byte(config), IncludeSource: true})
	c.Assert(err, qt.IsNil)
	return engine
}

func repetitionResult(t *testing.T, id, text, parameters, extra string) unswell.RunResult {
	t.Helper()
	c := qt.New(t)
	result, err := repetitionEngine(t, id, parameters, extra).Analyze(t.Context(),
		document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	return result
}

func TestNgramClustersPreserveEvidence(t *testing.T) {
	c := qt.New(t)
	text := "\ufeff" + strings.ReplaceAll(repeatedPhraseProse, ". ", ".\r\n\r\n")
	text = strings.Replace(text, "clear release notes", "clear **release** notes", 1)
	result := repetitionResult(t, "repetition.ngram-density", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Related, qt.HasLen, 2)
	c.Assert(finding.Primary.Snippet, qt.Equals, "clear **release** notes")
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, "clear **release** notes"))
	c.Assert(finding.Evidence.Activation, qt.Equals, 333)
	c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, float64(3))
	c.Assert(finding.Evidence.Metrics[1].Value, qt.Equals, float64(3))
	c.Assert(finding.Evidence.Metrics[2].Value, qt.Equals, float64(36))
	c.Assert(finding.Evidence.Metrics[3].Value, qt.Equals, float64(25))
	clean := repetitionResult(t, "repetition.ngram-density", text+"\n\nThe connection closes.", "", "")
	c.Assert(clean.Findings, qt.DeepEquals, result.Findings)
	c.Assert(clean.Assessments[:len(result.Assessments)], qt.DeepEquals, result.Assessments)
}

func TestNgramBoundariesAndExemptions(t *testing.T) {
	for _, row := range []struct{ name, text, parameters, extra string }{
		{"one occurrence", overlapParagraph, "", ""},
		{"window", repeatedPhraseProse, "{window_sentences: 2}", ""},
		{"punctuation", strings.ReplaceAll(repeatedPhraseProse, "release notes", "release, notes"), "", ""},
		{"protected code", strings.ReplaceAll(repeatedPhraseProse, "release", "`release`"), "", ""},
		{"approved term", repeatedPhraseProse, "", "vocabulary:\n  terms: [clear release notes]\n" +
			"  term_exemptions: [repetition.ngram-density]\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(repetitionResult(t, "repetition.ngram-density", row.text, row.parameters, row.extra).Findings, qt.HasLen, 0)
		})
	}
}

func TestNestedNgramClustersUseLongestEvidence(t *testing.T) {
	c := qt.New(t)
	text := "Readers value clear release notes with context for all of the upcoming changes. " +
		"Writers prepare clear release notes with context as part of the documentation process. " +
		"Editors review clear release notes with context to describe the release for everyone."
	result := repetitionResult(t, "repetition.ngram-density", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "clear release notes with context")
	c.Assert(result.Findings[0].Related, qt.HasLen, 2)
}

func TestRepetitionTechnicalDifferences(t *testing.T) {
	for _, row := range []struct{ name, before, after string }{
		{"negation", "opens", "never opens"},
		{"modality", "may open", "must open"},
		{"condition", "opens before", "opens after"},
		{"number", "waits 30 seconds for", "waits 60 seconds for"},
		{"numeric unit", "waits 30 seconds for", "waits 30 minutes for"},
		{"version", "uses version 1.2 to open", "uses version 1.3 to open"},
		{"identifier", "uses RetryFast to open", "uses RetrySlow to open"},
		{"state", "uses enabled settings to open", "uses disabled settings to open"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			left := strings.Replace(overlapParagraph, "opens", row.before, 1)
			right := strings.Replace(overlapParagraph, "opens", row.after, 1)
			for _, id := range []string{"repetition.paragraph-overlap", "repetition.summary-echo"} {
				text := left + "\n\n" + right
				if id == "repetition.summary-echo" {
					text = left + "\n\n## Summary\n\n" + right
				}
				result := repetitionResult(t, id, text, "{similarity: 0.5}", "")
				c.Assert(result.Findings, qt.HasLen, 0, qt.Commentf("%s", id))
			}
		})
	}
}

func TestRepetitionRulesRequireOptIn(t *testing.T) {
	c := qt.New(t)
	ids := []string{"repetition.ngram-density", "repetition.syntax-template", "repetition.paragraph-overlap",
		"repetition.heading-echo", "repetition.summary-echo"}
	seen := make(map[string]bool)
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		for _, id := range ids {
			if d.ID != id {
				continue
			}
			seen[id] = true
			c.Assert(d.Status, qt.Equals, "experimental")
			c.Assert(d.Defaults.Enabled, qt.IsFalse)
			c.Assert(d.Defaults.Gate, qt.Equals, "none")
			for _, profile := range []string{"technical", "strict", "minimal", "business", "reference", "custom"} {
				engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:" + profile + "]\n")})
				c.Assert(err, qt.IsNil)
				policy, err := engine.PolicyForFile("guide.md")
				c.Assert(err, qt.IsNil)
				c.Assert(policy.Rules[id].Enabled, qt.IsFalse)
			}
		}
	}
	c.Assert(seen, qt.HasLen, len(ids))
}
