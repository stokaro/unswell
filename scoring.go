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

func (e *Engine) assess(result *RunResult, doc document.Document, estimates probabilityRun) {
	for _, block := range doc.Blocks {
		if block.Words == 0 {
			continue
		}
		result.Assessments = append(
			result.Assessments,
			e.assessment(result.Findings, estimates, assessedUnit{doc.Name, "paragraph", block.ID, block.Span, block.Words}),
		)
		for _, sentence := range block.Sentences {
			if sentence.Words == 0 {
				continue
			}
			result.Assessments = append(
				result.Assessments,
				e.assessment(result.Findings, estimates,
					assessedUnit{doc.Name, "sentence", sentence.ID, sentence.Span, sentence.Words}),
			)
		}
	}
	e.summarize(result)
}

// assessedUnit names one scored target without widening the scoring signature.
type assessedUnit struct {
	path, scope string
	id          int
	span        document.Span
	words       int
}

func (e *Engine) assessment(findings []Finding, estimates probabilityRun, unit assessedUnit) Assessment {
	value, status, detail := estimates.probabilityFor(unit.scope, unit.span)
	assessment := Assessment{
		Path:              unit.path,
		Scope:             unit.scope,
		UnitID:            unit.id,
		Span:              unit.span,
		Words:             unit.words,
		Status:            "available",
		SlopProbability:   value,
		ProbabilityStatus: status,
		ProbabilityDetail: detail,
		Contributions:     []Contribution{},
	}
	scope, id := unit.scope, unit.id
	locals := e.localFindings(findings, scope, id)
	assessment.SlopScore, assessment.Contributions = e.scoreLocals(locals)
	assessment.EffectiveSlopScore = assessment.SlopScore
	if slices.ContainsFunc(locals, func(local localFinding) bool { return local.finding.Suppressed }) {
		suppressed := make(map[string]bool)
		for _, local := range locals {
			if local.finding.Suppressed {
				suppressed[local.finding.ID] = true
			}
		}
		active := slices.DeleteFunc(slices.Clone(locals), func(local localFinding) bool { return local.finding.Suppressed })
		assessment.EffectiveSlopScore, assessment.EffectiveContributions = e.scoreLocals(active)
		for _, local := range assessment.Contributions {
			if suppressed[local.FindingID] {
				local.Effective, local.Reason = 0, "source-suppression"
				assessment.EffectiveContributions = append(assessment.EffectiveContributions, local)
			}
		}
	}
	return assessment
}

func (e *Engine) scoreLocals(locals []localFinding) (float64, []Contribution) {
	contributions := []Contribution{}
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
		contributions = append(contributions, trace)
	}
	return float64(total) / 1000, contributions
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
