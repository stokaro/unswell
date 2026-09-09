package annotation

import (
	"crypto/sha256"
	"fmt"
	"slices"
)

func editorialDecision(unit Unit, judgments []Judgment, raters []string, final Adjudication, adjudicated bool) EditorialDecision {
	result := EditorialDecision{UnitID: unit.ID, Target: decisionTarget(unit), Status: "unresolved", Basis: "none",
		PrimaryJudgments: len(judgments), Categories: []string{}, MissingRaters: []string{}}
	answered := make(map[string]bool, len(judgments))
	for _, judgment := range judgments {
		answered[judgment.ActorID] = true
	}
	for _, id := range raters {
		if !answered[id] {
			result.MissingRaters = append(result.MissingRaters, id)
		}
	}
	if len(result.MissingRaters) > 0 {
		result.Reason = "missing_judgments"
		return result
	}
	if adjudicated {
		result.selectLabel(final.Label, final.Categories, "adjudication")
		return result
	}
	label, categories, unanimous := unanimousDecision(judgments)
	if !unanimous {
		result.Reason = "judgment_disagreement"
		return result
	}
	result.selectLabel(label, categories, "unanimous")
	return result
}

func unanimousDecision(judgments []Judgment) (string, []string, bool) {
	if len(judgments) < 2 {
		return "", nil, false
	}
	first := judgments[0]
	categories := sortedCategories(first.Categories)
	for _, judgment := range judgments[1:] {
		if judgment.Label != first.Label || !slices.Equal(sortedCategories(judgment.Categories), categories) {
			return "", nil, false
		}
	}
	return first.Label, categories, true
}

func (d *EditorialDecision) selectLabel(label string, categories []string, basis string) {
	d.Label, d.Categories, d.Basis = &label, sortedCategories(categories), basis
	if label == "uncertain" {
		d.Reason = "uncertain"
		return
	}
	d.Status = "resolved"
}

func sortedCategories(categories []string) []string {
	result := append([]string{}, categories...)
	slices.Sort(result)
	return result
}

func decisionTarget(unit Unit) DecisionTarget {
	allowed := slices.Clone(unit.Rights.AllowedUses)
	slices.Sort(allowed)
	return DecisionTarget{Kind: unit.Kind, Role: unit.Role, SourceSHA256: unit.Source.SHA256, SourceBytes: unit.Source.Bytes,
		SourceFormat: unit.Source.Language, ProseLanguage: unit.Source.ProseLanguage,
		TextSHA256:    fmt.Sprintf("%x", sha256.Sum256([]byte(unit.Text))),
		ContextSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(unit.Context))),
		Segments:      slices.Clone(unit.Source.Segments), Extraction: unit.Extraction, AllowedUses: allowed}
}
