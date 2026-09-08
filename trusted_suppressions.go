package unswell

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
)

func (e *Engine) filterTrustedSuppressions(ctx context.Context, result *RunResult, doc document.Document, builder *debtBuilder) error {
	if e.trustedSources == nil {
		return nil
	}
	previous := e.trustedSources[doc.Name]
	counts := builder.permissionCounts(result.Suppressions)
	trusted := make(map[string]bool, len(result.Suppressions))
	for i := range result.Suppressions {
		if err := ctx.Err(); err != nil {
			return err
		}
		permission := &result.Suppressions[i]
		fingerprint := builder.hash(builder.permissionIdentity(*permission))
		permission.TrustFingerprint = fingerprint
		permission.TrustState = permissionTrust(previous, doc.Hash, fingerprint, counts[fingerprint])
		trusted[permission.ID] = permission.TrustState == "trusted"
	}
	for i := range result.Findings {
		finding := &result.Findings[i]
		if slices.ContainsFunc(finding.SuppressionIDs, func(id string) bool { return !trusted[id] }) {
			// Matching requires every evidence segment. Keeping only the trusted
			// part of a permission union would incorrectly allow partial coverage.
			finding.Suppressed, finding.SuppressionIDs = false, nil
		}
	}
	return builder.err
}

func permissionTrust(previous sourceIdentities, sourceHash, fingerprint string, count int) string {
	if previous.document.SourceHash == sourceHash {
		return "trusted"
	}
	old := previous.permissions[fingerprint]
	if count > 1 || old > 1 {
		return "ambiguous"
	}
	if old == 1 {
		return "trusted"
	}
	return "untrusted"
}
