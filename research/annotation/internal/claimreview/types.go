// Package claimreview accounts for declared editorial claims without judging prose.
// It consumes the existing extractor's source map; it has no model or gate client.
package claimreview

import "github.com/stokaro/unswell/document"

// Version identifies the experimental research accounting contract.
const Version = "unswell-editorial-claims-v1"

// Reference identifies exact original bytes in one existing prose block.
type Reference struct {
	Block int           `json:"block"`
	Span  document.Span `json:"span"`
	Quote string        `json:"quote"`
}

// Origin names the run, parent finding, and separately declared actionable claim.
// Key is assigned by the reviewer, not inferred from overlapping source ranges.
type Origin struct {
	Run       string `json:"run"`
	Candidate string `json:"candidate"`
	Key       string `json:"key"`
}

// Specification declares one criticism and separates its targets from its context.
// Binding verifies the locations, not truth, atomicity, or editorial usefulness.
type Specification struct {
	Origin     Origin      `json:"origin"`
	Category   string      `json:"category"`
	Diagnostic string      `json:"diagnostic"`
	Reason     string      `json:"reason"`
	Suggestion string      `json:"suggestion"`
	Targets    []Reference `json:"targets"`
	Support    []Reference `json:"support"`
}

// Claim retains its immutable source-bound original throughout review stages.
type Claim struct {
	ID string `json:"id"`
	Specification
}

// Inventory is bound to the exact source and extraction result supplied by a caller.
// Fields are private so a model response cannot replace its original claims.
type Inventory struct {
	document document.Document
	claims   []Claim
}

// Decision accounts for exactly one original claim in one review stage.
// Status is retained, resolved, rejected, or uncertain. Only resolved cites an edit.
type Decision struct {
	ClaimID string `json:"claim_id"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	EditID  string `json:"edit_id"`
}

// Edit is an alternative for original target bytes, never support-only context.
type Edit struct {
	ID          string    `json:"id"`
	Target      Reference `json:"target"`
	Replacement string    `json:"replacement"`
}

// Duplicate declares a reviewed same-defect relationship without deleting claims.
type Duplicate struct {
	Claims         []string `json:"claims"`
	Representative string   `json:"representative"`
	Reason         string   `json:"reason"`
}

// Stage is untrusted review output. Source overlap grants no claim disposition.
type Stage struct {
	ID         string      `json:"id"`
	Decisions  []Decision  `json:"decisions"`
	Edits      []Edit      `json:"edits"`
	Duplicates []Duplicate `json:"duplicates"`
}

// Approval is a caller-supplied review, separate from the untrusted Stage.
// Resolution approves the exact edit digest for one claim and preserved meaning.
// Same-defect approves the exact member set. The library cannot authenticate a
// reviewer or establish the semantic truth of their decision.
type Approval struct {
	StageID          string   `json:"stage_id"`
	Kind             string   `json:"kind"`
	Claims           []string `json:"claims"`
	EditDigest       string   `json:"edit_digest"`
	Reviewer         string   `json:"reviewer"`
	Reason           string   `json:"reason"`
	MeaningPreserved bool     `json:"meaning_preserved"`
}

// Account is a complete stage disposition, including unchanged original claims.
// Uncertain dispositions make Complete false; accounting is not qualification.
type Account struct {
	Version            string      `json:"version"`
	SourceHash         string      `json:"source_sha256"`
	StageID            string      `json:"stage_id"`
	Claims             []Claim     `json:"claims"`
	Decisions          []Decision  `json:"decisions"`
	Edits              []Edit      `json:"edits"`
	Duplicates         []Duplicate `json:"duplicates"`
	Approvals          []Approval  `json:"approvals"`
	DisplayedClaims    []string    `json:"displayed_claims"`
	Complete           bool        `json:"complete"`
	EditorialQualified bool        `json:"editorial_qualified"`
}
