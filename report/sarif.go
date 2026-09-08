package report

import (
	"io"
	"net/url"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
)

func sarif(writer io.Writer, result unswell.RunResult) error {
	rules := make([]map[string]any, 0)
	indexes := make(map[string]int)
	for _, descriptor := range result.Manifest.Rules {
		indexes[descriptor.ID] = len(rules)
		rules = append(rules, map[string]any{"id": descriptor.ID, "shortDescription": map[string]string{"text": descriptor.Summary},
			"fullDescription": map[string]string{
				"text": descriptor.Description,
			}, "properties": map[string]string{"group": descriptor.Group, "status": descriptor.Status}})
	}
	for _, finding := range result.Findings {
		if _, ok := indexes[finding.RuleID]; ok {
			continue
		}
		indexes[finding.RuleID] = len(rules)
		rules = append(
			rules,
			map[string]any{"id": finding.RuleID, "shortDescription": map[string]string{"text": "Local editorial policy threshold"}},
		)
	}
	results := make([]map[string]any, 0, len(result.Findings))
	permissions := suppressionIndex(result.Suppressions)
	for _, finding := range result.Findings {
		entry := sarifFinding(finding, indexes[finding.RuleID])
		if finding.Suppressed {
			entry["suppressions"] = sarifSuppressions(finding, permissions)
		}
		results = append(results, entry)
	}
	notifications := make([]map[string]any, 0, len(result.Errors))
	for _, failure := range result.Errors {
		notifications = append(
			notifications,
			map[string]any{"level": "error", "message": map[string]string{"text": failure.Path + ": " + failure.Message}},
		)
	}
	return encodeJSON(writer, map[string]any{
		"version": "2.1.0", "$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"runs": []any{map[string]any{
			"tool": map[string]any{
				"driver": map[string]any{
					"name":           "Unswell",
					"version":        result.Manifest.ToolVersion,
					"informationUri": "https://github.com/stokaro/unswell",
					"rules":          rules,
				},
			},
			"columnKind": "unicodeCodePoints", "results": results,
			"invocations": []any{
				map[string]any{"executionSuccessful": result.Manifest.Complete, "toolExecutionNotifications": notifications},
			},
			"properties": runProperties(result),
		}},
	})
}

func filePolicies(result unswell.RunResult) map[string]any {
	policies := make(map[string]any, len(result.Documents))
	for _, doc := range result.Documents {
		if doc.ConfigHash != "" {
			policies[doc.Name] = map[string]any{"config_hash": doc.ConfigHash, "applied_overrides": doc.AppliedOverrides,
				"gate_mode": doc.GateMode}
		}
	}
	return policies
}

func sarifFinding(finding unswell.Finding, index int) map[string]any {
	related := make([]map[string]any, 0, len(finding.Related))
	for i, location := range finding.Related {
		entry := sarifLocation(location)
		entry["id"] = i + 1
		related = append(related, entry)
	}
	result := map[string]any{
		"ruleId":              finding.RuleID,
		"ruleIndex":           index,
		"level":               finding.Severity,
		"message":             map[string]string{"text": finding.Message},
		"locations":           []any{sarifLocation(finding.Primary)},
		"relatedLocations":    related,
		"partialFingerprints": map[string]string{"unswell/v1": finding.Fingerprint},
		"properties": map[string]any{"evidence": finding.Evidence, "group": finding.Group, "derived": finding.Derived,
			"suppression_ids": finding.SuppressionIDs, "change_state": finding.ChangeState, "change_fingerprint": finding.ChangeFingerprint},
	}
	if finding.BaselineFingerprint != "" {
		result["partialFingerprints"] = map[string]string{"unswell/v1": finding.Fingerprint,
			baseline.FingerprintVersion: finding.BaselineFingerprint}
	}
	if finding.BaselineState == "existing" {
		result["baselineState"] = "unchanged"
	} else if slices.Contains([]string{"new", "unchanged", "updated", "absent"}, finding.BaselineState) {
		result["baselineState"] = finding.BaselineState
	}
	return result
}

func sarifLocation(location unswell.Location) map[string]any {
	uri := url.URL{Path: strings.ReplaceAll(location.Path, "\\", "/")}
	region := map[string]any{
		"startLine":   location.Start.Line,
		"startColumn": location.Start.Column,
		"endLine":     location.End.Line,
		"endColumn":   location.End.Column,
		"byteOffset":  location.Span.Start,
		"byteLength":  location.Span.End - location.Span.Start,
	}
	if location.Snippet != "" {
		region["snippet"] = map[string]string{"text": location.Snippet}
	}
	return map[string]any{
		"physicalLocation": map[string]any{"artifactLocation": map[string]string{"uri": uri.String()}, "region": region},
	}
}

func runProperties(result unswell.RunResult) map[string]any {
	properties := map[string]any{"gate": result.Gate, "assessments": result.Assessments,
		"changes": result.Changes, "policy_comparison": result.PolicyComparison,
		"baseline": result.Baseline, "baseline_snapshot": result.BaselineSnapshot,
		"manifest": result.Manifest, "file_policies": filePolicies(result), "suppressions": result.Suppressions}
	if result.Features != nil {
		properties["features"] = result.Features
	}
	return properties
}
