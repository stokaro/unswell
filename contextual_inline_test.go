package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestContextualPatternsRetainProseAroundInlineCode(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"syntax.paired-contrast-density",
			"The `request` is rejected rather than queued. The `response` is saved rather than discarded."},
		{"syntax.not-only-density",
			"For `requests`, it not only reads but writes. For `responses`, it not only checks but validates."},
		{"filler.section-announcement",
			"In this section, we will describe `setup`. In this section, we will describe `deployment`."},
		{"filler.empty-transition",
			"With that being said, use `setup`. It goes without saying that `deployment` follows."},
		{"hype.metaphor-cluster",
			"The `service` offers a rich tapestry of possibilities. The `client` starts a journey of innovation."},
		{"syntax.whether-preface-density",
			"Whether you are a beginner or an expert, use `setup`. Whether you are a writer or a reader, use `deployment`."},
		{"syntax.triad-density",
			"A powerful, seamless, innovative platform starts `setup`. A robust, transformative, unparalleled experience runs `deployment`."},
		{"syntax.rhetorical-question-density",
			"Use `setup`. The result? A better experience. Use `deployment`. The benefit? A brighter future."},
	} {
		t.Run(row.id, func(t *testing.T) {
			c := qt.New(t)
			result := editorialResult(t, row.id, row.text, "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.RuleID, qt.Equals, row.id)
			for _, occurrence := range finding.Evidence.Occurrences {
				for _, span := range occurrence.Spans {
					c.Assert(row.text[span.Start:span.End], qt.Not(qt.Contains), "`")
				}
			}
		})
	}
}

func TestContextualCandidatesNeverConsumeProtectedTokens(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"syntax.paired-contrast-density",
			"The request uses `rather than` in its example. The response uses `rather than` in its example."},
		{"syntax.paired-contrast-density",
			"The request is rejected rather `example` than queued. The response is saved rather `example` than discarded."},
		{"syntax.not-only-density",
			"It not only reads `example` but also writes. It not only checks `example` but also validates."},
		{"syntax.not-only-density",
			"It not `only` reads but also writes. It not `only` checks but also validates."},
		{"filler.section-announcement",
			"In this `example` section, we will describe setup. In this `example` section, we will describe deployment."},
		{"syntax.whether-preface-density",
			"Whether you are a `beginner` or an expert, start here. Whether you are a `writer` or a reader, start there."},
		{"syntax.paired-contrast-density",
			"The request is rejected rather than queued.\n\n```text\nexample\n```\n\nThe response is saved rather than discarded."},
		{"syntax.paired-contrast-density",
			"The request is rejected rather than queued.\n\n## Another operation\n\nThe response is saved rather than discarded."},
	} {
		t.Run(row.id+"/"+row.text, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(editorialResult(t, row.id, row.text, "").Findings, qt.HasLen, 0)
		})
	}
}

func TestContrastWindowCountsOriginalSentences(t *testing.T) {
	first := "The `request` is rejected rather than queued. "
	last := "The `response` is saved rather than discarded."
	engine := editorialEngine(t, "syntax.paired-contrast-density", "")
	for _, row := range []struct {
		name   string
		middle string
		want   int
	}{
		{"within eight sentences", strings.Repeat("Check `state`. ", 6), 1},
		{"outside eight sentences", strings.Repeat("Check `state`. ", 7), 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := engine.Analyze(t.Context(), document.Source{
				Name: "guide.md", Format: document.Markdown, Bytes: []byte(first + row.middle + last),
			})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, row.want)
		})
	}
}

func TestContextualActivationsKeepCodeMentionInsideWindow(t *testing.T) {
	c := qt.New(t)
	id := "syntax.not-only-density"
	engine, err := unswell.New(unswell.Options{
		Features: []string{"activation/" + id},
		Config:   []byte("version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ": {enabled: true}\n"),
	})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{
		Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("It not only reads but writes.\n\nUse `code`.\n\nIt not only reads but writes."),
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	assertPhraseMeasurements(t, result, []string{"", "inapplicable/no_eligible_window", ""}, []float64{0.333, 0.333})
}

func TestDefaultContrastsAreAdvisoryAndConfigurable(t *testing.T) {
	// The two contrasts carry different facts. Their repeated form may be
	// worth editing, but detecting it must not claim either fact is redundant.
	text := "TLS encrypts transport rather than stored files. Backups preserve snapshots rather than live transaction history."
	for _, row := range []struct {
		profile, extra string
		want           int
	}{
		{"technical", "", 1}, {"strict", "", 1}, {"minimal", "", 0}, {"custom", "", 0},
		{"technical", "rules:\n  syntax.paired-contrast-density: {enabled: false}\n", 0},
	} {
		t.Run(row.profile+row.extra, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(
				"version: 1\nextends: [builtin:" + row.profile + "]\n" + row.extra,
			)})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{
				Name: "guide.md", Format: document.Markdown, Bytes: []byte(text),
			})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Gate.Passed, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, row.want)
			for _, finding := range result.Findings {
				c.Assert(finding.RuleID, qt.Equals, "syntax.paired-contrast-density")
				c.Assert(finding.Severity, qt.Equals, "note")
				c.Assert(finding.Gate, qt.Equals, "none")
			}
		})
	}
}

func TestContrastAlternativesShareOneBoundedWindow(t *testing.T) {
	for _, row := range []struct {
		name, text string
		want       int
	}{
		{"comma and lexical", "Adoption is a path, not a switch. The plan records `state` rather than guesses.", 1},
		{"opaque alternatives", "The command reads `source`, not `target`. The check reads `state`, not `history`.", 1},
		{"one useful contrast", "The digest covers the plan, not the clock.", 0},
		{"additive idiom", "The command reads files, not only directories. The check reads rows, not just tables.", 0},
		{"parenthetical idiom", "The command fails, not surprisingly. The check takes time, not to mention memory.", 0},
		{"incomplete alternatives", "The command reads files, not. The check reads rows, not.", 0},
		{"protected marker", "The command reads files, `not` directories. The check reads rows, `not` tables.", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := editorialResult(t, "syntax.paired-contrast-density", row.text, "")
			c.Assert(result.Findings, qt.HasLen, row.want)
		})
	}
}
