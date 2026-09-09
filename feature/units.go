package feature

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"

	"github.com/stokaro/unswell/nlp"
)

// UnitContract identifies descriptive formulas on explicitly prepared targets.
// It does not reinterpret the existing block collection or rule activations.
const UnitContract = "unswell-prepared-unit-features-v1"

// UnitCatalog returns the shared formulas with their explicit target scope.
func UnitCatalog(kind string) ([]Descriptor, error) {
	if !slices.Contains([]string{"sentence", "paragraph", "fragment"}, kind) {
		return nil, fmt.Errorf("unsupported feature target kind %q", kind)
	}
	result := catalog()
	for i := range result {
		result[i].Scope = kind
	}
	return result, nil
}

// MeasureUnit computes shared descriptive formulas on one prepared target.
// Identity must match its actual provider and requested capabilities. The caller
// supplies source/policy/vocabulary identities; no rule activation is inherited.
func MeasureUnit(ctx context.Context, unit nlp.PreparedUnit, identity Identity, limits Limits) (Measurements, error) {
	if err := ctx.Err(); err != nil {
		return Measurements{}, err
	}
	binding := unit.Binding()
	if binding.Contract != nlp.UnitContract {
		return Measurements{}, fmt.Errorf("invalid prepared feature unit")
	}
	if err := matchUnitNLP(unit, identity); err != nil {
		return Measurements{}, err
	}
	descriptors, err := UnitCatalog(binding.Kind)
	if err != nil {
		return Measurements{}, err
	}
	block := unit.Block()
	if err := validateInputs(ctx, block, identity, limits); err != nil {
		return Measurements{}, err
	}
	base, err := unitHash(block, identity)
	if err != nil {
		return Measurements{}, err
	}
	data, err := json.Marshal(struct {
		Contract, Base string
		Binding        nlp.UnitBinding
		Definitions    []Descriptor
	}{UnitContract, base, binding, descriptors})
	if err != nil {
		return Measurements{}, err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	return measureAvailable(ctx, block, identity, limits, hash)
}

func matchUnitNLP(unit nlp.PreparedUnit, identity Identity) error {
	actual := unit.Identity()
	actual.Capabilities = orderedCapabilities(actual.Capabilities)
	wanted := identity.NLP
	wanted.Capabilities = orderedCapabilities(wanted.Capabilities)
	if !reflect.DeepEqual(actual, wanted) ||
		!slices.Equal(orderedCapabilities(unit.Capabilities()), orderedCapabilities(identity.Capabilities)) {
		return fmt.Errorf("prepared unit NLP identity or capabilities do not match feature inputs")
	}
	return nil
}
