package unswell

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
)

type identityUnit struct {
	scope string
	id    int
}

type sourceIdentities struct {
	document    baseline.Document
	findings    map[string]string
	units       map[identityUnit]string
	permissions map[string]int
}

func (e *Engine) identityBuilder(ctx context.Context, doc document.Document, ids *sourceIdentities) (*debtBuilder, error) {
	if !e.collectBaseline && ids == nil && e.trustedSources == nil {
		return nil, nil
	}
	return newDebtBuilder(ctx, doc)
}

func (e *Engine) identifySource(
	ctx context.Context, result *RunResult, doc document.Document, ids *sourceIdentities, builder *debtBuilder,
) error {
	allUnits := ids != nil
	if ids == nil {
		ids = &sourceIdentities{}
	}
	ids.findings, ids.units = make(map[string]string), make(map[identityUnit]string)
	if err := builder.collectFindings(ctx, result, ids); err != nil {
		return err
	}
	if err := builder.collectAssessments(ctx, result, ids, allUnits); err != nil {
		return err
	}
	ids.document = baseline.Document{
		Path: doc.Name, Format: string(doc.Format), SourceHash: doc.Hash,
		PolicyHash: builder.hash(baselinePolicy(e.policy)), SuppressionHash: builder.suppressionHash(result.Suppressions),
	}
	ids.permissions = builder.permissionCounts(result.Suppressions)
	if result.BaselineSnapshot != nil {
		result.BaselineSnapshot.Documents = append(result.BaselineSnapshot.Documents, ids.document)
	}
	return builder.err
}

func (b *debtBuilder) collectFindings(ctx context.Context, result *RunResult, ids *sourceIdentities) error {
	for i := range result.Findings {
		if err := ctx.Err(); err != nil {
			return err
		}
		finding := &result.Findings[i]
		identity := b.finding(*finding)
		ids.findings[finding.ID] = b.fingerprint(identity)
		if result.BaselineSnapshot == nil {
			continue
		}
		finding.BaselineFingerprint = ids.findings[finding.ID]
		if !finding.Suppressed {
			finding.BaselineState = "new"
			result.BaselineSnapshot.Candidates = append(result.BaselineSnapshot.Candidates, identity)
		}
	}
	return b.err
}

func (b *debtBuilder) collectAssessments(ctx context.Context, result *RunResult, ids *sourceIdentities, allUnits bool) error {
	for i := range result.Assessments {
		if err := ctx.Err(); err != nil {
			return err
		}
		assessment := &result.Assessments[i]
		if assessment.EffectiveSlopScore == 0 && !allUnits {
			continue
		}
		identity := b.unit(assessment.Scope, assessment.UnitID)
		identity.EvidenceHash = b.assessmentEvidence(*assessment, ids.findings)
		fingerprint := b.fingerprint(identity)
		ids.units[identityUnit{assessment.Scope, assessment.UnitID}] = fingerprint
		if result.BaselineSnapshot == nil || assessment.EffectiveSlopScore == 0 {
			continue
		}
		assessment.BaselineFingerprint = fingerprint
		assessment.BaselineState = "new"
		result.BaselineSnapshot.Candidates = append(result.BaselineSnapshot.Candidates, identity)
	}
	return b.err
}

func (b *debtBuilder) assessmentEvidence(assessment Assessment, findings map[string]string) string {
	// Replace transient run IDs on owned copies. Scores and evidence, including
	// document repetition and source suppressions, remain part of unit identity.
	raw := stableContributions(assessment.Contributions, findings)
	effective := stableContributions(assessment.EffectiveContributions, findings)
	return b.hash(struct {
		Words             int
		Raw, Effective    float64
		Status            string
		Contributions     []Contribution
		EffectiveEvidence []Contribution
	}{assessment.Words, assessment.SlopScore, assessment.EffectiveSlopScore, assessment.Status, raw, effective})
}

func stableContributions(contributions []Contribution, findings map[string]string) []Contribution {
	result := slices.Clone(contributions)
	for i := range result {
		result[i].FindingID = findings[result[i].FindingID]
	}
	return result
}

type debtSuppression struct {
	Kind, Reason string
	Rules        []string
	Targets      []baseline.Identity
}

func (b *debtBuilder) suppressionHash(suppressions []Suppression) string {
	identities := make([]debtSuppression, 0, len(suppressions))
	for _, suppression := range suppressions {
		identities = append(identities, b.permissionIdentity(suppression))
	}
	return b.hash(identities)
}

func (b *debtBuilder) permissionIdentity(suppression Suppression) debtSuppression {
	identity := debtSuppression{Kind: suppression.Kind, Reason: suppression.Reason, Rules: suppression.RuleIDs}
	for _, target := range suppression.Targets {
		unit := baseline.Identity{Path: b.doc.Name, Kind: target.Scope}
		if target.Scope != "file" {
			unit = b.unit(target.Scope, target.UnitID)
		}
		identity.Targets = append(identity.Targets, unit)
	}
	return identity
}

func (b *debtBuilder) permissionCounts(suppressions []Suppression) map[string]int {
	counts := make(map[string]int, len(suppressions))
	for _, suppression := range suppressions {
		counts[b.hash(b.permissionIdentity(suppression))]++
	}
	return counts
}
