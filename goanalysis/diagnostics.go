package goanalysis

import (
	"context"
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell"
)

func resultDiagnostics(
	ctx context.Context, result unswell.RunResult, files map[string]*token.File, all bool,
) ([]analysis.Diagnostic, error) {
	if !result.Manifest.Complete || result.Status != "complete" {
		return nil, fmt.Errorf("incomplete Unswell analysis")
	}
	required := gateFindings(result)
	var diagnostics []analysis.Diagnostic
	for _, finding := range result.Findings {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !shouldReport(finding, required, all) {
			continue
		}
		diagnostic, err := findingDiagnostic(finding, files)
		if err != nil {
			return nil, err
		}
		diagnostics = append(diagnostics, diagnostic)
		delete(required, finding.ID)
	}
	if len(required) > 0 || (!result.Gate.Passed && len(diagnostics) == 0) {
		return nil, fmt.Errorf("editorial gate failure has no mappable finding")
	}
	return diagnostics, nil
}

func shouldReport(finding unswell.Finding, required map[string]bool, all bool) bool {
	return required[finding.ID] || (all && !finding.Suppressed)
}

func gateFindings(result unswell.RunResult) map[string]bool {
	required := make(map[string]bool, len(result.Gate.Reasons))
	if result.Gate.Passed {
		return required
	}
	for _, reason := range result.Gate.Reasons {
		required[reason.FindingID] = true
	}
	return required
}

func findingDiagnostic(finding unswell.Finding, files map[string]*token.File) (analysis.Diagnostic, error) {
	start, end, err := sourcePositions(finding.Primary, files)
	if err != nil {
		return analysis.Diagnostic{}, err
	}
	diagnostic := analysis.Diagnostic{Pos: start, End: end, Category: finding.RuleID,
		Message: finding.RuleID + ": " + finding.Message}
	for _, related := range finding.Related {
		start, end, err := sourcePositions(related, files)
		if err != nil {
			return analysis.Diagnostic{}, err
		}
		diagnostic.Related = append(diagnostic.Related, analysis.RelatedInformation{
			Pos: start, End: end, Message: finding.RuleID + ": related occurrence",
		})
	}
	return diagnostic, nil
}

func sourcePositions(location unswell.Location, files map[string]*token.File) (token.Pos, token.Pos, error) {
	file := files[location.Path]
	if file == nil || !location.Span.Valid(file.Size()) {
		return token.NoPos, token.NoPos, fmt.Errorf("invalid Unswell source range: %s", location.Path)
	}
	return file.Pos(location.Span.Start), file.Pos(location.Span.End), nil
}
