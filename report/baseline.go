package report

import (
	"fmt"

	"github.com/stokaro/unswell"
)

func baselineLines(result unswell.RunResult) []string {
	if result.Baseline == nil {
		return nil
	}
	existing, fresh := 0, 0
	for _, match := range result.Baseline.Matches {
		if match.State == "existing" {
			existing++
		} else {
			fresh++
		}
	}
	lines := []string{fmt.Sprintf("Baseline: %d existing candidates, %d new; %d gate reasons accepted in new mode.",
		existing, fresh, len(result.Gate.Accepted))}
	for _, entry := range result.Baseline.Stale {
		lines = append(lines, fmt.Sprintf("Baseline stale: %s %s %s (%s).",
			entry.Identity.Path, entry.Identity.Kind, entry.Identity.RuleID, entry.Fingerprint))
	}
	for _, entry := range result.Baseline.Unobserved {
		lines = append(lines, fmt.Sprintf("Baseline unobserved: %s %s %s (%s).",
			entry.Identity.Path, entry.Identity.Kind, entry.Identity.RuleID, entry.Fingerprint))
	}
	return lines
}
