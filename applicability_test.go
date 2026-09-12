package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

// cyrillicParagraph holds more than 20 letters, all outside the Latin script.
const cyrillicParagraph = "Клиент повторяет запрос после сбоя транспорта и ждет ответа сервера."

func TestNonLatinBlockIsExcludedAndTheRestIsAnalyzed(t *testing.T) {
	for _, row := range []struct {
		name   string
		source document.Source
	}{
		{"markdown", document.Source{Name: "draft.md", Format: document.Markdown, Bytes: []byte(
			"Certainly! The client opens connections.\n\n" +
				cyrillicParagraph + "\n\n" +
				"It is important to note that the client can retry the request.\n")}},
		{"go", document.Source{Name: "sample.go", Format: document.Go, Bytes: []byte(
			"package sample\n\n" +
				"// Certainly! The client opens connections.\n\n" +
				"// " + cyrillicParagraph + "\n\n" +
				"const message = \"It is important to note that the client can retry the request.\"\n")}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), row.source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Status, qt.Equals, "complete")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Errors, qt.HasLen, 0)
			c.Assert(ruleIDs(result), qt.DeepEquals, []string{"scaffold.chat-preamble", "filler.announced-importance"})
			doc := result.Documents[0]
			c.Assert(doc.Blocks, qt.Equals, 3)
			c.Assert(doc.Excluded, qt.HasLen, 1)
			c.Assert(doc.Excluded[0].Reason, qt.Equals, "non-latin-prose")
			span := doc.Excluded[0].Span
			c.Assert(strings.TrimSpace(string(row.source.Bytes[span.Start:span.End])), qt.Equals, cyrillicParagraph)
			// The skipped block keeps its position, so the block after it is still block 2.
			c.Assert(paragraphIDs(result), qt.DeepEquals, []int{0, 2})
		})
	}
}

func TestNonLatinProseThreshold(t *testing.T) {
	latin := strings.Repeat("a", 10)
	for _, row := range []struct {
		name     string
		probe    string
		excluded bool
	}{
		{"19 non-Latin letters stay analyzed", strings.Repeat("я", 19), false},
		{"20 non-Latin letters are excluded", strings.Repeat("я", 20), true},
		{"an even split stays analyzed", latin + strings.Repeat("я", 10), false},
		{"a non-Latin majority is excluded", latin + strings.Repeat("я", 11), true},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{})
			c.Assert(err, qt.IsNil)
			text := "The client retries only when enabled.\n\n" + row.probe + "\n"
			result, err := engine.Analyze(t.Context(), document.Source{Name: "draft.md", Format: document.Markdown, Bytes: []byte(text)})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Status, qt.Equals, "complete")
			c.Assert(result.Documents[0].Blocks, qt.Equals, 2)
			if row.excluded {
				c.Assert(exclusionReasons(result), qt.DeepEquals, []string{"non-latin-prose"})
				return
			}
			c.Assert(exclusionReasons(result), qt.HasLen, 0)
		})
	}
}

func TestNonLatinOnlyDocumentIsAnEmptyScan(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "draft.md", Format: document.Markdown,
		Bytes: []byte(cyrillicParagraph + "\n\n" + cyrillicParagraph + "\n")}
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.ErrorMatches, "scan contains no applicable English prose")
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Documents[0].ProseWords, qt.Equals, 0)
	c.Assert(exclusionReasons(result), qt.DeepEquals, []string{"non-latin-prose", "non-latin-prose"})

	permissive, err := unswell.New(unswell.Options{AllowEmpty: true})
	c.Assert(err, qt.IsNil)
	result, err = permissive.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 0)
	c.Assert(exclusionReasons(result), qt.DeepEquals, []string{"non-latin-prose", "non-latin-prose"})
}

func ruleIDs(result unswell.RunResult) []string {
	ids := make([]string, 0, len(result.Findings))
	for _, finding := range result.Findings {
		ids = append(ids, finding.RuleID)
	}
	return ids
}

func paragraphIDs(result unswell.RunResult) []int {
	ids := make([]int, 0, len(result.Assessments))
	for _, assessment := range result.Assessments {
		if assessment.Scope == "paragraph" {
			ids = append(ids, assessment.UnitID)
		}
	}
	return ids
}

func exclusionReasons(result unswell.RunResult) []string {
	reasons := make([]string, 0, len(result.Documents[0].Excluded))
	for _, exclusion := range result.Documents[0].Excluded {
		reasons = append(reasons, exclusion.Reason)
	}
	return reasons
}
