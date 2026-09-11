package corpus

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation"
)

// OriginDecisionsVersion identifies a decision set labeled from provenance.
const OriginDecisionsVersion = "unswell-origin-decisions-v1"

// OriginDecisions labels each candidate from the provenance its manifest
// declares. The rule is the origin profile of the annotation package. A
// generated controlled source with document scope yields endpoints. A
// historical cohort yields dated snapshots from before the boundary. Every
// other unit stays unresolved with a reason. The basis names declared
// provenance. It proves no authorship and qualifies nothing. Training and
// evaluation take the set in place of a blinded round.
func OriginDecisions(ctx context.Context, artifact Artifact) (annotation.DecisionSet, error) {
	if err := ctx.Err(); err != nil {
		return annotation.DecisionSet{}, err
	}
	if artifact.SHA256 == "" || artifact.Plan.ManifestSHA256 == "" {
		return annotation.DecisionSet{}, fmt.Errorf("origin decisions need a sealed candidate artifact")
	}
	sources := make(map[string]Source, len(artifact.Plan.Manifest.Sources))
	for _, source := range artifact.Plan.Manifest.Sources {
		sources[source.ID] = source
	}
	result := annotation.DecisionSet{Version: OriginDecisionsVersion,
		RoundID: "provenance:" + artifact.Plan.ManifestSHA256[:12], RoundSHA256: artifact.SHA256,
		PacketSHA256: artifact.Plan.ManifestSHA256, Purpose: "origin", Basis: "declared_provenance",
		Rubric: annotation.OriginRubric, ProfileSHA256: annotation.OriginProfileSHA256(),
		HumanCorpus: "not_qualified", PrimaryRaters: []string{},
		Units: make([]annotation.EditorialDecision, 0, len(artifact.Units))}
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return annotation.DecisionSet{}, err
		}
		source, exists := sources[candidate.SourceID]
		if !exists {
			return annotation.DecisionSet{}, fmt.Errorf("candidate %s names an unknown source", candidate.Unit.ID)
		}
		result.Units = append(result.Units, originDecision(candidate, source))
	}
	slices.SortFunc(result.Units, func(a, b annotation.EditorialDecision) int { return strings.Compare(a.UnitID, b.UnitID) })
	return annotation.FinishDecisions(ctx, result)
}

// originDecision applies the profile to one candidate. The cohort comes from
// the candidate, the origin from its source; neither is read from the text.
func originDecision(candidate Candidate, source Source) annotation.EditorialDecision {
	decision := annotation.EditorialDecision{UnitID: candidate.Unit.ID, Target: annotation.TargetOf(candidate.Unit),
		Status: "unresolved", Basis: "declared_provenance", Categories: []string{}, MissingRaters: []string{}}
	label, reason := originLabel(candidate.Cohort, source.Origin)
	if label == "" {
		decision.Reason = reason
		return decision
	}
	decision.Status, decision.Label = "resolved", &label
	return decision
}

func originLabel(cohort string, origin annotation.Origin) (label, reason string) {
	switch {
	case cohort == "controlled":
		return controlledLabel(origin)
	case strings.HasPrefix(cohort, "historical"):
		if origin.Label == "human" || origin.Label == "unknown" {
			return "human_snapshot", ""
		}
		return "", "origin/" + origin.Label
	case cohort == "contemporary":
		return "", "contemporary_snapshot"
	case cohort == "natural":
		return "", "natural_cohort"
	}
	return "", "no_cohort"
}

// controlledLabel applies the profile to a source of the controlled cohort:
// only a generated document with its generation record is an endpoint.
func controlledLabel(origin annotation.Origin) (label, reason string) {
	switch {
	case origin.Label == "generated" && origin.Scope == "document" && origin.GenerationRecord != "":
		return "endpoint_generated", ""
	case origin.Label == "human_ai_edited":
		return "", "polished_response"
	case origin.GenerationRecord == "":
		return "", "controlled_without_generation_record"
	}
	return "", "origin/" + origin.Label
}
