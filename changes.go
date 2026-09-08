package unswell

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
)

// GitSelection records application-verified committed source provenance. The
// engine never invokes Git. Clean describes the selected inputs, not every file.
type GitSelection struct {
	RequestedRef string `json:"requested_ref"`
	BaseCommit   string `json:"base_commit"`
	HeadCommit   string `json:"head_commit"`
	Clean        bool   `json:"clean"`
}

// ChangeSelection records structural comparison without granting baseline debt.
// Complete describes in-memory analysis; Git provenance belongs to the caller.
type ChangeSelection struct {
	Version       string            `json:"version"`
	Complete      bool              `json:"complete"`
	Documents     []ChangedDocument `json:"documents"`
	ComparedUnits int               `json:"compared_units"`
	SelectedUnits int               `json:"selected_units"`
}

// ChangedDocument identifies both versions and any reason for selecting a full file.
// A deleted document has no AfterHash and contributes no current source locations.
type ChangedDocument struct {
	Path       string `json:"path"`
	BeforeHash string `json:"before_hash,omitempty"`
	AfterHash  string `json:"after_hash,omitempty"`
	Status     string `json:"status"`
	FullReason string `json:"full_reason,omitempty"`
}

// AnalyzeChanged analyzes complete source versions under this engine's fixed
// policy. Only changed or ambiguous findings and scored units affect the gate.
// All current evidence remains visible. New paths are selected in full; identical
// duplicates cannot silently inherit an unchanged state. Callers own revision
// discovery, clean-source verification, and policy selection.
func (e *Engine) AnalyzeChanged(ctx context.Context, before, after []document.Source) (RunResult, error) {
	result, current, err := e.analyzeAll(ctx, after, true)
	result.Changes = &ChangeSelection{Version: "unswell-changes-v1", Documents: []ChangedDocument{}}
	if err != nil {
		return result, err
	}
	previousEngine := *e
	previousEngine.baselineFile, previousEngine.collectBaseline = nil, false
	previousEngine.gateMode, previousEngine.allowEmpty = "all", true
	_, previous, err := previousEngine.analyzeAll(ctx, before, true)
	if err != nil {
		return incomplete(result, fmt.Errorf("previous source analysis: %w", err))
	}
	comparisons, err := compareSources(ctx, previous, current)
	if err != nil {
		return incomplete(result, err)
	}
	for _, comparison := range comparisons {
		result.Changes.Documents = append(result.Changes.Documents, comparison.document)
	}
	slices.SortFunc(result.Changes.Documents, func(a, b ChangedDocument) int { return strings.Compare(a.Path, b.Path) })
	if err := selectChanges(ctx, &result, comparisons); err != nil {
		return incomplete(result, err)
	}
	result.Changes.Complete = true
	result.Gate.Passed = len(result.Gate.Reasons) == 0 || e.noGate
	return result, nil
}

type sourceComparison struct {
	document ChangedDocument
	before   sourceIdentities
	after    sourceIdentities
	counts   map[string]int
	old      map[string]int
}

func compareSources(ctx context.Context, before, after []sourceIdentities) (map[string]*sourceComparison, error) {
	result := make(map[string]*sourceComparison, len(before)+len(after))
	for _, source := range before {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result[source.document.Path] = &sourceComparison{before: source,
			document: ChangedDocument{Path: source.document.Path, BeforeHash: source.document.SourceHash, Status: "deleted"}}
	}
	for _, source := range after {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		comparison := result[source.document.Path]
		if comparison == nil {
			comparison = &sourceComparison{document: ChangedDocument{Path: source.document.Path, Status: "added", FullReason: "added"}}
			result[source.document.Path] = comparison
		} else {
			comparison.document.Status = "modified"
			comparison.document.FullReason = fullFileReason(comparison.before.document, source.document)
			if comparison.document.BeforeHash == source.document.SourceHash {
				comparison.document.Status = "unchanged"
			}
		}
		comparison.document.AfterHash = source.document.SourceHash
		comparison.after = source
		var err error
		comparison.old, err = identityCounts(ctx, comparison.before)
		if err != nil {
			return nil, err
		}
		comparison.counts, err = identityCounts(ctx, source)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func fullFileReason(before, after baseline.Document) string {
	switch {
	case before.Format != after.Format:
		return "format_changed"
	case before.PolicyHash != after.PolicyHash:
		return "policy_changed"
	case before.SuppressionHash != after.SuppressionHash:
		return "suppressions_changed"
	default:
		return ""
	}
}

func identityCounts(ctx context.Context, source sourceIdentities) (map[string]int, error) {
	result := make(map[string]int, len(source.findings)+len(source.units))
	for _, fingerprint := range source.findings {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result[fingerprint]++
	}
	for _, fingerprint := range source.units {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		result[fingerprint]++
	}
	return result, nil
}

func (c *sourceComparison) state(fingerprint string) string {
	if c.document.FullReason != "" {
		return "changed"
	}
	if c.document.Status == "unchanged" {
		return "unchanged"
	}
	if c.counts[fingerprint] > 1 || c.old[fingerprint] > 1 {
		return "ambiguous"
	}
	if c.old[fingerprint] == 1 {
		return "unchanged"
	}
	return "changed"
}
