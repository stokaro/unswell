package annotation

import (
	"crypto/sha256"
	"fmt"
)

// Tasks name what a binary label answers. The editorial task asks whether a
// unit needs revision under the published rubric. The origin task asks whether
// a unit is the endpoint of a generation record. That answer is provenance the
// corpus declares, never a judgment of the text.
const (
	TaskEditorial = "editorial_needs_revision"
	TaskOrigin    = "origin_endpoint"
)

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
	}
	return "", fmt.Errorf("unknown rubric %q", rubric)
}

// TargetOf binds a unit's exact source ranges and hashes without its prose.
func TargetOf(unit Unit) DecisionTarget { return decisionTarget(unit) }
