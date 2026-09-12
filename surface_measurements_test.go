package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestReadabilityCountingProtocol(t *testing.T) {
	c := qt.New(t)
	text := "Complexity requires verification. Complexity requires documentation."
	parameters := "{min_words: 1, min_sentences: 2, onset: 0, saturation: 40}"
	result := singleRuleResult(t, "readability.grade-metric", text, parameters, "")
	c.Assert(surfaceMetric(t, result, "prose-words"), qt.Equals, float64(6))
	c.Assert(surfaceMetric(t, result, "counted-characters"), qt.Equals, float64(61))
	c.Assert(surfaceMetric(t, result, "prose-sentences"), qt.Equals, float64(2))
	c.Assert(surfaceMetric(t, result, "automated-readability-index"), qt.Equals, 4.71*61/6+0.5*6/2-21.43)
	c.Assert(surfaceMetric(t, result, "automated-readability-index-prose"), qt.Equals, 4.71*61/6+0.5*6/2-21.43)
	c.Assert(surfaceMetric(t, result, "ari-words"), qt.Equals, float64(6))
	c.Assert(surfaceMetric(t, result, "ari-characters"), qt.Equals, float64(61))
	c.Assert(surfaceMetric(t, result, "ari-sentences"), qt.Equals, float64(2))
	c.Assert(surfaceMetric(t, result, "type-token-ratio"), qt.Equals, float64(4)/6)
	c.Assert(surfaceMetric(t, result, "hapax-token-ratio"), qt.Equals, float64(2)/6)
	c.Assert(surfaceMetric(t, result, "sentence-word-stddev"), qt.Equals, float64(0))
	for _, unit := range result.Assessments {
		c.Assert(unit.SlopScore, qt.Equals, float64(0))
	}
	protected := singleRuleResult(t, "readability.grade-metric", text+" `a b c 12345`", parameters, "")
	c.Assert(protected.Findings[0].Evidence.Metrics, qt.DeepEquals, result.Findings[0].Evidence.Metrics)
	c.Assert(singleRuleResult(t, "readability.grade-metric", text, "", "").Findings, qt.HasLen, 0)
}

// The grade metric gates on prose-only denominators: identifier-shaped tokens
// and code spans are not words, and each part of a hyphenated compound is one
// word. The shared feature keeps counting every eligible token.
func TestReadabilityProseDenominators(t *testing.T) {
	c := qt.New(t)
	text := "The feature-collection hash uses `sha256`. The TransportCacheEntry keeps config_v2 values."
	result := singleRuleResult(t, "readability.grade-metric", text, "{min_words: 1, min_sentences: 2, onset: 0, saturation: 40}", "")
	c.Assert(surfaceMetric(t, result, "ari-words"), qt.Equals, float64(8))
	c.Assert(surfaceMetric(t, result, "ari-characters"), qt.Equals, float64(42))
	c.Assert(surfaceMetric(t, result, "ari-sentences"), qt.Equals, float64(2))
	characters, words, sentences := 42.0, 8.0, 2.0
	c.Assert(surfaceMetric(t, result, "automated-readability-index-prose"), qt.Equals,
		4.71*characters/words+0.5*words/sentences-21.43)
	c.Assert(surfaceMetric(t, result, "prose-words"), qt.Equals, float64(9))
	c.Assert(surfaceMetric(t, result, "counted-characters"), qt.Equals, float64(69))
	characters, words = 69.0, 9.0
	c.Assert(surfaceMetric(t, result, "automated-readability-index"), qt.Equals,
		4.71*characters/words+0.5*words/sentences-21.43)
	c.Assert(surfaceMetric(t, result, "automated-readability-index") > 12, qt.IsTrue)
	c.Assert(singleRuleResult(t, "readability.grade-metric", text, "{min_words: 1, min_sentences: 2}", "").Findings, qt.HasLen, 0)
	long := strings.Repeat("The implementation requires comprehensive configuration, systematic verification, "+
		"and consistent documentation of operational prerequisites before production deployment. ", 4)
	c.Assert(singleRuleResult(t, "readability.grade-metric", long, "", "").Findings, qt.HasLen, 1)
}

func TestLongParagraphRequiresBothLengths(t *testing.T) {
	c := qt.New(t)
	parameters := "{onset: 10, saturation: 30, sentence_words: 6, min_long_sentences: 2}"
	text := "The client opens a connection to the server. The server sends a response to the client."
	result := singleRuleResult(t, "readability.long-paragraph", text, parameters, "")
	c.Assert(surfaceMetric(t, result, "long-sentences"), qt.Equals, float64(2))
	c.Assert(surfaceMetric(t, result, "paragraph-length"), qt.Equals, float64(16))
	for _, clean := range []string{strings.Repeat("The client waits. ", 10), "The client opens a connection to the server.",
		"The client opens a connection to the server.\n\nThe server sends a response to the client."} {
		c.Assert(singleRuleResult(t, "readability.long-paragraph", clean, parameters, "").Findings, qt.HasLen, 0)
	}
}

func TestParentheticalMappingAndNestedCounts(t *testing.T) {
	c := qt.New(t)
	text := "\ufeffThe client (which **may retry** (after failure)) waits.\r\n"
	result := singleRuleResult(t, "syntax.parenthetical-load", text, "{min_words: 1, onset: 1}", "")
	c.Assert(surfaceMetric(t, result, "parenthetical-words"), qt.Equals, float64(5))
	c.Assert(surfaceMetric(t, result, "balanced-insertions"), qt.Equals, float64(2))
	c.Assert(surfaceMetric(t, result, "maximum-insertion-depth"), qt.Equals, float64(2))
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "(which **may retry** (after failure))")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, "("))
	result = singleRuleResult(t, "syntax.parenthetical-load", "The client (may retry. It must wait.) opens a connection.",
		"{min_words: 1, onset: 1}", "")
	c.Assert(surfaceMetric(t, result, "parenthetical-words"), qt.Equals, float64(5))
	c.Assert(result.Findings[0].Related, qt.HasLen, 1)
}

func TestParentheticalBoundaries(t *testing.T) {
	for _, row := range []struct{ name, text, extra string }{
		{"unclosed", "The client (which may retry waits.", ""},
		{"mismatched", "The client (which may retry] waits.", ""},
		{"protected crossing", "The client (which `may` retry) waits.", ""},
		{"inside code", "The client `(which may retry)` waits.", ""},
		{"empty", "The client () [] waits.", ""},
		{"approved term", "The client (retry budget) applies.", "vocabulary:\n  terms: [retry budget]\n" +
			"  term_exemptions: [syntax.parenthetical-load]\n"},
		{"issue references", "Issues (#20), (#56), and (#57) track this work.", ""},
		{"license labels", "It uses regexp2 (MIT), yaml (MIT), and gonum (BSD-2-Clause).", ""},
		{"acronym gloss", "The application programming interface (API) is stable.", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, "syntax.parenthetical-load", row.text, "{min_words: 1, onset: 1}", row.extra).Findings, qt.HasLen, 0)
		})
	}
}

func TestParentheticalShortInsertionFloor(t *testing.T) {
	c := qt.New(t)
	references := "Issues (#20), (#56), and (#57) track this work."
	result := singleRuleResult(t, "syntax.parenthetical-load", references, "{min_words: 1, onset: 1, min_insertion_words: 0}", "")
	c.Assert(surfaceMetric(t, result, "parenthetical-words"), qt.Equals, float64(3))
	c.Assert(surfaceMetric(t, result, "balanced-insertions"), qt.Equals, float64(3))
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "(#20)")
	asides := "The client (see below) waits (as before)."
	result = singleRuleResult(t, "syntax.parenthetical-load", asides, "{min_words: 1, onset: 1}", "")
	c.Assert(surfaceMetric(t, result, "parenthetical-words"), qt.Equals, float64(4))
	c.Assert(surfaceMetric(t, result, "balanced-insertions"), qt.Equals, float64(2))
	c.Assert(surfaceMetric(t, result, "maximum-insertion-depth"), qt.Equals, float64(1))
	// The skipped inner reference adds no pair or depth; its word still lies inside the outer insertion.
	nested := "The client (which may retry (#56)) waits."
	result = singleRuleResult(t, "syntax.parenthetical-load", nested, "{min_words: 1, onset: 1}", "")
	c.Assert(surfaceMetric(t, result, "parenthetical-words"), qt.Equals, float64(4))
	c.Assert(surfaceMetric(t, result, "balanced-insertions"), qt.Equals, float64(1))
	c.Assert(surfaceMetric(t, result, "maximum-insertion-depth"), qt.Equals, float64(1))
}

func TestEmDashCountingAndSourceMapping(t *testing.T) {
	c := qt.New(t)
	text := "\ufeffThe client waits &mdash; then retries — if permitted.\r\n"
	result := singleRuleResult(t, "format.em-dash-density", text, "{min_words: 1}", "")
	c.Assert(surfaceMetric(t, result, "em-dash-count"), qt.Equals, float64(2))
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "&mdash;")
	c.Assert(result.Findings[0].Related[0].Snippet, qt.Equals, "—")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, "&mdash;"))
	for _, clean := range []string{"The client waits — then retries.", "The client waits – then retries - if permitted.",
		"The client waits `— —` then retries."} {
		c.Assert(singleRuleResult(t, "format.em-dash-density", clean, "{min_words: 0}", "").Findings, qt.HasLen, 0)
	}
}
