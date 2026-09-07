package unswell

import (
	"crypto/sha256"
	"fmt"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func (e *Engine) decide(result *RunResult, doc document.Document) {
	for _, finding := range result.Findings {
		if finding.Gate == "forbid" {
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
		if assessment.Words < threshold.MinWords || assessment.EffectiveSlopScore < float64(threshold.FailAt) {
			continue
		}
		finding := e.thresholdFinding(doc, assessment, threshold.FailAt)
		result.Findings = append(result.Findings, finding)
		result.Gate.Reasons = append(
			result.Gate.Reasons,
			GateReason{Code: finding.RuleID, Path: doc.Name, FindingID: finding.ID, Message: finding.Message},
		)
	}
}

func (e *Engine) thresholdFinding(doc document.Document, assessment Assessment, threshold int) Finding {
	ruleID := "gate." + assessment.Scope + "-score"
	fingerprint := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d:%s", doc.Name, ruleID, assessment.UnitID, doc.Hash))))
	location := Location{Path: doc.Name, Span: assessment.Span, Segments: []document.Span{assessment.Span}}
	// Assessment spans were validated during extraction and evidence collection.
	if start, err := document.Locate(doc.Source, assessment.Span.Start); err == nil {
		location.Start = start
	}
	if end, err := document.Locate(doc.Source, assessment.Span.End); err == nil {
		location.End = end
	}
	if e.includeSource {
		location.Snippet = string(doc.Source[assessment.Span.Start:assessment.Span.End])
	}
	return Finding{
		ID:          fingerprint[:16],
		RuleID:      ruleID,
		RuleVersion: "1",
		Severity:    "error",
		Gate:        "forbid",
		Group:       "policy",
		Scope:       assessment.Scope,
		Message: fmt.Sprintf(
			"%s index %.3g/100 reaches the configured threshold of %d.",
			assessment.Scope,
			assessment.EffectiveSlopScore,
			threshold,
		),
		Primary:       location,
		Related:       []Location{},
		Fingerprint:   fingerprint,
		BaselineState: "untracked",
		Derived:       true,
		Evidence: rule.Evidence{
			Kind:        "exact",
			Occurrences: []rule.Occurrence{},
			Metrics: []rule.Metric{
				{
					Name:       "effective-slop-score",
					Value:      assessment.EffectiveSlopScore,
					Unit:       "index-points",
					Onset:      float64(threshold),
					Saturation: 100,
				},
			},
		},
	}
}
