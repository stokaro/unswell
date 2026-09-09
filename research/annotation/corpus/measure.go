package corpus

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/rule"
)

// Measurements reproduces prepared features without accepting annotation labels.
type Measurements struct {
	Verification Verification
	Features     unswell.PreparedFeatureCollection
	Bindings     []FeatureBinding
}

// RuleMeasurements reproduces original-block activations without annotation labels.
type RuleMeasurements struct {
	Verification Verification
	Rules        []rule.Descriptor
	Features     unswell.FeatureCollection
	Bindings     []RuleFeatureBinding
}

// Measure verifies source groups and exact targets before computing prepared
// features. It shares Join's measurement path and accepts no editorial decisions.
func Measure(ctx context.Context, artifact Artifact, files map[string][]byte, features []string) (Measurements, error) {
	verification, err := verifyMeasurements(ctx, artifact, files)
	if err != nil {
		return Measurements{}, err
	}
	collection, err := measureCandidates(ctx, artifact.Plan, files, features)
	if err != nil {
		return Measurements{}, err
	}
	bindings, err := bindCandidates(ctx, artifact, collection)
	if err != nil {
		return Measurements{}, err
	}
	if err := ctx.Err(); err != nil {
		return Measurements{}, err
	}
	return Measurements{verification, collection, bindings}, nil
}

// MeasureRules verifies the same original blocks and policy as JoinRules. Partial
// target matches retain an absence reason. It never enables a requested rule.
func MeasureRules(ctx context.Context, artifact Artifact, files map[string][]byte,
	features []string, configuration []byte,
) (RuleMeasurements, error) {
	if len(configuration) == 0 || len(features) == 0 {
		return RuleMeasurements{}, fmt.Errorf("rule binding requires explicit configuration and activation IDs")
	}
	for _, id := range features {
		if !strings.HasPrefix(id, "activation/") {
			return RuleMeasurements{}, fmt.Errorf("rule binding accepts only activation IDs")
		}
	}
	engine, err := unswell.New(unswell.Options{Config: configuration, Features: features, Jobs: 1, AllowEmpty: true})
	if err != nil {
		return RuleMeasurements{}, err
	}
	verification, err := verifyMeasurements(ctx, artifact, files)
	if err != nil {
		return RuleMeasurements{}, err
	}
	collection, err := measureRuleSources(ctx, engine, artifact.Plan, files)
	if err != nil {
		return RuleMeasurements{}, err
	}
	bindings, err := bindRuleCandidates(ctx, artifact, collection)
	if err != nil {
		return RuleMeasurements{}, err
	}
	if err := ctx.Err(); err != nil {
		return RuleMeasurements{}, err
	}
	return RuleMeasurements{verification, engine.Catalog(), collection, bindings}, nil
}

func verifyMeasurements(ctx context.Context, artifact Artifact, files map[string][]byte) (Verification, error) {
	verification, err := Verify(ctx, artifact, files)
	if err != nil {
		return Verification{}, err
	}
	verification.Producer.Dependencies = slices.Clone(verification.Producer.Dependencies)
	verification.Producer.NLP.Capabilities = slices.Clone(verification.Producer.NLP.Capabilities)
	return verification, nil
}
