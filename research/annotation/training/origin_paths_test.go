package training_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// The lexical and rule-activation baselines take a prepared decision set the
// same way the prepared-feature path does, so every arm of a comparison can
// fit on provenance labels.
func TestLexicalAndRuleBaselinesFitOnProvenanceLabels(t *testing.T) {
	c := qt.New(t)
	artifact, files := originFixture(c)
	decisions, err := corpus.OriginDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)

	lexical := training.Options{Estimator: "logistic", Kind: "paragraph", MissingFeatures: "reject", Calibration: "isotonic",
		Fit: &training.Fit{L2: 1, Tolerance: 1e-8, MaxIterations: 100, MaxOperations: 10_000_000}}
	vocabulary := training.LexicalOptions{Counts: feature.LexicalOptions{WordMin: 1, WordMax: 2, CharMin: 3, CharMax: 4},
		MaxFeatures: 64, MinTargets: 2}
	fitted, err := training.RunLexicalDecisions(c.Context(), artifact, decisions, files, lexical, vocabulary)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Identity.Task, qt.Equals, annotation.TaskOrigin)
	c.Assert(fitted.Identity.FeatureSource, qt.Equals, "lexical_ngrams")
	c.Assert(fitted.Basis, qt.Equals, "declared_provenance")
	c.Assert(fitted.Partitions[0].Classes["endpoint_generated"] > 0, qt.IsTrue)

	configuration, err := os.ReadFile("testdata/rules.yaml")
	c.Assert(err, qt.IsNil)
	rules := fittingOptions()
	rules.AllowSimulation = false
	rules.Features = []string{"activation/policy.banned-phrases"}
	fitted, err = training.RunRulesDecisions(c.Context(), artifact, decisions, files, rules, configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Identity.Task, qt.Equals, annotation.TaskOrigin)
	c.Assert(fitted.Identity.FeatureSource, qt.Equals, "rule_activations")
	c.Assert(fitted.Basis, qt.Equals, "declared_provenance")

	foreign := decisions
	foreign.RoundSHA256 = "0"
	_, err = training.RunLexicalDecisions(c.Context(), artifact, foreign, files, lexical, vocabulary)
	c.Assert(err, qt.ErrorMatches, "decisions bind a different candidate artifact")
	_, err = training.RunRulesDecisions(c.Context(), artifact, foreign, files, rules, configuration)
	c.Assert(err, qt.ErrorMatches, "decisions bind a different candidate artifact")
}
