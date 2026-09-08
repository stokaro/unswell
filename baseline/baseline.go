// Package baseline records explicit acceptance of existing editorial debt.
// Callers supply complete engine snapshots and own all file and policy selection.
// This package neither analyzes prose nor performs I/O.
package baseline

// Version identifies the baseline storage contract.
const Version = "unswell-baseline-v1"

// FingerprintVersion identifies the structural identity and hashing contract.
const FingerprintVersion = "unswell-structural-v1"

// Limits bound artifact parsing and comparison independently of source limits.
const (
	MaxBytes     = 16 << 20
	MaxEntries   = 20000
	MaxDocuments = 10000
)

// Compatibility identifies analysis behavior, excluding presentation and gate mode.
// Hash fields are lowercase SHA-256 values; absent calibration still has an identity.
type Compatibility struct {
	PolicyHash      string `json:"policy_hash"`
	RulesHash       string `json:"rules_hash"`
	NLPHash         string `json:"nlp_hash"`
	FeatureContract string `json:"feature_contract"`
	ScoringContract string `json:"scoring_contract"`
	ModelHash       string `json:"model_hash"`
}

// Document records complete scan coverage and source-specific compatibility.
// SourceHash is provenance, not identity: unrelated edits can preserve accepted debt.
type Document struct {
	Path            string `json:"path"`
	Format          string `json:"format"`
	SourceHash      string `json:"source_hash"`
	PolicyHash      string `json:"policy_hash"`
	SuppressionHash string `json:"suppression_hash"`
}

// Identity describes one finding or scored unit without retaining its source prose.
// Kind is finding, sentence, or paragraph. Only findings carry a rule ID and version.
type Identity struct {
	Path          string `json:"path"`
	Kind          string `json:"kind"`
	RuleID        string `json:"rule_id,omitempty"`
	RuleVersion   string `json:"rule_version,omitempty"`
	StructureHash string `json:"structure_hash"`
	ContentHash   string `json:"content_hash"`
	EvidenceHash  string `json:"evidence_hash"`
}

// Entry is explicit acceptance of one exact identity under the file's compatibility.
// Status must be accepted; missing or unknown states never grant acceptance.
type Entry struct {
	Identity    Identity `json:"identity"`
	Fingerprint string   `json:"fingerprint"`
	Status      string   `json:"status"`
}

// File is the versioned accepted-debt artifact. Use Load to validate stored bytes.
type File struct {
	Version            string        `json:"version"`
	FingerprintVersion string        `json:"fingerprint_version"`
	Compatibility      Compatibility `json:"compatibility"`
	Documents          []Document    `json:"documents"`
	Entries            []Entry       `json:"entries"`
}

// Snapshot contains candidate debt from complete analysis of every listed document.
// It may cover a subset of baseline paths; omitted paths are never considered clean.
type Snapshot struct {
	Complete      bool          `json:"complete"`
	Compatibility Compatibility `json:"compatibility"`
	Documents     []Document    `json:"documents"`
	Candidates    []Identity    `json:"candidates"`
}

// Match records whether a current candidate has exact accepted debt.
// State is existing or new; a match never changes a finding or its score.
type Match struct {
	Fingerprint string `json:"fingerprint"`
	State       string `json:"state"`
}

// Comparison distinguishes current matches, resolved debt, and unexamined paths.
type Comparison struct {
	Matches    []Match `json:"matches"`
	Stale      []Entry `json:"stale"`
	Unobserved []Entry `json:"unobserved"`
}
