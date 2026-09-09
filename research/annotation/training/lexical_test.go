package training_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func lexicalOptions() training.LexicalOptions {
	return training.LexicalOptions{Counts: feature.LexicalOptions{WordMin: 1, WordMax: 2, CharMin: 3, CharMax: 4},
		MaxFeatures: 64, MinTargets: 1}
}

func lexicalFitting() training.Options {
	o := fittingOptions()
	o.Features = nil
	return o
}

func TestLexicalVocabularyTrainingPredictionAndRestoration(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Version, qt.Equals, training.Version)
	c.Assert(fitted.Lexical, qt.IsNotNil)
	c.Assert(fitted.Lexical.Terms, qt.HasLen, 64)
	c.Assert(fitted.Lexical.TrainingTargets, qt.Equals, 2)
	c.Assert(fitted.Identity.LexicalVocabularyHash, qt.Equals, fitted.Lexical.SHA256)
	c.Assert(fitted.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(fitted.Partitions[0].Rows, qt.HasLen, 2)
	c.Assert(fitted.Partitions[2].Rows, qt.HasLen, 2)
	loaded, err := training.Load(t.Context(), testfixture.Encode(t, fitted))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, fitted)
	predicted, err := training.Predict(t.Context(), candidates, input.Files, loaded, predictionPlan(loaded), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(predicted.Rows, qt.HasLen, 1)
	c.Assert(predicted.Rows[0].Status, qt.Equals, "available")
	restored, err := training.LoadPredictions(t.Context(), testfixture.Encode(t, predicted))
	c.Assert(err, qt.IsNil)
	c.Assert(restored, qt.DeepEquals, predicted)
	again, err := training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, fitted)
	loaded.Lexical.Terms[0].Key = "changed"
	c.Assert(fitted.Lexical.Terms[0].Key, qt.Not(qt.Equals), "changed")
}

func TestLexicalVocabularyDoesNotLearnReservedText(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	before, err := training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.IsNil)
	for _, index := range []int{2, 3, 4, 5} {
		input.ReplaceSource(index, fmt.Sprintf("Reserved%d ", index)+strings.Repeat("Reservedxyzzz ", 40)+"remains outside vocabulary fitting.")
	}
	candidates, round = input.Compile(t)
	after, err := training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(after.CorpusSHA256, qt.Not(qt.Equals), before.CorpusSHA256)
	c.Assert(after.Lexical, qt.DeepEquals, before.Lexical)
	c.Assert(after.Identity, qt.DeepEquals, before.Identity)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
	c.Assert(after.Calibration.InputSHA256, qt.Not(qt.Equals), before.Calibration.InputSHA256)
}

func TestLexicalTrainingRequiresPermissionAndSimulationOptIn(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := lexicalFitting()
	options.AllowSimulation = false
	_, err := training.RunLexical(t.Context(), candidates, round, input.Files, options, lexicalOptions())
	c.Assert(err, qt.ErrorMatches, ".*allow_simulation.*")
	input.Manifest.Sources[0].Rights.AllowedUses = []string{"annotation", "evaluation"}
	candidates, round = input.Compile(t)
	_, err = training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.ErrorMatches, ".*training permission.*")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = training.RunLexical(ctx, candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestLexicalModelRejectsCorruptVocabulary(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.RunLexical(t.Context(), candidates, round, input.Files, lexicalFitting(), lexicalOptions())
	c.Assert(err, qt.IsNil)
	for _, edit := range []func(*training.Artifact){
		func(a *training.Artifact) { a.Version = "unswell-editorial-training-v1" },
		func(a *training.Artifact) { a.Version, a.Lexical, a.Identity.FeatureSource = training.Version, nil, "" },
		func(a *training.Artifact) { a.Lexical = nil },
		func(a *training.Artifact) { a.Lexical.Terms[0].Key = "c:corrupt" },
		func(a *training.Artifact) { a.Lexical.Options.MaxFeatures = 129 },
		func(a *training.Artifact) { a.Lexical.TrainingTargets++ },
	} {
		var changed training.Artifact
		c.Assert(json.Unmarshal(testfixture.Encode(t, fitted), &changed), qt.IsNil)
		edit(&changed)
		if changed.Lexical != nil {
			changed.Lexical.SHA256 = ""
			changed.Lexical.SHA256 = fmt.Sprintf("%x", sha256.Sum256(testfixture.Encode(t, changed.Lexical)))
			changed.Identity.LexicalVocabularyHash = changed.Lexical.SHA256
		}
		changed.SHA256 = ""
		changed.SHA256 = fmt.Sprintf("%x", sha256.Sum256(testfixture.Encode(t, changed)))
		_, err := training.Load(t.Context(), testfixture.Encode(t, changed))
		c.Assert(err, qt.IsNotNil)
	}
}
