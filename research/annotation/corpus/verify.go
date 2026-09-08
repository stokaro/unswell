package corpus

import (
	"context"
	"fmt"
	"reflect"

	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// LoadPlan rejects ambiguous JSON and a modified or stale assignment artifact.
func LoadPlan(ctx context.Context, data []byte) (Plan, error) {
	var plan Plan
	if err := jsoninput.Decode(ctx, data, MaxManifestBytes, &plan, inputLimits()); err != nil {
		return Plan{}, err
	}
	if err := ValidatePlan(ctx, plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// LoadArtifact checks integrity and grouping, but does not verify source files.
// Call Verify before treating source maps or candidate contents as reproduced.
func LoadArtifact(ctx context.Context, data []byte) (Artifact, error) {
	var artifact Artifact
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &artifact, inputLimits()); err != nil {
		return Artifact{}, err
	}
	if err := artifact.validate(ctx); err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func (a Artifact) validate(ctx context.Context) error {
	if a.Version != Version || a.Status != "unlabeled_candidates" || len(a.Units) == 0 || len(a.Units) > MaxUnits {
		return fmt.Errorf("unsupported or empty candidate artifact")
	}
	if err := ValidatePlan(ctx, a.Plan); err != nil {
		return err
	}
	want := a.SHA256
	a.SHA256 = ""
	got, err := digest(a)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("candidate artifact digest mismatch")
	}
	return nil
}

// Verification proves reproduction on the supplied exact sources, not annotation
// quality, permissions, provenance, English fluency, or unrecorded relationships.
type Verification struct {
	Status         string   `json:"status"`
	ArtifactSHA256 string   `json:"artifact_sha256"`
	Producer       Pipeline `json:"producer"`
	Verifier       Pipeline `json:"verifier"`
	Units          int      `json:"units"`
	SourceCount    int      `json:"source_count"`
	HumanCorpus    string   `json:"human_corpus"`
}

// Verify repeats extraction and compares all candidates, mappings, and exclusions.
// Compiler/VCS metadata may differ, and both identities remain in the output.
func Verify(ctx context.Context, artifact Artifact, files map[string][]byte) (Verification, error) {
	if err := artifact.validate(ctx); err != nil {
		return Verification{}, err
	}
	rebuilt, err := Build(ctx, artifact.Plan, files)
	if err != nil {
		return Verification{}, err
	}
	if !samePipeline(artifact.Pipeline, rebuilt.Pipeline) || !reflect.DeepEqual(artifact.Sources, rebuilt.Sources) ||
		!reflect.DeepEqual(artifact.Units, rebuilt.Units) {
		return Verification{}, fmt.Errorf("candidate contents, mappings, or pipeline are not reproduced")
	}
	return Verification{Status: "source_and_candidates_reproduced", ArtifactSHA256: artifact.SHA256,
		Producer: artifact.Pipeline, Verifier: rebuilt.Pipeline, Units: len(rebuilt.Units),
		SourceCount: len(rebuilt.Sources), HumanCorpus: "not_qualified"}, nil
}

func samePipeline(a, b Pipeline) bool {
	return a.Version == b.Version && reflect.DeepEqual(a.NLP, b.NLP) && reflect.DeepEqual(a.Dependencies, b.Dependencies)
}
