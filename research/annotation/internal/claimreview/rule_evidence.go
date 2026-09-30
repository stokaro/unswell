package claimreview

import (
	"fmt"
	"math"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/rule"
)

// ExplainRule preserves the observable basis of an existing rule diagnostic.
// An absent evidence suggestion uses the engine's diagnostic as the review action;
// it never invents replacement prose. Measurements do not establish edit necessity.
// The caller must separately bind the source and review claim boundaries and roles.
func ExplainRule(finding unswell.Finding) (reason, action string, err error) {
	if err := validateOriginalRule(finding); err != nil {
		return "", "", err
	}
	evidence := finding.Evidence
	if !validRuleActivation(evidence.Kind, evidence.Activation) {
		return "", "", fmt.Errorf("invalid evidence kind or activation")
	}
	parts := []string{fmt.Sprintf("Rule %s v%s emitted %s evidence. Source occurrences: %d; activation: %d/1000.",
		finding.RuleID, finding.RuleVersion, evidence.Kind, len(evidence.Occurrences), evidence.Activation)}
	if evidence.Message != "" {
		if !validText(evidence.Message) {
			return "", "", fmt.Errorf("invalid original evidence explanation")
		}
		parts = append(parts, evidence.Message)
	}
	measurements, err := describeRuleMetrics(evidence.Metrics)
	if err != nil {
		return "", "", err
	}
	parts = append(parts, measurements...)
	parts = append(parts, "These observations require editorial review; they do not prove a wording defect.")
	reason = strings.Join(parts, " ")
	action = evidence.Suggestion
	if action == "" {
		action = finding.Message
	}
	if !validText(reason) || !validText(action) {
		return "", "", fmt.Errorf("rule explanation or action exceeds the claim contract")
	}
	return reason, action, nil
}

func validRuleActivation(kind string, activation int) bool {
	return (kind == "exact" || kind == "heuristic" || kind == "statistical") && activation >= 0 && activation <= 1000
}

func describeRuleMetrics(metrics []rule.Metric) ([]string, error) {
	measurements := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		if !validRuleMetric(metric.Name, metric.Unit, metric.Value, metric.Onset, metric.Saturation) {
			return nil, fmt.Errorf("invalid original evidence metric")
		}
		measurements = append(measurements, fmt.Sprintf("Measured %s=%g %s; onset=%g, saturation=%g.",
			metric.Name, metric.Value, metric.Unit, metric.Onset, metric.Saturation))
	}
	return measurements, nil
}

func validateOriginalRule(finding unswell.Finding) error {
	for _, text := range []string{finding.ID, finding.RuleID, finding.RuleVersion, finding.Message} {
		if !validText(text) {
			return fmt.Errorf("missing original rule identity or diagnostic")
		}
	}
	if finding.Derived || len(finding.Evidence.Metrics) == 0 || len(finding.Evidence.Metrics) > 64 ||
		len(finding.Evidence.Occurrences) == 0 {
		return fmt.Errorf("missing original measurable rule evidence")
	}
	return nil
}

func validRuleMetric(name, unit string, value, onset, saturation float64) bool {
	return validText(name) && validText(unit) && finiteNumber(value) && finiteNumber(onset) && finiteNumber(saturation)
}

func finiteNumber(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
