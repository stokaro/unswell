package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

const denialTriplet = "Backups are not a checkbox. They are your last line of defense. " +
	"Monitoring is not a dashboard. It is the foundation of operational confidence. " +
	"Testing is not a phase. It is a commitment to quality."

func TestRhetoricalFramesExposeBothClauses(t *testing.T) {
	c := qt.New(t)
	result := singleRuleResult(t, "syntax.repeated-reframing", denialTriplet, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Evidence.Metrics[0].Name, qt.Equals, "denial-redefinition")
	c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, float64(3))
	c.Assert(finding.Primary.Snippet, qt.Equals, "Backups are not a checkbox")
	c.Assert(finding.Related, qt.HasLen, 5)
	c.Assert(finding.Related[0].Snippet, qt.Equals, "They are your last line of defense")
	c.Assert(finding.Related[4].Snippet, qt.Equals, "It is a commitment to quality")
}

func TestRhetoricalFramesGeneralize(t *testing.T) {
	for _, text := range []string{
		"Reviews aren't a ritual; they're a safeguard. Ownership isn't a title; it's a responsibility.",
		"Recovery is not a slogan. That is a promise. Reviews are not a ritual. Those are a safeguard.",
		"Reviews weren’t a ritual. They were a safeguard. Ownership wasn’t a title. It was a responsibility.",
		"Recovery is not just about speed — it is about trust. Reliability is not merely about uptime — it is about confidence.",
		"Resilience is not a purchase. Resilience is a practice. Maintenance is not a burden. Maintenance is an investment.",
		"`Backup()` is not a checkbox. It is a safeguard. `Monitor()` is not a dashboard. It is a commitment.",
		"Backups are not a **checkbox**. They are a safeguard.\r\n\r\nTesting is not a phase. It is a commitment.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "syntax.repeated-reframing", text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Related, qt.HasLen, 3)
			c.Assert(result.Findings[0].Evidence.Metrics[0].Value, qt.Equals, float64(2))
		})
	}
}

func TestRhetoricalFramesRespectBoundaries(t *testing.T) {
	pair := "Backups are not a checkbox. They are a safeguard."
	for _, row := range []struct{ name, text, parameters string }{
		{"single technical distinction", "A dev database is not the target. It is a disposable replay target.", ""},
		{"prohibition", "The client must not retry. It is a safety constraint. The server must not cache. It is a privacy constraint.", ""},
		{"different subjects", "Backups are not a ritual. Logs are a record. Tests are not a phase. Reviews are a safeguard.", ""},
		{"negated redefinition", "Backups are not a ritual. They are not a guarantee. Tests are not a phase. They are not a guarantee.", ""},
		{"additive", "Backups are not only a safeguard. They are a requirement. Tests are not only a phase. They are a commitment.", ""},
		{"states", "The client is not ready. It is busy. The server is not ready. It is busy.", ""},
		{"question", "Backups are not a checkbox? They are a safeguard. Testing is not a phase? It is a commitment.", ""},
		{"window", denialTriplet, "{window_sentences: 2}"},
		{"heading", pair + "\n\n## Another section\n\n" + pair, ""},
		{"fence", pair + "\n\n```go\nvar x = 1\n```\n\n" + pair, ""},
		{"omitted comment", pair + "\n\n<!-- boundary -->\n\n" + pair, ""},
		{"list", "- " + pair + "\n- " + pair, ""},
		{"protected operator", strings.ReplaceAll(denialTriplet, "not", "`not`"), ""},
		{"split pair", "Backups are not a checkbox.\n\nThey are a safeguard.\n\nTesting is not a phase.\n\nIt is a commitment.", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, "syntax.repeated-reframing", row.text, row.parameters, "").Findings, qt.HasLen, 0)
		})
	}
}

func TestDocumentMetadiscourseFrames(t *testing.T) {
	for _, text := range []string{
		"This page defines all four; other pages link here instead of redefining them.",
		"This guide explains the `--dev-url` option.",
		"The following sections described recovery and rollback.",
		"In this chapter, we will discuss the retry protocol.",
		"Throughout this guide we demonstrate the deployment process.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "filler.document-metadiscourse", text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Evidence.Metrics[0].Name, qt.Equals, "document-metadiscourse")
			c.Assert(result.Findings[0].Primary.Span.End > strings.Index(text, " ")+1, qt.IsTrue)
		})
	}
}

func TestDocumentMetadiscourseControls(t *testing.T) {
	for _, text := range []string{
		"See the configuration reference for the complete list of environment variables.",
		"The page contains 8 KiB. The document is stored in the cache.",
		"The client defines the retry policy. This API describes database state.",
		"This page does not describe authentication.",
		"This `page defines` all four database roles.",
		"This page `defines` all four database roles.",
		"Does this guide explain the retry protocol?",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, "filler.document-metadiscourse", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestRhetoricalFramesDefaultToUnscoredNotes(t *testing.T) {
	for _, profile := range []string{"technical", "strict"} {
		t.Run(profile, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte("version: 1\nextends: [builtin:" + profile + "]\n")})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
				Bytes: []byte(denialTriplet + "\n\nThis page defines the four roles.")})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 2)
			c.Assert(result.Gate.Passed, qt.IsTrue)
			for _, finding := range result.Findings {
				c.Assert(finding.Severity, qt.Equals, "note")
				c.Assert(finding.Gate, qt.Equals, "none")
			}
		})
	}
}

func TestRhetoricalFramesPreserveMappedProse(t *testing.T) {
	c := qt.New(t)
	text := "\ufeffReviews aren't a **ritual**; they're a safeguard.\r\n" +
		"`Backup()` isn't a checkbox; it's a commitment."
	result := singleRuleResult(t, "syntax.repeated-reframing", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(finding.Primary.Snippet, qt.Equals, "Reviews aren't a **ritual")
	c.Assert(finding.Related[1].Snippet, qt.Equals, "`Backup()` isn't a checkbox")
}

func TestRhetoricalFramesTechnicalRepetitionStaysAdvisory(t *testing.T) {
	c := qt.New(t)
	text := "A dev database is not the target. It is a disposable replay target. " +
		"A shadow database is not the target. It is a disposable verification target."
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleID, qt.Equals, "syntax.repeated-reframing")
	c.Assert(result.Gate.Passed, qt.IsTrue)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.SlopScore, qt.Equals, float64(0))
	}
}

func TestRhetoricalFramesRevisionsAndExemptions(t *testing.T) {
	id := "syntax.repeated-reframing"
	c := qt.New(t)
	revised := "Backups are your last line of defense. Monitoring is the foundation of operational confidence. " +
		"Testing is a commitment to quality."
	c.Assert(singleRuleResult(t, id, revised, "", "").Findings, qt.HasLen, 0)
	text := "Backups are not a checkbox. They are a safeguard. Testing is not a phase. It is a commitment."
	exempt := windowTerm(id, "Backups are not a checkbox")
	c.Assert(singleRuleResult(t, id, text, "", exempt).Findings, qt.HasLen, 0)
	meta := "filler.document-metadiscourse"
	c.Assert(singleRuleResult(t, meta, "This page defines all four roles.", "", windowTerm(meta, "This page defines all four roles")).Findings,
		qt.HasLen, 0)
}

func TestRhetoricalFramesExposeBlockActivations(t *testing.T) {
	for _, id := range []string{"syntax.repeated-reframing", "filler.document-metadiscourse"} {
		t.Run(id, func(t *testing.T) {
			checkWindowAbsence(t, id, []windowAbsenceCase{
				{"The client starts.", "", "", ""},
				{"# The client starts\n", "", "", "unsupported_unit"},
			})
		})
	}
	compareLocalExample(t, "syntax.repeated-reframing", rule.Example{Text: denialTriplet, Match: true})
	compareLocalExample(t, "filler.document-metadiscourse", rule.Example{Text: "This page defines all four roles.", Match: true})
}
