package patterns_test

import (
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

// screenClasses admits both rules in comments, so both are screened.
func screenClasses() corpus.RuleClasses {
	classes := classes()
	classes.Rules[0].Roles = []string{"comment"}
	return classes
}

// screenSource is the controlled source ID of one response of a run.
func screenSource(run, response string) string {
	return generation.ControlledID(run, "r", generation.ResponsePath(run, response))
}

func screenUnit(source, cohort string, a, b int) corpus.UnitFindings {
	findings := []corpus.FindingRecord{}
	for range a {
		findings = append(findings, corpus.FindingRecord{RuleID: "a.rule"})
	}
	for range b {
		findings = append(findings, corpus.FindingRecord{RuleID: "b.rule"})
	}
	return corpus.UnitFindings{UnitID: "u-" + source, SourceID: source, Cohort: cohort, Kind: "paragraph", Role: "comment",
		Findings: findings}
}

// screenFixture has four tasks in four components. Three sit in the
// development partition and one in training. Family f answers every task
// under generate/neutral; family g answers only the training task. a.rule
// fires in every response and in no original; b.rule fires everywhere. The
// polish response and a sentence unit carry a.rule and enter no arm.
func screenFixture() (corpus.DatasetPlan, []patterns.ScreenRun, corpus.FindingsArtifact) {
	plan := corpus.DatasetPlan{Version: corpus.DatasetVersion, Dataset: corpus.Dataset{ID: "d"}, DatasetSHA256: "sha"}
	tasks := generation.Tasks{Version: generation.TasksVersion, Cohort: "historical"}
	records := generation.Generation{Version: generation.GenerationVersion, Run: "run-1", Family: "f", Model: "m1"}
	other := generation.Generation{Version: generation.GenerationVersion, Run: "run-2", Family: "g", Model: "m2"}
	art := corpus.FindingsArtifact{Version: corpus.FindingsVersion, HumanCorpus: "not_qualified", SHA256: "abc", Policy: policy()}
	for i, partition := range []string{"development", "development", "development", "training"} {
		group := string(rune('A' + i))
		original, task, response := "h"+group, "t"+group, "r"+group
		plan.Sources = append(plan.Sources, corpus.DatasetSource{ID: original, Group: group, Partition: partition},
			corpus.DatasetSource{ID: screenSource("run-1", response), Group: group, Partition: partition},
			corpus.DatasetSource{ID: screenSource("run-1", response+"p"), Group: group, Partition: partition},
			corpus.DatasetSource{ID: screenSource("run-2", response), Group: group, Partition: partition})
		tasks.Tasks = append(tasks.Tasks, generation.Task{ID: task, SourceID: original, UnitID: "u", GroupID: group,
			Role: "comment", Repository: "r", Cohort: "historical"})
		records.Records = append(records.Records,
			generation.Record{ResponseID: response, TaskID: task, Operation: "generate", Prompt: "neutral", Status: "complete"},
			generation.Record{ResponseID: response + "p", TaskID: task, Operation: "polish", Prompt: "neutral", Status: "complete"})
		art.Units = append(art.Units, screenUnit(original, "historical", 0, 1),
			screenUnit(screenSource("run-1", response), "controlled", 1, 1),
			screenUnit(screenSource("run-1", response+"p"), "controlled", 1, 1))
	}
	other.Records = append(other.Records,
		generation.Record{ResponseID: "rD", TaskID: "tD", Operation: "generate", Prompt: "neutral", Status: "complete"})
	art.Units = append(art.Units, screenUnit(screenSource("run-2", "rD"), "controlled", 1, 1))
	sentence := screenUnit("hA", "historical", 1, 0)
	sentence.Kind = "sentence"
	art.Units = append(art.Units, sentence)
	runs := []patterns.ScreenRun{{Records: records, Tasks: tasks}, {Records: other, Tasks: tasks}}
	return plan, runs, art
}

func TestScreeningContrastsFamiliesOnThePartition(t *testing.T) {
	c := qt.New(t)
	plan, runs, art := screenFixture()
	screening, err := patterns.AnalyzeScreen(t.Context(), plan, runs, []corpus.FindingsArtifact{art}, screenClasses(),
		patterns.ScreenOptions{})
	c.Assert(err, qt.IsNil)
	c.Assert(screening.Version, qt.Equals, patterns.ScreeningVersion)
	c.Assert(screening.Partition, qt.Equals, "development")
	c.Assert(screening.Unit, qt.Equals, "paragraph")
	c.Assert(screening.Role, qt.Equals, "comment")
	c.Assert(screening.H0Cohort, qt.Equals, "historical")
	c.Assert(screening.BaselineLoad, qt.Equals, "not_checked")
	// Family f has three responses in the partition; family g has none there.
	c.Assert(screening.Families, qt.DeepEquals, []patterns.ScreenFamily{
		{Family: "f", Models: []string{"m1"}, Runs: []string{"run-1"}, Responses: 4, Documents: 3, Units: 3},
		{Family: "g", Models: []string{"m2"}, Runs: []string{"run-2"}, Responses: 1},
	})
	c.Assert(screening.Components, qt.Equals, 3)
	c.Assert(screening.Rules, qt.HasLen, 2)
	// a.rule: every controlled unit and no original carries it, so D is one
	// in every replicate and the p-value sits at the floor.
	a := screening.Rules[0].Families[0]
	c.Assert(a.Family, qt.Equals, "f")
	c.Assert(a.Controlled, qt.DeepEquals, patterns.ScreenArm{Units: 3, WithFinding: 3, Components: 3, Support: 3, Prevalence: ptr(1)})
	c.Assert(a.H0, qt.DeepEquals, patterns.ScreenArm{Units: 3, Components: 3, Prevalence: ptr(0)})
	c.Assert(a.Models, qt.DeepEquals, []patterns.ScreenModel{{Model: "m1", ScreenArm: a.Controlled}})
	c.Assert(*a.Difference.Value, qt.Equals, 1.0)
	c.Assert(*a.PValue, qt.Equals, 0.0001)
	c.Assert(a.PStatus, qt.Equals, "bootstrap_two_sided")
	c.Assert(a.State, qt.Equals, "inconclusive")
	c.Assert(a.Reason, qt.Equals, "zero_count")
	c.Assert(a.SampleSize, qt.IsNotNil)
	c.Assert(a.SampleSize.DesignEffect, qt.Equals, 1.0)
	// b.rule: equal prevalence gives D zero everywhere, so p is one and the
	// three components fail the cluster minimum.
	b := screening.Rules[1].Families[0]
	c.Assert(*b.Difference.Value, qt.Equals, 0.0)
	c.Assert(*b.PValue, qt.Equals, 1.0)
	c.Assert(b.State, qt.Equals, "inconclusive")
	c.Assert(b.Reason, qt.Equals, "cluster_minimum")
	// Family g counted nothing in the partition: no p-value, no sample size.
	g := screening.Rules[0].Families[1]
	c.Assert(g.Controlled, qt.DeepEquals, patterns.ScreenArm{})
	c.Assert(g.PValue, qt.IsNil)
	c.Assert(g.PStatus, qt.Equals, "nothing_counted")
	c.Assert(g.QValue, qt.IsNil)
	c.Assert(g.SampleSize, qt.IsNil)
	// Two tests entered the adjustment; only a.rule on f is a candidate.
	c.Assert(screening.Tests, qt.Equals, 2)
	c.Assert(*a.QValue, qt.Equals, 0.0002)
	c.Assert(a.PassesFDR, qt.IsTrue)
	c.Assert(b.PassesFDR, qt.IsFalse)
	c.Assert(screening.Candidates, qt.DeepEquals, []patterns.ScreenCandidate{{RuleID: "a.rule", Family: "f", Difference: 1, QValue: 0.0002}})
}

func ptr(value float64) *float64 { return &value }

func TestScreeningTakesPartitionsFromThePlanAndRejectsBadInputs(t *testing.T) {
	c := qt.New(t)
	plan, runs, art := screenFixture()
	// Screening the training partition sees the one task there, in both families.
	screening, err := patterns.AnalyzeScreen(t.Context(), plan, runs, []corpus.FindingsArtifact{art}, screenClasses(),
		patterns.ScreenOptions{Partition: "training"})
	c.Assert(err, qt.IsNil)
	c.Assert(screening.Families[0].Units, qt.Equals, 1)
	c.Assert(screening.Families[1].Units, qt.Equals, 1)
	c.Assert(screening.Rules[0].Families[0].H0.Units, qt.Equals, 1)
	c.Assert(screening.Rules[0].Families[0].PStatus, qt.Equals, "fewer_than_two_components")
	// A rule that admits no comment is not screened.
	screening, err = patterns.AnalyzeScreen(t.Context(), plan, runs, []corpus.FindingsArtifact{art}, classes(), patterns.ScreenOptions{})
	c.Assert(err, qt.IsNil)
	c.Assert(screening.Rules, qt.HasLen, 1)
	c.Assert(screening.Rules[0].RuleID, qt.Equals, "b.rule")
	// A unit whose source the plan does not know is an error, not a drop.
	short := plan
	short.Sources = plan.Sources[1:]
	_, err = patterns.AnalyzeScreen(t.Context(), short, runs, []corpus.FindingsArtifact{art}, screenClasses(), patterns.ScreenOptions{})
	c.Assert(err, qt.ErrorMatches, "source hA of a finding artifact is not in the dataset plan")
	for _, row := range []struct {
		name string
		edit func(runs []patterns.ScreenRun, options *patterns.ScreenOptions)
	}{
		{"unknown task", func(runs []patterns.ScreenRun, _ *patterns.ScreenOptions) { runs[1].Records.Records[0].TaskID = "t9" }},
		{"mixed cohorts", func(runs []patterns.ScreenRun, _ *patterns.ScreenOptions) { runs[1].Tasks.Cohort = "natural" }},
		{"no family", func(runs []patterns.ScreenRun, _ *patterns.ScreenOptions) { runs[0].Records.Family = "" }},
		{"fdr", func(_ []patterns.ScreenRun, options *patterns.ScreenOptions) { options.FDR = 1 }},
		{"mid", func(_ []patterns.ScreenRun, options *patterns.ScreenOptions) { options.MID = -1 }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, runs, art := screenFixture()
			options := patterns.ScreenOptions{}
			row.edit(runs, &options)
			_, err := patterns.AnalyzeScreen(t.Context(), plan, runs, []corpus.FindingsArtifact{art}, screenClasses(), options)
			c.Assert(err, qt.IsNotNil)
		})
	}
	_, err = patterns.AnalyzeScreen(t.Context(), plan, nil, []corpus.FindingsArtifact{art}, screenClasses(), patterns.ScreenOptions{})
	c.Assert(err, qt.IsNotNil)
}

func TestBenjaminiHochbergOnAHandCheckedSet(t *testing.T) {
	c := qt.New(t)
	q := patterns.BenjaminiHochberg([]float64{0.01, 0.04, 0.03, 0.20})
	c.Assert(q, qt.HasLen, 4)
	for i, want := range []float64{0.04, 0.16 / 3, 0.16 / 3, 0.20} {
		c.Assert(math.Abs(q[i]-want) < 1e-12, qt.IsTrue, qt.Commentf("q[%d] = %v", i, q[i]))
	}
	c.Assert(patterns.BenjaminiHochberg(nil), qt.HasLen, 0)
	c.Assert(patterns.BenjaminiHochberg([]float64{0.9}), qt.DeepEquals, []float64{0.9})
}

func TestRequiredUnitsOnAHandComputedExample(t *testing.T) {
	c := qt.New(t)
	// p0 = 0.02, p1 = 0.05: (1.95996 + 0.84162)^2 * (0.0196 + 0.0475) / 0.0009 = 585.2.
	units := patterns.RequiredUnits(0.02, 0.03, 0.05)
	c.Assert(math.Abs(units-585.18) < 0.05, qt.IsTrue, qt.Commentf("%v", units))
	c.Assert(int(math.Ceil(units)), qt.Equals, 586)
	// With design effect 1 and ten units per component, an arm needs 59 components.
	c.Assert(int(math.Ceil(units*1/10)), qt.Equals, 59)
	c.Assert(patterns.RequiredUnits(0.02, 0.03, 0.05/12) > units, qt.IsTrue)
}
