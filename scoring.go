package unswell

import (
	"cmp"
	"slices"

	"github.com/stokaro/unswell/document"
)

type localFinding struct {
	finding Finding
	spans   []document.Span
	points  int
}

func (e *Engine) assess(result *RunResult, doc document.Document) {
	for _, block := range doc.Blocks {
		if block.Words == 0 {
			continue
		}
		result.Assessments = append(
			result.Assessments,
			e.assessment(result.Findings, doc.Name, "paragraph", block.ID, block.Span, block.Words),
		)
		for _, sentence := range block.Sentences {
			if sentence.Words == 0 {
				continue
			}
			result.Assessments = append(
				result.Assessments,
				e.assessment(result.Findings, doc.Name, "sentence", sentence.ID, sentence.Span, sentence.Words),
			)
		}
	}
	e.summarize(result)
}

func (e *Engine) assessment(findings []Finding, path, scope string, id int, span document.Span, words int) Assessment {
	assessment := Assessment{
		Path:              path,
		Scope:             scope,
		UnitID:            id,
		Span:              span,
		Words:             words,
		Status:            "available",
		ProbabilityStatus: "calibration_unavailable",
		Contributions:     []Contribution{},
	}
	locals := e.localFindings(findings, scope, id)
	slices.SortStableFunc(locals, func(a, b localFinding) int { return cmp.Compare(b.points, a.points) })
	accepted := make([]localFinding, 0)
	ruleTotals := make(map[string]int)
	groupTotals := make(map[string]int)
	total := 0
	for _, local := range locals {
		settings := e.policy.Rules[local.finding.RuleID]
		groupCap, exists := e.policy.GroupCaps[local.finding.Group]
		if !exists {
			groupCap = 40
		}
		trace := Contribution{FindingID: local.finding.ID, RuleID: local.finding.RuleID, Group: local.finding.Group,
			Weight: settings.Score.Weight, Activation: local.finding.Evidence.Activation, Raw: float64(local.points) / 1000,
			RuleCap: settings.Score.Cap, GroupCap: groupCap, Reason: "contributed"}
		points := local.points
		if correlated(local, accepted) {
			points = 0
			trace.Reason = "duplicate-correlated-evidence"
		} else {
			accepted = append(accepted, local)
		}
		before := points
		points = min(points, max(0, settings.Score.Cap*1000-ruleTotals[local.finding.RuleID]))
		ruleTotals[local.finding.RuleID] += points
		if points < before {
			trace.Reason = "rule-cap"
		}
		before = points
		points = min(points, max(0, groupCap*1000-groupTotals[local.finding.Group]))
		groupTotals[local.finding.Group] += points
		if points < before {
			trace.Reason = "group-cap"
		}
		before = points
		points = min(points, max(0, 100000-total))
		if points < before {
			trace.Reason = "index-cap"
		}
		total += points
		trace.Effective = float64(points) / 1000
		assessment.Contributions = append(assessment.Contributions, trace)
	}
	assessment.SlopScore = float64(total) / 1000
	assessment.EffectiveSlopScore = assessment.SlopScore
	return assessment
}

func (e *Engine) localFindings(findings []Finding, scope string, id int) []localFinding {
	result := make([]localFinding, 0)
	for _, finding := range findings {
		spans := make([]document.Span, 0)
		for _, occurrence := range finding.Evidence.Occurrences {
			matches := scope == "paragraph" && occurrence.BlockID == id || scope == "sentence" && occurrence.SentenceID == id
			if matches {
				spans = append(spans, occurrence.Spans...)
			}
		}
		if len(spans) == 0 {
			continue
		}
		settings := e.policy.Rules[finding.RuleID]
		result = append(
			result,
			localFinding{finding: finding, spans: spans, points: settings.Score.Weight * finding.Evidence.Activation},
		)
	}
	return result
}

func correlated(candidate localFinding, accepted []localFinding) bool {
	span := document.Bounds(candidate.spans)
	for _, other := range accepted {
		if other.finding.Group != candidate.finding.Group {
			continue
		}
		otherSpan := document.Bounds(other.spans)
		if span.Start >= otherSpan.Start && span.End <= otherSpan.End || otherSpan.Start >= span.Start && otherSpan.End <= span.End {
			return true
		}
	}
	return false
}

func (e *Engine) summarize(result *RunResult) {
	values := make([]float64, 0)
	eligible, flagged := 0, 0
	for _, assessment := range result.Assessments {
		if assessment.Scope != "paragraph" {
			continue
		}
		values = append(values, assessment.SlopScore)
		if assessment.Words >= e.policy.Gate.Paragraph.MinWords {
			eligible++
			if assessment.SlopScore >= float64(e.policy.Gate.Paragraph.FailAt) {
				flagged++
			}
		}
	}
	if len(values) == 0 {
		return
	}
	slices.Sort(values)
	doc := &result.Documents[0]
	doc.Maximum = values[len(values)-1]
	doc.Median = values[(len(values)-1)/2]
	doc.P90 = values[(len(values)*9+9)/10-1]
	if eligible > 0 {
		doc.FlaggedFraction = float64(flagged) / float64(eligible)
	}
}
