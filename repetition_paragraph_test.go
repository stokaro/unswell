package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
)

const paragraphOpenerSample = "Measured on PostgreSQL 18.6. The client may wait up to 30 seconds.\n\n" +
	"Measured on PostgreSQL 18.7. The client must not wait more than 60 seconds.\n\n" +
	"Measured on PostgreSQL 18.8. The client retries only after the connection closes."

func TestParagraphOpenersKeepSentenceBoundariesAndTechnicalFacts(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{CollectBaseline: true, IncludeSource: true,
		Features: []string{"activation/repetition.paragraph-openers"}, Config: []byte(
			"version: 1\nextends: [builtin:custom]\nrules:\n  repetition.paragraph-openers: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	text := "\ufeff" + strings.ReplaceAll(paragraphOpenerSample, "\n", "\r\n")
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(string(source.Bytes), qt.Equals, text)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Documents[0].Source, qt.Equals, text)
	c.Assert(result.Documents[0].Sentences, qt.Equals, 6)
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.RuleVersion, qt.Equals, "3")
	c.Assert(finding.Primary.Snippet, qt.Equals, "Measured on PostgreSQL 18.6.")
	c.Assert(finding.Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(finding.Related, qt.HasLen, 2)
	c.Assert(finding.Related[0].Snippet, qt.Equals, "Measured on PostgreSQL 18.7.")
	c.Assert(finding.Related[0].Start.Line, qt.Equals, 3)
	c.Assert(finding.Related[1].Snippet, qt.Equals, "Measured on PostgreSQL 18.8.")
	c.Assert(finding.Related[1].Start.Line, qt.Equals, 5)
	assertPhraseMeasurements(t, result, []string{"", "", ""}, []float64{0.333, 0.333, 0.333})
	c.Assert(result.Features.Sources[0].RulesetHash, qt.Equals, result.Manifest.RulesetHash)
	c.Assert(result.BaselineSnapshot, qt.IsNotNil)
	candidates := result.BaselineSnapshot.Candidates
	c.Assert(candidates[0].RuleVersion, qt.Equals, "3")
	previous := candidates[0]
	previous.RuleVersion = "1"
	oldFingerprint, err := baseline.Fingerprint(previous)
	c.Assert(err, qt.IsNil)
	c.Assert(finding.BaselineFingerprint, qt.Not(qt.Equals), oldFingerprint)
}

func TestParagraphOpenersSeparateParagraphAndSentenceMinima(t *testing.T) {
	longOpeners := strings.NewReplacer("18.6. The", "18.6, where the", "18.7. The", "18.7, where the",
		"18.8. The", "18.8, where the").Replace(paragraphOpenerSample)
	for _, test := range []struct {
		name, text            string
		paragraphs, sentences int
	}{
		{"short opening sentences", paragraphOpenerSample, 1, 0},
		{"long opening sentences", longOpeners, 1, 1},
		{"short paragraphs", "Measured on PostgreSQL 18.6.\n\nMeasured on PostgreSQL 18.7.\n\nMeasured on PostgreSQL 18.8.", 0, 0},
		{"protected prefixes", "`Measured on PostgreSQL 18.6`. The client may wait up to 30 seconds.\n\n" +
			"`Measured on PostgreSQL 18.7`. The client must not wait more than 60 seconds.\n\n" +
			"`Measured on PostgreSQL 18.8`. The client retries only after the connection closes.", 0, 0},
		{"insufficient first prefix", "Hi. " + strings.ReplaceAll(longOpeners, "\n\n", "\n\nHi. "), 0, 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			for _, check := range []struct {
				id    string
				count int
			}{
				{"repetition.paragraph-openers", test.paragraphs}, {"repetition.sentence-openers", test.sentences},
			} {
				result := singleRuleResult(t, check.id, test.text, "", "")
				c.Assert(result.Findings, qt.HasLen, check.count, qt.Commentf("%s", check.id))
			}
		})
	}
}
