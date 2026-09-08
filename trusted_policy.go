package unswell

import (
	"context"
	"encoding/hex"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
)

// ChangeOptions selects comparison against a caller-declared trusted engine policy.
// PolicyChanges describe external resource differences; the engine never loads them.
type ChangeOptions struct {
	TrustedPolicy bool
	PolicyChanges []PolicyChange
}

// PolicyChange identifies a resource or structural permission change. Hashes are
// SHA-256 identities; an empty side means that the resource or permission is absent.
type PolicyChange struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	BeforeHash string `json:"before_hash,omitempty"`
	AfterHash  string `json:"after_hash,omitempty"`
}

// PolicyComparison audits trusted-policy use. FullScan selects all current units
// for resource changes; permission changes can select individual complete files.
// This record proves neither organizational approval nor the checker's integrity.
type PolicyComparison struct {
	Version  string         `json:"version"`
	Complete bool           `json:"complete"`
	FullScan bool           `json:"full_scan"`
	Changes  []PolicyChange `json:"changes"`
}

// AnalyzeChangedWithOptions optionally uses this engine's policy and the before
// sources as trusted inputs. Only matching base permissions can suppress current
// findings. Resource changes trigger full selection under the trusted policy;
// they do not apply the candidate policy or automatically accept new baseline debt.
// Callers own resource loading, revision verification, and the trust assertion.
func (e *Engine) AnalyzeChangedWithOptions(
	ctx context.Context, before, after []document.Source, options ChangeOptions,
) (RunResult, error) {
	changes, err := validatePolicyChanges(options)
	if err != nil {
		return incomplete(e.emptyResult(), err)
	}
	if !options.TrustedPolicy {
		return e.AnalyzeChanged(ctx, before, after)
	}
	previousEngine := *e
	previousEngine.baselineFile, previousEngine.collectBaseline = nil, false
	previousEngine.gateMode, previousEngine.allowEmpty = "all", true
	previousEngine.trustedSources = nil
	_, previous, previousErr := previousEngine.analyzeAll(ctx, before, true)
	currentEngine := *e
	currentEngine.trustedSources = make(map[string]sourceIdentities, len(previous))
	for _, source := range previous {
		currentEngine.trustedSources[source.document.Path] = source
	}
	result, current, err := currentEngine.analyzeAll(ctx, after, true)
	result.Changes = &ChangeSelection{Version: "unswell-changes-v1", Documents: []ChangedDocument{}}
	result.PolicyComparison = &PolicyComparison{Version: "unswell-policy-comparison-v1", Changes: changes, FullScan: len(changes) > 0}
	if err != nil {
		return result, err
	}
	if previousErr != nil {
		return incomplete(result, fmt.Errorf("previous source analysis: %w", previousErr))
	}
	return e.finishChanged(ctx, result, previous, current)
}

func validatePolicyChanges(options ChangeOptions) ([]PolicyChange, error) {
	if len(options.PolicyChanges) > 256 || (!options.TrustedPolicy && len(options.PolicyChanges) > 0) {
		return nil, fmt.Errorf("policy changes require trusted comparison and at most 256 resources")
	}
	changes := append([]PolicyChange{}, options.PolicyChanges...)
	seen := make(map[string]bool, len(changes))
	for _, change := range changes {
		if !validPolicyChange(change) || seen[change.Kind+":"+change.Path] {
			return nil, fmt.Errorf("invalid or duplicate policy change: %s", change.Path)
		}
		seen[change.Kind+":"+change.Path] = true
	}
	sortPolicyChanges(changes)
	return changes, nil
}

func validPolicyChange(change PolicyChange) bool {
	if change.Path == "" || path.Clean(change.Path) != change.Path || strings.ContainsAny(change.Path, "\\\x00\r\n") ||
		strings.HasPrefix(change.Path, "/") || change.Path == ".." || strings.HasPrefix(change.Path, "../") {
		return false
	}
	kinds := []string{"config", "dictionary", "ruleset", "baseline", "model", "config-discovery", "ruleset-directory"}
	if !slices.Contains(kinds, change.Kind) {
		return false
	}
	return change.BeforeHash != change.AfterHash && policyHash(change.BeforeHash) && policyHash(change.AfterHash)
}

func policyHash(value string) bool {
	if value == "" {
		return true
	}
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sortPolicyChanges(changes []PolicyChange) {
	slices.SortFunc(changes, func(a, b PolicyChange) int {
		if order := strings.Compare(a.Path, b.Path); order != 0 {
			return order
		}
		return strings.Compare(a.Kind, b.Kind)
	})
}

func compareTrustedPolicy(result *RunResult, comparisons map[string]*sourceComparison) {
	if result.PolicyComparison == nil {
		return
	}
	for _, comparison := range comparisons {
		before, after := permissionHash(comparison.before), permissionHash(comparison.after)
		if before != after {
			result.PolicyComparison.Changes = append(result.PolicyComparison.Changes, PolicyChange{
				Path: comparison.document.Path, Kind: "source-permissions", BeforeHash: before, AfterHash: after,
			})
		}
		if result.PolicyComparison.FullScan && comparison.document.AfterHash != "" {
			comparison.document.FullReason = "policy_changed"
		}
	}
	sortPolicyChanges(result.PolicyComparison.Changes)
}

func permissionHash(source sourceIdentities) string {
	if len(source.permissions) == 0 {
		return ""
	}
	return source.document.SuppressionHash
}
