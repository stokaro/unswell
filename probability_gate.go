package unswell

import (
	"crypto/sha256"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/probability"
	"github.com/stokaro/unswell/rule"
)

// expectedAbstentions are the pack's own declared applicability limits. A gate
// that requires an estimate still passes on these, because the pack states in
// advance that it does not qualify such a unit.
var expectedAbstentions = []string{probability.StatusUnsupportedUnit, probability.StatusInsufficientEvidence}

// decideProbability fails the gate on a calibrated estimate at or above the
// configured threshold, and on an absent estimate the pack did not predict.
func (e *Engine) decideProbability(result *RunResult, doc document.Document, assessment Assessment) {
	gate := e.policy.Gate.Probability
	if gate == nil || e.pack == nil || assessment.Scope != e.pack.Kind() {
		return
	}
	if assessment.SlopProbability != nil {
		if *assessment.SlopProbability >= gate.FailAt {
			e.failGate(result, e.probabilityFinding(doc, assessment, gate.FailAt), doc.Name)
		}
		return
	}
	if gate.Require && !slices.Contains(expectedAbstentions, assessment.ProbabilityStatus) {
		e.failGate(result, e.unavailableFinding(doc, assessment), doc.Name)
	}
}

func (e *Engine) probabilityFinding(doc document.Document, assessment Assessment, threshold float64) Finding {
	finding := e.gateFinding(doc, assessment, "probability")
	finding.Message = fmt.Sprintf("%s revision probability %.3g reaches the configured threshold of %.3g.",
		assessment.Scope, *assessment.SlopProbability, threshold)
	finding.Evidence.Metrics = []rule.Metric{{Name: "revision-probability", Value: *assessment.SlopProbability,
		Unit: "probability", Onset: threshold, Saturation: 1}}
	return finding
}

func (e *Engine) unavailableFinding(doc document.Document, assessment Assessment) Finding {
	finding := e.gateFinding(doc, assessment, "probability-unavailable")
	finding.Message = fmt.Sprintf(
		"%s requires a calibrated revision probability; the configured model reported %s%s.",
		assessment.Scope, assessment.ProbabilityStatus, probabilityDetail(assessment))
	return finding
}

func probabilityDetail(assessment Assessment) string {
	if assessment.ProbabilityDetail == "" {
		return ""
	}
	return " (" + assessment.ProbabilityDetail + ")"
}

// gateFinding builds the shared derived diagnostic for one scored unit.
func (e *Engine) gateFinding(doc document.Document, assessment Assessment, kind string) Finding {
	if assessment.BaselineState == "" {
		assessment.BaselineState = "untracked"
	}
	ruleID := "gate." + assessment.Scope + "-" + kind
	fingerprint := fmt.Sprintf("%x",
		sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d:%s", doc.Name, ruleID, assessment.UnitID, doc.Hash))))
	return Finding{
		BaselineFingerprint: assessment.BaselineFingerprint,
		BaselineState:       assessment.BaselineState,
		ID:                  fingerprint[:16],
		RuleID:              ruleID,
		RuleVersion:         "1",
		Severity:            "error",
		Gate:                "forbid",
		Group:               "policy",
		Scope:               assessment.Scope,
		Primary:             e.unitLocation(doc, assessment),
		Related:             []Location{},
		Fingerprint:         fingerprint,
		Derived:             true,
		Evidence:            rule.Evidence{Kind: "exact", Occurrences: []rule.Occurrence{}, Metrics: []rule.Metric{}},
	}
}

func (e *Engine) unitLocation(doc document.Document, assessment Assessment) Location {
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
	return location
}
