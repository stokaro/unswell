package corpus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/rule"
)

// RuleJoinedVersion identifies explicit original-block activation bindings.
const RuleJoinedVersion = "unswell-corpus-rule-bindings-v1"

// RuleFeatureBinding relates a candidate to a complete original block or records
// why it cannot inherit that block's activation. InputHash binds measured input;
// Kind is the annotation target, while the feature descriptors keep block scope.
type RuleFeatureBinding struct {
	UnitID    string `json:"unit_id"`
	SourceID  string `json:"source_id"`
	Path      string `json:"path"`
	GroupID   string `json:"group_id"`
	Partition string `json:"partition"`
	Kind      string `json:"kind"`
	BlockID   *int   `json:"block_id"`
	InputHash string `json:"input_hash,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// RuleJoinedArtifact keeps original-block measurements separate from annotation
// targets. Document rules retain their surrounding source context. Missing target
// matches or activation values never imply zero or qualified human labels.
type RuleJoinedArtifact struct {
	Version      string                    `json:"version"`
	Status       string                    `json:"status"`
	SHA256       string                    `json:"sha256,omitempty"`
	HumanCorpus  string                    `json:"human_corpus"`
	Context      string                    `json:"context"`
	Verification Verification              `json:"verification"`
	Decisions    annotation.DecisionSet    `json:"decisions"`
	Rules        []rule.Descriptor         `json:"rules"`
	Features     unswell.FeatureCollection `json:"features"`
	Bindings     []RuleFeatureBinding      `json:"bindings"`
}

// JoinRules reproduces exact targets and runs selected registered rule activations
// under an explicit inline configuration. It never enables requested rules or
// overrides extraction to make a target match. Partial blocks remain unmatched.
// Errors return no partial artifact. This result is not a training authorization.
func JoinRules(ctx context.Context, artifact Artifact, round *annotation.Round, files map[string][]byte,
	features []string, configuration []byte,
) (RuleJoinedArtifact, error) {
	measured, err := MeasureRules(ctx, artifact, files, features, configuration)
	if err != nil {
		return RuleJoinedArtifact{}, err
	}
	decisions, err := targetDecisions(ctx, artifact, round)
	if err != nil {
		return RuleJoinedArtifact{}, err
	}
	result := RuleJoinedArtifact{Version: RuleJoinedVersion, Status: "verified_targets_with_block_activations",
		HumanCorpus: "not_qualified", Context: "source_document", Verification: measured.Verification, Decisions: decisions,
		Rules: measured.Rules, Features: measured.Features, Bindings: measured.Bindings}
	return finishRuleJoin(ctx, result)
}

// JoinRulesDecisions binds a prepared decision set to block activations
// instead of a round; the decisions must name this artifact and its targets.
func JoinRulesDecisions(ctx context.Context, artifact Artifact, decisions annotation.DecisionSet, files map[string][]byte,
	features []string, configuration []byte,
) (RuleJoinedArtifact, error) {
	if err := MatchDecisionTargets(ctx, artifact, decisions); err != nil {
		return RuleJoinedArtifact{}, err
	}
	measured, err := MeasureRules(ctx, artifact, files, features, configuration)
	if err != nil {
		return RuleJoinedArtifact{}, err
	}
	result := RuleJoinedArtifact{Version: RuleJoinedVersion, Status: "verified_targets_with_block_activations",
		HumanCorpus: "not_qualified", Context: "source_document", Verification: measured.Verification, Decisions: decisions,
		Rules: measured.Rules, Features: measured.Features, Bindings: measured.Bindings}
	return finishRuleJoin(ctx, result)
}

func finishRuleJoin(ctx context.Context, result RuleJoinedArtifact) (RuleJoinedArtifact, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return RuleJoinedArtifact{}, err
	}
	if len(encoded)+128 > MaxArtifactBytes {
		return RuleJoinedArtifact{}, fmt.Errorf("rule bindings exceed artifact size limit")
	}
	result.SHA256 = hashBytes(encoded)
	if err := ctx.Err(); err != nil {
		return RuleJoinedArtifact{}, err
	}
	return result, nil
}
