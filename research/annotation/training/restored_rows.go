package training

import (
	"fmt"
	"slices"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func validateRestoredPartitions(a Artifact) error {
	names := []string{"training", "development", "calibration", "final_test"}
	if len(a.Partitions) != len(names) {
		return fmt.Errorf("training artifact requires all four partitions")
	}
	seen := make(map[string]bool)
	for i, partition := range a.Partitions {
		if partition.Name != names[i] || partition.Candidates < 0 || partition.Sources < 0 || partition.Groups < 0 {
			return fmt.Errorf("training artifact has invalid partition counts or order")
		}
		if err := validateRestoredPartition(a.Identity.Task, partition, a.Options, seen); err != nil {
			return err
		}
	}
	return nil
}

func validateRestoredPartition(task string, partition Partition, options Options, seen map[string]bool) error {
	reserved := partitionReserved(partition.Name, options)
	if reserved && (len(partition.Rows) != 0 || len(partition.Classes) != 0) {
		return fmt.Errorf("reserved partition cannot contain fitted rows or class statistics")
	}
	if err := validateFittedCounts(task, partition); err != nil {
		return err
	}
	for _, counts := range []map[string]int{partition.ZeroFilled, partition.Unavailable} {
		for _, count := range counts {
			if count < 0 {
				return fmt.Errorf("training artifact has a negative abstention count")
			}
		}
	}
	for _, row := range partition.Rows {
		if !validFittedRow(row) || seen[row.UnitID] {
			return fmt.Errorf("fitted rows require unique IDs and valid source/group/input identities")
		}
		seen[row.UnitID] = true
	}
	return nil
}

func validateFittedTargets(a Artifact, candidates corpus.Artifact) error {
	targets := make(map[string]corpus.Candidate, len(candidates.Units))
	for _, candidate := range candidates.Units {
		targets[candidate.Unit.ID] = candidate
	}
	reserved := reservedGroups(a.Options)
	for _, partition := range a.Partitions {
		for _, row := range partition.Rows {
			target, exists := targets[row.UnitID]
			if !exists || target.Partition != partition.Name || target.Unit.Kind != a.Options.Kind ||
				target.SourceID != row.SourceID || target.GroupID != row.GroupID {
				return fmt.Errorf("fitted row is inconsistent with the frozen source-group partition")
			}
			if reserved[row.GroupID] {
				return fmt.Errorf("fitted row %s belongs to a reserved group", row.UnitID)
			}
		}
	}
	return nil
}

func validateFittedCounts(task string, partition Partition) error {
	negative, positive, err := annotation.Labels(task)
	if err != nil {
		return err
	}
	count := 0
	for label, value := range partition.Classes {
		if value < 0 || value > len(partition.Rows) || !slices.Contains([]string{negative, positive}, label) {
			return fmt.Errorf("training artifact has invalid fitted class counts")
		}
		count += value
	}
	if count != len(partition.Rows) || count > partition.Candidates {
		return fmt.Errorf("fitted class counts do not match rows or candidate count")
	}
	return nil
}

func validFittedRow(row Row) bool {
	return row.UnitID != "" && row.SourceID != "" && row.GroupID != "" && validDigest(row.FeatureInputHash)
}

func partitionReserved(name string, options Options) bool {
	return name == "development" || name == "final_test" || (name == "calibration" && options.Calibration == "none")
}
