package e2e_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCorpusRuleTrainingUsesExplicitPolicy(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	artifact := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	args := []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--feature", "activation/policy.banned-phrases", "--calibration", "isotonic", "--allow-simulation"}
	c.Assert(researchCommand(t, binary, args, artifact, 2), qt.HasLen, 0)
	args = append(args, "--rule-config", fixture+"rules.yaml")
	output := researchCommand(t, binary, args, artifact, 0)
	checkFittedResult(t, output, 0.5, 0.5)
	c.Assert(researchCommand(t, binary, args, artifact, 0), qt.DeepEquals, output)
	checkRuleTrainingIdentity(t, output)
	for _, row := range []struct{ flag, value string }{
		{"--rule-config", fixture + "absent.yaml"}, {"--feature", "prose-words"},
		{"--missing-features", "zero"}, {"--allow-simulation", "false"},
	} {
		t.Run(row.flag, func(t *testing.T) {
			c := qt.New(t)
			changed := append(slices.Clone(args), row.flag+"="+row.value)
			c.Assert(researchCommand(t, binary, changed, artifact, 2), qt.HasLen, 0)
		})
	}
	join := []string{"join", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--feature", "activation/policy.banned-phrases", "--rule-config", fixture + "rules.yaml"}
	bindings := researchCommand(t, binary, join, artifact, 0)
	c.Assert(string(bindings), qt.Contains, `"status": "verified_targets_with_block_activations"`)
	c.Assert(string(bindings), qt.Contains, `"context": "source_document"`)
}

func checkRuleTrainingIdentity(t *testing.T, output []byte) {
	t.Helper()
	c := qt.New(t)
	var result struct {
		Identity struct {
			FeatureSource      string `json:"feature_source"`
			Context            string `json:"context"`
			UnitContract       string `json:"unit_contract"`
			ActivationContract string `json:"activation_contract"`
			RulesetHash        string `json:"ruleset_hash"`
			RuleConfigSHA256   string `json:"rule_config_sha256"`
			IncludeStructure   bool   `json:"include_structure"`
		} `json:"identity"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Identity.FeatureSource, qt.Equals, "rule_activations")
	c.Assert(result.Identity.Context, qt.Equals, "source_document")
	c.Assert(result.Identity.IncludeStructure, qt.IsTrue)
	c.Assert(result.Identity.UnitContract, qt.Equals, "mapped-block-v1")
	c.Assert(result.Identity.ActivationContract, qt.Equals, "unswell-rule-activations-v1")
	c.Assert(result.Identity.RulesetHash, qt.HasLen, 64)
	c.Assert(result.Identity.RuleConfigSHA256, qt.HasLen, 64)
}
