package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestBaselineRepetitionRechecksUnchangedUnits(t *testing.T) {
	c := qt.New(t)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.exact-sentence: {enabled: true}\n" +
		"gate:\n  paragraph_score: {fail_at: 1, min_words: 1}\n  sentence_score: {fail_at: 1, min_words: 1}\n"
	prose := "The client opens a new connection after the server closes the previous connection."
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# First\n\n" + prose + "\n\n# Second\n\n" + prose)}
	data, before := captureDebt(t, policy, source)
	c.Assert(before.Gate.Passed, qt.IsFalse)
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	source.Bytes = append(source.Bytes, []byte("\n\n# Third\n\n"+prose)...)
	after, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(after.Gate.Passed, qt.IsFalse)
	for _, assessment := range after.Assessments {
		if assessment.SlopScore > 0 {
			c.Assert(assessment.BaselineState, qt.Equals, "new")
		}
	}
}

func TestBaselineSuppressionBindingIsPolicySensitive(t *testing.T) {
	c := qt.New(t)
	permission := "<!-- unswell-disable-next-block filler.announced-importance -- Required contract wording. -->\n\n"
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# First\n\n" + permission + debtProse + "\n\n# Second\n\n" + debtProse)}
	data, _ := captureDebt(t, debtPolicy, source)
	engine, err := unswell.New(unswell.Options{Config: []byte(debtPolicy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	source.Bytes = append([]byte("A clean introduction.\n\n"), source.Bytes...)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "untracked")
	source.Bytes = []byte("# First\n\n" + debtProse + "\n\n# Second\n\n" + permission + debtProse)
	result, err = engine.Analyze(t.Context(), source)
	c.Assert(err, qt.ErrorMatches, ".*suppressions changed.*")
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestBaselineOverridesAndPartialCoverage(t *testing.T) {
	c := qt.New(t)
	policy := debtPolicy + "overrides:\n  - files: ['new.md']\n    gate: {mode: new}\n"
	source := document.Source{Name: "old.md", Format: document.Markdown, Bytes: []byte(debtProse)}
	data, _ := captureDebt(t, policy, source)
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), Baseline: data})
	c.Assert(err, qt.IsNil)
	other := document.Source{Name: "new.md", Format: document.Markdown, Bytes: []byte(debtProse)}
	result, err := engine.Analyze(t.Context(), other)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Baseline.Unobserved, qt.HasLen, 3)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	changedPolicy := strings.ReplaceAll(policy, "gate: {mode: new}", "rules:\n      filler.announced-importance: {enabled: false}")
	engine, err = unswell.New(unswell.Options{Config: []byte(changedPolicy), Baseline: data})
	c.Assert(err, qt.IsNil)
	result, err = engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "existing")
	c.Assert(result.Gate.Passed, qt.IsFalse)
}
