package unswell_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

func TestWindowPhraseActivationsPreserveCatalogExamples(t *testing.T) {
	ids := []string{"filler.section-announcement", "filler.empty-transition", "hype.metaphor-cluster"}
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(ids, d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(d.BlockObservations, qt.IsTrue)
			for _, example := range d.Examples {
				compareLocalExample(t, d.ID, example)
			}
		})
	}
}

func TestWindowPhraseActivationFailuresAndConcurrentCalls(t *testing.T) {
	c := qt.New(t)
	config := "version: 1\nextends: [builtin:custom]\nrules:\n  filler.section-announcement: {enabled: true}\n"
	// The collection fits three entries; four phrase attempts still exhaust the rule budget.
	failure := "version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 3}\nrules:\n" +
		"  filler.section-announcement: {enabled: true, parameters: {phrases: [in alpha, in beta, in gamma, in delta]}}\n"
	options := unswell.Options{NoGate: true, Config: []byte(failure),
		Features: []string{"activation/filler.section-announcement"}}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
		"In this section, we will describe setup. In this section, we will describe deployment.")}
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("In")})
	c.Assert(err, qt.ErrorMatches, ".*editorial pattern checks exceed max_candidates.*")
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	assertPhraseMeasurements(t, result, []string{"evaluation_failed"}, nil)
	options.Config = []byte(config)
	engine, err = unswell.New(options)
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
		})
	}
}

func TestWindowPhraseActivationsRespectCandidateBoundaries(t *testing.T) {
	for _, test := range []struct {
		name, text, id, parameters, extra string
		reasons                           []string
		numbers                           []float64
		findings                          int
	}{
		{"cluster crosses paragraphs", "Alpha beta.\n\nAlpha beta.\n\nGamma delta.", "", "", "",
			[]string{"", "", ""}, []float64{0.333, 0.333, 0}, 1},
		{"one occurrence is allowed", "Alpha beta.", "", "", "", []string{""}, []float64{0}, 0},
		{"short fragment", "Alpha", "", "", "", []string{"inapplicable/no_eligible_window"}, nil, 0},
		{"empty dictionary", "Alpha beta.", "", "{phrases: []}", "", []string{"inapplicable/no_patterns"}, nil, 0},
		{"protected sentence", "Alpha beta with `code`.", "", "", "", []string{"inapplicable/no_eligible_window"}, nil, 0},
		{"exempt opening", "Alpha beta.", "", "", "vocabulary:\n  terms: [alpha beta]\n" +
			"  term_exemptions: [filler.empty-transition]\n", []string{"inapplicable/no_eligible_window"}, nil, 0},
		{"longer window extends past term", "Alpha beta gamma.", "", "{phrases: [alpha beta, alpha beta delta]}",
			"vocabulary:\n  terms: [alpha beta]\n  term_exemptions: [filler.empty-transition]\n",
			[]string{""}, []float64{0}, 0},
		{"eligible sentence after exemption", "Alpha beta. Gamma delta.", "", "",
			"vocabulary:\n  terms: [alpha beta]\n  term_exemptions: [filler.empty-transition]\n",
			[]string{""}, []float64{0}, 0},
		{"opening position only", "Gamma alpha beta. Gamma alpha beta.", "", "", "", []string{""}, []float64{0}, 0},
		{"metaphor inside sentence", "Gamma alpha beta. Gamma alpha beta.", "hype.metaphor-cluster", "", "",
			[]string{""}, []float64{0.333}, 1},
		{"heading breaks transition window", "Alpha beta.\n\n# Setup\n\nAlpha beta.", "", "", "",
			[]string{"", "inapplicable/unsupported_unit", ""}, []float64{0, 0}, 0},
		{"section window bridges heading", "Alpha beta.\n\n# Setup\n\nAlpha beta.", "filler.section-announcement", "", "",
			[]string{"", "inapplicable/unsupported_unit", ""}, []float64{0.333, 0.333}, 1},
		{"protected sentence breaks cluster", "Alpha beta.\n\nUse `code`.\n\nAlpha beta.", "", "", "",
			[]string{"", "inapplicable/no_eligible_window", ""}, []float64{0, 0}, 0},
		{"excluded fence breaks cluster", "Alpha beta.\n\n```text\nAlpha beta.\n```\n\nAlpha beta.", "", "", "",
			[]string{"", ""}, []float64{0, 0}, 0},
		{"list is unsupported", "- Alpha beta.", "", "", "", []string{"inapplicable/unsupported_unit"}, nil, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			id, parameters := test.id, test.parameters
			if id == "" {
				id = "filler.empty-transition"
			}
			if parameters == "" {
				parameters = "{phrases: [alpha beta]}"
			}
			options := unswell.Options{Features: []string{"activation/" + id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ":\n    enabled: true\n    parameters: " +
					parameters + "\n" + test.extra)}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(test.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, test.findings)
			assertPhraseMeasurements(t, result, test.reasons, test.numbers)
			options.Features = nil
			ordinary, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			without, err := ordinary.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			result.Features = nil
			c.Assert(result, qt.DeepEquals, without)
		})
	}
}
