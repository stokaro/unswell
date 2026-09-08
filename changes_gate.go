package unswell

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/document"
)

type changedRange struct {
	path, scope string
	span        document.Span
}

type changedUnit struct {
	fingerprint, state string
}

func selectChanges(ctx context.Context, result *RunResult, comparisons map[string]*sourceComparison) error {
	units, err := selectChangedUnits(ctx, result, comparisons)
	if err != nil {
		return err
	}
	unchanged := make(map[string]bool)
	for i := range result.Findings {
		if err := ctx.Err(); err != nil {
			return err
		}
		finding := &result.Findings[i]
		if finding.Derived {
			unit := units[changedRange{finding.Primary.Path, finding.Scope, finding.Primary.Span}]
			finding.ChangeFingerprint, finding.ChangeState = unit.fingerprint, unit.state
		} else {
			comparison := comparisons[finding.Primary.Path]
			finding.ChangeFingerprint = comparison.after.findings[finding.ID]
			finding.ChangeState = comparison.state(finding.ChangeFingerprint)
		}
		unchanged[finding.ID] = finding.ChangeState == "unchanged"
	}
	result.Gate.Reasons = slices.DeleteFunc(result.Gate.Reasons, func(reason GateReason) bool {
		if unchanged[reason.FindingID] {
			result.Gate.Unchanged = append(result.Gate.Unchanged, reason)
			return true
		}
		return false
	})
	return nil
}

func selectChangedUnits(
	ctx context.Context, result *RunResult, comparisons map[string]*sourceComparison,
) (map[changedRange]changedUnit, error) {
	units := make(map[changedRange]changedUnit, len(result.Assessments))
	for i := range result.Assessments {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		assessment := &result.Assessments[i]
		comparison := comparisons[assessment.Path]
		fingerprint := comparison.after.units[identityUnit{assessment.Scope, assessment.UnitID}]
		assessment.ChangeFingerprint, assessment.ChangeState = fingerprint, comparison.state(fingerprint)
		units[changedRange{assessment.Path, assessment.Scope, assessment.Span}] = changedUnit{fingerprint, assessment.ChangeState}
		result.Changes.ComparedUnits++
		if assessment.ChangeState != "unchanged" {
			result.Changes.SelectedUnits++
		}
	}
	return units, nil
}
