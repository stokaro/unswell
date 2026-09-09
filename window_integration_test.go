package unswell_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

func TestWindowActivationsRetainResourceFailures(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(windowActivationIDs(), d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{NoGate: true, Features: []string{"activation/" + d.ID}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 1}\nrules:\n  " + d.ID + ": {enabled: true}\n")})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
				Bytes: []byte(d.Examples[0].Text)})
			c.Assert(err, qt.ErrorMatches, ".*shared feature checks exceed max_candidates.*")
			c.Assert(result.Manifest.Complete, qt.IsFalse)
			c.Assert(result.Gate.Passed, qt.IsFalse)
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}

func TestWindowActivationsRetainBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, id, text string
		reasons        []string
		numbers        []float64
	}{
		{"contrast across paragraphs", "syntax.paired-contrast-density", "It is not about speed.\n\nIt is about impact.",
			[]string{"inapplicable/no_eligible_window", "inapplicable/no_eligible_window"}, nil},
		{"answer across paragraphs", "syntax.rhetorical-question-density", "The result?\n\nAn answer.",
			[]string{"inapplicable/no_eligible_window", "inapplicable/no_eligible_window"}, nil},
		{"not-only across heading", "syntax.not-only-density", "It not only reads but writes.\n\n# Setup\n\nIt not only reads but writes.",
			[]string{"", "inapplicable/unsupported_unit", ""}, []float64{0, 0}},
		{"not-only across protected sentence", "syntax.not-only-density",
			"It not only reads but writes.\n\nUse `code`.\n\nIt not only reads but writes.",
			[]string{"", "inapplicable/no_eligible_window", ""}, []float64{0, 0}},
		{"not-only across excluded fence", "syntax.not-only-density",
			"It not only reads but writes.\n\n```text\nIgnored\n```\n\nIt not only reads but writes.",
			[]string{"", ""}, []float64{0, 0}},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			options := unswell.Options{Features: []string{"activation/" + test.id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nrules:\n  " + test.id + ": {enabled: true}\n")}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(test.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 0)
			assertPhraseMeasurements(t, result, test.reasons, test.numbers)
			options.Features = nil
			ordinary, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			want, err := ordinary.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			result.Features = nil
			c.Assert(result, qt.DeepEquals, want)
		})
	}
}

func TestWindowActivationsRetainCancellationAndConcurrentOwnership(t *testing.T) {
	checkConcurrentActivations(t, windowActivationIDs(),
		"It not only reads but writes. It not only checks but validates.\n\nThe result? A better experience. The benefit? A brighter future.", 2)
}
