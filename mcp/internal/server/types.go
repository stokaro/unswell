package server

import (
	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// Source is client-supplied data. Name labels evidence and never opens a file.
type Source struct {
	Name   string          `json:"name"             jsonschema:"Project-relative filename for detection, exceptions, and evidence."`
	Format document.Format `json:"format,omitempty" jsonschema:"Optional syntax name; omitted uses filename or shebang detection."`
	Text   string          `json:"text"             jsonschema:"Original UTF-8 contents, analyzed as data without execution."`
}

// CheckInput submits a bounded batch to the engine under the startup policy.
type CheckInput struct {
	Sources []Source `json:"sources" jsonschema:"Between 1 and 256 documents; policy byte and analysis limits also apply."`
}

// CheckOutput keeps style failures separate from operational errors.
type CheckOutput struct {
	Outcome string            `json:"outcome" jsonschema:"pass, policy_failure, or error. Only pass is a complete successful check."`
	Result  unswell.RunResult `json:"result"  jsonschema:"Engine findings, source coordinates, gate decision, exclusions, and manifest."`
}

// DescribeInput has no options that could alter the fixed policy.
type DescribeInput struct {
	File string `json:"file,omitempty" jsonschema:"Optional project-relative filename for policy overrides. Does not open the file."`
}

// Format describes syntax and available prose contexts.
type Format struct {
	Name     document.Format `json:"name"`
	Contexts []string        `json:"contexts"`
}

// Description exposes the same catalog and effective policy used for checking.
type Description struct {
	PreparedFeatures []string          `json:"prepared_features,omitempty"`
	PreparedKinds    []string          `json:"prepared_kinds,omitempty"`
	Features         []string          `json:"features,omitempty"`
	BaselineLoaded   bool              `json:"baseline_loaded"`
	GateMode         string            `json:"gate_mode"`
	Version          string            `json:"version"`
	Commit           string            `json:"commit"`
	Formats          []Format          `json:"formats"`
	Policy           config.Policy     `json:"policy"`
	Rules            []rule.Descriptor `json:"rules"`
}
