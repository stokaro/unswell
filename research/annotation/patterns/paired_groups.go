package patterns

import (
	"fmt"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// pairedGroups uses one global plan for both sides of every pair and H0.
// Archived tasks and individual shards may contain narrower local groups.
func pairedGroups(plan corpus.DatasetPlan, tasks generation.Tasks, inputs []corpus.FindingsArtifact) (map[string]string, error) {
	if plan.Version != corpus.DatasetVersion {
		return nil, fmt.Errorf("paired analysis requires a global dataset plan")
	}
	groups := make(map[string]string, len(plan.Sources))
	for _, source := range plan.Sources {
		if source.ID == "" || source.Group == "" || groups[source.ID] != "" {
			return nil, fmt.Errorf("paired analysis requires unique source IDs with global groups")
		}
		groups[source.ID] = source.Group
	}
	for _, task := range tasks.Tasks {
		if groups[task.SourceID] == "" {
			return nil, fmt.Errorf("task %s names a source outside the global dataset plan", task.ID)
		}
	}
	for _, input := range inputs {
		if err := checkPairedSourceGroups(input, groups); err != nil {
			return nil, err
		}
	}
	return groups, nil
}

func checkPairedSourceGroups(input corpus.FindingsArtifact, groups map[string]string) error {
	for _, doc := range input.Documents {
		if groups[doc.SourceID] == "" {
			return fmt.Errorf("document %s is outside the global dataset plan", doc.SourceID)
		}
	}
	for _, unit := range input.Units {
		if groups[unit.SourceID] == "" {
			return fmt.Errorf("unit %s names a source outside the global dataset plan", unit.UnitID)
		}
	}
	return nil
}
