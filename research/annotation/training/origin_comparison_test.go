package training_test

import (
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/training"
)

// Two trials fitted on provenance labels compare against the same labels
// without a round, and a decision set for another artifact is refused.
func TestComparisonScoresProvenanceLabels(t *testing.T) {
	c := qt.New(t)
	artifact, files := originFixture(c)
	decisions, err := corpus.OriginDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	protocol := strings.Repeat("b", 64)
	threshold := 0.5

	trials := make([][]byte, 0, 2)
	digests := make([]string, 0, 2)
	for _, features := range [][]string{{"prose-words"}, {"prose-words", "prose-sentences"}} {
		options := fittingOptions()
		options.AllowSimulation, options.Features = false, features
		fitted, err := training.RunDecisions(c.Context(), artifact, decisions, files, options)
		c.Assert(err, qt.IsNil)
		plan := training.PredictionPlan{Version: training.PredictionVersion, ID: "origin-trial", ProtocolSHA256: protocol,
			ModelSHA256: fitted.SHA256, CorpusSHA256: artifact.SHA256, Partition: "final_test",
			Context: "prepared_piece", Response: "logistic", Threshold: &threshold}
		predictions, err := training.Predict(c.Context(), artifact, files, fitted, plan, nil)
		c.Assert(err, qt.IsNil)
		encoded, err := json.Marshal(predictions)
		c.Assert(err, qt.IsNil)
		trials, digests = append(trials, encoded), append(digests, predictions.SHA256)
	}
	plan := evaluation.ComparisonPlan{Version: evaluation.ComparisonVersion, ID: "origin-pair", ProtocolSHA256: protocol,
		CandidateSHA256: digests[0], ComparatorSHA256: digests[1]}

	result, err := evaluation.RunComparisonDecisions(c.Context(), trials[0], trials[1], plan, artifact, decisions, false)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Basis, qt.Equals, "declared_provenance")
	c.Assert(result.RoundSHA256, qt.Equals, artifact.SHA256)
	c.Assert(result.Candidates > 0, qt.IsTrue)

	evaluated, err := evaluation.RunDecisions(c.Context(), trials[0], artifact, decisions, false)
	c.Assert(err, qt.IsNil)
	c.Assert(evaluated.Basis, qt.Equals, "declared_provenance")
	c.Assert(evaluated.Candidates, qt.Equals, result.Candidates)

	foreign := decisions
	foreign.RoundSHA256 = "0"
	_, err = evaluation.RunComparisonDecisions(c.Context(), trials[0], trials[1], plan, artifact, foreign, false)
	c.Assert(err, qt.ErrorMatches, "decisions bind a different candidate artifact")
	_, err = evaluation.RunDecisions(c.Context(), trials[0], artifact, foreign, false)
	c.Assert(err, qt.ErrorMatches, "decisions bind a different candidate artifact")
}
