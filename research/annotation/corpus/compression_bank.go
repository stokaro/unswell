package corpus

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

// CompressionBankVersion identifies explicit cohort selection and group reservation.
const CompressionBankVersion = "unswell-compression-bank-v1"

// Compression-bank limits bound explicit selection and the serialized developer artifact.
const (
	MaxCompressionSelectionBytes = 256 << 10
	MaxCompressionBankBytes      = 4 << 20
)

// CompressionCohort preserves an explicit reference order and endpoint policy.
// Origin accepts human or generated under a round, historical or contemporary
// under cohort labels, and mixed (a mixture of the two classes of one source).
type CompressionCohort struct {
	ID      string   `json:"id"`
	Origin  string   `json:"origin"`
	UnitIDs []string `json:"unit_ids"`
}

// CompressionBankOptions freezes common compressor settings and target kind.
// Tutorial origin claims require explicit AllowSimulation and stay simulated.
type CompressionBankOptions struct {
	Version         string                     `json:"version"`
	Kind            string                     `json:"kind"`
	Compression     feature.CompressionOptions `json:"compression"`
	Cohorts         []CompressionCohort        `json:"cohorts"`
	AllowSimulation bool                       `json:"allow_simulation"`
}

// CompressionReferenceUnit binds one ordered reference entry to its original source.
// Start and End address the reference string, not the original file. Label is
// the class the unit matched under the bank's label source.
type CompressionReferenceUnit struct {
	UnitID   string            `json:"unit_id"`
	SourceID string            `json:"source_id"`
	GroupID  string            `json:"group_id"`
	Path     string            `json:"path"`
	Source   annotation.Source `json:"source"`
	Binding  nlp.UnitBinding   `json:"binding"`
	Origin   annotation.Origin `json:"origin"`
	Label    string            `json:"label,omitempty"`
	Rights   annotation.Rights `json:"rights"`
	Notices  []Notice          `json:"notices"`
	Start    int               `json:"start"`
	End      int               `json:"end"`
}

// CompressionReference contains source prose intentionally, as an explicit model input.
// Reference excludes the final LF added by the shared compression primitive.
type CompressionReference struct {
	ID        string                      `json:"id"`
	Origin    string                      `json:"origin"`
	Reference string                      `json:"reference"`
	Identity  feature.CompressionIdentity `json:"identity"`
	Units     []CompressionReferenceUnit  `json:"units"`
}

// CompressionReservation excludes a related target even when it is not a seed.
type CompressionReservation struct {
	UnitID   string `json:"unit_id"`
	SourceID string `json:"source_id"`
	GroupID  string `json:"group_id"`
}

// CompressionBank is an unqualified developer artifact containing reference prose.
// ReservedGroups and ReservedTargets are the common exclusion union for all cohorts.
// The original plan is unchanged. Consumers must apply this reservation when fitting.
type CompressionBank struct {
	Version                string                   `json:"version"`
	SHA256                 string                   `json:"sha256,omitempty"`
	HumanCorpus            string                   `json:"human_corpus"`
	Basis                  string                   `json:"basis"`
	RoundSHA256            string                   `json:"round_sha256"`
	OriginsSHA256          string                   `json:"origins_sha256"`
	Verification           Verification             `json:"verification"`
	Options                CompressionBankOptions   `json:"options"`
	Cohorts                []CompressionReference   `json:"cohorts"`
	ReservedGroups         []Group                  `json:"reserved_groups"`
	ReservedTargets        []CompressionReservation `json:"reserved_targets"`
	RemainingTrainingUnits int                      `json:"remaining_training_units"`
}

// BuildCompressionBank reproduces sources, selects training references, and reserves
// all connected groups across cohorts. It never examines editorial quality labels.
// Caller-owned inputs remain immutable; errors return no partial bank.
func BuildCompressionBank(ctx context.Context, artifact Artifact, round *annotation.Round, files map[string][]byte,
	options CompressionBankOptions,
) (CompressionBank, error) {
	if err := ctx.Err(); err != nil {
		return CompressionBank{}, err
	}
	if err := options.Validate(); err != nil {
		return CompressionBank{}, err
	}
	prepared, err := Prepare(ctx, artifact, files)
	if err != nil {
		return CompressionBank{}, err
	}
	origins, err := compressionOrigins(ctx, artifact, round, options.AllowSimulation)
	if err != nil {
		return CompressionBank{}, err
	}
	labels := compressionLabels{basis: origins.Basis, round: origins.RoundSHA256, set: origins.SHA256,
		units: make(map[string]string, len(origins.Units)), origins: make(map[string]annotation.Origin, len(origins.Units)),
		policies: []string{"human", "generated"}, missing: "a curated unit-scoped endpoint claim"}
	for _, claim := range origins.Units {
		if claim.Origin.Scope == "unit" && slices.Contains(labels.policies, claim.Origin.Label) {
			labels.units[claim.UnitID], labels.origins[claim.UnitID] = claim.Origin.Label, claim.Origin
		}
	}
	return buildCompressionBank(ctx, artifact, prepared, labels, options)
}

// BuildCompressionBankDecisions selects references under a prepared decision
// set instead of a round, such as the cohort labels of the pattern protocol.
// A historical_snapshot label admits a unit to a historical cohort. A
// contemporary_snapshot label admits it to a contemporary one. A mixed cohort
// takes both. The reservation rule is the same as under a round.
func BuildCompressionBankDecisions(ctx context.Context, artifact Artifact, decisions annotation.DecisionSet,
	files map[string][]byte, options CompressionBankOptions,
) (CompressionBank, error) {
	if err := ctx.Err(); err != nil {
		return CompressionBank{}, err
	}
	if err := options.Validate(); err != nil {
		return CompressionBank{}, err
	}
	if err := MatchDecisionTargets(ctx, artifact, decisions); err != nil {
		return CompressionBank{}, err
	}
	if decisions.Basis == "simulation" && !options.AllowSimulation {
		return CompressionBank{}, fmt.Errorf("simulated reference claims require explicit allow_simulation")
	}
	prepared, err := Prepare(ctx, artifact, files)
	if err != nil {
		return CompressionBank{}, err
	}
	return buildCompressionBank(ctx, artifact, prepared, cohortSeedLabels(artifact, decisions), options)
}

// cohortSeedLabels maps every resolved snapshot label to its cohort policy,
// with the unit's declared origin alongside.
func cohortSeedLabels(artifact Artifact, decisions annotation.DecisionSet) compressionLabels {
	labels := compressionLabels{basis: decisions.Basis, round: decisions.RoundSHA256, set: decisions.SHA256,
		units: make(map[string]string, len(decisions.Units)), origins: make(map[string]annotation.Origin, len(decisions.Units)),
		policies: []string{"historical", "contemporary"}, missing: "a resolved historical or contemporary label"}
	declared := make(map[string]annotation.Origin, len(artifact.Units))
	for _, candidate := range artifact.Units {
		declared[candidate.Unit.ID] = candidate.Unit.Origin
	}
	for _, decision := range decisions.Units {
		if decision.Status != "resolved" || decision.Label == nil {
			continue
		}
		if policy, ok := strings.CutSuffix(*decision.Label, "_snapshot"); ok && slices.Contains(labels.policies, policy) {
			labels.units[decision.UnitID], labels.origins[decision.UnitID] = policy, declared[decision.UnitID]
		}
	}
	return labels
}

func buildCompressionBank(ctx context.Context, artifact Artifact, prepared Prepared, labels compressionLabels,
	options CompressionBankOptions,
) (CompressionBank, error) {
	builder := newCompressionBankBuilder(artifact, prepared, labels, options)
	for _, cohort := range options.Cohorts {
		if err := builder.addCohort(ctx, cohort); err != nil {
			return CompressionBank{}, err
		}
	}
	if err := builder.reserve(ctx, artifact); err != nil {
		return CompressionBank{}, err
	}
	return finishCompressionBank(ctx, builder.result)
}

// LoadCompressionBank restores a bank and checks its digest, version, and
// options. It does not verify the reference prose against any source.
func LoadCompressionBank(ctx context.Context, data []byte) (CompressionBank, error) {
	var result CompressionBank
	limits := jsoninput.Limits{Array: MaxUnits, Object: 64}
	if err := jsoninput.Decode(ctx, data, MaxCompressionBankBytes, &result, limits); err != nil {
		return CompressionBank{}, err
	}
	want := result.SHA256
	result.SHA256 = ""
	rebuilt, err := finishCompressionBank(ctx, result)
	if err != nil {
		return CompressionBank{}, err
	}
	if rebuilt.SHA256 != want || rebuilt.Version != CompressionBankVersion || rebuilt.HumanCorpus != "not_qualified" {
		return CompressionBank{}, fmt.Errorf("compression bank digest, version, or status mismatch")
	}
	if err := validateLoadedBank(rebuilt); err != nil {
		return CompressionBank{}, err
	}
	return rebuilt, nil
}

func validateLoadedBank(bank CompressionBank) error {
	if err := bank.Options.Validate(); err != nil {
		return err
	}
	if len(bank.Cohorts) != len(bank.Options.Cohorts) {
		return fmt.Errorf("compression bank cohorts do not match its options")
	}
	for i, cohort := range bank.Cohorts {
		if cohort.ID != bank.Options.Cohorts[i].ID || cohort.Reference == "" ||
			cohort.Identity.ReferenceSHA256 != hashBytes([]byte(cohort.Reference+"\n")) {
			return fmt.Errorf("compression bank reference %s does not match its identity", cohort.ID)
		}
	}
	return nil
}

func finishCompressionBank(ctx context.Context, result CompressionBank) (CompressionBank, error) {
	data, err := json.Marshal(result)
	if err != nil {
		return CompressionBank{}, err
	}
	if len(data)+128 > MaxCompressionBankBytes {
		return CompressionBank{}, fmt.Errorf("compression bank exceeds %d bytes", MaxCompressionBankBytes)
	}
	// Detach every nested provenance and policy slice from caller-owned inputs.
	var owned CompressionBank
	if err := json.Unmarshal(data, &owned); err != nil {
		return CompressionBank{}, err
	}
	owned.SHA256 = hashBytes(data)
	if err := ctx.Err(); err != nil {
		return CompressionBank{}, err
	}
	return owned, nil
}
