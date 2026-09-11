package corpus

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation"
)

// CohortDecisionsVersion identifies a decision set labeled from cohorts.
const CohortDecisionsVersion = "unswell-cohort-decisions-v1"

// CohortDecisions labels each candidate from its cohort under the cohort
// profile of the annotation package. A contemporary snapshot is the positive
// class and a historical snapshot the negative one. Controlled responses,
// natural snapshots, and sources without a cohort stay unresolved with a
// reason. The basis names declared provenance. The set feeds the comparison
// of feature families on the pattern cohorts; it proves nothing about origin
// and never becomes a pack.
func CohortDecisions(ctx context.Context, artifact Artifact) (annotation.DecisionSet, error) {
	if err := ctx.Err(); err != nil {
		return annotation.DecisionSet{}, err
	}
	if artifact.SHA256 == "" || artifact.Plan.ManifestSHA256 == "" {
		return annotation.DecisionSet{}, fmt.Errorf("cohort decisions need a sealed candidate artifact")
	}
	result := annotation.DecisionSet{Version: CohortDecisionsVersion,
		RoundID: "cohort:" + artifact.Plan.ManifestSHA256[:12], RoundSHA256: artifact.SHA256,
		PacketSHA256: artifact.Plan.ManifestSHA256, Purpose: "cohort", Basis: "declared_provenance",
		Rubric: annotation.CohortRubric, ProfileSHA256: annotation.CohortProfileSHA256(),
		HumanCorpus: "not_qualified", PrimaryRaters: []string{},
		Units: make([]annotation.EditorialDecision, 0, len(artifact.Units))}
	for _, candidate := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return annotation.DecisionSet{}, err
		}
		result.Units = append(result.Units, cohortDecision(candidate))
	}
	slices.SortFunc(result.Units, func(a, b annotation.EditorialDecision) int { return strings.Compare(a.UnitID, b.UnitID) })
	return annotation.FinishDecisions(ctx, result)
}

func cohortDecision(candidate Candidate) annotation.EditorialDecision {
	decision := annotation.EditorialDecision{UnitID: candidate.Unit.ID, Target: annotation.TargetOf(candidate.Unit),
		Status: "unresolved", Basis: "declared_provenance", Categories: []string{}, MissingRaters: []string{}}
	label, reason := cohortLabel(candidate.Cohort)
	if label == "" {
		decision.Reason = reason
		return decision
	}
	decision.Status, decision.Label = "resolved", &label
	return decision
}

func cohortLabel(cohort string) (label, reason string) {
	switch {
	case cohort == "contemporary":
		return "contemporary_snapshot", ""
	case strings.HasPrefix(cohort, "historical"):
		return "historical_snapshot", ""
	case cohort == "controlled":
		return "", "controlled_response"
	case cohort == "natural":
		return "", "natural_cohort"
	}
	return "", "no_cohort"
}
