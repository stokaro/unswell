// Package feature computes versioned descriptive measurements from extracted prose.
// It does not extract text, infer authorship, apply policy, or load resources.
package feature

import (
	"errors"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

// Contract identifies the block measurement definitions and missing-value rules.
const Contract = "unswell-block-features-v1"

// ErrTokenLimit identifies an exhausted token visit budget, including set totals.
var ErrTokenLimit = errors.New("features exceed max_tokens")

// Descriptor specifies a formula independently of its consumers' thresholds.
type Descriptor struct {
	ID, Version, Family, Type, Unit, Scope string
	Formula, Normalization, Limitations    string
	Requires                               []nlp.Capability
	MinWords                               int
	Missing                                string
}

// Value contains either a finite number or a machine-readable absence reason.
// Number is nil for unavailable values; a present zero is an observed value.
type Value struct {
	ID      string   `json:"id"`
	Version string   `json:"version"`
	Unit    string   `json:"unit"`
	Number  *float64 `json:"number"`
	Reason  string   `json:"reason,omitempty"`
}

// Identity records the effective inputs, including capabilities actually requested
// from NLP. Provider support alone does not establish that a value was computed.
// Policy and vocabulary identities must change when their effective inputs change.
type Identity struct {
	NLP           nlp.Identity
	Capabilities  []nlp.Capability
	Source        string
	Policy        string
	Vocabulary    string
	Preprocessing string
}

// Limits bounds one computation. Set construction applies MaxTokens and
// MaxBytes across all included blocks. Every input token consumes one visit,
// including punctuation and protected tokens; only eligible words enter ratios.
type Limits struct {
	MaxTokens      int
	MaxUniqueWords int
	MaxBytes       int
	MaxBlocks      int
}

// Counts contains exact counts of eligible prose and the logical token cost.
// Available is false when the required token/sentence representation is absent.
type Counts struct {
	Available                                 bool
	Words, Characters, Sentences, TokenVisits int
}

type number struct {
	value  float64
	reason string
}

// Measurements owns one block's values, sentence lengths, and original segments.
// Its zero value has no available measurements. All methods permit concurrent use.
type Measurements struct {
	counts  Counts
	lengths []int
	spans   []document.Span
	values  []number
	hash    string
}

// Counts returns eligible prose counts and the logical token visit cost.
func (m Measurements) Counts() Counts { return m.counts }

// SentenceLengths returns owned lengths of sentences containing eligible words.
func (m Measurements) SentenceLengths() []int { return slices.Clone(m.lengths) }

// Spans returns owned original byte segments of counted tokens, excluding gaps.
func (m Measurements) Spans() []document.Span { return slices.Clone(m.spans) }

// Hash covers the input unit, effective identities, and feature definitions.
func (m Measurements) Hash() string { return m.hash }

// Value returns one known measurement. Unknown IDs return an error.
func (m Measurements) Value(id string) (Value, error) {
	for i, descriptor := range catalog() {
		if descriptor.ID == id {
			return m.value(i, descriptor), nil
		}
	}
	return Value{}, fmt.Errorf("unknown feature %q", id)
}

// Values returns owned values in the stable contract order.
func (m Measurements) Values() []Value {
	descriptors := catalog()
	values := make([]Value, len(descriptors))
	for i, descriptor := range descriptors {
		values[i] = m.value(i, descriptor)
	}
	return values
}

func (m Measurements) value(index int, d Descriptor) Value {
	value := Value{ID: d.ID, Version: d.Version, Unit: d.Unit}
	if index >= len(m.values) {
		value.Reason = "not_computed"
		return value
	}
	n := m.values[index]
	value.Reason = n.reason
	if n.reason == "" {
		value.Number = &n.value
	}
	return value
}

// Catalog returns owned definitions in the stable measurement order.
func Catalog() []Descriptor { return catalog() }

func catalog() []Descriptor {
	var result []Descriptor
	for _, row := range definitions() {
		d := Descriptor{ID: row.id, Version: "1", Family: row.family, Type: "number", Unit: row.unit, Scope: "block",
			Formula: row.formula, MinWords: row.minimum, Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences},
			Normalization: "Eligible Word tokens, excluding Protected; frequencies use provider Normal unchanged.",
			Missing:       "No imputation: empty ratios and absent capabilities have no numeric value.",
			Limitations:   "Descriptive only; length and technical terminology can affect these values without indicating a defect."}
		if row.family == "pos" {
			d.Requires = append(d.Requires, nlp.POS)
		}
		result = append(result, d)
	}
	return result
}

type definition struct {
	id, family, unit, formula string
	minimum                   int
}

func definitions() []definition {
	return []definition{
		{"prose-words", "structure", "words", "Count eligible words.", 0},
		{"counted-characters", "structure", "letters-and-digits", "Count Unicode letters and digits in eligible words.", 0},
		{"prose-sentences", "structure", "sentences", "Count sentences with at least one eligible word.", 0},
		{"mean-sentence-words", "structure", "words/sentence", "Mean eligible word count of nonempty sentences.", 1},
		{"sentence-word-stddev", "structure", "words", "Population standard deviation of nonempty sentence lengths using Welford.", 1},
		{"shortest-sentence", "structure", "words", "Minimum nonempty sentence length.", 1},
		{"longest-sentence", "structure", "words", "Maximum nonempty sentence length.", 1},
		{"type-token-ratio", "lexical", "ratio", "Distinct normalized eligible words / eligible words.", 1},
		{"hapax-token-ratio", "lexical", "ratio", "Normalized types occurring once / eligible words.", 1},
		{"noun-token-ratio", "pos", "ratio", "Eligible tokens with tag prefix NN / eligible words.", 1},
		{"verb-token-ratio", "pos", "ratio", "Eligible tokens with tag prefix VB / eligible words.", 1},
		{"adjective-token-ratio", "pos", "ratio", "Eligible tokens with tag prefix JJ / eligible words.", 1},
		{"adverb-token-ratio", "pos", "ratio", "Eligible tokens with tag prefix RB / eligible words.", 1},
		{"automated-readability-index", "readability", "ARI-formula-units", "4.71 * characters / words + 0.5 * words / sentences - 21.43.", 1},
	}
}
