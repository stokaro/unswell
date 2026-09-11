package evaluation_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/evaluation"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// Arms follow the response ID of a source's generation record; a source
// without a record has no arm, and a record the file does not hold is an
// error rather than a missing stratum.
func TestArmsFromRecordsFollowGenerationReferences(t *testing.T) {
	c := qt.New(t)
	candidates := corpus.Artifact{Plan: corpus.Plan{Manifest: corpus.Manifest{Sources: []corpus.Source{
		{ID: "s0", Origin: annotation.Origin{Label: "generated", GenerationRecord: "records.json#r1"}},
		{ID: "s1", Origin: annotation.Origin{Label: "human_ai_edited", GenerationRecord: "records.json#r2"}},
		{ID: "s2", Origin: annotation.Origin{Label: "unknown"}},
	}}}}
	records := generation.Generation{Records: []generation.Record{
		{ResponseID: "r1", Operation: "generate", Prompt: "neutral", Family: "anthropic-claude"},
		{ResponseID: "r2", Operation: "polish", Prompt: "plain", Family: "anthropic-claude"},
	}}
	arms, err := evaluation.ArmsFromRecords(candidates, records)
	c.Assert(err, qt.IsNil)
	c.Assert(arms, qt.DeepEquals, map[string]evaluation.Arm{
		"s0": {Operation: "generate", Prompt: "neutral", Family: "anthropic-claude"},
		"s1": {Operation: "polish", Prompt: "plain", Family: "anthropic-claude"},
	})

	candidates.Plan.Manifest.Sources[0].Origin.GenerationRecord = "records.json#r9"
	_, err = evaluation.ArmsFromRecords(candidates, records)
	c.Assert(err, qt.ErrorMatches, "source s0 names response r9, which the records do not hold")
	candidates.Plan.Manifest.Sources[0].Origin.GenerationRecord = "records.json"
	_, err = evaluation.ArmsFromRecords(candidates, records)
	c.Assert(err, qt.ErrorMatches, "source s0 names a generation record without a response id")
}

// A trial paired with itself yields the candidate estimate of the paired
// procedure with an interval over its groups; one group gives none.
func TestSingleIntervalsResampleGroups(t *testing.T) {
	c := qt.New(t)
	rows := []evaluation.Observation{
		observed("a", "one", 1, 0.9), observed("b", "one", 0, 0.2), observed("c", "two", 1, 0.4),
		observed("d", "two", 0, 0.1), observed("e", "three", 1, 0.8), observed("f", "three", 0, 0.6),
	}
	intervals, err := evaluation.SingleIntervals(t.Context(), rows, 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(intervals.Groups, qt.Equals, 3)
	c.Assert(intervals.Replicates, qt.Equals, 10000)
	c.Assert(intervals.Metrics, qt.Not(qt.HasLen), 0)
	var recall *evaluation.MetricEstimate
	for i := range intervals.Metrics {
		if intervals.Metrics[i].ID == "recall" {
			recall = &intervals.Metrics[i]
		}
	}
	c.Assert(recall, qt.IsNotNil)
	c.Assert(recall.Micro.Value, qt.IsNotNil)
	c.Assert(*recall.Micro.Value, qt.Equals, 2.0/3)
	c.Assert(recall.Micro.Interval.Lower, qt.IsNotNil)
	c.Assert(recall.Micro.Interval.Upper, qt.IsNotNil)

	single, err := evaluation.SingleIntervals(t.Context(), rows[:2], 0.5)
	c.Assert(err, qt.IsNil)
	c.Assert(single.Groups, qt.Equals, 1)
	for _, metric := range single.Metrics {
		c.Assert(metric.Micro.Interval.Lower, qt.IsNil)
	}
}
