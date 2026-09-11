package corpus_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestCohortDecisionsLabelFromCohorts(t *testing.T) {
	c := qt.New(t)
	manifest, files := provenanceSample()
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)

	decisions, err := corpus.CohortDecisions(t.Context(), artifact)
	c.Assert(err, qt.IsNil)
	c.Assert(decisions.Version, qt.Equals, corpus.CohortDecisionsVersion)
	c.Assert(decisions.Rubric, qt.Equals, annotation.CohortRubric)
	c.Assert(decisions.ProfileSHA256, qt.Equals, annotation.CohortProfileSHA256())
	c.Assert(decisions.Basis, qt.Equals, "declared_provenance")
	c.Assert(decisions.RoundSHA256, qt.Equals, artifact.SHA256)
	c.Assert(decisions.Units, qt.HasLen, len(artifact.Units))

	bySource := map[string]corpus.Candidate{}
	for _, candidate := range artifact.Units {
		bySource[candidate.Unit.ID] = candidate
	}
	outcomes := map[string]map[string]int{}
	for _, decision := range decisions.Units {
		candidate := bySource[decision.UnitID]
		c.Assert(decision.Target, qt.DeepEquals, annotation.TargetOf(candidate.Unit))
		key := decision.Reason
		if decision.Label != nil {
			key = *decision.Label
		}
		if outcomes[candidate.SourceID] == nil {
			outcomes[candidate.SourceID] = map[string]int{}
		}
		outcomes[candidate.SourceID][key]++
	}
	c.Assert(outcomes["s0"], qt.DeepEquals, map[string]int{"controlled_response": 2})
	c.Assert(outcomes["s1"], qt.DeepEquals, map[string]int{"controlled_response": 2})
	c.Assert(outcomes["s2"], qt.DeepEquals, map[string]int{"historical_snapshot": 2})
	c.Assert(outcomes["s3"], qt.DeepEquals, map[string]int{"contemporary_snapshot": 2})

	joined, err := corpus.JoinDecisions(t.Context(), artifact, decisions, files, []string{"prose-words"})
	c.Assert(err, qt.IsNil)
	c.Assert(joined.Decisions.Rubric, qt.Equals, annotation.CohortRubric)
	_, err = corpus.CohortDecisions(t.Context(), corpus.Artifact{})
	c.Assert(err, qt.ErrorMatches, "cohort decisions need a sealed candidate artifact")
	task, err := annotation.TaskForRubric(annotation.CohortRubric)
	c.Assert(err, qt.IsNil)
	c.Assert(task, qt.Equals, annotation.TaskCohort)
	negative, positive, err := annotation.Labels(annotation.TaskCohort)
	c.Assert(err, qt.IsNil)
	c.Assert(negative, qt.Equals, "historical_snapshot")
	c.Assert(positive, qt.Equals, "contemporary_snapshot")
}
