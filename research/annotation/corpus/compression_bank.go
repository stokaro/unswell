package corpus

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/research/annotation"
)

// CompressionBankVersion identifies explicit cohort selection and group reservation.
const CompressionBankVersion = "unswell-compression-bank-v1"

// Compression-bank limits bound explicit selection and the serialized developer artifact.
const (
	MaxCompressionSelectionBytes = 256 << 10
	MaxCompressionBankBytes      = 4 << 20
)

// CompressionCohort preserves an explicit reference order and endpoint policy.
// Origin accepts human, generated, or mixed (a mixture of the two endpoint classes).
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
// Start and End address the reference string, not the original file.
type CompressionReferenceUnit struct {
	UnitID   string            `json:"unit_id"`
	SourceID string            `json:"source_id"`
	GroupID  string            `json:"group_id"`
	Path     string            `json:"path"`
	Source   annotation.Source `json:"source"`
	Binding  nlp.UnitBinding   `json:"binding"`
	Origin   annotation.Origin `json:"origin"`
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
	builder := newCompressionBankBuilder(artifact, prepared, origins, options)
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
