package training

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
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
	columns   Identity
	decisions map[string]annotation.EditorialDecision
	sources   map[string]unswell.PreparedFeatureSource
	units     map[string]unswell.PreparedFeatureUnit
	result    selection
}

func selectRows(ctx context.Context, plan corpus.Plan, joined corpus.JoinedArtifact, options Options) (selection, error) {
	columns, err := columnIdentity(joined, options.Kind)
	if err != nil {
		return selection{}, err
	}
	selector := rowSelector{options: options, columns: columns, decisions: make(map[string]annotation.EditorialDecision),
		sources: make(map[string]unswell.PreparedFeatureSource), units: make(map[string]unswell.PreparedFeatureUnit),
		result: selection{partitions: partitionCounts(plan)}}
	for _, decision := range joined.Decisions.Units {
		selector.decisions[decision.UnitID] = decision
	}
	for _, source := range joined.Features.Sources {
		selector.sources[source.Path] = source
		for _, unit := range source.Units {
			selector.units[measurementKey(source.Path, unit.InputHash)] = unit
		}
	}
	bindings := slices.Clone(joined.Bindings)
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
	unit, exists := s.units[measurementKey(binding.Path, binding.FeatureInputHash)]
	if !exists {
		return fmt.Errorf("missing reproduced measurement for %s", binding.UnitID)
	}
	if unit.Binding.Kind != s.options.Kind {
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

func (s *rowSelector) addResolved(binding corpus.FeatureBinding, unit unswell.PreparedFeatureUnit,
	decision annotation.EditorialDecision, partition *Partition,
) error {
	if !slices.Contains(decision.Target.AllowedUses, "training") {
		return fmt.Errorf("unit %s lacks a declared training permission", binding.UnitID)
	}
	label, err := binaryLabel(decision)
	if err != nil {
		return err
	}
	identity := sourceIdentity(s.columns, s.sources[binding.Path])
	if err := s.acceptIdentity(identity); err != nil {
		return err
	}
	values, reason, err := numericValues(unit.Values, identity.Columns)
	if err != nil {
		return err
	}
	if reason != "" {
		if s.options.MissingFeatures == "reject" {
			return fmt.Errorf("unit %s has unavailable feature %s", binding.UnitID, reason)
		}
		partition.Excluded["feature/"+reason]++
		return nil
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
