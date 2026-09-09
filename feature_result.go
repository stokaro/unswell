package unswell

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

// FeatureCollectionVersion identifies the optional saved collection structure.
const FeatureCollectionVersion = "unswell-feature-collection-v1"

// FeatureBlockBindingContract identifies complete mapped-text block bindings.
const FeatureBlockBindingContract = "mapped-block-v1"

// FeatureBlockBinding identifies every mapped byte of an original extracted block.
// TextSHA256 includes whitespace, punctuation, and protected separators. Segments
// cover mapped text, including the source ranges of protected placeholders, not
// only counted tokens. Missing bindings cannot be inferred from token segments.
// Trimmed fields apply the same outer-whitespace trim as corpus unit preparation;
// original fields retain the actual input used by rules.
type FeatureBlockBinding struct {
	Contract        string          `json:"contract"`
	TextSHA256      string          `json:"text_sha256"`
	Segments        []document.Span `json:"segments"`
	TrimmedSHA256   string          `json:"trimmed_sha256"`
	TrimmedSegments []document.Span `json:"trimmed_segments"`
}

// FeatureCollection contains requested descriptive measurements, not model scores.
// The enclosing RunResult completion state applies: partial runs are not complete
// training or inference inputs. Values never include normalized words or keys.
type FeatureCollection struct {
	ActivationContract string          `json:"activation_contract,omitempty"`
	Version            string          `json:"version"`
	BlockContract      string          `json:"block_contract"`
	Requested          []string        `json:"requested"`
	Sources            []FeatureSource `json:"sources"`
}

// FeatureSource identifies the actual representation and policy used for a source.
type FeatureSource struct {
	RulesetHash    string           `json:"ruleset_hash,omitempty"`
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
// unsupported measurement kind has no input hash or counted segments. Rule
// activations retain the source, context, and rule identities instead.
type FeatureUnit struct {
	Binding     *FeatureBlockBinding `json:"binding,omitempty"`
	Scope       string               `json:"scope"`
	UnitID      int                  `json:"unit_id"`
	Kind        string               `json:"kind"`
	ContextHash string               `json:"context_hash"`
	Span        document.Span        `json:"span"`
	Segments    []document.Span      `json:"segments"`
	InputHash   string               `json:"input_hash,omitempty"`
	Values      []feature.Value      `json:"values"`
}
