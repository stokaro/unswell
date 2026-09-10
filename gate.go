package unswell

import (
	"fmt"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func (e *Engine) decide(result *RunResult, doc document.Document) {
	for _, finding := range result.Findings {
		if finding.Gate == "forbid" && !finding.Suppressed {
			result.Gate.Reasons = append(
				result.Gate.Reasons,
				GateReason{Code: "forbidden-phrase", Path: doc.Name, FindingID: finding.ID, Message: finding.Message},
			)
		}
	}
	for _, assessment := range result.Assessments {
		threshold := e.policy.Gate.Paragraph
		if assessment.Scope == "sentence" {
			threshold = e.policy.Gate.Sentence
		}
		if assessment.Words >= threshold.MinWords && assessment.EffectiveSlopScore >= float64(threshold.FailAt) {
			e.failGate(result, e.thresholdFinding(doc, assessment, threshold.FailAt), doc.Name)
		}
		e.decideProbability(result, doc, assessment)
	}
}

func (e *Engine) failGate(result *RunResult, finding Finding, path string) {
	result.Findings = append(result.Findings, finding)
	result.Gate.Reasons = append(
		result.Gate.Reasons,
		GateReason{Code: finding.RuleID, Path: path, FindingID: finding.ID, Message: finding.Message},
	)
}

func (e *Engine) thresholdFinding(doc document.Document, assessment Assessment, threshold int) Finding {
	finding := e.gateFinding(doc, assessment, "score")
	finding.Message = fmt.Sprintf(
		"%s index %.3g/100 reaches the configured threshold of %d.",
		assessment.Scope,
		assessment.EffectiveSlopScore,
		threshold,
	)
	finding.Evidence.Metrics = []rule.Metric{
		{
			Name:       "effective-slop-score",
			Value:      assessment.EffectiveSlopScore,
			Unit:       "index-points",
			Onset:      float64(threshold),
			Saturation: 100,
		},
	}
	return finding
}
