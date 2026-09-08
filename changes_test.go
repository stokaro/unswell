package unswell_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestChangedAnalysisRetainsUnchangedDebtAfterLineMovement(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	before := changedSource("# Retries\n\n" + debtProse)
	after := changedSource("# **Retries**\n\nA clean introduction.\n\n" + debtProse)
	result, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Gate.Unchanged, qt.HasLen, 3)
	c.Assert(result.Baseline, qt.IsNil)
	c.Assert(result.BaselineSnapshot, qt.IsNil)
	c.Assert(result.Changes.Complete, qt.IsTrue)
	c.Assert(result.Changes.Documents[0].Status, qt.Equals, "modified")
	for _, finding := range result.Findings {
		c.Assert(finding.ChangeState, qt.Equals, "unchanged")
		c.Assert(finding.BaselineState, qt.Equals, "untracked")
	}
	for _, assessment := range result.Assessments {
		if assessment.SlopScore > 0 {
			c.Assert(assessment.SlopScore, qt.Equals, float64(15))
			c.Assert(assessment.EffectiveSlopScore, qt.Equals, float64(15))
			c.Assert(assessment.ChangeState, qt.Equals, "unchanged")
		}
	}
}

func TestChangedParagraphIncludesUneditedSentencesAndDeletedSeparators(t *testing.T) {
	for _, suffix := range []string{" A new condition applies.", "\nThe client must not retry."} {
		c := qt.New(t)
		engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
		c.Assert(err, qt.IsNil)
		before := changedSource(debtProse + "\n\nThe client must not retry.")
		after := changedSource(debtProse + suffix)
		result, err := engine.AnalyzeChanged(t.Context(), before, after)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Gate.Passed, qt.IsFalse)
		c.Assert(result.Gate.Reasons, qt.HasLen, 3)
		for _, assessment := range result.Assessments {
			c.Assert(assessment.ChangeState, qt.Equals, "changed")
		}
	}
}

func TestChangedRepetitionSelectsOlderOccurrences(t *testing.T) {
	c := qt.New(t)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.exact-sentence: {enabled: true}\n" +
		"gate:\n  paragraph_score: {fail_at: 1, min_words: 1}\n  sentence_score: {fail_at: 1, min_words: 1}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	prose := "The client opens a new connection after the server closes the previous connection."
	before := changedSource("# First\n\n" + prose + "\n\n# Second\n\n" + prose)
	after := changedSource(string(before[0].Bytes) + "\n\n# Third\n\n" + prose)
	result, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	for _, assessment := range result.Assessments {
		if assessment.SlopScore > 0 {
			c.Assert(assessment.ChangeState, qt.Equals, "changed")
		}
	}
}

func TestChangedAmbiguityRequiresACompleteSourceMatch(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	before := changedSource(debtProse + "\n\n" + debtProse)
	unchanged, err := engine.AnalyzeChanged(t.Context(), before, before)
	c.Assert(err, qt.IsNil)
	c.Assert(unchanged.Gate.Passed, qt.IsTrue)
	after := changedSource("A new introduction.\n\n" + string(before[0].Bytes))
	result, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	for _, finding := range result.Findings {
		c.Assert(finding.ChangeState, qt.Equals, "ambiguous")
	}
}

func TestChangedAnalysisKeepsBaselineAcceptanceSeparate(t *testing.T) {
	c := qt.New(t)
	before := changedSource("# First\n\n" + debtProse)
	data, _ := captureDebt(t, debtPolicy, before[0])
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	after := changedSource(string(before[0].Bytes) + "\n\n# Second\n\n" + debtProse)
	result, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Gate.Accepted, qt.HasLen, 3)
	c.Assert(result.Gate.Reasons, qt.HasLen, 3)
	c.Assert(result.Findings[0].ChangeState, qt.Equals, "unchanged")
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "existing")
	c.Assert(result.Findings[1].ChangeState, qt.Equals, "changed")
	c.Assert(result.Findings[1].BaselineState, qt.Equals, "new")
}

func TestChangedPermissionSelectsTheEntireDocument(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy)})
	c.Assert(err, qt.IsNil)
	permission := "<!-- unswell-disable-next-block filler.announced-importance -- Required contract wording. -->\n\n"
	before := changedSource("# First\n\n" + permission + debtProse + "\n\n# Second\n\n" + debtProse)
	after := changedSource("# First\n\n" + debtProse + "\n\n# Second\n\n" + permission + debtProse)
	result, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Changes.Documents[0].FullReason, qt.Equals, "suppressions_changed")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.ChangeState, qt.Equals, "changed")
	}
}

func TestChangedAnalysisRejectsIncompletePreviousSource(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), NoGate: true})
	c.Assert(err, qt.IsNil)
	result, err := engine.AnalyzeChanged(t.Context(), changedSource(string([]byte{0xff})), changedSource(debtProse))
	c.Assert(err, qt.ErrorMatches, "previous source analysis: .*UTF-8.*")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Changes.Complete, qt.IsFalse)
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = engine.AnalyzeChanged(ctx, nil, changedSource(debtProse))
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestChangedAnalysisIsConcurrentAndOwnsItsResults(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Jobs: 4})
	c.Assert(err, qt.IsNil)
	before, after := changedSource(debtProse), changedSource(strings.ReplaceAll(debtProse, "may retry", "may not retry"))
	want, err := engine.AnalyzeChanged(t.Context(), before, after)
	c.Assert(err, qt.IsNil)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			result, err := engine.AnalyzeChanged(t.Context(), before, after)
			c.Check(err, qt.IsNil)
			c.Check(result, qt.DeepEquals, want)
			result.Changes.Documents[0].Path = "altered.md"
		})
	}
	workers.Wait()
}

func changedSource(text string) []document.Source {
	return []document.Source{{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}}
}
