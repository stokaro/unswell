package training

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func predictionPlan(fitted Artifact) PredictionPlan {
	threshold := 0.5
	context := "prepared_piece"
	if fitted.Identity.FeatureSource == "rule_activations" {
		context = "source_document"
	}
	return PredictionPlan{Version: PredictionVersion, ID: "simulated-test-v1", ProtocolSHA256: strings.Repeat("a", 64),
		ModelSHA256: fitted.SHA256, CorpusSHA256: fitted.CorpusSHA256, Partition: "final_test",
		Context: context, Response: "logistic", Threshold: &threshold}
}

func TestPredictRestoresFrozenPreparedModel(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	candidates, round := input.compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.files, fittingOptions())
	c.Assert(err, qt.IsNil)
	before := encode(t, fitted)
	result, err := Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Rows, qt.HasLen, 1)
	row := result.Rows[0]
	c.Assert(row.SourceID, qt.Equals, "d000006")
	c.Assert(row.Status, qt.Equals, "available")
	c.Assert(row.Positive, qt.IsNotNil)
	c.Assert(row.Response, qt.DeepEquals, row.LogisticResponse)
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(encode(t, fitted), qt.DeepEquals, before)
	loaded, err := LoadPredictions(t.Context(), encode(t, result))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, result)
	result.Model.Logistic.Means[0] = 100
	c.Assert(fitted.Logistic.Means[0], qt.Equals, float64(6))
	c.Assert(loaded.Model.Logistic.Means[0], qt.Equals, float64(6))
}

func TestPredictDoesNotReadOrFitEvaluationLabels(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	candidates, round := input.compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.files, fittingOptions())
	c.Assert(err, qt.IsNil)
	before, err := Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	// The prediction API cannot receive this round, including malformed labels.
	input.round["judgments"] = json.RawMessage(`{"invalid":"evaluation labels are not prediction input"}`)
	after, err := Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(after, qt.DeepEquals, before)
}

func TestPredictRulesUsesExactFrozenConfiguration(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	candidates, round := input.compile(t)
	configuration, err := os.ReadFile("testdata/rules.yaml")
	c.Assert(err, qt.IsNil)
	options := fittingOptions()
	options.Features = []string{"activation/policy.banned-phrases"}
	fitted, err := RunRules(t.Context(), candidates, round, input.files, options, configuration)
	c.Assert(err, qt.IsNil)
	result, err := Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Plan.Context, qt.Equals, "source_document")
	c.Assert(result.Rows, qt.HasLen, 1)
	c.Assert(result.Rows[0].Status, qt.Equals, "available")
	_, err = Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), append(configuration, '\n'))
	c.Assert(err, qt.ErrorMatches, ".*configuration digest mismatch")
}

func TestPredictCalibrationRangeIsNotAZeroResponse(t *testing.T) {
	c := qt.New(t)
	input := fixtureInput(t)
	input.replaceSource(5, strings.Repeat("Reserved ", 100)+"target remains unchanged during inference.")
	candidates, round := input.compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.files, fittingOptions())
	c.Assert(err, qt.IsNil)
	plan := predictionPlan(fitted)
	plan.Response = "isotonic"
	result, err := Predict(t.Context(), candidates, input.files, fitted, plan, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Rows, qt.HasLen, 1)
	row := result.Rows[0]
	c.Assert(row.Status, qt.Equals, "inapplicable")
	c.Assert(row.Reason, qt.Equals, "calibration/out_of_range")
	c.Assert(row.Response, qt.IsNil)
	c.Assert(row.Positive, qt.IsNil)
	c.Assert(row.LinearScore, qt.IsNotNil)
	_, err = LoadPredictions(t.Context(), encode(t, result))
	c.Assert(err, qt.IsNil)
}

func TestPredictionPlanAndModelRejections(t *testing.T) {
	input := fixtureInput(t)
	candidates, round := input.compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.files, fittingOptions())
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*PredictionPlan)
	}{
		{"training partition", func(p *PredictionPlan) { p.Partition = "training" }},
		{"calibration partition", func(p *PredictionPlan) { p.Partition = "calibration" }},
		{"wrong context", func(p *PredictionPlan) { p.Context = "source_document" }},
		{"misstated target-only NLP", func(p *PredictionPlan) { p.Context = "target_only" }},
		{"missing threshold", func(p *PredictionPlan) { p.Threshold = nil }},
		{"different model", func(p *PredictionPlan) { p.ModelSHA256 = strings.Repeat("b", 64) }},
		{"different corpus", func(p *PredictionPlan) { p.CorpusSHA256 = strings.Repeat("b", 64) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			plan := predictionPlan(fitted)
			row.edit(&plan)
			result, err := Predict(t.Context(), candidates, input.files, fitted, plan, nil)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, Predictions{})
		})
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = Predict(ctx, candidates, input.files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	fitted.Logistic.Weights[0]++
	_, err = Predict(t.Context(), candidates, input.files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorMatches, ".*digest mismatch")
}
