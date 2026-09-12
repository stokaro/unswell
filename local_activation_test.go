package unswell_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
)

func localActivationIDs() []string {
	return []string{"syntax.long-sentence", "hype.modifier-cluster", "density.connective-overuse",
		"syntax.nominalization-chain", "syntax.noun-stack", "syntax.parenthetical-load", "format.em-dash-density"}
}

func TestLocalActivationsPreserveCatalogExamples(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(localActivationIDs(), d.ID) {
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

func compareLocalExample(t *testing.T, id string, example rule.Example) {
	t.Helper()
	c := qt.New(t)
	policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ": {enabled: true}\n")
	options := unswell.Options{Config: policy, Features: []string{"activation/" + id}}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	format := example.Format
	if format == "" {
		format = document.Plain
	}
	source := document.Source{Name: "example.txt", Format: format, Bytes: []byte(example.Text)}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(len(result.Findings) > 0, qt.Equals, example.Match, qt.Commentf("%s: %s", id, example.Text))
	positive := false
	for _, unit := range result.Features.Sources[0].Units {
		value := unit.Values[0]
		c.Assert(value.Reason, qt.Not(qt.Equals), "applicability_unknown")
		positive = positive || value.Number != nil && *value.Number > 0
	}
	c.Assert(positive, qt.Equals, example.Match)
	options.Features = nil
	ordinary, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	want, err := ordinary.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	result.Features = nil
	c.Assert(result, qt.DeepEquals, want)
}

func localActivation(t *testing.T, id, text, parameters, extra string) feature.Value {
	t.Helper()
	c := qt.New(t)
	settings := "{enabled: true}"
	if parameters != "" {
		settings = "{enabled: true, parameters: " + parameters + "}"
	}
	engine, err := unswell.New(unswell.Options{AllowEmpty: true, Features: []string{"activation/" + id}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nrules:\n  " + id + ": " + settings + "\n" + extra)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 0)
	c.Assert(result.Features.Sources, qt.HasLen, 1)
	c.Assert(result.Features.Sources[0].Units, qt.HasLen, 1)
	return result.Features.Sources[0].Units[0].Values[0]
}

func TestLocalActivationsMeasureZero(t *testing.T) {
	for _, id := range localActivationIDs() {
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			text := "The client sends a request to the server and waits for a reply. " +
				"The server checks the request and sends a reply to the client. " +
				"The client reads the reply and then closes the connection to the server. The server frees its memory."
			value := localActivation(t, id, text, "", "")
			c.Assert(value.Reason, qt.Equals, "")
			c.Assert(value.Number, qt.IsNotNil)
			c.Assert(*value.Number, qt.Equals, float64(0))
		})
	}
}

func TestLocalActivationsRetainApplicabilityReasons(t *testing.T) {
	for _, test := range []struct{ id, text, parameters, extra, reason string }{
		{"syntax.long-sentence", ".", "", "", "no_prose_words"},
		{"hype.modifier-cluster", "The robust estimator tolerates outliers.", "{phrases: []}", "", "no_patterns"},
		{"hype.modifier-cluster", "robust", "", "vocabulary:\n  terms: [robust]\n  term_exemptions: [hype.modifier-cluster]\n",
			"no_eligible_tokens"},
		{"hype.modifier-cluster", "`robust`.", "", "", "no_eligible_tokens"},
		{"syntax.nominalization-chain", "We perform an evaluation of the implementation.", "{verbs: []}", "", "no_patterns"},
		{"syntax.nominalization-chain", "We perform an evaluation of the implementation.", "{nouns: []}", "", "no_patterns"},
		{"syntax.nominalization-chain", "# The client starts\n", "", "", "unsupported_unit"},
		{"syntax.nominalization-chain", "evaluation", "",
			"vocabulary:\n  terms: [evaluation]\n  term_exemptions: [syntax.nominalization-chain]\n", "no_eligible_tokens"},
		{"syntax.noun-stack", "# The client starts\n", "", "", "unsupported_unit"},
		{"syntax.noun-stack", "TransportCacheEntry", "", "", "no_eligible_tokens"},
		{"syntax.noun-stack", "request", "",
			"vocabulary:\n  terms: [request]\n  term_exemptions: [syntax.noun-stack]\n", "no_eligible_tokens"},
		{"density.connective-overuse", "The client starts.", "", "", "insufficient_words"},
		{"density.connective-overuse", "The client starts.", "{min_words: 0, phrases: []}", "", "no_patterns"},
		{"density.connective-overuse", "Moreover, the client starts.", "{min_words: 0}",
			"vocabulary:\n  terms: [moreover]\n  term_exemptions: [density.connective-overuse]\n", "no_eligible_tokens"},
		{"syntax.parenthetical-load", "# The client starts\n", "", "", "unsupported_unit"},
		{"syntax.parenthetical-load", "The client starts.", "", "", "insufficient_words"},
		{"syntax.parenthetical-load", ".", "{min_words: 0}", "", "no_prose_words"},
		{"syntax.parenthetical-load", "client", "{min_words: 0}",
			"vocabulary:\n  terms: [client]\n  term_exemptions: [syntax.parenthetical-load]\n", "no_eligible_tokens"},
		{"readability.grade-metric", "The TransportCacheEntry keeps config_v2 values.", "{min_words: 4}", "", "insufficient_words"},
		{"readability.grade-metric", "TransportCacheEntry `config` v2", "{min_words: 1}", "", "no_prose_words"},
		{"format.em-dash-density", "# The client starts\n", "", "", "unsupported_unit"},
		{"format.em-dash-density", "The client starts.", "", "", "insufficient_words"},
		{"format.em-dash-density", "—", "{min_words: 0}", "", "no_prose_words"},
	} {
		t.Run(test.id+"/"+test.reason, func(t *testing.T) {
			c := qt.New(t)
			value := localActivation(t, test.id, test.text, test.parameters, test.extra)
			c.Assert(value.Number, qt.IsNil)
			c.Assert(value.Reason, qt.Equals, "inapplicable/"+test.reason)
		})
	}
}

func TestNounActivationRejectsMalformedChunks(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Features: []string{"activation/syntax.noun-stack"},
		NLP: malformedChunkNLP{Provider: provider}, NoGate: true, Config: []byte(
			"version: 1\nextends: [builtin:custom]\nrules:\n  syntax.noun-stack: {enabled: true}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(nounProse)})
	c.Assert(err, qt.ErrorMatches, ".*invalid NP chunk token range.*")
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	value := result.Features.Sources[0].Units[0].Values[0]
	c.Assert(value.Number, qt.IsNil)
	c.Assert(value.Reason, qt.Equals, "evaluation_failed")
}
