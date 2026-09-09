package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/research/annotation"
)

// JoinedVersion identifies reproduced annotation and prepared feature bindings.
const JoinedVersion = "unswell-corpus-feature-bindings-v1"

// FeatureBinding joins one exact candidate to its measured target. Split metadata
// comes from the reproduced corpus plan, never from labels or rule outcomes.
type FeatureBinding struct {
	UnitID           string `json:"unit_id"`
	SourceID         string `json:"source_id"`
	Path             string `json:"path"`
	GroupID          string `json:"group_id"`
	Partition        string `json:"partition"`
	FeatureInputHash string `json:"feature_input_hash"`
}

// JoinedArtifact retains decisions separately from descriptive measurements.
// It contains every corpus target, including units outside the round and units
// without resolved labels. HumanCorpus remains not_qualified. SHA256 covers
// compact Go JSON with that field omitted and does not prove honest provenance.
type JoinedArtifact struct {
	Version      string                            `json:"version"`
	Status       string                            `json:"status"`
	SHA256       string                            `json:"sha256,omitempty"`
	HumanCorpus  string                            `json:"human_corpus"`
	Verification Verification                      `json:"verification"`
	Decisions    annotation.DecisionSet            `json:"decisions"`
	Features     unswell.PreparedFeatureCollection `json:"features"`
	Bindings     []FeatureBinding                  `json:"bindings"`
}

// Join reproduces a corpus and measures its exact targets with the shared engine.
// The supplied round may cover a subset. Features are an explicit nonempty set
// of prepared feature IDs; kinds and extraction policy come from the frozen plan.
// It never promotes declarations or simulation into qualified training labels.
// Sources are caller-owned immutable bytes. Errors return no partial artifact.
func Join(ctx context.Context, artifact Artifact, round *annotation.Round, files map[string][]byte,
	features []string,
) (JoinedArtifact, error) {
	verification, err := Verify(ctx, artifact, files)
	if err != nil {
		return JoinedArtifact{}, err
	}
	verification.Producer.Dependencies = slices.Clone(verification.Producer.Dependencies)
	verification.Producer.NLP.Capabilities = slices.Clone(verification.Producer.NLP.Capabilities)
	expected := make([]annotation.Unit, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		expected = append(expected, candidate.Unit)
	}
	if err := round.MatchTargets(ctx, expected); err != nil {
		return JoinedArtifact{}, err
	}
	decisions, err := round.Decisions(ctx)
	if err != nil {
		return JoinedArtifact{}, err
	}
	collection, err := measureCandidates(ctx, artifact.Plan, files, features)
	if err != nil {
		return JoinedArtifact{}, err
	}
	bindings, err := bindCandidates(ctx, artifact, collection)
	if err != nil {
		return JoinedArtifact{}, err
	}
	result := JoinedArtifact{Version: JoinedVersion, Status: "verified_targets_with_measured_features",
		HumanCorpus: "not_qualified", Verification: verification, Decisions: decisions, Features: collection, Bindings: bindings}
	return finishJoin(ctx, result)
}

func finishJoin(ctx context.Context, result JoinedArtifact) (JoinedArtifact, error) {
	encoded, err := json.Marshal(result)
	if err != nil {
		return JoinedArtifact{}, err
	}
	if len(encoded)+128 > MaxArtifactBytes {
		return JoinedArtifact{}, fmt.Errorf("joined artifact exceeds %d bytes", MaxArtifactBytes)
	}
	result.SHA256 = hashBytes(encoded)
	if err := ctx.Err(); err != nil {
		return JoinedArtifact{}, err
	}
	return result, nil
}
