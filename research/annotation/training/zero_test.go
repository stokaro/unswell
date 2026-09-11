package training_test

import (
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// A rule that cannot fire on a unit counts as zero under the zero policy, the
// artifact records how often, and prediction fills the same zeros. Every
// other feature source refuses the policy, at fitting and at restoration.
func TestZeroPolicyCountsAbstainingRulesAsZero(t *testing.T) {
	c := qt.New(t)
	artifact, files := originFixture(c)
	decisions, err := corpus.OriginDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	configuration := []byte("version: 1\nextends: [builtin:strict-v1]\nrules:\n" +
		"  hype.vague-praise: {enabled: true}\n  filler.wordy-phrase: {enabled: true}\n")
	options := fittingOptions()
	options.AllowSimulation = false
	options.Features = []string{"activation/hype.vague-praise", "activation/filler.wordy-phrase"}

	options.MissingFeatures = "exclude"
	excluded, err := training.RunRulesDecisions(c.Context(), artifact, decisions, files, options, configuration)
	if err == nil {
		c.Assert(excluded.Partitions[0].Excluded["feature/activation/hype.vague-praise/inapplicable/no_eligible_window"] > 0, qt.IsTrue)
	}

	options.MissingFeatures = "zero"
	fitted, err := training.RunRulesDecisions(c.Context(), artifact, decisions, files, options, configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Options.MissingFeatures, qt.Equals, "zero")
	zeroed := 0
	for _, count := range fitted.Partitions[0].ZeroFilled {
		zeroed += count
	}
	c.Assert(zeroed > 0, qt.IsTrue)
	c.Assert(fitted.Partitions[0].Excluded["feature/activation/hype.vague-praise/inapplicable/no_eligible_window"], qt.Equals, 0)

	encoded, err := json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	restored, err := training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(restored.Options.MissingFeatures, qt.Equals, "zero")

	threshold := 0.5
	plan := training.PredictionPlan{Version: training.PredictionVersion, ID: "zero-trial", ProtocolSHA256: protocolDigest(),
		ModelSHA256: fitted.SHA256, CorpusSHA256: artifact.SHA256, Partition: "final_test",
		Context: "source_document", Response: "logistic", Threshold: &threshold}
	predictions, err := training.Predict(c.Context(), artifact, files, fitted, plan, configuration)
	c.Assert(err, qt.IsNil)
	available := 0
	for _, row := range predictions.Rows {
		if row.Status == "available" {
			available++
		}
	}
	c.Assert(available, qt.Equals, len(predictions.Rows))

	prepared := fittingOptions()
	prepared.AllowSimulation, prepared.MissingFeatures = false, "zero"
	_, err = training.RunDecisions(c.Context(), artifact, decisions, files, prepared)
	c.Assert(err, qt.ErrorMatches, "the zero missing-feature policy applies to rule activations only")

	fitted.Identity.FeatureSource = ""
	fitted.SHA256 = ""
	encoded, err = json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	_, err = training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNotNil)
}

func protocolDigest() string {
	return "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
}
