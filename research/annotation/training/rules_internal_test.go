package training

// White-box tests: Change intermediate rule identities and inspect selected partitions;
// RunRules generates the join itself and does not accept altered intermediate feature collections.

import (
	"context"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
)

func ruleOptions() Options {
	options := fittingOptions()
	options.Features = []string{"activation/policy.banned-phrases"}
	return options
}

func ruleConfig(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile("testdata/rules.yaml")
	c := qt.New(t)
	c.Assert(err, qt.IsNil)
	return data
}

func TestRuleBaselineUsesActualActivationsAndSharedFitting(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	configuration := ruleConfig(t)
	result, err := RunRules(t.Context(), candidates, round, input.Files, ruleOptions(), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Identity.FeatureSource, qt.Equals, "rule_activations")
	c.Assert(result.Identity.Context, qt.Equals, "source_document")
	c.Assert(result.Identity.IncludeStructure, qt.IsTrue)
	c.Assert(result.Identity.Kind, qt.Equals, "paragraph")
	c.Assert(result.Identity.Columns[0].Scope, qt.Equals, "block")
	c.Assert(result.Identity.RulesetHash, qt.HasLen, 64)
	c.Assert(result.Identity.RuleConfigSHA256, qt.HasLen, 64)
	c.Assert(result.Logistic.Means, qt.DeepEquals, []float64{0.5})
	c.Assert(result.Logistic.Scales, qt.DeepEquals, []float64{0.5})
	c.Assert(result.Calibration.Responses, qt.DeepEquals, []float64{0, 1})
	c.Assert(result.Partitions[0].Rows, qt.HasLen, 2)
	c.Assert(result.Partitions[2].Rows, qt.HasLen, 2)
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	again, err := RunRules(t.Context(), candidates, round, input.Files, ruleOptions(), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, result)
	result.Identity.Columns[0].ID = "changed"
	result.Options.Features[0] = "changed"
	result.Logistic.Means[0] = 99
	third, err := RunRules(t.Context(), candidates, round, input.Files, ruleOptions(), configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(third, qt.DeepEquals, again)
}

func TestRuleBaselineKeepsCalibrationAndReservedInputsSeparate(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options, configuration := ruleOptions(), ruleConfig(t)
	before, err := RunRules(t.Context(), candidates, round, input.Files, options, configuration)
	c.Assert(err, qt.IsNil)
	input.ReplaceSource(3, "Keep the timeout condition.")
	input.ReplaceSource(4, "It is important to note that development remains reserved.")
	input.ReplaceSource(5, "In order to check evaluation, keep this source reserved.")
	candidates, round = input.Compile(t)
	after, err := RunRules(t.Context(), candidates, round, input.Files, options, configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
	c.Assert(after.Calibration.InputSHA256, qt.Not(qt.Equals), before.Calibration.InputSHA256)
	for _, index := range []int{1, 3} {
		c.Assert(after.Partitions[index].Classes, qt.HasLen, 0)
		c.Assert(after.Partitions[index].Rows, qt.HasLen, 0)
		c.Assert(after.Partitions[index].Excluded["reserved_partition"], qt.Equals, 2)
	}
	options.Calibration = "none"
	after, err = RunRules(t.Context(), candidates, round, input.Files, options, configuration)
	c.Assert(err, qt.IsNil)
	c.Assert(after.Calibration, qt.IsNil)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
}

func TestRuleBaselineRetainsPolicyIdentityWithoutChangingRawValues(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	configuration := ruleConfig(t)
	before, err := RunRules(t.Context(), candidates, round, input.Files, ruleOptions(), configuration)
	c.Assert(err, qt.IsNil)
	changed := []byte(strings.Replace(string(configuration), "enabled: true", "enabled: true\n    severity: error", 1))
	after, err := RunRules(t.Context(), candidates, round, input.Files, ruleOptions(), changed)
	c.Assert(err, qt.IsNil)
	c.Assert(after.Logistic, qt.DeepEquals, before.Logistic)
	c.Assert(after.Identity.RuleConfigSHA256, qt.Not(qt.Equals), before.Identity.RuleConfigSHA256)
	c.Assert(after.Identity.PolicyHash, qt.Not(qt.Equals), before.Identity.PolicyHash)
	c.Assert(after.Partitions[0].Rows[0].FeatureInputHash, qt.Not(qt.Equals), before.Partitions[0].Rows[0].FeatureInputHash)
}

func TestRuleSelectionPreservesMissingTargetsAndDisabledValues(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "testdata")
	candidates, round := input.Compile(t)
	options := ruleOptions()
	joined, err := corpus.JoinRules(t.Context(), candidates, round, input.Files, options.Features, ruleConfig(t))
	c.Assert(err, qt.IsNil)
	for i := range joined.Bindings {
		binding := &joined.Bindings[i]
		if binding.UnitID == "u000002" {
			binding.BlockID, binding.InputHash, binding.Reason = nil, "", "no_complete_block_match"
		}
	}
	_, err = selectRuleRows(t.Context(), candidates.Plan, joined, options, "config-hash")
	c.Assert(err, qt.ErrorMatches, ".*unavailable target no_complete_block_match.*")
	options.MissingFeatures = "exclude"
	selected, err := selectRuleRows(t.Context(), candidates.Plan, joined, options, "config-hash")
	c.Assert(err, qt.IsNil)
	c.Assert(selected.training, qt.HasLen, 1)
	c.Assert(selected.training[0].Values, qt.DeepEquals, []float64{1})
	c.Assert(selected.partitions[0].Excluded["target/no_complete_block_match"], qt.Equals, 1)
	joined, err = corpus.JoinRules(t.Context(), candidates, round, input.Files, options.Features,
		[]byte("version: 1\nextends: [builtin:custom]\n"))
	c.Assert(err, qt.IsNil)
	selected, err = selectRuleRows(t.Context(), candidates.Plan, joined, options, "config-hash")
	c.Assert(err, qt.IsNil)
	c.Assert(selected.training, qt.HasLen, 0)
	c.Assert(selected.partitions[0].Excluded["feature/activation/policy.banned-phrases/disabled"], qt.Equals, 2)
}

func TestRuleBaselineRejectsInvalidOrUnauthorizedInputs(t *testing.T) {
	for _, name := range []string{"simulation", "rights", "tamper", "cancel", "configuration", "disabled", "prepared feature", "kind"} {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			input := testfixture.Load(t, "testdata")
			if name == "rights" {
				input.Manifest.Sources[0].Rights.AllowedUses = []string{"annotation", "evaluation"}
			}
			candidates, round := input.Compile(t)
			options, configuration := ruleOptions(), ruleConfig(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			switch name {
			case "simulation":
				options.AllowSimulation = false
			case "tamper":
				input.Files["d000001.txt"][0]++
			case "cancel":
				cancel()
			case "configuration":
				configuration = nil
			case "disabled":
				configuration = []byte("version: 1\nextends: [builtin:custom]\n")
			case "prepared feature":
				options.Features = []string{"prose-words"}
			case "kind":
				options.Kind = "sentence"
			}
			result, err := RunRules(ctx, candidates, round, input.Files, options, configuration)
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, Artifact{})
		})
	}
}
