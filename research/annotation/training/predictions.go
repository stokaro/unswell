package training

import "github.com/stokaro/unswell/research/annotation/corpus"

// PredictionVersion identifies the frozen plan and label-free predictions.
const PredictionVersion = "unswell-research-predictions-v1"

// PredictionPlan freezes a numerical trial before its evaluation labels are read.
// ProtocolSHA256 binds separately retained protocol bytes; it is not an attestation
// of pre-registration or a substitute for the research protocol's full run manifest.
// Context is prepared_piece or source_document. Response is logistic or isotonic.
type PredictionPlan struct {
	Version        string   `json:"version"`
	ID             string   `json:"id"`
	ProtocolSHA256 string   `json:"protocol_sha256"`
	ModelSHA256    string   `json:"model_sha256"`
	CorpusSHA256   string   `json:"corpus_sha256"`
	Partition      string   `json:"partition"`
	Context        string   `json:"context"`
	Response       string   `json:"response"`
	Threshold      *float64 `json:"threshold"`
}

// Prediction retains one selected target, including missing target/feature or
// calibration support. Numerical outputs are null when unavailable. A threshold
// decision is an experimental classifier response, never the product gate.
type Prediction struct {
	UnitID           string   `json:"unit_id"`
	SourceID         string   `json:"source_id"`
	GroupID          string   `json:"group_id"`
	FeatureInputHash string   `json:"feature_input_hash"`
	Status           string   `json:"status"`
	Reason           string   `json:"reason"`
	LinearScore      *float64 `json:"linear_score"`
	LogisticResponse *float64 `json:"logistic_response"`
	Response         *float64 `json:"response"`
	Positive         *bool    `json:"positive"`
}

// Predictions is an immutable-by-convention saved run without source text or
// labels. Loading validates the digest, not provenance or execution authenticity.
type Predictions struct {
	Version           string              `json:"version"`
	SHA256            string              `json:"sha256,omitempty"`
	Status            string              `json:"status"`
	ProbabilityStatus string              `json:"probability_status"`
	Plan              PredictionPlan      `json:"plan"`
	PlanSHA256        string              `json:"plan_sha256"`
	Model             Artifact            `json:"model"`
	Verification      corpus.Verification `json:"verification"`
	Rows              []Prediction        `json:"rows"`
}
