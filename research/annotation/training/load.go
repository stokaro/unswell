package training

import (
	"context"
	"encoding/hex"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// Load restores a bounded numerical artifact and verifies its digest and model
// shape. It does not attest that fitting occurred or qualify editorial accuracy.
func Load(ctx context.Context, data []byte) (Artifact, error) {
	var result Artifact
	limits := jsoninput.Limits{Array: corpus.MaxUnits, Object: 256,
		Arrays: map[string]int{"scores": model.MaxCalibrationSamples, "responses": model.MaxCalibrationSamples}}
	if err := jsoninput.Decode(ctx, data, MaxArtifactBytes, &result, limits); err != nil {
		return Artifact{}, err
	}
	want := result.SHA256
	result.SHA256 = ""
	rebuilt, err := finish(ctx, result)
	if err != nil {
		return Artifact{}, err
	}
	if rebuilt.SHA256 != want {
		return Artifact{}, fmt.Errorf("training artifact digest mismatch")
	}
	if err := validateRestored(rebuilt); err != nil {
		return Artifact{}, err
	}
	return rebuilt, nil
}

func validateRestored(a Artifact) error {
	if !validArtifactStatus(a) {
		return fmt.Errorf("unsupported training artifact status or version")
	}
	if err := validateRestoredDigests(a); err != nil {
		return err
	}
	if err := validateRestoredContract(a); err != nil {
		return err
	}
	if err := validateRestoredPartitions(a); err != nil {
		return err
	}
	_, _, err := restoreModels(a)
	return err
}

func validArtifactStatus(a Artifact) bool {
	return a.Version == Version && a.Status == "experimental_numerical_fit" &&
		a.HumanCorpus == "not_qualified" &&
		a.ProbabilityStatus == "unavailable_unqualified_model" &&
		slices.Contains([]string{"simulation", "declared_human", "declared_provenance"}, a.Basis)
}

func validModelShape(a Artifact) bool {
	return consistentTask(a.Identity) && a.Identity.Kind == a.Options.Kind && a.Identity.Rubric != "" &&
		len(a.Options.Features) == len(a.Identity.Columns)
}

func validateRestoredContract(a Artifact) error {
	if !validModelShape(a) {
		return fmt.Errorf("training artifact has incompatible target, columns, or algorithm")
	}
	if !validSelectionOptions(a) {
		return fmt.Errorf("training artifact has unsupported selection options")
	}
	columnsHash, err := hashJSON(a.Identity.Columns)
	if err != nil || columnsHash != a.Identity.ColumnsSHA256 {
		return fmt.Errorf("training column contract digest mismatch")
	}
	if err := validateEstimatorOptions(a.Options); err != nil {
		return err
	}
	if err := validateColumnOrder(a); err != nil {
		return err
	}
	if err := validateRestoredReservation(a.Options.Reservation); err != nil {
		return err
	}
	if err := validateLexicalArtifact(a); err != nil {
		return err
	}
	return validateCompressionArtifact(a)
}

func validSelectionOptions(a Artifact) bool {
	return slices.Contains([]string{"sentence", "paragraph", "fragment"}, a.Options.Kind) &&
		slices.Contains([]string{"reject", "exclude", "zero"}, a.Options.MissingFeatures) &&
		(a.Options.MissingFeatures != "zero" || a.Identity.FeatureSource == "rule_activations")
}

func validateRestoredReservation(r *Reservation) error {
	if r != nil && (!validDigest(r.BankSHA256) || len(r.Groups) == 0 ||
		!slices.IsSorted(r.Groups) || len(slices.Compact(slices.Clone(r.Groups))) != len(r.Groups)) {
		return fmt.Errorf("training artifact has an invalid reservation")
	}
	return nil
}

func validDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 32 && hex.EncodeToString(decoded) == value
}

func validateRestoredDigests(a Artifact) error {
	for _, value := range []string{a.CorpusSHA256, a.ManifestSHA256, a.RoundSHA256, a.JoinedSHA256,
		a.Identity.ColumnsSHA256, a.Identity.ProfileSHA256} {
		if !validDigest(value) {
			return fmt.Errorf("training artifact requires complete SHA-256 identities")
		}
	}
	return nil
}

func validateColumnOrder(a Artifact) error {
	for i, id := range a.Options.Features {
		if id != a.Identity.Columns[i].ID || (i > 0 && id <= a.Options.Features[i-1]) {
			return fmt.Errorf("training features require matching sorted unique columns")
		}
	}
	return nil
}
