package training

import (
	"fmt"

	"github.com/stokaro/unswell/research/annotation"
)

// consistentTask reports whether an identity's task answers its rubric. A
// fitted artifact names both, and a restored one is refused when they
// disagree, so a pack cannot carry editorial rows under the origin task.
func consistentTask(identity Identity) bool {
	task, err := annotation.TaskForRubric(identity.Rubric)
	return err == nil && task == identity.Task
}

// classLabel maps one resolved decision to the binary class of a task.
func classLabel(task string, decision annotation.EditorialDecision) (int, error) {
	negative, positive, err := annotation.Labels(task)
	if err != nil {
		return 0, err
	}
	if decision.Label != nil {
		switch *decision.Label {
		case negative:
			return 0, nil
		case positive:
			return 1, nil
		}
	}
	return 0, fmt.Errorf("resolved unit %s requires a %s or %s label", decision.UnitID, negative, positive)
}

// classCounts returns the fitted negative and positive counts of a partition
// under the labels of a task.
func classCounts(task string, partition Partition) (negative, positive int, err error) {
	negativeLabel, positiveLabel, err := annotation.Labels(task)
	if err != nil {
		return 0, 0, err
	}
	return partition.Classes[negativeLabel], partition.Classes[positiveLabel], nil
}
