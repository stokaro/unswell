package training

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// RunRules fits the rule-activation baseline using the ordinary engine and an
// explicit inline policy. Only complete matching targets inherit block values.
// It shares Run's selection, permissions, optimizer, and calibration contracts.
// Configuration bytes are hashed, not included in the experimental result.
func RunRules(ctx context.Context, candidates corpus.Artifact, round *annotation.Round, files map[string][]byte,
	options Options, configuration []byte,
) (Artifact, error) {
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Artifact{}, err
	}
	joined, err := corpus.JoinRules(ctx, candidates, round, files, options.Features, configuration)
	if err != nil {
		return Artifact{}, err
	}
	return fitRules(ctx, candidates, joined, options, configuration)
}

// RunRulesDecisions fits the rule-activation baseline from a prepared decision
// set, such as the provenance labels of the origin task, instead of a round.
func RunRulesDecisions(ctx context.Context, candidates corpus.Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, options Options, configuration []byte,
) (Artifact, error) {
	if err := validateOptions(ctx, candidates, options); err != nil {
		return Artifact{}, err
	}
	joined, err := corpus.JoinRulesDecisions(ctx, candidates, decisions, files, options.Features, configuration)
	if err != nil {
		return Artifact{}, err
	}
	return fitRules(ctx, candidates, joined, options, configuration)
}

func fitRules(ctx context.Context, candidates corpus.Artifact, joined corpus.RuleJoinedArtifact, options Options,
	configuration []byte,
) (Artifact, error) {
	if joined.Decisions.Basis == "simulation" && !options.AllowSimulation {
		return Artifact{}, fmt.Errorf("tutorial training requires explicit allow_simulation")
	}
	options.Features = slices.Clone(joined.Features.Requested)
	configHash := fmt.Sprintf("%x", sha256.Sum256(configuration))
	selected, err := selectRuleRows(ctx, candidates.Plan, joined, options, configHash)
	if err != nil {
		return Artifact{}, err
	}
	return fitSelected(ctx, candidates, joined.Decisions, joined.SHA256, options, selected)
}
