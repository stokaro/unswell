package report

import (
	"fmt"

	"github.com/stokaro/unswell"
)

func findingMessage(finding unswell.Finding) string {
	if finding.Suppressed {
		return finding.Message + " [suppressed]"
	}
	return finding.Message
}

func suppressionLines(result unswell.RunResult) []string {
	var lines []string
	for _, record := range result.Suppressions {
		lines = append(lines, fmt.Sprintf("Suppression %s: %s:%d:%d [%s] %s (matches: %d).",
			record.Status, record.Directive.Path, record.Directive.Start.Line, record.Directive.Start.Column,
			record.Kind, record.Reason, len(record.FindingIDs)))
	}
	for _, assessment := range result.Assessments {
		if len(assessment.EffectiveContributions) != 0 {
			lines = append(lines, fmt.Sprintf("%s %s %d: raw index %.3g; effective index %.3g.",
				assessment.Path, assessment.Scope, assessment.UnitID, assessment.SlopScore, assessment.EffectiveSlopScore))
		}
	}
	return lines
}

func suppressionIndex(records []unswell.Suppression) map[string]unswell.Suppression {
	byID := make(map[string]unswell.Suppression, len(records))
	for _, record := range records {
		byID[record.ID] = record
	}
	return byID
}

func sarifSuppressions(finding unswell.Finding, byID map[string]unswell.Suppression) []map[string]any {
	var result []map[string]any
	for _, id := range finding.SuppressionIDs {
		record := byID[id]
		result = append(result, map[string]any{"kind": "inSource", "status": "accepted", "justification": record.Reason})
	}
	// Older saved reports retain their original suppression flag without records.
	if len(result) == 0 {
		result = append(result, map[string]any{"kind": "inSource", "status": "accepted"})
	}
	return result
}
