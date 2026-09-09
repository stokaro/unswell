package corpus_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

const ruleBindingConfig = "version: 1\nextends: [builtin:custom]\nrules:\n" +
	"  policy.banned-phrases: {enabled: true, parameters: {phrases: [cache]}}\n"

func TestRuleJoinMatchesCompleteTargetsAndRetainsMissingLabels(t *testing.T) {
	c := qt.New(t)
	artifact, files := candidateFixture(t)
	round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
	features := []string{"activation/policy.banned-phrases"}
	result, err := corpus.JoinRules(t.Context(), artifact, round, files, features, []byte(ruleBindingConfig))
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "verified_targets_with_block_activations")
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.Context, qt.Equals, "source_document")
	c.Assert(result.Decisions.Units[0].Label, qt.IsNil)
	c.Assert(result.Bindings, qt.HasLen, len(artifact.Units))
	matched, unmatched := 0, 0
	for i, binding := range result.Bindings {
		unit := artifact.Units[i].Unit
		c.Assert(binding.UnitID, qt.Equals, unit.ID)
		c.Assert(binding.Kind, qt.Equals, unit.Kind)
		if unit.Text == "Keep" || unit.Text == "unchanged." {
			c.Assert(binding.Reason, qt.Equals, "no_complete_block_match")
			c.Assert(binding.BlockID, qt.IsNil)
			c.Assert(binding.InputHash, qt.Equals, "")
			unmatched++
		} else {
			c.Assert(binding.Reason, qt.Equals, "", qt.Commentf("target %q", unit.Text))
			c.Assert(binding.BlockID, qt.IsNotNil)
			c.Assert(binding.InputHash, qt.HasLen, 64)
			matched++
		}
	}
	c.Assert(matched > 0, qt.IsTrue)
	c.Assert(unmatched, qt.Equals, 2)
	again, err := corpus.JoinRules(t.Context(), artifact, round, files, features, []byte(ruleBindingConfig))
	c.Assert(err, qt.IsNil)
	c.Assert(result, qt.DeepEquals, again)
	result.Features.Sources[0].Units[0].Binding.TextSHA256 = "changed"
	c.Assert(again.Features.Sources[0].Units[0].Binding.TextSHA256, qt.Not(qt.Equals), "changed")
}

func TestRuleJoinDoesNotEnableRulesOrInventZeros(t *testing.T) {
	c := qt.New(t)
	artifact, files := candidateFixture(t)
	round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
	result, err := corpus.JoinRules(t.Context(), artifact, round, files, []string{"activation/policy.banned-phrases"},
		[]byte("version: 1\nextends: [builtin:custom]\n"))
	c.Assert(err, qt.IsNil)
	for _, source := range result.Features.Sources {
		for _, unit := range source.Units {
			c.Assert(unit.Values[0].Number, qt.IsNil)
			c.Assert(unit.Values[0].Reason, qt.Equals, "disabled")
		}
	}
}

func TestRuleJoinRejectsConfigurationAndExtractionChanges(t *testing.T) {
	for _, configuration := range []string{"", "version: 1\nunknown: true\n", ruleBindingConfig + "extraction: {contexts: []}\n",
		ruleBindingConfig + "analysis: {include_quotes: true}\n"} {
		t.Run(strings.ReplaceAll(configuration, "\n", " "), func(t *testing.T) {
			c := qt.New(t)
			artifact, files := candidateFixture(t)
			round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
			result, err := corpus.JoinRules(t.Context(), artifact, round, files,
				[]string{"activation/policy.banned-phrases"}, []byte(configuration))
			c.Assert(err, qt.IsNotNil)
			c.Assert(result, qt.DeepEquals, corpus.RuleJoinedArtifact{})
		})
	}
}

func TestRuleJoinRejectsSourceTamperingAndCancellation(t *testing.T) {
	c := qt.New(t)
	artifact, files := candidateFixture(t)
	round := candidateRound(t, []annotation.Unit{artifact.Units[0].Unit})
	files["readme.md"][0]++
	result, err := corpus.JoinRules(t.Context(), artifact, round, files,
		[]string{"activation/policy.banned-phrases"}, []byte(ruleBindingConfig))
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, corpus.RuleJoinedArtifact{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.JoinRules(ctx, artifact, round, files, []string{"activation/policy.banned-phrases"}, []byte(ruleBindingConfig))
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
