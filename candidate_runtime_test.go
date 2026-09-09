package unswell_test

import (
	"context"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type candidateBudgetRule struct{ rule.Rule }

// Evaluate constrains the real rule after the engine has prepared shared features.
func (r candidateBudgetRule) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	view.MaxCandidates = 1
	return r.Rule.Evaluate(ctx, view, emit)
}

func TestCandidateActivationsDiscardFailedRuleValues(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(candidateActivationIDs(), d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{candidateBudgetRule{implementation}}, NoGate: true,
				Features: []string{"activation/" + d.ID}, Config: []byte("version: 1\nrules:\n  " + d.ID + ": {enabled: true}\n")})
			c.Assert(err, qt.IsNil)
			format := d.Examples[0].Format
			if format == "" {
				format = document.Plain
			}
			result, err := engine.Analyze(t.Context(), document.Source{Name: "example", Format: format,
				Bytes: []byte(d.Examples[0].Text)})
			c.Assert(err, qt.ErrorMatches, ".*(max_candidates|budget).*")
			c.Assert(result.Manifest.Complete, qt.IsFalse)
			c.Assert(result.Gate.Passed, qt.IsFalse)
			c.Assert(len(result.Features.Sources[0].Units) > 0, qt.IsTrue)
			for _, unit := range result.Features.Sources[0].Units {
				c.Assert(unit.Values[0].Number, qt.IsNil)
				c.Assert(unit.Values[0].Reason, qt.Equals, "evaluation_failed")
			}
		})
	}
}

func TestCandidateActivationsKeepExcludedProseOut(t *testing.T) {
	c := qt.New(t)
	options := unswell.Options{AllowEmpty: true}
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n"
	for _, id := range candidateActivationIDs() {
		options.Features = append(options.Features, "activation/"+id)
		policy += "  " + id + ": {enabled: true}\n"
	}
	options.Config = []byte(policy)
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("- `Clear reports`\n\n```go\nx()\n```\n")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 0)
	c.Assert(result.Features.Sources, qt.HasLen, 1)
	c.Assert(result.Features.Sources[0].Units, qt.HasLen, 0)
}

func TestCandidateActivationsRetainCancellationAndConcurrentOwnership(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"format.list-fragmentation", fragmentedLists},
		{"repetition.heading-echo", "# A practical approach to the delivery process\n\nA practical approach to the delivery process."},
		{"repetition.near-sentence", overlapParagraph + "\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)},
		{"repetition.ngram-density", repeatedPhraseProse},
		{"repetition.paragraph-overlap", overlapParagraph + "\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)},
		{"repetition.summary-echo", overlapParagraph + "\n\n# Summary\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)},
		{"repetition.syntax-template", "The careful writer describes the simple process for the entire local team. " +
			"The curious reader reviews the clear procedure for the entire small group. " +
			"The skilled editor explains the useful approach for the entire new audience."},
	} {
		t.Run(row.id, func(t *testing.T) {
			checkConcurrentActivations(t, []string{row.id}, row.text, 1)
		})
	}
}
