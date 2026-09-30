package claimreview

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/research/annotation/internal/commandio"
	"github.com/stokaro/unswell/rule"
)

// RuleEvidenceVersion identifies the source report's observational projection.
const RuleEvidenceVersion = "unswell-rule-evidence-v1"

type ruleObservation struct {
	ID               string             `json:"id"`
	SourceHash       string             `json:"source_sha256"`
	RuleID           string             `json:"rule_id"`
	RuleVersion      string             `json:"rule_version"`
	Diagnostic       string             `json:"diagnostic"`
	Reason           string             `json:"reason"`
	Action           string             `json:"action"`
	ActionBasis      string             `json:"action_basis"`
	Primary          unswell.Location   `json:"primary"`
	Related          []unswell.Location `json:"related"`
	Evidence         rule.Evidence      `json:"evidence"`
	Suppressed       bool               `json:"suppressed"`
	EditorialVerdict string             `json:"editorial_verdict"`
}

type ruleEvidenceOutput struct {
	Version            string            `json:"version"`
	ToolCommit         string            `json:"tool_commit"`
	ConfigHash         string            `json:"config_hash"`
	Observations       []ruleObservation `json:"observations"`
	EditorialQualified bool              `json:"editorial_qualified"`
}

// RunRules projects every original finding in one complete canonical saved report.
// It preserves evidence and locations without source snippets, judgments, or edits.
// It does not reread sources, infer target/support roles, or execute a model.
func RunRules(ctx context.Context, reader io.Reader, writer io.Writer) error {
	result, err := commandio.Await(ctx, func() (unswell.RunResult, error) { return report.Read(reader) })
	if err != nil {
		return err
	}
	if result.Status != "complete" {
		return fmt.Errorf("rule evidence requires a complete source report")
	}
	if len(result.Findings) > maximumClaims || len(result.Documents) > maximumClaims {
		return fmt.Errorf("rule observation inventory exceeds %d entries", maximumClaims)
	}
	out := ruleEvidenceOutput{Version: RuleEvidenceVersion, ToolCommit: result.Manifest.ToolCommit,
		ConfigHash: result.Manifest.ConfigHash, Observations: make([]ruleObservation, 0, len(result.Findings))}
	sources, err := ruleSourceHashes(result.Documents)
	if err != nil {
		return err
	}
	seen := make(map[string]bool)
	for _, finding := range result.Findings {
		if err := ctx.Err(); err != nil {
			return err
		}
		if seen[finding.ID] {
			return fmt.Errorf("duplicate original rule finding")
		}
		seen[finding.ID] = true
		observation, err := observeRule(finding, sources)
		if err != nil {
			return err
		}
		out.Observations = append(out.Observations, observation)
	}
	_, err = commandio.Await(ctx, func() (struct{}, error) { return struct{}{}, json.NewEncoder(writer).Encode(out) })
	return err
}

func ruleSourceHashes(documents []unswell.DocumentResult) (map[string]string, error) {
	sources := make(map[string]string)
	for _, doc := range documents {
		if _, exists := sources[doc.Name]; exists {
			return nil, fmt.Errorf("duplicate source identity in rule report")
		}
		sources[doc.Name] = doc.SourceHash
	}
	return sources, nil
}

func observeRule(finding unswell.Finding, sources map[string]string) (ruleObservation, error) {
	hash, exists := sources[finding.Primary.Path]
	if !exists || hash == "" {
		return ruleObservation{}, fmt.Errorf("rule finding lacks its original source identity")
	}
	reason, action, err := ExplainRule(finding)
	if err != nil {
		return ruleObservation{}, fmt.Errorf("finding %s: %w", finding.ID, err)
	}
	primary, related := finding.Primary, append([]unswell.Location{}, finding.Related...)
	primary.Snippet = ""
	for i := range related {
		if related[i].Path != primary.Path {
			return ruleObservation{}, fmt.Errorf("rule observation crosses source documents")
		}
		related[i].Snippet = ""
	}
	basis := "evidence_suggestion"
	if finding.Evidence.Suggestion == "" {
		basis = "engine_diagnostic"
	}
	return ruleObservation{ID: finding.ID, SourceHash: hash, RuleID: finding.RuleID, RuleVersion: finding.RuleVersion,
		Diagnostic: finding.Message, Reason: reason, Action: action, ActionBasis: basis,
		Primary: primary, Related: related, Evidence: finding.Evidence, Suppressed: finding.Suppressed,
		EditorialVerdict: "unreviewed"}, nil
}
