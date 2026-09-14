// Package probability decodes explicitly supplied editorial revision-probability
// packs and decides per-unit applicability before any estimate is reported.
// It performs no I/O, fits nothing, and cannot verify a declared acceptance.
package probability

import (
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
)

// Version identifies the pack contract. A training artifact is not a pack.
const Version = "unswell-probability-pack-v1"

// Task names what a pack estimates. The two targets are independent: a revision
// probability answers whether a unit needs editing, and an origin estimate is a
// similarity to a defined training class. Neither implies the other, and an
// origin estimate is never a percentage of a text written by any tool.
const (
	Task       = "editorial_needs_revision"
	TaskOrigin = "origin_endpoint"
)

// MaxBytes bounds one decoded pack. MaxColumns bounds its numerical width.
const (
	MaxBytes   = 8 << 20
	MaxColumns = model.MaxFeatures
)

// Applicability statuses. Only StatusAvailable carries an estimate; every other
// status is an explicit abstention with an unavailable probability.
const (
	StatusAvailable            = "available"
	StatusUnsupportedUnit      = "unsupported_unit"
	StatusInsufficientEvidence = "insufficient_evidence"
	StatusMissingFeature       = "missing_feature"
	StatusCalibrationRange     = "calibration_range"
	// StatusIncompatible reports that a pack does not match a source's
	// effective measurement inputs and was therefore not evaluated.
	StatusIncompatible = "incompatible_model"
	// StatusUnavailable is the model-free status of existing results. It stays
	// the reported status when no pack is configured for a run.
	StatusUnavailable = "calibration_unavailable"
)

// Limits are applicability bounds the pack author selected from validation data.
// This package records them; it cannot check how they were chosen.
type Limits struct {
	MinWords int `json:"min_words"`
	// MaxWords is the longest unit the fit admitted. A model fitted inside a
	// word band says nothing about a unit outside it, so the pack abstains
	// there rather than extrapolating. Zero leaves the bound open.
	MaxWords int `json:"max_words,omitempty"`
}

// Logistic holds normalized linear parameters in exact column order.
type Logistic struct {
	Means     []float64 `json:"means"`
	Scales    []float64 `json:"scales"`
	Weights   []float64 `json:"weights"`
	Intercept float64   `json:"intercept"`
}

// Calibration holds an increasing mapping fitted on a separate partition.
// Its presence is not evidence of held-out calibration quality.
type Calibration struct {
	Algorithm string    `json:"algorithm"`
	Scores    []float64 `json:"scores"`
	Responses []float64 `json:"responses"`
}

// Contract binds the effective run inputs that change measured column values.
// It deliberately excludes gate thresholds, severities, and report selection,
// which do not change prepared measurements.
type Contract struct {
	FeatureContract  string               `json:"feature_contract"`
	UnitContract     string               `json:"unit_contract"`
	Columns          []feature.Descriptor `json:"columns"`
	ColumnsSHA256    string               `json:"columns_sha256"`
	NLP              nlp.Identity         `json:"nlp"`
	Capabilities     []nlp.Capability     `json:"capabilities"`
	PreparationHash  string               `json:"preparation_hash"`
	IncludeQuotes    bool                 `json:"include_quotes"`
	IncludeStructure bool                 `json:"include_structure"`
}

// File is the serialized pack. Use Load to validate stored bytes.
// DeclaredStatus, HumanCorpus, and Evaluation are author declarations that this
// package records and checks for consistency, never for truth.
type File struct {
	Version        string      `json:"version"`
	SHA256         string      `json:"sha256"`
	ID             string      `json:"id"`
	DeclaredStatus string      `json:"declared_status"`
	HumanCorpus    string      `json:"human_corpus"`
	Evaluation     string      `json:"evaluation"`
	Task           string      `json:"task"`
	Rubric         string      `json:"rubric"`
	Kind           string      `json:"kind"`
	Contract       Contract    `json:"contract"`
	Limits         Limits      `json:"limits"`
	Estimator      string      `json:"estimator"`
	Logistic       *Logistic   `json:"logistic,omitempty"`
	Calibration    Calibration `json:"calibration"`
	// Vocabulary is present exactly when the pack's columns are n-gram counts
	// rather than named prepared features. Its terms are source-derived text,
	// ordered to match Contract.Columns one for one.
	Vocabulary *Vocabulary `json:"vocabulary,omitempty"`
}

// Vocabulary carries the frozen n-gram keys a lexical pack counts, and the
// counting options that decide which n-grams a unit yields at all. A caller
// measuring for such a pack must use these options; anything else produces a
// different column for the same key.
type Vocabulary struct {
	Options feature.LexicalOptions `json:"options"`
	Terms   []string               `json:"terms"`
}

// Run describes the effective analysis inputs a pack must match.
type Run struct {
	FeatureContract string
	// LexicalContract names the n-gram counting contract this build computes.
	// A pack of named features ignores it; a pack of n-gram counts is refused
	// unless it matches, because a different counter gives a different column
	// for the same key.
	LexicalContract  string
	UnitContract     string
	NLP              nlp.Identity
	Capabilities     []nlp.Capability
	PreparationHash  string
	IncludeQuotes    bool
	IncludeStructure bool
}

// Unit is one prepared target with its measured values in exact column order.
type Unit struct {
	Kind   string
	Words  int
	Values []feature.Value
}

// Estimate is one unit's decision. Probability is nil unless Status is available.
// LinearScore is an uncalibrated numerical score, never a probability.
type Estimate struct {
	Status      string   `json:"status"`
	Detail      string   `json:"detail,omitempty"`
	Probability *float64 `json:"probability"`
	LinearScore *float64 `json:"linear_score,omitempty"`
}

// Pack is an immutable validated pack. Concurrent use is safe.
type Pack struct {
	file        File
	classifier  *model.Logistic
	calibration *model.Isotonic
}
