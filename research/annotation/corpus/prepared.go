package corpus

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/nlp"
)

// PreparedTarget associates a reproduced candidate ID with its original NLP unit.
// Unit accessors expose source-derived prose; this is explicit developer data.
type PreparedTarget struct {
	UnitID string
	Unit   nlp.PreparedUnit
}

// Prepared contains the same units used by Build, after complete verification.
// Targets retain candidate order. Neither labels nor inferred permissions enter it.
type Prepared struct {
	Policy       config.Policy
	Verification Verification
	Targets      []PreparedTarget
}

// Prepare repeats the shared corpus extraction and NLP path, then verifies every
// candidate before returning its prepared unit. It does not retokenize targets.
// Retained NLP data has a conservative 256 MiB accounting limit in addition to
// Build's limits. Errors, including cancellation, return no partial targets.
func Prepare(ctx context.Context, artifact Artifact, files map[string][]byte) (Prepared, error) {
	if err := artifact.validate(ctx); err != nil {
		return Prepared{}, err
	}
	var units []nlp.PreparedUnit
	var retained int64
	rebuilt, err := build(ctx, artifact.Plan, files, func(unit nlp.PreparedUnit) error {
		block := unit.Block()
		retained += int64(len(block.Text))*32 + int64(len(unit.Context()))
		for _, sentence := range block.Sentences {
			retained += int64(len(sentence.Tokens))*256 + int64(len(sentence.Text))
		}
		if retained > 256<<20 {
			return fmt.Errorf("prepared corpus exceeds retained NLP budget")
		}
		units = append(units, unit)
		return ctx.Err()
	})
	if err != nil {
		return Prepared{}, err
	}
	verification, err := verifyRebuilt(artifact, rebuilt)
	if err != nil {
		return Prepared{}, err
	}
	policy, err := preparedPolicy(artifact.Plan)
	if err != nil {
		return Prepared{}, err
	}
	result := Prepared{Policy: policy, Verification: verification, Targets: make([]PreparedTarget, len(units))}
	for i, unit := range units {
		result.Targets[i] = PreparedTarget{UnitID: rebuilt.Units[i].Unit.ID, Unit: unit}
	}
	if err := ctx.Err(); err != nil {
		return Prepared{}, err
	}
	return result, nil
}

func preparedPolicy(plan Plan) (config.Policy, error) {
	configuration, err := measurementConfig(plan)
	if err != nil {
		return config.Policy{}, err
	}
	engine, err := unswell.New(unswell.Options{Config: configuration, Jobs: 1, AllowEmpty: true})
	if err != nil {
		return config.Policy{}, err
	}
	return engine.PolicyForFile("")
}
