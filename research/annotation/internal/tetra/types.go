// Package tetra imports explicitly supplied TETRA XML into grouped research pairs.
package tetra

import (
	"errors"

	"github.com/stokaro/unswell/document"
)

// Version identifies the input and reconstructed-section output contract.
const Version = "unswell-tetra-import-v1"

// ErrIncomplete means source issues quarantined at least one complete paper group.
var ErrIncomplete = errors.New("TETRA import contains quarantined source groups")

const (
	maximumFiles       = 256
	maximumXMLBytes    = 256 << 10
	maximumSourceBytes = 8 << 20
	maximumInputBytes  = 16 << 20
	maximumOutputBytes = 64 << 20
)

// Input supplies a trusted acquisition manifest and exact local source bytes.
// Hash validation proves content integrity, not membership in the claimed commit.
type Input struct {
	Version    string       `json:"version"`
	Repository string       `json:"repository"`
	Revision   string       `json:"revision"`
	License    string       `json:"license"`
	IncludeXML bool         `json:"include_xml"`
	Files      []SourceFile `json:"files"`
}

// SourceFile preserves raw XML as base64 in JSON, including malformed source bytes.
type SourceFile struct {
	Path    string `json:"path"`
	SHA256  string `json:"sha256"`
	GitBlob string `json:"git_blob"`
	XML     []byte `json:"xml_base64"`
}

// Issue is an explicit source failure; it never becomes an editorial label.
type Issue struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// Edit retains the upstream change and its positions in both complete sections.
// Empty spans are valid insertions or deletions, not missing observations.
type Edit struct {
	ID            string        `json:"id"`
	ElementIndex  int           `json:"element_index"`
	Types         []string      `json:"types"`
	TypeAvailable bool          `json:"type_available"`
	OriginalType  string        `json:"original_type"`
	Comment       *string       `json:"comment"`
	Before        string        `json:"before"`
	After         string        `json:"after"`
	OriginalSpan  document.Span `json:"original_span"`
	RevisedSpan   document.Span `json:"revised_span"`
	Unchanged     bool          `json:"unchanged"`
}

// Section keeps all of one editor's same-section changes together.
// Coordinates refer to reconstructed UTF-8 sections, never XML/publication bytes.
type Section struct {
	Name           string `json:"name"`
	Original       string `json:"original"`
	Revised        string `json:"revised"`
	OriginalSHA256 string `json:"original_sha256"`
	RevisedSHA256  string `json:"revised_sha256"`
	Edits          []Edit `json:"edits"`
}

// File records every supplied source, including invalid and quarantined files.
type File struct {
	Path             string            `json:"path"`
	Paper            string            `json:"paper"`
	Editor           string            `json:"editor"`
	SHA256           string            `json:"sha256"`
	GitBlob          string            `json:"git_blob"`
	Bytes            int               `json:"bytes"`
	XMLBase64        *string           `json:"xml_base64,omitempty"`
	DeclaredMetadata map[string]string `json:"declared_metadata"`
	Status           string            `json:"status"`
	SourceEligible   bool              `json:"source_eligible"`
	Issues           []Issue           `json:"issues"`
	Sections         []Section         `json:"sections"`
}

// Group binds all filename versions of one paper before any future partition.
type Group struct {
	Paper          string   `json:"paper"`
	Files          []string `json:"files"`
	SourceEligible bool     `json:"source_eligible"`
}

// Counts reports sources and edits without estimating independence or quality.
type Counts struct {
	Files             int `json:"files"`
	ParsedFiles       int `json:"parsed_files"`
	InvalidFiles      int `json:"invalid_files"`
	IdentityConflicts int `json:"identity_conflicts"`
	QuarantinedFiles  int `json:"quarantined_files"`
	Groups            int `json:"groups"`
	EligibleGroups    int `json:"eligible_groups"`
	Edits             int `json:"edits"`
	EligibleEdits     int `json:"eligible_edits"`
}

// Report is an unqualified import, not a needs_revision training corpus.
type Report struct {
	Version            string  `json:"version"`
	Repository         string  `json:"repository"`
	Revision           string  `json:"revision"`
	License            string  `json:"license"`
	Rendering          string  `json:"rendering"`
	Target             string  `json:"target"`
	Files              []File  `json:"files"`
	Groups             []Group `json:"groups"`
	Counts             Counts  `json:"counts"`
	Complete           bool    `json:"complete"`
	TaskLabelsAssigned bool    `json:"task_labels_assigned"`
	ProductQualified   bool    `json:"product_qualified"`
}
