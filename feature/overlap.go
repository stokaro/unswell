package feature

import (
	"context"
	"fmt"

	"github.com/stokaro/unswell/nlp"
)

// Overlap contains observed intersection and union sizes. An empty union has no
// defined Jaccard value; it is not evidence of identical prose.
type Overlap struct {
	intersection, union int
	available           bool
}

// CompareWords measures two available sets without changing them. Source kind,
// negation, number, identifier, term, and window guards belong to the caller.
// It does not qualify pairs as semantically equivalent or editorial defects.
func CompareWords(ctx context.Context, left, right WordSet) (Overlap, error) {
	if err := ctx.Err(); err != nil {
		return Overlap{}, err
	}
	if !left.Available() || !right.Available() {
		return Overlap{}, fmt.Errorf("lexical comparison requires two available word sets")
	}
	if left.Len() > right.Len() {
		left, right = right, left
	}
	shared := 0
	for word := range left.words {
		if err := ctx.Err(); err != nil {
			return Overlap{}, err
		}
		if _, exists := right.words[word]; exists {
			shared++
		}
	}
	return Overlap{intersection: shared, union: left.Len() + right.Len() - shared, available: true}, nil
}

// Available distinguishes an observed comparison from the zero value.
func (o Overlap) Available() bool { return o.available }

// Intersection returns the observed shared word count; check Available first.
func (o Overlap) Intersection() int { return o.intersection }

// Union returns the observed union size; check Available first.
func (o Overlap) Union() int { return o.union }

// Values returns the three observed pair features in LexicalCatalog order.
func (o Overlap) Values() []Value {
	if !o.available {
		return []Value{
			{ID: "word-intersection", Version: "1", Unit: "distinct-words", Reason: "not_computed"},
			{ID: "word-union", Version: "1", Unit: "distinct-words", Reason: "not_computed"},
			{ID: "word-jaccard", Version: "1", Unit: "ratio", Reason: "not_computed"},
		}
	}
	intersection, union := float64(o.intersection), float64(o.union)
	ratio := Value{ID: "word-jaccard", Version: "1", Unit: "ratio", Reason: "empty_union"}
	if o.union > 0 {
		value := intersection / union
		ratio.Number, ratio.Reason = &value, ""
	}
	return []Value{
		{ID: "word-intersection", Version: "1", Unit: "distinct-words", Number: &intersection},
		{ID: "word-union", Version: "1", Unit: "distinct-words", Number: &union}, ratio,
	}
}

// LexicalCatalog returns owned definitions for observed word-set pair features.
// Input word selection and normalization must be identified by the caller.
func LexicalCatalog() []Descriptor {
	var result []Descriptor
	for _, row := range []struct{ id, unit, formula string }{
		{"word-intersection", "distinct-words", "Size of the intersection of two distinct normalized word sets."},
		{"word-union", "distinct-words", "Size of the union of two distinct normalized word sets."},
		{"word-jaccard", "ratio", "Intersection size / union size; undefined when the union is empty."},
	} {
		result = append(result, Descriptor{ID: row.id, Version: "1", Family: "repetition", Type: "number", Unit: row.unit,
			Scope: "pair", Formula: row.formula, Requires: []nlp.Capability{nlp.Tokens},
			Normalization: "Caller-selected normalized words; candidate stopwords remain in the denominator.",
			Limitations:   "Lexical overlap is not semantic equivalence or evidence of an editorial defect.",
			Missing:       "Unavailable input sets are errors; Jaccard is absent with empty_union for an empty union."})
	}
	return result
}
