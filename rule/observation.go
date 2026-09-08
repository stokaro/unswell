package rule

import "github.com/stokaro/unswell/feature"

// Observer validates applicability recorded during the current rule evaluation.
// Errors latch in the engine, even if an implementation ignores a returned error.
type Observer interface {
	Observe(feature.BlockObservation) error
}

// Observe reports applicability when the caller requested activation collection.
// No observer means the optional collection was not requested.
func (v View) Observe(observation feature.BlockObservation) error {
	if v.Observer == nil {
		return nil
	}
	return v.Observer.Observe(observation)
}
