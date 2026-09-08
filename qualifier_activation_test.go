package unswell_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
)

func qualifierActivationIDs() []string {
	return []string{"hype.vague-praise", "hype.absolute-claim", "filler.weak-intensifiers", "filler.stacked-hedging"}
}

func TestQualifierActivationsPreserveCatalogExamples(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(qualifierActivationIDs(), d.ID) {
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

func TestQualifierActivationsDistinguishMissingFromZero(t *testing.T) {
	for _, test := range []struct{ id, text, parameters, extra, reason string }{
		{"hype.vague-praise", "The client sends requests.", "", "", ""},
		{"hype.vague-praise", "Alpha", "{phrases: [alpha beta]}", "", "no_eligible_window"},
		{"hype.vague-praise", "Alpha `code` beta", "{phrases: [alpha beta]}", "", "no_eligible_window"},
		{"hype.vague-praise", "Alpha beta", "{phrases: [alpha beta]}",
			"vocabulary:\n  terms: [alpha beta]\n  term_exemptions: [hype.vague-praise]\n", "no_eligible_window"},
		{"hype.vague-praise", "Gamma delta. Alpha `code` beta.", "{phrases: [alpha beta]}", "", ""},
		{"hype.vague-praise", "Alpha beta.", "{phrases: []}", "", "no_patterns"},
		{"hype.vague-praise", "Is this a game-changing solution?", "", "", "no_eligible_window"},
		{"hype.vague-praise", "This is not a game-changing solution.", "", "", "no_eligible_window"},
		{"hype.absolute-claim", "Under these conditions, the service guarantees complete safety.", "", "", "no_eligible_window"},
		{"hype.absolute-claim", "The client sends requests.", "", "", ""},
		{"hype.absolute-claim", "# Complete safety\n", "", "", "unsupported_unit"},
		{"filler.weak-intensifiers", "The client starts.", "", "", "insufficient_words"},
		{"filler.weak-intensifiers", "The client", "{min_words: 0}", "", ""},
		{"filler.weak-intensifiers", "very", "{min_words: 0}", "", "no_eligible_window"},
		{"filler.weak-intensifiers", "very first", "{min_words: 0}", "", "no_eligible_window"},
		{"filler.weak-intensifiers", "very `powerful`", "{min_words: 0}", "", "no_eligible_window"},
		{"filler.weak-intensifiers", "very powerful", "{min_words: 0}",
			"vocabulary:\n  terms: [very]\n  term_exemptions: [filler.weak-intensifiers]\n", "no_eligible_window"},
		{"filler.weak-intensifiers", "very powerful", "{min_words: 0, phrases: []}", "", "no_patterns"},
		{"filler.weak-intensifiers", ".", "{min_words: 0}", "", "no_prose_words"},
		{"filler.weak-intensifiers", "very powerful", "{min_words: 0, onset: 100, saturation: 200}", "", ""},
		{"filler.stacked-hedging", "The client starts.", "", "", ""},
		{"filler.stacked-hedging", "may", "", "", ""},
		{"filler.stacked-hedging", "`may possibly`.", "", "", "no_eligible_tokens"},
		{"filler.stacked-hedging", "if and unless", "", "", "no_eligible_tokens"},
		{"filler.stacked-hedging", "may possibly perhaps", "",
			"vocabulary:\n  terms: [may possibly perhaps]\n  term_exemptions: [filler.stacked-hedging]\n", "no_eligible_tokens"},
		{"filler.stacked-hedging", "The client starts.", "{phrases: []}", "", "no_patterns"},
		{"filler.stacked-hedging", "# The client starts\n", "", "", "unsupported_unit"},
	} {
		t.Run(test.id+"/"+test.text, func(t *testing.T) {
			c := qt.New(t)
			value := localActivation(t, test.id, test.text, test.parameters, test.extra)
			if test.reason != "" {
				c.Assert(value.Number, qt.IsNil)
				c.Assert(value.Reason, qt.Equals, "inapplicable/"+test.reason)
			} else {
				c.Assert(value.Reason, qt.Equals, "")
				c.Assert(value.Number, qt.IsNotNil)
				c.Assert(*value.Number, qt.Equals, float64(0))
			}
		})
	}
}

func TestQualifierActivationsPreservePhrasesBesideProtectedCode(t *testing.T) {
	for _, id := range []string{"hype.vague-praise", "hype.absolute-claim"} {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			options := unswell.Options{Features: []string{"activation/" + id}, Config: []byte(
				"version: 1\nextends: [builtin:custom]\nrules:\n  " + id +
					": {enabled: true, parameters: {phrases: [alpha beta]}}\n")}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Use `code` with alpha **beta**.")}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 1)
			assertPhraseMeasurements(t, result, []string{""}, []float64{1})
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
