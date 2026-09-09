package training_test

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
	"github.com/stokaro/unswell/research/annotation/training"
)

func FuzzResearchPredictionInputs(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"version":"unswell-research-predictions-v1","threshold":null}`))
	f.Add([]byte(`{"rows":[{"response":0,"positive":null}]}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		c := qt.New(t)
		if plan, err := training.LoadPredictionPlan(t.Context(), data); err == nil {
			encoded, err := json.Marshal(plan)
			c.Assert(err, qt.IsNil)
			again, err := training.LoadPredictionPlan(t.Context(), encoded)
			c.Assert(err, qt.IsNil)
			c.Assert(again, qt.DeepEquals, plan)
		}
		if result, err := training.LoadPredictions(t.Context(), data); err == nil {
			encoded, err := json.Marshal(result)
			c.Assert(err, qt.IsNil)
			again, err := training.LoadPredictions(t.Context(), encoded)
			c.Assert(err, qt.IsNil)
			c.Assert(again, qt.DeepEquals, result)
		}
	})
}

func TestConcurrentPredictionsOwnResults(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	fitted, err := training.Run(t.Context(), candidates, round, input.Files, fittingOptions())
	c.Assert(err, qt.IsNil)
	type outcome struct {
		predictions training.Predictions
		err         error
	}
	results := make(chan outcome, 2)
	for range 2 {
		go func() {
			result, err := training.Predict(t.Context(), candidates, input.Files, fitted, predictionPlan(fitted), nil)
			results <- outcome{result, err}
		}()
	}
	first, second := <-results, <-results
	c.Assert(first.err, qt.IsNil)
	c.Assert(second.err, qt.IsNil)
	c.Assert(first.predictions, qt.DeepEquals, second.predictions)
	first.predictions.Rows[0].Reason = "changed"
	*first.predictions.Rows[0].Response = 0
	c.Assert(second.predictions.Rows[0].Reason, qt.Equals, "")
	c.Assert(*second.predictions.Rows[0].Response > 0, qt.IsTrue)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = training.Predict(ctx, candidates, input.Files, fitted, predictionPlan(fitted), nil)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
