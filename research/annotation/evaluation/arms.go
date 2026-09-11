package evaluation

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// Arm names the generation arm a controlled source came from: the operation,
// the prompt condition, and the generator family of its record. A source
// without a generation record has no arm.
type Arm struct {
	Operation string
	Prompt    string
	Family    string
}

// Options carry what an evaluation cannot read from its inputs: whether a
// simulated round is allowed, and the arms of the generation records that
// produced the controlled sources, keyed by source ID.
type Options struct {
	AllowSimulation bool
	Arms            map[string]Arm
}

// ArmsFromRecords maps every source whose origin names a generation record to
// the arm of that record. The reference has the form path#response_id; only
// the response ID is matched. A reference to a response the records do not
// hold is an error, so a stale record file cannot drop an arm in silence.
func ArmsFromRecords(candidates corpus.Artifact, records generation.Generation) (map[string]Arm, error) {
	byResponse := make(map[string]Arm, len(records.Records))
	for _, record := range records.Records {
		byResponse[record.ResponseID] = Arm{Operation: record.Operation, Prompt: record.Prompt, Family: record.Family}
	}
	arms := make(map[string]Arm)
	for _, source := range candidates.Plan.Manifest.Sources {
		reference := source.Origin.GenerationRecord
		if reference == "" {
			continue
		}
		_, response, found := strings.Cut(reference, "#")
		if !found || response == "" {
			return nil, fmt.Errorf("source %s names a generation record without a response id", source.ID)
		}
		arm, exists := byResponse[response]
		if !exists {
			return nil, fmt.Errorf("source %s names response %s, which the records do not hold", source.ID, response)
		}
		arms[source.ID] = arm
	}
	return arms, nil
}

// Intervals are cluster bootstrap intervals over provenance groups for the
// metrics of one trial. The trial is paired with itself, so the candidate side
// of the paired procedure is the estimate; groups are the resampling unit, and
// fewer than two groups give no interval.
type Intervals struct {
	Method     string           `json:"method"`
	Replicates int              `json:"replicates"`
	Seed       int              `json:"seed"`
	Groups     int              `json:"groups"`
	Metrics    []MetricEstimate `json:"metrics"`
}

// MetricEstimate is one metric with its unit-weighted and group-averaged
// estimate and interval.
type MetricEstimate struct {
	ID         string   `json:"id"`
	Micro      Estimate `json:"micro"`
	GroupMacro Estimate `json:"group_macro"`
}

// SingleIntervals resamples the source groups of one trial.
func SingleIntervals(ctx context.Context, rows []Observation, constant float64) (Intervals, error) {
	pairs := make([]PairedObservation, len(rows))
	for i, row := range rows {
		pairs[i] = PairedObservation{Candidate: row, Comparator: row}
	}
	slices.SortFunc(pairs, func(a, b PairedObservation) int { return strings.Compare(a.Candidate.UnitID, b.Candidate.UnitID) })
	groups := pairGroups(pairs, [2]float64{constant, constant})
	comparisons, err := bootstrap(ctx, groups)
	if err != nil {
		return Intervals{}, err
	}
	result := Intervals{Method: "source-group-pcg17-percentile-v1", Replicates: bootstrapReplicates, Seed: 17,
		Groups: len(groups), Metrics: make([]MetricEstimate, 0, len(comparisons))}
	for _, comparison := range comparisons {
		result.Metrics = append(result.Metrics, MetricEstimate{ID: comparison.ID,
			Micro: comparison.Micro.Candidate, GroupMacro: comparison.GroupMacro.Candidate})
	}
	return result, nil
}
