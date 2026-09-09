package training_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func predictionPlan(fitted training.Artifact) training.PredictionPlan {
	threshold := 0.5
	context := "prepared_piece"
	if fitted.Identity.FeatureSource == "rule_activations" {
		context = "source_document"
	}
	return training.PredictionPlan{Version: training.PredictionVersion, ID: "simulated-test-v1", ProtocolSHA256: strings.Repeat("a", 64),
		ModelSHA256: fitted.SHA256, CorpusSHA256: fitted.CorpusSHA256, Partition: "final_test",
		Context: context, Response: "logistic", Threshold: &threshold}
}

func TestPredictRestoresFrozenPreparedModel(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	before := testfixture.Encode(t, fitted)
	result, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Rows, qt.HasLen, 1)
	row := result.Rows[0]
	c.Assert(row.SourceID, qt.Equals, "d000006")
	c.Assert(row.Status, qt.Equals, "available")
	c.Assert(row.Positive, qt.IsNotNil)
	c.Assert(row.Response, qt.DeepEquals, row.LogisticResponse)
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(testfixture.Encode(t, fitted), qt.DeepEquals, before)
	loaded, err := training.LoadPredictions(t.Context(), testfixture.Encode(t, result))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	result.Model.Logistic.Means[0] = 100
	c.Assert(fitted.Logistic.Means[0], qt.Equals, float64(6))
	c.Assert(loaded.Model.Logistic.Means[0], qt.Equals, float64(6))
}

func TestPredictDoesNotReadOrFitEvaluationLabels(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	before, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	// The prediction API cannot receive this round, including malformed labels.
	input.Round["judgments"] = json.RawMessage(`{"invalid":"evaluation labels are not prediction input"}`)
	after, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(after, qt.DeepEquals, before)
}

func TestPredictRulesUsesExactFrozenConfiguration(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	configuration, err := os.ReadFile("testdata/rules.yaml")
	c.Assert(err, qt.IsNil)
	options := fittingOptions()
	options.Features = []string{"activation/policy.banned-phrases"}
	fitted, err := training.RunRules(t.Context(), candidates, round, input.Files, options, configuration)
	c.Assert(err, qt.IsNil)
	result, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Plan.Context, qt.Equals, "source_document")
	c.Assert(result.Rows, qt.HasLen, 1)
	c.Assert(result.Rows[0].Status, qt.Equals, "available")
	_, err = training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), append(configuration, '\n'))
	c.Assert(err, qt.ErrorMatches, ".*configuration digest mismatch")
}

func TestPredictCalibrationRangeIsNotAZeroResponse(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	input.ReplaceSource(5, strings.Repeat("Reserved ", 100)+"target remains unchanged during inference.")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	plan := predictionPlan(fitted)
	plan.Response = "isotonic"
	result, err := training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Rows, qt.HasLen, 1)
	row := result.Rows[0]
	c.Assert(row.Status, qt.Equals, "inapplicable")
	c.Assert(row.Reason, qt.Equals, "calibration/out_of_range")
	c.Assert(row.Response, qt.IsNil)
	c.Assert(row.Positive, qt.IsNil)
	c.Assert(row.LinearScore, qt.IsNotNil)
	_, err = training.LoadPredictions(t.Context(), testfixture.Encode(t, result))
	c.Assert(err, qt.IsNil)
}

func TestPredictionPlanAndModelRejections(t *testing.T) {
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*training.PredictionPlan)
	}{
		{"training partition", func(p *training.PredictionPlan) { p.Partition = "training" }},
		{"calibration partition", func(p *training.PredictionPlan) { p.Partition = "calibration" }},
		{"wrong context", func(p *training.PredictionPlan) { p.Context = "source_document" }},
		{"misstated target-only NLP", func(p *training.PredictionPlan) { p.Context = "target_only" }},
		{"missing threshold", func(p *training.PredictionPlan) { p.Threshold = nil }},
		{"different model", func(p *training.PredictionPlan) { p.ModelSHA256 = strings.Repeat("b", 64) }},
		{"different corpus", func(p *training.PredictionPlan) { p.CorpusSHA256 = strings.Repeat("b", 64) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			plan := predictionPlan(fitted)
			row.edit(&plan)
			result, err := training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, training.Predictions{})
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = training.Predict(ctx, candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	fitted.Logistic.Weights[0]++
	_, err = training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorMatches, ".*digest mismatch")
}
