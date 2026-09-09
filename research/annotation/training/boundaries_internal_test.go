package training

// White-box tests: Inject unavailable and malformed measurements into row selection;
// Run constructs those measurements internally and does not accept a precomputed join.

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
)

func TestRunRejectsInvalidSelectionsAndLimits(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*Options)
	}{
		{"kind", func(o *Options) { o.Kind = "document" }},
		{"missing policy", func(o *Options) { o.MissingFeatures = "zero" }},
		{"calibration", func(o *Options) { o.Calibration = "invented" }},
		{"no columns", func(o *Options) { o.Features = nil }},
		{"unknown column", func(o *Options) { o.Features = []string{"unknown"} }},
		{"block activation", func(o *Options) { o.Features = []string{"activation/syntax.noun-stack"} }},
		{"duplicate column", func(o *Options) { o.Features = []string{"prose-words", "prose-words"} }},
		{"nonfinite option", func(o *Options) { o.Fit.L2 = math.NaN() }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			input := testfixture.Load(t, "testdata")
			candidates, round := input.Compile(t)
			options := fittingOptions()
			row.edit(&options)
			result, err := Run(t.Context(), candidates, round, input.Files, options)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, Artifact{})
		})
	}
}

func TestNoParentLabelsAndNoPartialBudgetResult(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := fittingOptions()
	options.Kind = "sentence"
	result, err := Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.ErrorMatches, "training partition: .*")
	c.Assert(result, qt.DeepEquals, Artifact{})
	options = fittingOptions()
	options.Fit.MaxOperations = 1
	result, err = Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.ErrorIs, model.ErrBudget)
	c.Assert(result, qt.DeepEquals, Artifact{})
}

func TestSelectionMakesMissingFeaturesExplicit(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := fittingOptions()
	joined, err := corpus.Join(t.Context(), candidates, round, input.Files, options.Features)
	c.Assert(err, qt.IsNil)
	for i := range joined.Features.Sources[0].Units {
		unit := &joined.Features.Sources[0].Units[i]
		if unit.Binding.Kind == "paragraph" {
			unit.Values[0].Number = nil
			unit.Values[0].Reason = "insufficient_evidence"
		}
	}
	_, err = selectRows(t.Context(), candidates.Plan, joined, options)
	c.Assert(err, qt.ErrorMatches, ".*unavailable feature prose-words/insufficient_evidence.*")
	options.MissingFeatures = "exclude"
	selected, err := selectRows(t.Context(), candidates.Plan, joined, options)
	c.Assert(err, qt.IsNil)
	c.Assert(selected.training, qt.HasLen, 1)
	c.Assert(selected.partitions[0].Excluded["feature/prose-words/insufficient_evidence"], qt.Equals, 1)
	c.Assert(selected.training[0].Values, qt.DeepEquals, []float64{9})
}

func TestNumericContractRejectsInvalidAndPreservesZero(t *testing.T) {
	c := qt.New(t)
	columns, err := feature.UnitCatalog("paragraph")
	c.Assert(err, qt.IsNil)
	columns = columns[:1]
	zero, nonfinite := 0.0, math.Inf(1)
	for _, row := range []struct {
		name  string
		value feature.Value
	}{
		{"wrong ID", feature.Value{ID: "other", Version: "1", Unit: "words", Number: &zero}},
		{"wrong version", feature.Value{ID: "prose-words", Version: "2", Unit: "words", Number: &zero}},
		{"nonfinite", feature.Value{ID: "prose-words", Version: "1", Unit: "words", Number: &nonfinite}},
		{"value and reason", feature.Value{ID: "prose-words", Version: "1", Unit: "words", Number: &zero, Reason: "missing"}},
		{"unexplained missing", feature.Value{ID: "prose-words", Version: "1", Unit: "words"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, _, err := numericValues([]feature.Value{row.value}, columns)
			c.Assert(err, qt.IsNotNil)
		})
	}
	values, reason, err := numericValues([]feature.Value{{ID: "prose-words", Version: "1", Unit: "words", Number: &zero}}, columns)
	c.Assert(err, qt.IsNil)
	c.Assert(reason, qt.Equals, "")
	c.Assert(values, qt.DeepEquals, []float64{0})
}
