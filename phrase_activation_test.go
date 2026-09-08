package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestPhraseActivationSeparatesComparisonsFromSkippedWindows(t *testing.T) {
	for _, test := range []struct {
		name, text, parameters, vocabulary string
		reasons                            []string
		numbers                            []float64
		findings                           int
	}{
		{"positive and negative", "Alpha beta.\n\nGamma delta.", "{phrases: [alpha beta]}", "",
			[]string{"", ""}, []float64{1, 0}, 1},
		{"too short", "Alpha.", "{phrases: [alpha beta gamma]}", "",
			[]string{"inapplicable/no_eligible_window"}, nil, 0},
		{"empty dictionary", "Alpha beta.", "{phrases: []}", "",
			[]string{"inapplicable/no_patterns"}, nil, 0},
		{"document start includes heading", "# Alpha beta\n\nAlpha beta.\n\nGamma delta.",
			"{phrases: [alpha beta], positions: [document-start]}", "",
			[]string{"", "inapplicable/no_eligible_window", "inapplicable/no_eligible_window"}, []float64{1}, 1},
		{"document end includes list item", "# Alpha beta\n\nGamma delta.\n\n- Alpha beta.",
			"{phrases: [alpha beta], positions: [document-end]}", "",
			[]string{"inapplicable/no_eligible_window", "inapplicable/no_eligible_window", ""}, []float64{1}, 1},
		{"protected window", "Alpha `code` beta.", "{phrases: [alpha beta], positions: [sentence-start]}", "",
			[]string{"inapplicable/no_eligible_window"}, nil, 0},
		{"term exemption", "Alpha beta.", "{phrases: [alpha beta], positions: [sentence-start]}",
			"vocabulary:\n  terms: [alpha beta]\n  term_exemptions: [policy.banned-phrases]\n",
			[]string{"inapplicable/no_eligible_window"}, nil, 0},
		{"eligible sentence after exemption", "Alpha beta. Gamma delta.",
			"{phrases: [alpha beta], positions: [sentence-start]}",
			"vocabulary:\n  terms: [alpha beta]\n  term_exemptions: [policy.banned-phrases]\n",
			[]string{""}, []float64{0}, 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			policy := []byte("version: 1\nextends: [builtin:custom]\nrules:\n" +
				"  policy.banned-phrases:\n    enabled: true\n    parameters: " + test.parameters + "\n" + test.vocabulary)
			options := unswell.Options{Config: policy, Features: []string{"activation/policy.banned-phrases"}}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(test.text)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Manifest.Complete, qt.IsTrue)
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

func assertPhraseMeasurements(t *testing.T, result unswell.RunResult, reasons []string, numbers []float64) {
	t.Helper()
	c := qt.New(t)
	units := result.Features.Sources[0].Units
	c.Assert(units, qt.HasLen, len(reasons))
	var got []float64
	for i, reason := range reasons {
		value := units[i].Values[0]
		c.Assert(value.Reason, qt.Equals, reason)
		if reason != "" {
			c.Assert(value.Number, qt.IsNil)
		} else {
			c.Assert(value.Number, qt.IsNotNil)
			got = append(got, *value.Number)
		}
	}
	c.Assert(got, qt.DeepEquals, numbers)
}

func TestPhraseActivationUsesPerFilePolicy(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Jobs: 4, Features: []string{"activation/policy.banned-phrases"}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nrules:\n  policy.banned-phrases:\n" +
			"    enabled: true\n    parameters: {phrases: [alpha beta]}\noverrides:\n" +
			"  - files: [empty.md]\n    rules:\n      policy.banned-phrases:\n        parameters: {phrases: []}\n" +
			"  - files: [disabled.md]\n    rules: {policy.banned-phrases: {enabled: false}}\n")})
	c.Assert(err, qt.IsNil)
	var sources []document.Source
	for _, name := range []string{"positive.md", "empty.md", "disabled.md"} {
		sources = append(sources, document.Source{Name: name, Format: document.Markdown, Bytes: []byte("Alpha beta.")})
	}
	result, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Path, qt.Equals, "positive.md")
	c.Assert(result.Features.Sources, qt.HasLen, 3)
	for _, source := range result.Features.Sources {
		value := source.Units[0].Values[0]
		switch source.Path {
		case "disabled.md":
			c.Assert(value.Number, qt.IsNil)
			c.Assert(value.Reason, qt.Equals, "disabled")
		case "empty.md":
			c.Assert(value.Number, qt.IsNil)
			c.Assert(value.Reason, qt.Equals, "inapplicable/no_patterns")
		case "positive.md":
			c.Assert(value.Number, qt.IsNotNil)
			c.Assert(*value.Number, qt.Equals, float64(1))
			c.Assert(value.Reason, qt.Equals, "")
		default:
			t.Fatalf("unexpected feature source %q", source.Path)
		}
	}
}
