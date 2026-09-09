package training

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/model"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

type selection struct {
	identity              Identity
	partitions            []Partition
	training, calibration []model.Example
	identitySet           bool
}

type rowSelector struct {
	options   Options
	decisions map[string]annotation.EditorialDecision
	measure   func(corpus.FeatureBinding) (measurement, bool)
	result    selection
}

type measurement struct {
	kind     string
	identity Identity
	values   []feature.Value
	reason   string
}

func selectMeasuredRows(ctx context.Context, plan corpus.Plan, decisions annotation.DecisionSet,
	bindings []corpus.FeatureBinding, selector rowSelector,
) (selection, error) {
	selector.decisions = make(map[string]annotation.EditorialDecision)
	selector.result = selection{partitions: partitionCounts(plan)}
	for _, decision := range decisions.Units {
		selector.decisions[decision.UnitID] = decision
	}
	bindings = slices.Clone(bindings)
	slices.SortFunc(bindings, func(a, b corpus.FeatureBinding) int { return strings.Compare(a.UnitID, b.UnitID) })
	for _, binding := range bindings {
		if err := ctx.Err(); err != nil {
			return selection{}, err
		}
		if err := selector.add(binding); err != nil {
			return selection{}, err
		}
	}
	return selector.result, ctx.Err()
}

func (s *rowSelector) add(binding corpus.FeatureBinding) error {
	index := partitionIndex(binding.Partition)
	if index < 0 {
		return fmt.Errorf("unknown partition for %s", binding.UnitID)
	}
	partition := &s.result.partitions[index]
	partition.Candidates++
	if binding.Partition == "development" || binding.Partition == "final_test" ||
		(binding.Partition == "calibration" && s.options.Calibration == "none") {
		partition.Excluded["reserved_partition"]++
		return nil
	}
	unit, exists := s.measure(binding)
	if !exists {
		return fmt.Errorf("missing reproduced measurement for %s", binding.UnitID)
	}
	if unit.kind != s.options.Kind {
		partition.Excluded["unselected_kind"]++
		return nil
	}
	decision, exists := s.decisions[binding.UnitID]
	if !exists {
		partition.Excluded["unannotated"]++
		return nil
	}
	if decision.Status != "resolved" {
		partition.Excluded["decision/"+decision.Reason]++
		return nil
	}
	return s.addResolved(binding, unit, decision, partition)
}

func (s *rowSelector) addResolved(binding corpus.FeatureBinding, unit measurement,
	decision annotation.EditorialDecision, partition *Partition,
) error {
	if !slices.Contains(decision.Target.AllowedUses, "training") {
		return fmt.Errorf("unit %s lacks a declared training permission", binding.UnitID)
	}
	label, err := binaryLabel(decision)
	if err != nil {
		return err
	}
	if unit.reason != "" {
		return s.unavailable(binding.UnitID, "target/"+unit.reason, partition)
	}
	if err := s.acceptIdentity(unit.identity); err != nil {
		return err
	}
	values, reason, err := numericValues(unit.values, unit.identity.Columns)
	if err != nil {
		return err
	}
	if reason != "" {
		return s.unavailable(binding.UnitID, "feature/"+reason, partition)
	}
	example := model.Example{Values: values, Label: label}
	if binding.Partition == "training" {
		s.result.training = append(s.result.training, example)
	} else {
		s.result.calibration = append(s.result.calibration, example)
	}
	partition.Rows = append(partition.Rows, Row{UnitID: binding.UnitID, SourceID: binding.SourceID,
		GroupID: binding.GroupID, FeatureInputHash: binding.FeatureInputHash})
	partition.Classes[*decision.Label]++
	return nil
}

func (s *rowSelector) unavailable(id, reason string, partition *Partition) error {
	if s.options.MissingFeatures == "reject" {
		category, detail, _ := strings.Cut(reason, "/")
		return fmt.Errorf("unit %s has unavailable %s %s", id, category, detail)
	}
	partition.Excluded[reason]++
	return nil
}

func binaryLabel(decision annotation.EditorialDecision) (int, error) {
	if decision.Label != nil {
		switch *decision.Label {
		case "acceptable":
			return 0, nil
		case "needs_revision":
			return 1, nil
		}
	}
	return 0, fmt.Errorf("resolved unit %s requires an acceptable or needs_revision label", decision.UnitID)
}

func measurementKey(path, hash string) string { return path + "\x00" + hash }

func partitionIndex(name string) int {
	return slices.Index([]string{"training", "development", "calibration", "final_test"}, name)
}

func partitionCounts(plan corpus.Plan) []Partition {
	result := make([]Partition, 0, 4)
	for _, name := range []string{"training", "development", "calibration", "final_test"} {
		result = append(result, Partition{Name: name, Excluded: make(map[string]int), Rows: []Row{}, Classes: make(map[string]int)})
	}
	for _, group := range plan.Groups {
		p := &result[partitionIndex(group.Partition)]
		p.Groups++
		p.Sources += len(group.Sources)
	}
	return result
}
