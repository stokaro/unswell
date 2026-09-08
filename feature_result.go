package unswell

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

// FeatureCollectionVersion identifies the optional saved collection structure.
const FeatureCollectionVersion = "unswell-feature-collection-v1"

// FeatureCollection contains requested descriptive measurements, not model scores.
// The enclosing RunResult completion state applies: partial runs are not complete
// training or inference inputs. Values never include normalized words or keys.
type FeatureCollection struct {
	Version       string          `json:"version"`
	BlockContract string          `json:"block_contract"`
	Requested     []string        `json:"requested"`
	Sources       []FeatureSource `json:"sources"`
}

// FeatureSource identifies the actual representation and policy used for a source.
type FeatureSource struct {
	Path           string           `json:"path"`
	SourceHash     string           `json:"source_hash"`
	PolicyHash     string           `json:"policy_hash"`
	VocabularyHash string           `json:"vocabulary_hash"`
	Preprocessing  string           `json:"preprocessing"`
	NLP            nlp.Identity     `json:"nlp"`
	Capabilities   []nlp.Capability `json:"capabilities"`
	Units          []FeatureUnit    `json:"units"`
}

// FeatureUnit records one original block and its requested numeric values.
// Segments contain counted source tokens, without filling protected gaps. An
// unsupported kind has no input hash or segments and only absent values.
type FeatureUnit struct {
	Scope       string          `json:"scope"`
	UnitID      int             `json:"unit_id"`
	Kind        string          `json:"kind"`
	ContextHash string          `json:"context_hash"`
	Span        document.Span   `json:"span"`
	Segments    []document.Span `json:"segments"`
	InputHash   string          `json:"input_hash,omitempty"`
	Values      []feature.Value `json:"values"`
}
