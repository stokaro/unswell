package feature

import (
	"fmt"
	"slices"

	"github.com/stokaro/unswell/nlp"
)

// CompressionCatalog describes the optional reference-dependent measurements.
// These IDs are not available in Catalog or model-free feature collection.
func CompressionCatalog(kind string) ([]Descriptor, error) {
	if !slices.Contains([]string{"sentence", "paragraph", "fragment"}, kind) {
		return nil, fmt.Errorf("unsupported compression target kind %q", kind)
	}
	result := make([]Descriptor, 0, 2)
	for _, row := range []struct{ id, formula string }{
		{"compression.incremental-bytes", "(C(reference+LF+target)-C(reference+LF))/UTF8Bytes(target)."},
		{"compression.reference-gain", "((C(LF+target)-C(LF))-(C(reference+LF+target)-C(reference+LF)))/UTF8Bytes(target)."},
	} {
		result = append(result, Descriptor{ID: row.id, Version: "1", Family: "compression", Type: "number",
			Unit: "compressed-bytes/input-byte", Scope: kind, Formula: row.formula, Normalization: CompressionContract,
			Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}, MinWords: 1,
			Missing:     "No value for nonpositive increments; malformed targets, incompatible identities, and resource failures are errors.",
			Limitations: "Reference-dependent byte compression, not semantic similarity, editorial quality, or origin probability."})
	}
	return result, nil
}
