package patterns_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

func sentence(id, group, cohort, role, text string, words int) corpus.Candidate {
	return corpus.Candidate{SourceID: id, GroupID: group, Partition: "training", Cohort: cohort, Words: words,
		Unit: annotation.Unit{ID: "u-" + id, Kind: "sentence", Role: role, Text: text}}
}

// Two cohorts of comments. The historical sentences open with "the client
// opens" once; the contemporary ones open with it three times across two
// components, so the opener and its trigram rank first among the contrasts.
func frequencyFixture() []corpus.Artifact {
	first := corpus.Artifact{Version: corpus.Version, SHA256: "one", Units: []corpus.Candidate{
		sentence("h1", "A", "historical", "comment", "The client opens a connection to the server.", 8),
		sentence("h2", "A", "historical", "comment", "The server closes the connection after a timeout.", 8),
		sentence("h3", "B", "historical", "comment", "A timeout closes the connection.", 5),
		sentence("h4", "B", "historical", "readme", "Install the client with one command.", 6),
		sentence("p1", "A", "historical", "comment", "Ignored paragraph.", 2),
	}}
	first.Units[4].Unit.Kind = "paragraph"
	second := corpus.Artifact{Version: corpus.Version, SHA256: "two", Units: []corpus.Candidate{
		sentence("c1", "C", "contemporary", "comment", "The client opens a connection to the server.", 8),
		sentence("c2", "C", "contemporary", "comment", "The client opens a stream to the server.", 8),
		sentence("c3", "D", "contemporary", "comment", "The client opens a channel, then it waits.", 8),
		sentence("c4", "D", "contemporary", "comment", "The stream stops after 3 attempts.", 5),
		sentence("n1", "E", "", "comment", "No cohort, no count.", 4),
	}}
	return []corpus.Artifact{first, second}
}

func frequencyTables(t *testing.T) patterns.FrequencyTables {
	t.Helper()
	c := qt.New(t)
	frequencies, err := patterns.NewFrequencies(patterns.FrequencyOptions{MinCount: 2, MinComponents: 1, Top: 20})
	c.Assert(err, qt.IsNil)
	inputs := frequencyFixture()
	for _, artifact := range inputs {
		c.Assert(frequencies.Add(t.Context(), artifact), qt.IsNil)
	}
	frequencies.Select()
	for _, artifact := range inputs {
		c.Assert(frequencies.Support(t.Context(), artifact), qt.IsNil)
	}
	return frequencies.Tables()
}

func measuresOf(tables patterns.FrequencyTables) map[string]patterns.FrequencyMeasure {
	measures := map[string]patterns.FrequencyMeasure{}
	for _, measure := range tables.Measures {
		measures[measure.Measure] = measure
	}
	return measures
}

func TestFrequenciesCountStrataAndItems(t *testing.T) {
	c := qt.New(t)
	tables := frequencyTables(t)
	c.Assert(tables.Version, qt.Equals, patterns.FrequencyVersion)
	c.Assert(tables.Unit, qt.Equals, "sentence")
	c.Assert(tables.Baseline, qt.Equals, "historical")
	c.Assert(tables.MinCount, qt.Equals, 2)
	c.Assert(tables.Targets, qt.DeepEquals, []string{"contemporary"})
	c.Assert(tables.Inputs, qt.HasLen, 2)
	// Strata in order of first sight; the paragraph and the cohort-less unit count nowhere.
	c.Assert(tables.Strata, qt.DeepEquals, []patterns.Stratum{
		{Cohort: "historical", Role: "comment", Sentences: 3, Words: 21, Components: 2},
		{Cohort: "historical", Role: "readme", Sentences: 1, Words: 6, Components: 1},
		{Cohort: "contemporary", Role: "comment", Sentences: 4, Words: 29, Components: 2}})
	measures := measuresOf(tables)
	c.Assert(measures, qt.HasLen, 5)
	var opener patterns.FrequencyItem
	for _, item := range measures[patterns.MeasureOpener].Items {
		if item.Key == "the client opens" {
			opener = item
		}
	}
	c.Assert(opener.Key, qt.Equals, "the client opens")
	// The historical count of one stays below the minimum, so only the contemporary cell is listed;
	// an item is listed at all only because the key enters a contrast.
	c.Assert(opener.Cells, qt.DeepEquals, map[string]patterns.FrequencyCell{
		"contemporary/comment": {Count: 3, PerThousandWords: 1000 * 3.0 / 29, Components: 2}})
	// A number breaks the n-gram window: "after 3 attempts" yields no trigram.
	for _, item := range measures[patterns.MeasureWord3].Items {
		c.Assert(item.Key, qt.Not(qt.Contains), "3")
	}
	c.Assert(measures[patterns.MeasureTemplate].Items, qt.Not(qt.HasLen), 0)
	for _, item := range measures[patterns.MeasureTemplate].Items {
		c.Assert(item.Key, qt.Contains, " ")
	}
}

func TestFrequenciesRankContrasts(t *testing.T) {
	c := qt.New(t)
	measures := measuresOf(frequencyTables(t))
	openers := measures[patterns.MeasureOpener]
	c.Assert(openers.Contrasts, qt.Not(qt.HasLen), 0)
	top := openers.Contrasts[0]
	c.Assert(top.Cohort, qt.Equals, "contemporary")
	c.Assert(top.Role, qt.Equals, "comment")
	c.Assert(top.Key, qt.Equals, "the client opens")
	// The baseline holds the opener once, below the minimum, so the ratio stays absent.
	c.Assert(top.RatioStatus, qt.Equals, "baseline_below_minimum")
	c.Assert(top.Ratio, qt.IsNil)
	c.Assert(top.Baseline.Count, qt.Equals, 1)
	c.Assert(top.Target.Count, qt.Equals, 3)
	words := measuresOf(frequencyTables(t))[patterns.MeasureWord1]
	var the patterns.FrequencyContrast
	for _, contrast := range words.Contrasts {
		if contrast.Key == "the" {
			the = contrast
		}
	}
	c.Assert(the.RatioStatus, qt.Equals, "defined")
	c.Assert(*the.Ratio, qt.Equals, (1000*float64(the.Target.Count)/29)/(1000*float64(the.Baseline.Count)/21))
	c.Assert(words.Contrasts[0].RatioStatus, qt.Equals, "defined")
	// The readme stratum has no contemporary counterpart, so it contrasts nothing,
	// and a key the baseline never carries stays without a ratio.
	statuses := map[string]string{}
	for _, contrast := range measures[patterns.MeasureWord1].Contrasts {
		c.Assert(contrast.Role, qt.Equals, "comment")
		statuses[contrast.Key] = contrast.RatioStatus
	}
	c.Assert(statuses["stream"], qt.Equals, "absent_in_baseline")
}

func TestFrequenciesRefuseSupportBeforeSelection(t *testing.T) {
	c := qt.New(t)
	frequencies, err := patterns.NewFrequencies(patterns.FrequencyOptions{})
	c.Assert(err, qt.IsNil)
	c.Assert(frequencies.Support(t.Context(), frequencyFixture()[0]), qt.ErrorMatches, "select keys before .*")
	tables := frequencies.Tables()
	c.Assert(tables.MinCount, qt.Equals, 5)
	c.Assert(tables.MinComponents, qt.Equals, 3)
	c.Assert(tables.Top, qt.Equals, 200)
	c.Assert(tables.Measures, qt.HasLen, 5)
}

// A task names the paragraph of h1 and h2; a generated response answers it.
// Paired, those two sentences count under historical-paired and the response
// under controlled-generate, while every other unit keeps its cohort.
func TestFrequenciesPairTasksAndResponses(t *testing.T) {
	c := qt.New(t)
	inputs := frequencyFixture()
	paragraph := "The client opens a connection to the server.  The server closes the connection after a timeout."
	inputs[0].Units[0].Unit.Context = paragraph
	inputs[0].Units[1].Unit.Context = paragraph
	inputs[0].Units[1].SourceID = "h1"
	response := sentence("r1", "R", "controlled", "comment", "The client opens a connection and waits.", 7)
	response.Unit.Source.Reference = "generation:run/r1"
	inputs[1].Units = append(inputs[1].Units, response)
	frequencies, err := patterns.NewFrequencies(patterns.FrequencyOptions{Baseline: "historical-paired", MinCount: 1, MinComponents: 1})
	c.Assert(err, qt.IsNil)
	tasks := generation.Tasks{Cohort: "historical", Tasks: []generation.Task{{ID: "t1", SourceID: "h1", UnitID: "p1", Text: paragraph}}}
	records := generation.Generation{Records: []generation.Record{{ResponseID: "r1", TaskID: "t1", Operation: "generate"}}}
	c.Assert(frequencies.Pair(tasks, records), qt.IsNil)
	stray := generation.Generation{Records: []generation.Record{{ResponseID: "r2", TaskID: "t9", Operation: "polish"}}}
	c.Assert(frequencies.Pair(tasks, stray), qt.ErrorMatches, "record r2 names task t9 .*")
	for _, artifact := range inputs {
		c.Assert(frequencies.Add(t.Context(), artifact), qt.IsNil)
	}
	frequencies.Select()
	for _, artifact := range inputs {
		c.Assert(frequencies.Support(t.Context(), artifact), qt.IsNil)
	}
	tables := frequencies.Tables()
	c.Assert(tables.Pairing, qt.DeepEquals, &patterns.FrequencyPairing{Cohort: "historical", Tasks: 2, Responses: 1})
	sizes := map[string]int{}
	for _, stratum := range tables.Strata {
		sizes[stratum.Cohort+"/"+stratum.Role] = stratum.Sentences
	}
	c.Assert(sizes, qt.DeepEquals, map[string]int{"historical-paired/comment": 2, "historical/comment": 1, "historical/readme": 1,
		"contemporary/comment": 4, "controlled-generate/comment": 1})
	c.Assert(tables.Targets, qt.DeepEquals, []string{"contemporary", "controlled-generate", "historical"})
	var opener patterns.FrequencyContrast
	for _, contrast := range tables.Measures[3].Contrasts {
		if contrast.Cohort == "controlled-generate" && contrast.Key == "the client opens" {
			opener = contrast
		}
	}
	c.Assert(opener.Key, qt.Equals, "the client opens")
	c.Assert(opener.Baseline.Count, qt.Equals, 1)
	c.Assert(opener.Target.Count, qt.Equals, 1)
}
