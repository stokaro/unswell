package training_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func TestForestArtifactRejectsInconsistentMetadata(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, forestFitting())
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*training.Artifact)
	}{
		{"old version", func(a *training.Artifact) { a.Version = "unswell-editorial-training-v2" }},
		{"estimator", func(a *training.Artifact) { a.Options.Estimator = "logistic" }},
		{"extra options", func(a *training.Artifact) { a.Options.Fit = fittingOptions().Fit }},
		{"extra model", func(a *training.Artifact) { a.Logistic = &training.Logistic{} }},
		{"missing model", func(a *training.Artifact) { a.Forest = nil }},
		{"missing options", func(a *training.Artifact) { a.Options.Forest = nil }},
		{"feature width", func(a *training.Artifact) { a.Forest.Parameters.Features++ }},
		{"bad tree", func(a *training.Artifact) { a.Forest.Parameters.Trees[0][0].Right = 0 }},
		{"root samples", func(a *training.Artifact) { a.Forest.Parameters.Trees[0] = a.Forest.Parameters.Trees[0][1:2] }},
		{"tree count", func(a *training.Artifact) { a.Options.Forest.Trees++ }},
		{"node count", func(a *training.Artifact) { a.Forest.Nodes++ }},
		{"node budget", func(a *training.Artifact) { a.Options.Forest.MaxNodes = 3 }},
		{"operation budget", func(a *training.Artifact) { a.Forest.Operations = a.Options.Forest.MaxOperations + 1 }},
		{"child size", func(a *training.Artifact) { a.Options.Forest.MinLeaf = 2 }},
		{"column subset", func(a *training.Artifact) { a.Options.Forest.FeaturesPerSplit = 2 }},
		{"calibration channel", func(a *training.Artifact) { a.Calibration.ScoreKind = "linear_score" }},
		{"calibration domain", func(a *training.Artifact) { a.Calibration.Scores[0] = -1 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var changed training.Artifact
			c.Assert(json.Unmarshal(testfixture.Encode(t, fitted), &changed), qt.IsNil)
			row.edit(&changed)
			changed.SHA256 = ""
			changed.SHA256 = fmt.Sprintf("%x", sha256.Sum256(testfixture.Encode(t, changed)))
			result, err := training.Load(t.Context(), testfixture.Encode(t, changed))
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, training.Artifact{})
		})
	}
}

func TestForestPredictionsRejectMixedChannels(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, forestFitting())
	c.Assert(err, qt.IsNil)
	plan := predictionPlan(fitted)
	_, err = training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
	c.Assert(err, qt.ErrorMatches, ".*response does not match.*")
	plan.Response = "forest"
	result, err := training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
	c.Assert(err, qt.IsNil)
	for _, row := range []struct {
		name string
		edit func(*training.Prediction)
	}{
		{"wrong value", func(r *training.Prediction) { value := 0.25; r.ForestResponse = &value }},
		{"mixed channels", func(r *training.Prediction) { r.LogisticResponse = r.ForestResponse }},
		{"missing raw value", func(r *training.Prediction) { r.ForestResponse = nil }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var changed training.Predictions
			c.Assert(json.Unmarshal(testfixture.Encode(t, result), &changed), qt.IsNil)
			row.edit(&changed.Rows[0])
			changed.SHA256 = ""
			changed.SHA256 = fmt.Sprintf("%x", sha256.Sum256(testfixture.Encode(t, changed)))
			_, err := training.LoadPredictions(t.Context(), testfixture.Encode(t, changed))
			c.Assert(err, qt.IsNotNil)
		})
	}
}
