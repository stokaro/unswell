package unswell_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp/english"
)

func TestRepetitionLimitsAndCapabilities(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	text := overlapParagraph + "\n\n## Summary\n\n" + overlapParagraph
	for _, id := range []string{"repetition.ngram-density", "repetition.syntax-template", "repetition.paragraph-overlap",
		"repetition.heading-echo", "repetition.summary-echo"} {
		engine := repetitionEngine(t, id, "", "analysis: {max_candidates: 1}\n")
		result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(err, qt.ErrorMatches, ".*max_candidates.*", qt.Commentf("%s", id))
		c.Assert(result.Gate.Passed, qt.IsFalse)
		config := []byte("version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ": {enabled: true}\n")
		_, err = unswell.New(unswell.Options{Config: config, NLP: limitedPolicyNLP{Provider: provider}})
		c.Assert(err, qt.ErrorMatches, "rule "+id+" requires unavailable capability pos")
	}
	engine := repetitionEngine(t, "repetition.ngram-density", "", "")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err := engine.Analyze(ctx, document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(repeatedPhraseProse)})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestRepetitionParameterValidation(t *testing.T) {
	for _, row := range []struct{ id, parameters string }{
		{"repetition.ngram-density", "min_ngram_words: 2"},
		{"repetition.ngram-density", "max_ngram_words: 9"},
		{"repetition.ngram-density", "min_ngram_words: 7, max_ngram_words: 4"},
		{"repetition.paragraph-overlap", "window_blocks: 0"},
		{"repetition.summary-echo", "window_blocks: 129"},
	} {
		t.Run(row.id+"/"+row.parameters, func(t *testing.T) {
			c := qt.New(t)
			config := "version: 1\nrules:\n  " + row.id + ": {enabled: true, parameters: {" + row.parameters + "}}\n"
			_, err := unswell.New(unswell.Options{Config: []byte(config)})
			c.Assert(err, qt.ErrorMatches, ".*invalid .* parameter")
		})
	}
}

func TestStructureRequirementFollowsFileOverrides(t *testing.T) {
	c := qt.New(t)
	config := []byte("version: 1\nextends: [builtin:custom]\noverrides:\n  - files: [summary.md]\n" +
		"    rules:\n      repetition.summary-echo: {enabled: true, gate: forbid}\n")
	engine, err := unswell.New(unswell.Options{Config: config})
	c.Assert(err, qt.IsNil)
	text := overlapParagraph + "\n\n## Summary\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)
	for _, name := range []string{"summary.md", "other.md"} {
		result, err := engine.Analyze(t.Context(), document.Source{Name: name, Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(err, qt.IsNil)
		if name == "summary.md" {
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].RuleID, qt.Equals, "repetition.summary-echo")
		} else {
			c.Assert(result.Findings, qt.HasLen, 0)
		}
		encoded, err := json.Marshal(result)
		c.Assert(err, qt.IsNil)
		c.Assert(string(encoded), qt.Not(qt.Contains), "heading-2:")
		c.Assert(result.Documents[0].Source, qt.Equals, "")
	}
}

func TestRepeatedFactsAndCorrelatedEvidence(t *testing.T) {
	c := qt.New(t)
	text := ""
	for _, value := range []string{"30", "60", "90"} {
		text += "The clear release notes describe the timeout of " + value + " seconds for the new client. "
	}
	for _, id := range []string{"repetition.ngram-density", "repetition.syntax-template"} {
		c.Assert(repetitionResult(t, id, text, "", "").Findings, qt.HasLen, 0)
	}
	config := "version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  repetition.ngram-density: {enabled: true, score: {weight: 100, cap: 100}}\n" +
		"  repetition.paragraph-overlap: {enabled: true, score: {weight: 100, cap: 100}}\n" +
		"  repetition.exact-sentence: {enabled: true, score: {weight: 100, cap: 100}}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(config)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte(strings.Repeat(overlapParagraph+"\n\n", 3))})
	c.Assert(err, qt.IsNil)
	c.Assert(len(result.Findings) >= 3, qt.IsTrue)
	for _, unit := range result.Assessments {
		c.Assert(unit.SlopScore, qt.Equals, float64(45))
	}
}

func TestRepetitionConcurrentReuse(t *testing.T) {
	c := qt.New(t)
	engine := repetitionEngine(t, "repetition.paragraph-overlap", "", "")
	text := overlapParagraph + "\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
			c.Assert(string(source.Bytes), qt.Equals, text)
		})
	}
}

func TestSyntaxTemplatesDistinguishInstructionsAndExactCopies(t *testing.T) {
	const first = "The careful writer describes the simple process for the entire local team."
	const second = "The curious reader reviews the clear procedure for the entire small group."
	const third = "The skilled editor explains the useful approach for the entire new audience."
	c := qt.New(t)
	result := repetitionResult(t, "repetition.syntax-template", first+" "+second+" "+third, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Related, qt.HasLen, 2)
	c.Assert(result.Findings[0].Evidence.Metrics[1].Value, qt.Equals, float64(3))
	for _, text := range []string{
		strings.Repeat(first+" ", 3),
		"- " + first + "\n- " + second + "\n- " + third,
		strings.ReplaceAll(first+" "+second+" "+third, "process", "`process`"),
	} {
		c.Assert(repetitionResult(t, "repetition.syntax-template", text, "", "").Findings, qt.HasLen, 0)
	}
}
