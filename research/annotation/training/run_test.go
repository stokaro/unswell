package training_test

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func TestRunFitsTrainingAndSeparateCalibration(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	result, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "experimental_numerical_fit")
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(result.Basis, qt.Equals, "simulation")
	c.Assert(result.Identity.Task, qt.Equals, "editorial_needs_revision")
	c.Assert(result.Identity.Kind, qt.Equals, "paragraph")
	c.Assert(result.Identity.Columns[0].ID, qt.Equals, "prose-words")
	c.Assert(result.Logistic.Means, qt.DeepEquals, []float64{6})
	c.Assert(result.Logistic.Scales, qt.HasLen, 1)
	c.Assert(math.Abs(result.Logistic.Scales[0]-3) < 1e-12, qt.IsTrue)
	c.Assert(result.Logistic.GradientNorm <= result.Options.Fit.Tolerance, qt.IsTrue)
	c.Assert(result.Calibration, qt.IsNotNil)
	c.Assert(result.Calibration.Samples, qt.Equals, 2)
	c.Assert(result.Calibration.Responses, qt.DeepEquals, []float64{0, 1})
	c.Assert(result.Calibration.ScoreKind, qt.Equals, "linear_score")
	c.Assert(result.Calibration.MeanSquaredError, qt.Equals, float64(0))
	c.Assert(result.Partitions[0].Rows, qt.HasLen, 2)
	c.Assert(result.Partitions[2].Rows, qt.HasLen, 2)
	for _, index := range []int{1, 3} {
		c.Assert(result.Partitions[index].Rows, qt.HasLen, 0)
		c.Assert(result.Partitions[index].Classes, qt.HasLen, 0)
		c.Assert(result.Partitions[index].Excluded["reserved_partition"], qt.Equals, 2)
	}
	again, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(result, qt.DeepEquals, again)
	result.Logistic.Means[0] = 100
	c.Assert(again.Logistic.Means, qt.DeepEquals, []float64{6})
}

func TestReservedInputsDoNotAffectFittedParameters(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	before, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	input.ReplaceSource(4, strings.Repeat("Development ", 1000)+"remains reserved.")
	input.ReplaceSource(5, strings.Repeat("Evaluation ", 1500)+"remains reserved.")
	var judgments []annotation.Judgment
	c.Assert(json.Unmarshal(input.Round["judgments"], &judgments), qt.IsNil)
	for i := range judgments {
		if judgments[i].UnitID == "u000012" {
			judgments[i].Label = "acceptable"
			judgments[i].Categories = []string{}
		}
	}
	input.Round["judgments"] = testfixture.Encode(t, judgments)
	candidates, round = input.Compile(t)
	after, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(after.CorpusSHA256, qt.Not(qt.Equals), before.CorpusSHA256)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
	c.Assert(after.Calibration, qt.DeepEquals, before.Calibration)
	c.Assert(after.Identity, qt.DeepEquals, before.Identity)
}

func TestCalibrationRowsDoNotTrainTheClassifier(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	before, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	input.ReplaceSource(3, strings.Repeat("Calibration ", 30)+"changes its independent score.")
	candidates, round = input.Compile(t)
	after, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
	c.Assert(after.Calibration.InputSHA256, qt.Not(qt.Equals), before.Calibration.InputSHA256)
	options := fittingOptions()
	options.Calibration = "none"
	uncalibrated, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(uncalibrated.Calibration, qt.IsNil)
	c.Assert(uncalibrated.Partitions[2].Rows, qt.HasLen, 0)
	c.Assert(uncalibrated.Logistic, qt.DeepEquals, before.Logistic)
}

func TestRunRejectsSimulationAndMissingRights(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := fittingOptions()
	options.AllowSimulation = false
	result, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.ErrorMatches, ".*allow_simulation.*")
	c.Assert(result, qt.DeepEquals, training.Artifact{})
	input.Manifest.Sources[0].Rights.AllowedUses = []string{"annotation", "evaluation"}
	candidates, round = input.Compile(t)
	result, err = training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.ErrorMatches, ".*training permission.*")
	c.Assert(result, qt.DeepEquals, training.Artifact{})
}

func TestRunSourceFailureAndCancellation(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	input.Files["d000001.txt"][0]++
	result, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, training.Artifact{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = training.Run(ctx, candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, training.Artifact{})
}
