package training

// White-box tests: Rehash corrupted models and predictions to reach semantic loader checks;
// public artifact producers cannot create these invalid states with matching digests.

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
)

func TestLoadedModelRejectsInvalidRehashedParameters(t *testing.T) {
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	c := qt.New(t)
	original, err := Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*Artifact)
	}{
		{"scale", func(a *Artifact) { a.Logistic.Scales[0] = 0 }},
		{"dimensions", func(a *Artifact) { a.Logistic.Weights = nil }},
		{"unknown feature", func(a *Artifact) { a.Options.Features[0] = "unknown" }},
		{"columns", func(a *Artifact) { a.Identity.Columns[0].Version = "unknown" }},
		{"kind", func(a *Artifact) { a.Options.Kind = "document" }},
		{"missing mode", func(a *Artifact) { a.Options.MissingFeatures = "zero" }},
		{"score kind", func(a *Artifact) { a.Calibration.ScoreKind = "editorial_index" }},
		{"knots", func(a *Artifact) { a.Calibration.Scores[1] = a.Calibration.Scores[0] }},
		{"qualification", func(a *Artifact) { a.ProbabilityStatus = "available" }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed, err := Load(t.Context(), testfixture.Encode(t, original))
			c.Assert(err, qt.IsNil)
			row.edit(&changed)
			changed.SHA256 = ""
			changed, err = finish(t.Context(), changed)
			c.Assert(err, qt.IsNil)
			result, err := Load(t.Context(), testfixture.Encode(t, changed))
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, Artifact{})
		})
	}
}

func TestPredictPartialTargetsRemainUnavailable(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	input.ReplaceSource(5, "Keep the first sentence. Preserve the second sentence.")
	// Reuse simulated labels for the same single-sentence training/calibration targets.
	var judgments []annotation.Judgment
	c.Assert(json.Unmarshal(input.Round["judgments"], &judgments), qt.IsNil)
	ids := map[string]string{"u000002": "u000001", "u000004": "u000003", "u000006": "u000005", "u000008": "u000007"}
	for i := range judgments {
		if id, exists := ids[judgments[i].UnitID]; exists {
			judgments[i].UnitID = id
		}
	}
	input.Round["judgments"] = testfixture.Encode(t, judgments)
	candidates, round := input.Compile(t)
	configuration, err := os.ReadFile("testdata/rules.yaml")
	c.Assert(err, qt.IsNil)
	options := fittingOptions()
	options.Kind, options.Features = "sentence", []string{"activation/policy.banned-phrases"}
	fitted, err := RunRules(t.Context(), candidates, round, input.Files, options, configuration)
	c.Assert(err, qt.IsNil)
	result, err := Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Rows, qt.HasLen, 2)
	for _, row := range result.Rows {
		c.Assert(row.Status, qt.Equals, "inapplicable")
		c.Assert(row.Reason, qt.Equals, "target/no_complete_block_match")
		c.Assert(row.FeatureInputHash, qt.Equals, "")
		c.Assert(row.Response, qt.IsNil)
		c.Assert(row.LinearScore, qt.IsNil)
	}
	_, err = LoadPredictions(t.Context(), testfixture.Encode(t, result))
	c.Assert(err, qt.IsNil)
}

func TestPredictRequiresEvaluationPermissionAndOriginalSources(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	input.Manifest.Sources[5].Rights.AllowedUses = []string{"annotation"}
	candidates, round := input.Compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	result, err := Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorMatches, ".*evaluation permission")
	c.Assert(result, qt.DeepEquals, Predictions{})
	input.Files["d000006.txt"][0]++
	result, err = Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, Predictions{})
}

func TestPredictionLoaderRejectsInconsistentRows(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	original, err := Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*Predictions)
	}{
		{"duplicate", func(p *Predictions) { p.Rows = append(p.Rows, p.Rows[0]) }},
		{"null response", func(p *Predictions) { p.Rows[0].Response = nil }},
		{"decision", func(p *Predictions) { *p.Rows[0].Positive = !*p.Rows[0].Positive }},
		{"fake absence", func(p *Predictions) { p.Rows[0].Status = "inapplicable" }},
		{"plan hash", func(p *Predictions) { p.Plan.ID = "changed-trial" }},
		{"verification", func(p *Predictions) { p.Verification.ArtifactSHA256 = strings.Repeat("b", 64) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed, err := LoadPredictions(t.Context(), testfixture.Encode(t, original))
			c.Assert(err, qt.IsNil)
			row.edit(&changed)
			changed.SHA256 = ""
			changed, err = finishPredictions(t.Context(), changed)
			c.Assert(err, qt.IsNil)
			_, err = LoadPredictions(t.Context(), testfixture.Encode(t, changed))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestResearchLoadersRejectAmbiguousJSON(t *testing.T) {
	for _, data := range []string{`{"version":1,"version":2}`, `{"unknown":true}`, `{} {}`, `{"id":"\ud800"}`} {
		t.Run(data, func(t *testing.T) {
			c := qt.New(t)
			_, err := Load(t.Context(), []byte(data))
			c.Assert(err, qt.IsNotNil)
			_, err = LoadPredictionPlan(t.Context(), []byte(data))
			c.Assert(err, qt.IsNotNil)
			_, err = LoadPredictions(t.Context(), []byte(data))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

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
