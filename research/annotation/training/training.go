// Package training fits experimental Go models from reproduced annotation targets.
// It does not qualify a corpus, expose editorial probabilities, or load resources.
package training

import (
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/nlp"
)

// Version identifies the experimental training artifact and selection semantics.
const Version = "unswell-editorial-training-v2"

// MaxArtifactBytes bounds the compact serialized training result.
const MaxArtifactBytes = 16 << 20

// Options selects one scope and declares fitting, missing-data, and simulation policy.
// MissingFeatures accepts reject or exclude; Calibration accepts none or isotonic.
type Options struct {
	Kind            string   `json:"kind"`
	Features        []string `json:"features"`
	MissingFeatures string   `json:"missing_features"`
	Calibration     string   `json:"calibration"`
	AllowSimulation bool     `json:"allow_simulation"`
	Fit             Fit      `json:"fit"`
}

// Fit is the serialized configuration of the shared numerical optimizer.
type Fit struct {
	L2            float64 `json:"l2"`
	Tolerance     float64 `json:"tolerance"`
	MaxIterations int     `json:"max_iterations"`
	MaxOperations int64   `json:"max_operations"`
}

func (f Fit) numerical() model.FitOptions {
	return model.FitOptions{L2: f.L2, Tolerance: f.Tolerance, MaxIterations: f.MaxIterations, MaxOperations: f.MaxOperations}
}

// Identity binds the training target and effective representations. Columns use
// the shared descriptor serialization; ColumnsSHA256 freezes its exact definition.
type Identity struct {
	LexicalVocabularyHash string               `json:"lexical_vocabulary_hash,omitempty"`
	FeatureSource         string               `json:"feature_source,omitempty"`
	Context               string               `json:"context,omitempty"`
	ActivationContract    string               `json:"activation_contract,omitempty"`
	RulesetHash           string               `json:"ruleset_hash,omitempty"`
	RuleConfigSHA256      string               `json:"rule_config_sha256,omitempty"`
	Preprocessing         string               `json:"preprocessing,omitempty"`
	Task                  string               `json:"task"`
	Rubric                string               `json:"rubric"`
	ProfileSHA256         string               `json:"profile_sha256"`
	Kind                  string               `json:"kind"`
	FeatureContract       string               `json:"feature_contract"`
	UnitContract          string               `json:"unit_contract"`
	Columns               []feature.Descriptor `json:"columns"`
	ColumnsSHA256         string               `json:"columns_sha256"`
	NLP                   nlp.Identity         `json:"nlp"`
	Capabilities          []nlp.Capability     `json:"capabilities"`
	PolicyHash            string               `json:"policy_hash"`
	VocabularyHash        string               `json:"vocabulary_hash"`
	ExtractionPolicyHash  string               `json:"extraction_policy_hash"`
	PreparationHash       string               `json:"preparation_hash"`
	IncludeQuotes         bool                 `json:"include_quotes"`
	IncludeStructure      bool                 `json:"include_structure"`
}

// Row identifies an actually fitted row without exposing its label or vector.
type Row struct {
	UnitID           string `json:"unit_id"`
	SourceID         string `json:"source_id"`
	GroupID          string `json:"group_id"`
	FeatureInputHash string `json:"feature_input_hash"`
}

// Partition reports selection and exclusions. Class counts describe fitted rows
// only; reserved partitions have no class statistics or predictions.
type Partition struct {
	Name       string         `json:"name"`
	Candidates int            `json:"candidates"`
	Sources    int            `json:"sources"`
	Groups     int            `json:"groups"`
	Excluded   map[string]int `json:"excluded"`
	Rows       []Row          `json:"rows"`
	Classes    map[string]int `json:"classes"`
}

// Logistic contains numerical parameters and the observed convergence result.
type Logistic struct {
	Algorithm    string    `json:"algorithm"`
	InputSHA256  string    `json:"input_sha256"`
	Means        []float64 `json:"means"`
	Scales       []float64 `json:"scales"`
	Weights      []float64 `json:"weights"`
	Intercept    float64   `json:"intercept"`
	Iterations   int       `json:"iterations"`
	Operations   int64     `json:"operations"`
	Loss         float64   `json:"training_loss"`
	GradientNorm float64   `json:"gradient_norm"`
}

// Calibration records a fit over separate linear scores, never held-out quality.
type Calibration struct {
	Algorithm        string    `json:"algorithm"`
	InputSHA256      string    `json:"input_sha256"`
	ScoreKind        string    `json:"score_kind"`
	Scores           []float64 `json:"scores"`
	Responses        []float64 `json:"responses"`
	Samples          int       `json:"samples"`
	DistinctScores   int       `json:"distinct_scores"`
	Pools            int       `json:"pools"`
	MeanSquaredError float64   `json:"calibration_fit_mean_squared_error"`
}

// Artifact is a numerical experiment, not a model pack accepted by normal scans.
// SHA256 covers compact Go JSON with that field omitted. It is not an attestation.
type Artifact struct {
	Version           string       `json:"version"`
	Status            string       `json:"status"`
	SHA256            string       `json:"sha256,omitempty"`
	HumanCorpus       string       `json:"human_corpus"`
	ProbabilityStatus string       `json:"probability_status"`
	Basis             string       `json:"basis"`
	CorpusSHA256      string       `json:"corpus_sha256"`
	ManifestSHA256    string       `json:"manifest_sha256"`
	RoundSHA256       string       `json:"round_sha256"`
	JoinedSHA256      string       `json:"joined_sha256"`
	Options           Options      `json:"options"`
	Identity          Identity     `json:"identity"`
	Partitions        []Partition  `json:"partitions"`
	Logistic          Logistic     `json:"logistic"`
	Calibration       *Calibration `json:"calibration"`
	Lexical           *Vocabulary  `json:"lexical,omitempty"`
}
