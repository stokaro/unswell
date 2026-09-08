package report

import (
	"fmt"

	"github.com/stokaro/unswell"
)

func auditLines(result unswell.RunResult) []string {
	lines := append(suppressionLines(result), baselineLines(result)...)
	lines = append(lines, changeLines(result)...)
	return append(lines, policyLines(result)...)
}

func policyLines(result unswell.RunResult) []string {
	comparison := result.PolicyComparison
	if comparison == nil {
		return nil
	}
	lines := []string{fmt.Sprintf("Trusted policy: complete: %t; full scan: %t; %d changes.",
		comparison.Complete, comparison.FullScan, len(comparison.Changes))}
	for _, change := range comparison.Changes {
		lines = append(lines, fmt.Sprintf("Policy change: %s [%s]; %s to %s.",
			change.Path, change.Kind, change.BeforeHash, change.AfterHash))
	}
	return lines
}

func changeLines(result unswell.RunResult) []string {
	if result.Changes == nil {
		return nil
	}
	lines := []string{fmt.Sprintf("Changed units: %d selected of %d compared; complete: %t; %d unchanged gate reasons retained.",
		result.Changes.SelectedUnits, result.Changes.ComparedUnits, result.Changes.Complete, len(result.Gate.Unchanged))}
	if git := result.Manifest.Git; git != nil {
		lines = append(lines, fmt.Sprintf("Committed selection: %s to %s; selected inputs clean: %t.", git.BaseCommit, git.HeadCommit, git.Clean))
	}
	for _, doc := range result.Changes.Documents {
		lines = append(lines, fmt.Sprintf("Changed document: %s (%s); full selection reason: %s.", doc.Path, doc.Status, doc.FullReason))
	}
	return lines
}
