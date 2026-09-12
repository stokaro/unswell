package unswell_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestNonLatinExclusionsAcrossFeatureConsumers(t *testing.T) {
	for _, kind := range []string{"markdown", "comment", "string"} {
		for position := range 3 {
			t.Run(fmt.Sprintf("%s/%d", kind, position), func(t *testing.T) {
				for _, mode := range []string{"readability.long-paragraph", "readability.grade-metric", "features", "prepared", "combined"} {
					t.Run(mode, func(t *testing.T) {
						checkNonLatinFeatures(t, kind, position, mode)
					})
				}
			})
		}
	}
}

func checkNonLatinFeatures(t *testing.T, kind string, position int, mode string) {
	t.Helper()
	c := qt.New(t)
	options := exclusionFeatureOptions(mode)
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	source := mixedFeatureSource(kind, position)
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Errors, qt.HasLen, 0)
	c.Assert(ruleIDs(result), qt.DeepEquals, []string{"filler.announced-importance"})
	c.Assert(result.Documents[0].Blocks, qt.Equals, 3)
	c.Assert(exclusionReasons(result), qt.DeepEquals, []string{"non-latin-prose"})
	wantIDs := slices.Delete([]int{0, 1, 2}, position, position+1)
	c.Assert(paragraphIDs(result), qt.DeepEquals, wantIDs)
	if result.Features != nil {
		units := result.Features.Sources[0].Units
		c.Assert(units, qt.HasLen, 3)
		assertExcludedFeatures(t, units[position])
		c.Assert(units[position].Span, qt.Equals, result.Documents[0].Excluded[0].Span)
		for _, id := range wantIDs {
			c.Assert(units[id].Excluded, qt.IsFalse)
		}
	}
	if result.PreparedFeatures != nil {
		units := result.PreparedFeatures.Sources[0].Units
		c.Assert(len(units) > 0, qt.IsTrue)
		for _, unit := range units {
			c.Assert(wantIDs, qt.Contains, unit.Binding.BlockID)
		}
	}
}

func exclusionFeatureOptions(mode string) unswell.Options {
	options := unswell.Options{Config: []byte("version: 1\nextends: [builtin:technical]\nrules:\n" +
		"  filler.announced-importance: {gate: forbid}\n")}
	if strings.HasPrefix(mode, "readability.") {
		options.Config = append(options.Config, []byte("  "+mode+": {enabled: true}\n")...)
	}
	if mode == "combined" {
		options.Config = append(options.Config, []byte("  readability.long-paragraph: {enabled: true}\n"+
			"  readability.grade-metric: {enabled: true}\n")...)
	}
	if mode == "features" || mode == "combined" {
		options.Features = []string{"prose-words", "type-token-ratio", "activation/filler.announced-importance",
			"activation/readability.long-paragraph", "activation/readability.grade-metric"}
	}
	if mode == "prepared" || mode == "combined" {
		options.PreparedFeatures = []string{"prose-words"}
		options.PreparedKinds = []string{"sentence", "paragraph", "fragment"}
	}
	return options
}

func mixedFeatureSource(kind string, position int) document.Source {
	parts := slices.Insert([]string{"The client retries only when enabled.",
		"It is important to note that the client retries."}, position, cyrillicParagraph)
	source := document.Source{Name: "mixed.md", Format: document.Markdown}
	if kind == "markdown" {
		source.Bytes = []byte(strings.Join(parts, "\n\n") + "\n")
		return source
	}
	source.Name, source.Format = "mixed.go", document.Go
	for i, text := range parts {
		parts[i] = "// " + text
		if kind == "string" {
			parts[i] = fmt.Sprintf("const message%d = %q", i, text)
		}
	}
	source.Bytes = []byte("package sample\n\n" + strings.Join(parts, "\n\n") + "\n")
	return source
}

func assertExcludedFeatures(t *testing.T, unit unswell.FeatureUnit) {
	t.Helper()
	c := qt.New(t)
	c.Assert(unit.Excluded, qt.IsTrue)
	c.Assert(unit.Segments, qt.HasLen, 0)
	c.Assert(unit.Binding, qt.IsNotNil)
	for _, value := range unit.Values {
		reason := "excluded_unit"
		if strings.HasPrefix(value.ID, "activation/") {
			reason = "inapplicable/excluded_unit"
		}
		c.Assert(value.Number, qt.IsNil)
		c.Assert(value.Reason, qt.Equals, reason)
	}
}

func TestNonLatinFeatureCollectionRetainsEmptyScanPolicy(t *testing.T) {
	for _, allowEmpty := range []bool{false, true} {
		t.Run(fmt.Sprint(allowEmpty), func(t *testing.T) {
			c := qt.New(t)
			options := exclusionFeatureOptions("features")
			options.AllowEmpty = allowEmpty
			options.PreparedFeatures, options.PreparedKinds = []string{"prose-words"}, []string{"paragraph"}
			engine, err := unswell.New(options)
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "only.md", Format: document.Markdown,
				Bytes: []byte(cyrillicParagraph)})
			if allowEmpty {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, "scan contains no applicable English prose")
			}
			c.Assert(result.Manifest.Complete, qt.Equals, allowEmpty)
			c.Assert(result.Gate.Passed, qt.Equals, allowEmpty)
			c.Assert(result.Findings, qt.HasLen, 0)
			c.Assert(result.Documents[0].ProseWords, qt.Equals, 0)
			assertExcludedFeatures(t, result.Features.Sources[0].Units[0])
			c.Assert(result.PreparedFeatures.Sources[0].Units, qt.HasLen, 0)
		})
	}
}

func TestExcludedActivationsKeepTheirReasonAfterAbstention(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{
		Config:   []byte("version: 1\nextends: [builtin:custom]\nrules:\n  test.abstain: {enabled: true}\n"),
		Rules:    []rule.Rule{abstainingRule{reason: rule.ReasonBudgetExhausted}},
		Features: []string{"prose-words", "activation/test.abstain"},
	})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "mixed.md", Format: document.Markdown,
		Bytes: []byte(cyrillicParagraph + "\n\nThe client retries.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Abstentions, qt.HasLen, 1)
	assertExcludedFeatures(t, result.Features.Sources[0].Units[0])
	activation := result.Features.Sources[0].Units[1].Values[0]
	c.Assert(activation.Number, qt.IsNil)
	c.Assert(activation.Reason, qt.Equals, "inapplicable/budget_exhausted")
}
