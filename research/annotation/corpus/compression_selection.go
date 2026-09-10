package corpus

import (
	"context"
	_ "embed"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/internal/jsoninput"
)

//go:embed compression-bank.schema.json
var compressionSelectionSchema []byte

// LoadCompressionBankOptions requires explicit settings and rejects ambiguous JSON.
func LoadCompressionBankOptions(ctx context.Context, data []byte) (CompressionBankOptions, error) {
	var options CompressionBankOptions
	limits := jsoninput.Limits{Array: 128, Object: 16, Arrays: map[string]int{"cohorts": 8}}
	if err := jsoninput.Decode(ctx, data, MaxCompressionSelectionBytes, &options, limits); err != nil {
		return CompressionBankOptions{}, err
	}
	if err := jsoninput.Schema(data, compressionSelectionSchema, "urn:unswell:compression-bank:v1"); err != nil {
		return CompressionBankOptions{}, err
	}
	if err := options.Validate(); err != nil {
		return CompressionBankOptions{}, err
	}
	if err := ctx.Err(); err != nil {
		return CompressionBankOptions{}, err
	}
	return options, nil
}

// Validate checks explicit cohort IDs, endpoint policies, and compressor limits.
func (o CompressionBankOptions) Validate() error {
	if o.Version != CompressionBankVersion || !slices.Contains([]string{"sentence", "paragraph", "fragment"}, o.Kind) {
		return fmt.Errorf("unsupported compression bank version or target kind")
	}
	if err := o.Compression.Validate(); err != nil {
		return err
	}
	return validateCompressionCohorts(o.Cohorts)
}

func validateCompressionCohorts(cohorts []CompressionCohort) error {
	if len(cohorts) == 0 || len(cohorts) > 8 {
		return fmt.Errorf("compression bank requires 1 to 8 cohorts")
	}
	seen := make(map[string]bool)
	for _, cohort := range cohorts {
		if !text(cohort.ID) || len(cohort.ID) > 128 || seen[cohort.ID] {
			return fmt.Errorf("compression cohort IDs must be nonempty, unique, and at most 128 bytes")
		}
		seen[cohort.ID] = true
		if err := cohort.validate(); err != nil {
			return err
		}
	}
	return nil
}

func (c CompressionCohort) validate() error {
	if !slices.Contains([]string{"human", "generated", "mixed"}, c.Origin) || len(c.UnitIDs) == 0 || len(c.UnitIDs) > 128 {
		return fmt.Errorf("compression cohort requires an endpoint policy and 1 to 128 unit IDs")
	}
	seen := make(map[string]bool)
	for _, id := range c.UnitIDs {
		if !text(id) || len(id) > 128 || seen[id] {
			return fmt.Errorf("reference unit IDs must be nonempty, unique, and at most 128 bytes")
		}
		seen[id] = true
	}
	return nil
}

func compressionOrigins(ctx context.Context, artifact Artifact, round *annotation.Round,
	allowSimulation bool,
) (annotation.OriginSet, error) {
	expected := make([]annotation.Unit, 0, len(artifact.Units))
	for _, candidate := range artifact.Units {
		expected = append(expected, candidate.Unit)
	}
	if err := round.MatchTargets(ctx, expected); err != nil {
		return annotation.OriginSet{}, err
	}
	origins, err := round.Origins(ctx)
	if err != nil {
		return annotation.OriginSet{}, err
	}
	if origins.Basis == "simulation" && !allowSimulation {
		return annotation.OriginSet{}, fmt.Errorf("simulated reference claims require explicit allow_simulation")
	}
	return origins, nil
}
