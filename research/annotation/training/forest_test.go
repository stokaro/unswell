package training_test

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func forestFitting() training.Options {
	o := fittingOptions()
	o.Estimator, o.Fit = "forest", nil
	o.Forest = &model.ForestOptions{Seed: 7, Trees: 3, MaxDepth: 3, MinLeaf: 1,
		MaxNodes: 127, MaxOperations: 1000000}
	return o
}

func TestForestCorpusFitAndCalibration(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := forestFitting()
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Logistic, qt.IsNil)
	c.Assert(fitted.Forest.Parameters.Trees, qt.HasLen, 3)
	// The only training vectors are 3 and 9 words, with labels 0 and 1.
	c.Assert(fitted.Forest.Parameters.Trees[0], qt.DeepEquals, []model.ForestNode{
		{Feature: 0, Threshold: 6, Left: 1, Right: 2, Positive: 1, Samples: 2},
		{Feature: -1, Left: -1, Right: -1, Samples: 1},
		{Feature: -1, Left: -1, Right: -1, Positive: 1, Samples: 1},
	})
	c.Assert(fitted.Calibration.ScoreKind, qt.Equals, "forest_response")
	c.Assert(fitted.Calibration.Scores, qt.DeepEquals, []float64{0, 1})
	c.Assert(fitted.Calibration.Responses, qt.DeepEquals, []float64{0, 1})
	c.Assert(fitted.Partitions[0].Rows, qt.HasLen, 2)
	c.Assert(fitted.Partitions[2].Rows, qt.HasLen, 2)
	for _, index := range []int{1, 3} {
		c.Assert(fitted.Partitions[index].Rows, qt.HasLen, 0)
		c.Assert(fitted.Partitions[index].Classes, qt.HasLen, 0)
	}
	loaded, err := training.Load(t.Context(), testfixture.Encode(t, fitted))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, fitted)
	again, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, fitted)
	options.Forest.Seed++
	c.Assert(fitted.Options.Forest.Seed, qt.Equals, uint64(7))
	for _, channel := range []string{"forest", "isotonic"} {
		plan := predictionPlan(fitted)
		plan.Response = channel
		result, err := training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Rows[0].LinearScore, qt.IsNil)
		c.Assert(result.Rows[0].LogisticResponse, qt.IsNil)
		c.Assert(*result.Rows[0].ForestResponse, qt.Equals, 1.0)
		c.Assert(*result.Rows[0].Response, qt.Equals, 1.0)
		c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
		restored, err := training.LoadPredictions(t.Context(), testfixture.Encode(t, result))
		c.Assert(err, qt.IsNil)
		c.Assert(restored, qt.DeepEquals, result)
	}
}

func TestForestReservedTextCannotTrainTrees(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	before, err := training.Run(t.Context(), candidates, round, input.Files, forestFitting())
	c.Assert(err, qt.IsNil)
	for _, index := range []int{2, 3, 4, 5} {
		input.ReplaceSource(index, fmt.Sprintf("Reserved%d ", index)+strings.Repeat("Reserved ", 20)+"remains outside tree fitting.")
	}
	candidates, round = input.Compile(t)
	after, err := training.Run(t.Context(), candidates, round, input.Files, forestFitting())
	c.Assert(err, qt.IsNil)
	c.Assert(after.CorpusSHA256, qt.Not(qt.Equals), before.CorpusSHA256)
	c.Assert(after.Forest, qt.DeepEquals, before.Forest)
	c.Assert(after.Calibration.InputSHA256, qt.Not(qt.Equals), before.Calibration.InputSHA256)
	options := forestFitting()
	options.Calibration = "none"
	uncalibrated, err := training.Run(t.Context(), candidates, round, input.Files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(uncalibrated.Forest, qt.DeepEquals, before.Forest)
	c.Assert(uncalibrated.Partitions[2].Rows, qt.HasLen, 0)
}

func TestForestCalibrationAbstainsOutsideObservedScores(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	input.ReplaceSource(2, "Short calibration example.")
	input.ReplaceSource(3, "Another short example.")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, forestFitting())
	c.Assert(err, qt.IsNil)
	plan := predictionPlan(fitted)
	plan.Response = "isotonic"
	result, err := training.Predict(t.Context(), candidates, input.Files, fitted, plan, nil)
	c.Assert(err, qt.IsNil)
	row := result.Rows[0]
	c.Assert(row.Reason, qt.Equals, "calibration/out_of_range")
	c.Assert(row.Response, qt.IsNil)
	c.Assert(row.Positive, qt.IsNil)
	c.Assert(*row.ForestResponse, qt.Equals, 1.0)
	_, err = training.LoadPredictions(t.Context(), testfixture.Encode(t, result))
	c.Assert(err, qt.IsNil)
}
