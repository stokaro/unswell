package unswell_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp/english"
)

func TestEditorialEvidencePreservesCoordinatesAndIsolation(t *testing.T) {
	const text = "\ufeffCafé 🙂. In this section, we will describe setup.\r\n\r\n" +
		"In this **section**, we will describe deployment."
	// Section announcements require sentence starts; the prefix is a separate sentence.
	result := editorialResult(t, "filler.section-announcement", text, "")
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Evidence.Activation, qt.Equals, 333)
	c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, float64(2))
	c.Assert(finding.Related, qt.HasLen, 1)
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, "In this section"))
	c.Assert(finding.Primary.Snippet, qt.Equals, "In this section, we will")
	c.Assert(finding.Related[0].Snippet, qt.Equals, "In this **section**, we will")
	c.Assert(len(finding.Related[0].Segments) > 1, qt.IsTrue)
	clean := editorialResult(t, "filler.section-announcement", text+"\n\nThe client sends requests.", "")
	c.Assert(clean.Findings, qt.DeepEquals, result.Findings)
	for i := range result.Assessments {
		c.Assert(clean.Assessments[i], qt.DeepEquals, result.Assessments[i])
	}
}

func TestEditorialTermsReduceCountsBeforeScoring(t *testing.T) {
	c := qt.New(t)
	text := "It is important. In this section, we will describe setup. In this section, we will describe deployment."
	result := editorialResult(t, "filler.section-announcement", text,
		"vocabulary:\n  terms: ['In this section, we will describe setup']\n  term_exemptions: [filler.section-announcement]\n")
	c.Assert(result.Findings, qt.HasLen, 0)
}

func TestEditorialLimitsAndCancellationDoNotPass(t *testing.T) {
	c := qt.New(t)
	engine := editorialEngine(t, "filler.section-announcement", "analysis: {max_candidates: 1}\n")
	text := "In this section, we will describe setup. In this section, we will describe deployment."
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	_, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.ErrorMatches, ".*max_candidates.*")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = unswell.New(unswell.Options{Config: []byte("version: 1\nrules:\n  syntax.rhetorical-question-density:\n" +
		"    enabled: true\n    parameters: {max_answer_words: 0}\n")})
	c.Assert(err, qt.ErrorMatches, ".*max_answer_words.*")
}

func TestEditorialDensityReportsTheMeasuredDenominator(t *testing.T) {
	text := "The very powerful client offers an incredibly innovative interface and an extremely useful command " +
		"for every person using the service."
	result := editorialResult(t, "filler.weak-intensifiers", text, "")
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 1)
	metrics := result.Findings[0].Evidence.Metrics
	c.Assert(metrics[0].Unit, qt.Equals, "matches/100-prose-words")
	c.Assert(metrics[0].Value, qt.Equals, metrics[1].Value*100/metrics[2].Value)
	c.Assert(metrics[1].Value, qt.Equals, float64(3))
	c.Assert(metrics[2].Value, qt.Equals, float64(result.Documents[0].ProseWords))
	short := editorialResult(t, "filler.weak-intensifiers", "A very powerful and incredibly innovative client.", "")
	c.Assert(short.Findings, qt.HasLen, 0)
}

func TestEditorialCorrelationCapsAndConcurrentResults(t *testing.T) {
	c := qt.New(t)
	config := "    score: {weight: 100, cap: 100}\n  hype.modifier-cluster:\n    enabled: true\n" +
		"    score: {weight: 100, cap: 100}\n"
	engine := editorialEngine(t, "syntax.triad-density", config)
	text := "A robust, transformative, unparalleled experience follows. A robust, transformative, unparalleled platform starts."
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(len(want.Findings) >= 3, qt.IsTrue)
	for _, unit := range want.Assessments {
		c.Assert(unit.SlopScore <= 35, qt.IsTrue)
	}
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result, qt.DeepEquals, want)
		})
	}
}

func TestEditorialPOSCapabilityCannotBeSubstituted(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	for _, id := range []string{"filler.weak-intensifiers", "filler.stacked-hedging", "syntax.triad-density"} {
		_, err := unswell.New(unswell.Options{NLP: limitedPolicyNLP{Provider: provider}, Config: []byte(
			"version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ": {enabled: true}\n")})
		c.Assert(err, qt.ErrorMatches, "rule "+id+" requires unavailable capability pos")
	}
}
