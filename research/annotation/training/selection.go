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
	task      string
	decisions map[string]annotation.EditorialDecision
	measure   func(corpus.FeatureBinding) (measurement, bool)
	reserved  map[string]bool
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
	task, err := annotation.TaskForRubric(decisions.Rubric)
	if err != nil {
		return selection{}, err
	}
	selector.task = task
	selector.decisions = make(map[string]annotation.EditorialDecision)
	selector.reserved = reservedGroups(selector.options)
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
	if s.reservedRow(binding, partition) {
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

// reservedRow counts a row of a reserved partition or a reserved reference
// group and reports whether it stays out.
func (s *rowSelector) reservedRow(binding corpus.FeatureBinding, partition *Partition) bool {
	if binding.Partition == "development" || binding.Partition == "final_test" ||
		(binding.Partition == "calibration" && s.options.Calibration == "none") {
		partition.Excluded["reserved_partition"]++
		return true
	}
	if s.reserved[binding.GroupID] {
		partition.Excluded["reserved_reference_group"]++
		return true
	}
	return false
}

func (s *rowSelector) addResolved(binding corpus.FeatureBinding, unit measurement,
	decision annotation.EditorialDecision, partition *Partition,
) error {
	if !slices.Contains(decision.Target.AllowedUses, "training") {
		return fmt.Errorf("unit %s lacks a declared training permission", binding.UnitID)
	}
	label, err := classLabel(s.task, decision)
	if err != nil {
		return err
	}
	if unit.reason != "" {
		return s.unavailable(binding.UnitID, "target/"+unit.reason, partition)
	}
	if err := s.acceptIdentity(unit.identity); err != nil {
		return err
	}
	values, reason, err := s.values(unit, partition)
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

// values resolves a unit's vector under the missing-feature policy. The zero
// policy fills unavailable activations with zero and counts them on the
// partition by feature and reason.
func (s *rowSelector) values(unit measurement, partition *Partition) ([]float64, string, error) {
	values, reason, err := numericValues(unit.values, unit.identity.Columns)
	if err != nil {
		return nil, "", err
	}
	if reason != "" && s.options.MissingFeatures == "zero" {
		var filled map[string]int
		values, filled = zeroValues(unit.values)
		if partition.ZeroFilled == nil {
			partition.ZeroFilled = make(map[string]int)
		}
		for key, count := range filled {
			partition.ZeroFilled[key] += count
		}
		reason = ""
	}
	if reason != "" {
		countUnavailable(unit.values, partition)
	}
	return values, reason, nil
}

// countUnavailable records every unavailable value of a row the policy
// excludes, so a partition shows each feature's abstentions and not only
// the first column's.
func countUnavailable(values []feature.Value, partition *Partition) {
	if partition.Unavailable == nil {
		partition.Unavailable = make(map[string]int)
	}
	for _, value := range values {
		if value.Number == nil {
			partition.Unavailable[value.ID+"/"+value.Reason]++
		}
	}
}

func (s *rowSelector) unavailable(id, reason string, partition *Partition) error {
	if s.options.MissingFeatures == "reject" {
		category, detail, _ := strings.Cut(reason, "/")
		return fmt.Errorf("unit %s has unavailable %s %s", id, category, detail)
	}
	partition.Excluded[reason]++
	return nil
}

// reservedGroups is the set of source groups a fit excludes under its
// reservation; a fit without one excludes nothing.
func reservedGroups(options Options) map[string]bool {
	if options.Reservation == nil {
		return nil
	}
	result := make(map[string]bool, len(options.Reservation.Groups))
	for _, group := range options.Reservation.Groups {
		result[group] = true
	}
	return result
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
