package probability_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/probability"
)

func TestCompatibleRequiresMatchingMeasurementInputs(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	c.Assert(pack.Compatible(matchingRun(c, file)), qt.IsNil)
	for _, row := range []struct {
		name, message string
		edit          func(*probability.Run)
	}{
		{"feature contract", ".*expects feature contract.*", func(r *probability.Run) { r.FeatureContract = "other-v1" }},
		{"unit contract", ".*expects feature contract.*", func(r *probability.Run) { r.UnitContract = "other-v1" }},
		{"preparation", ".*extraction and preparation policy.*", func(r *probability.Run) { r.PreparationHash = "" }},
		{"quotes", ".*include_quotes=false.*", func(r *probability.Run) { r.IncludeQuotes = true }},
		{"structure", ".*include_structure=false.*", func(r *probability.Run) { r.IncludeStructure = true }},
		{"provider", ".*expects NLP provider builtin-en.*", func(r *probability.Run) { r.NLP.Version = "other" }},
		{"requested capabilities", ".*expects the .* capabilities.*", func(r *probability.Run) {
			r.Capabilities = []nlp.Capability{nlp.Tokens}
		}},
		{"extra capabilities", ".*expects the .* capabilities.*", func(r *probability.Run) {
			r.Capabilities = append(r.Capabilities, nlp.POS)
		}},
		{"provider support", ".*requires the sentences capability from the provider.*", func(r *probability.Run) {
			r.NLP.Capabilities = []nlp.Capability{nlp.Tokens}
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			run := matchingRun(c, file)
			row.edit(&run)
			c.Assert(pack.Compatible(run), qt.ErrorMatches, row.message)
		})
	}
}

func TestEstimateAbstainsWithAMachineReadableStatus(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	for _, row := range []struct {
		name, status, detail string
		unit                 probability.Unit
	}{
		{"unsupported unit", probability.StatusUnsupportedUnit, "sentence",
			probability.Unit{Kind: "paragraph", Words: 40, Values: measured(c, file, number(12), number(0.6))}},
		{"short target", probability.StatusInsufficientEvidence, "min_words=5",
			probability.Unit{Kind: "sentence", Words: 4, Values: measured(c, file, number(12), number(0.6))}},
		{"missing measurement", probability.StatusMissingFeature, "type-token-ratio/no_prose_words",
			probability.Unit{Kind: "sentence", Words: 12, Values: measured(c, file, number(12), nil)}},
		{"outside calibration", probability.StatusCalibrationRange, "",
			probability.Unit{Kind: "sentence", Words: 40, Values: measured(c, file, number(100), number(0.6))}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			estimate, err := pack.Estimate(c.Context(), row.unit)
			c.Assert(err, qt.IsNil)
			c.Assert(estimate.Status, qt.Equals, row.status)
			c.Assert(estimate.Detail, qt.Equals, row.detail)
			c.Assert(estimate.Probability, qt.IsNil)
		})
	}
}

func TestEstimateReportsACalibratedValueAndItsUncalibratedScore(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	estimate, err := pack.Estimate(c.Context(),
		probability.Unit{Kind: "sentence", Words: 12, Values: measured(c, file, number(12), number(0.6))})
	c.Assert(err, qt.IsNil)
	c.Assert(estimate.Status, qt.Equals, probability.StatusAvailable)
	c.Assert(estimate.Detail, qt.Equals, "")
	c.Assert(math.Abs(*estimate.LinearScore-0.2) < 1e-12, qt.IsTrue)
	c.Assert(math.Abs(*estimate.Probability-0.5) < 1e-12, qt.IsTrue)
	outside, err := pack.Estimate(c.Context(),
		probability.Unit{Kind: "sentence", Words: 40, Values: measured(c, file, number(100), number(0.6))})
	c.Assert(err, qt.IsNil)
	c.Assert(*outside.LinearScore > 1, qt.IsTrue)
	c.Assert(outside.Probability, qt.IsNil)
}

func TestEstimateRejectsVectorsThatDoNotMatchItsColumns(t *testing.T) {
	c := qt.New(t)
	file := packFile(c, "sentence", "prose-words", "type-token-ratio")
	pack := loaded(c, file)
	complete := measured(c, file, number(12), number(0.6))
	for _, row := range []struct {
		name, message string
		values        []feature.Value
	}{
		{"short vector", ".*requires 2 measured columns.*", complete[:1]},
		{"identity", ".*expects column prose-words version 1 at position 0.*",
			[]feature.Value{{ID: "counted-characters", Version: "1", Number: number(1)}, complete[1]}},
		{"version", ".*expects column type-token-ratio version 1 at position 1.*",
			[]feature.Value{complete[0], {ID: "type-token-ratio", Version: "2", Number: number(0.6)}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			estimate, err := pack.Estimate(c.Context(), probability.Unit{Kind: "sentence", Words: 12, Values: row.values})
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(estimate, qt.DeepEquals, probability.Estimate{})
		})
	}
	ctx, cancel := context.WithCancel(c.Context())
	cancel()
	estimate, err := pack.Estimate(ctx, probability.Unit{Kind: "sentence", Words: 12, Values: complete})
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(estimate, qt.DeepEquals, probability.Estimate{})
}
