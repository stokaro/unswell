package annotation

import (
	"crypto/sha256"
	"fmt"
)

// Tasks name what a binary label answers.
//   - editorial: the unit needs revision under the published rubric.
//   - origin: the unit is the endpoint of a generation record.
//   - cohort: the unit comes from a contemporary snapshot, not a historical one.
//
// The last two are provenance the corpus declares, never a judgment of the
// text, and a cohort artifact never becomes a pack.
const (
	TaskEditorial = "editorial_needs_revision"
	TaskOrigin    = "origin_endpoint"
	TaskCohort    = "cohort_membership"
)

// CohortRubric identifies the cohort labeling rule of the cohort task.
const CohortRubric = "unswell-cohort-membership-v1"

// CohortProfile is the frozen text of that rule.
const CohortProfile = `contemporary_snapshot: a source in the contemporary cohort.
historical_snapshot: a source in a historical cohort, dated before the boundary.
unresolved, with a reason: a controlled response, a natural snapshot, or a
source without a cohort.
No label comes from the text.`

// CohortProfileSHA256 is the digest of CohortProfile.
func CohortProfileSHA256() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(CohortProfile)))
}

// OriginRubric identifies the provenance labeling rule of the origin task.
const OriginRubric = "unswell-origin-endpoint-v1"

// OriginProfile is the frozen text of that rule. Its digest fills the profile
// field of a decision set, so a training artifact records which rule labeled
// its rows.
const OriginProfile = `endpoint_generated: a controlled source, origin generated, document scope, and
a named generation record.
human_snapshot: a historical source, origin human or unknown, dated before the
boundary.
unresolved: every other unit, with a reason. This covers a polished response,
a contemporary or natural snapshot, a mixed or edited origin, and a controlled
source with no generation record.
No label comes from the text.`

// OriginProfileSHA256 is the digest of OriginProfile.
func OriginProfileSHA256() string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(OriginProfile)))
}

// Labels returns the negative and positive label of a task.
func Labels(task string) (negative, positive string, err error) {
	switch task {
	case TaskEditorial:
		return "acceptable", "needs_revision", nil
	case TaskOrigin:
		return "human_snapshot", "endpoint_generated", nil
	case TaskCohort:
		return "historical_snapshot", "contemporary_snapshot", nil
	}
	return "", "", fmt.Errorf("unknown task %q", task)
}

// TaskForRubric returns the task a rubric answers.
func TaskForRubric(rubric string) (string, error) {
	switch rubric {
	case Rubric:
		return TaskEditorial, nil
	case OriginRubric:
		return TaskOrigin, nil
	case CohortRubric:
		return TaskCohort, nil
	}
	return "", fmt.Errorf("unknown rubric %q", rubric)
}

// TargetOf binds a unit's exact source ranges and hashes without its prose.
func TargetOf(unit Unit) DecisionTarget { return decisionTarget(unit) }
