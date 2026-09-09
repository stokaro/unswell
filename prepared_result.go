package unswell

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

// PreparedFeatureCollectionVersion identifies optional prepared measurements.
const PreparedFeatureCollectionVersion = "unswell-prepared-feature-collection-v1"

// PreparedFeatureCollection contains descriptive values, never model scores.
// The enclosing run's completion state applies to every collection consumer.
type PreparedFeatureCollection struct {
	Version         string                  `json:"version"`
	FeatureContract string                  `json:"feature_contract"`
	UnitContract    string                  `json:"unit_contract"`
	Requested       []string                `json:"requested"`
	Kinds           []string                `json:"kinds"`
	Sources         []PreparedFeatureSource `json:"sources"`
}

// PreparedFeatureSource binds actual source, policy, extraction, and NLP inputs.
// PreparationHash covers UnitContract, extraction policy, and both switches.
// PolicyHash and VocabularyHash conservatively bind the entire effective policy.
type PreparedFeatureSource struct {
	TargetCount          int                   `json:"target_count"`
	Path                 string                `json:"path"`
	SourceHash           string                `json:"source_hash"`
	PolicyHash           string                `json:"policy_hash"`
	VocabularyHash       string                `json:"vocabulary_hash"`
	ExtractionPolicyHash string                `json:"extraction_policy_hash"`
	PreparationHash      string                `json:"preparation_hash"`
	IncludeQuotes        bool                  `json:"include_quotes"`
	IncludeStructure     bool                  `json:"include_structure"`
	NLP                  nlp.Identity          `json:"nlp"`
	Capabilities         []nlp.Capability      `json:"capabilities"`
	Units                []PreparedFeatureUnit `json:"units"`
}

// PreparedFeatureUnit keeps complete target/context mappings separate from the
// counted token segments. Its parent block ID is not a unique target identifier.
type PreparedFeatureUnit struct {
	Binding   nlp.UnitBinding `json:"binding"`
	InputHash string          `json:"input_hash"`
	Segments  []document.Span `json:"segments"`
	Values    []feature.Value `json:"values"`
}
