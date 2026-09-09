package unswell

import (
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func (e *Engine) selectPreparedFeatures(ids, kinds []string) error {
	if len(ids) == 0 && len(kinds) == 0 {
		return nil
	}
	if len(ids) == 0 || len(ids) > len(feature.Catalog()) || len(kinds) == 0 || len(kinds) > 3 {
		return fmt.Errorf("prepared collection requires feature IDs and 1 through 3 target kinds")
	}
	if err := validatePreparedKinds(kinds); err != nil {
		return err
	}
	e.preparedIDs, e.preparedKinds = slices.Clone(ids), slices.Clone(kinds)
	slices.Sort(e.preparedIDs)
	slices.Sort(e.preparedKinds)
	return e.planPreparedCapabilities()
}

func validatePreparedKinds(kinds []string) error {
	for i, kind := range kinds {
		if _, err := feature.UnitCatalog(kind); err != nil {
			return err
		}
		if slices.Contains(kinds[:i], kind) {
			return fmt.Errorf("duplicate prepared kind %q", kind)
		}
	}
	return nil
}

func (e *Engine) planPreparedCapabilities() error {
	catalog, supported := feature.Catalog(), e.nlp.Identity().Capabilities
	for i, id := range e.preparedIDs {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		if index < 0 || (i > 0 && id == e.preparedIDs[i-1]) {
			return fmt.Errorf("unknown or duplicate prepared feature %q", id)
		}
		for _, capability := range catalog[index].Requires {
			if !slices.Contains(supported, capability) {
				return fmt.Errorf("prepared feature %s requires unavailable capability %s", id, capability)
			}
			e.preparedCapabilities = append(e.preparedCapabilities, capability)
		}
	}
	slices.Sort(e.preparedCapabilities)
	e.preparedCapabilities = slices.Compact(e.preparedCapabilities)
	return nil
}

// PreparedFeatureIDs returns owned, canonical IDs selected for prepared targets.
func (e *Engine) PreparedFeatureIDs() []string { return slices.Clone(e.preparedIDs) }

// PreparedUnitKinds returns owned, canonical kinds selected for preparation.
func (e *Engine) PreparedUnitKinds() []string { return slices.Clone(e.preparedKinds) }

func (e *Engine) preparedCollection() *PreparedFeatureCollection {
	if len(e.preparedIDs) == 0 {
		return nil
	}
	return &PreparedFeatureCollection{Version: PreparedFeatureCollectionVersion, FeatureContract: feature.UnitContract,
		UnitContract: nlp.UnitContract, Requested: e.PreparedFeatureIDs(), Kinds: e.PreparedUnitKinds()}
}

func (e *Engine) configureResultCollection(options Options) error {
	if err := e.selectPreparedFeatures(options.PreparedFeatures, options.PreparedKinds); err != nil {
		return err
	}
	return e.configureBaseline(options)
}
