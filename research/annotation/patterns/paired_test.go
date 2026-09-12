package patterns_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
	"github.com/stokaro/unswell/research/annotation/patterns"
)

// Two tasks in two components. Task 1's original carries a.rule; its
// generate response does not, and its polish response does. Task 2's original
// carries nothing; its generate response carries a.rule, and its polish
// response was refused. b.rule fires nowhere in the arms. An earlier run's
// response to task 1 shares the path of this run's generate response and
// carries both rules; it belongs to another run, so the analysis skips it.
func pairedFixture() (generation.Generation, generation.Tasks, corpus.FindingsArtifact) {
	tasks := generation.Tasks{Version: generation.TasksVersion, Cohort: "historical", Tasks: []generation.Task{
		{ID: "t1", SourceID: "h1", UnitID: "u-h1-p", GroupID: "A", Role: "comment", Words: 20},
		{ID: "t2", SourceID: "h2", UnitID: "u-h2-p", GroupID: "B", Role: "comment", Words: 20},
	}}
	records := generation.Generation{Version: generation.GenerationVersion, Run: "run-1", Family: "f", Model: "m",
		Records: []generation.Record{
			{ResponseID: "r1g", TaskID: "t1", Operation: "generate", Prompt: "neutral", Status: "complete"},
			{ResponseID: "r1p", TaskID: "t1", Operation: "polish", Prompt: "neutral", Status: "complete"},
			{ResponseID: "r2g", TaskID: "t2", Operation: "generate", Prompt: "neutral", Status: "complete"},
			{ResponseID: "r2p", TaskID: "t2", Operation: "polish", Prompt: "neutral", Status: "refused"},
		}}
	art := corpus.FindingsArtifact{Version: corpus.FindingsVersion, HumanCorpus: "not_qualified", SHA256: "abc", Policy: policy(),
		Documents: []corpus.DocumentFindings{
			doc("h1", "A", "historical", "comment", 100, 1, 0),
			doc("h2", "B", "historical", "comment", 100, 0, 0),
			doc("h3", "C", "historical", "comment", 100, 1, 0),
			{SourceID: generation.ControlledID("run-0", "", "generated/run-0/r1g.md"), Path: "generated/run-0/r1g.md", GroupID: "A",
				Partition: "training", Cohort: "controlled", Role: "comment",
				Status: "measured", ProseWords: 20, Findings: 2, ByRule: map[string]int{"a.rule": 1, "b.rule": 1}},
			{SourceID: generation.ControlledID("run-1", "", "generated/run-1/r1g.md"), Path: "generated/run-1/r1g.md", GroupID: "A",
				Partition: "training", Cohort: "controlled", Role: "comment",
				Status: "measured", ProseWords: 20, ByRule: map[string]int{"a.rule": 0, "b.rule": 0}},
			{SourceID: generation.ControlledID("run-1", "", "generated/run-1/r1p.md"), Path: "generated/run-1/r1p.md", GroupID: "A",
				Partition: "training", Cohort: "controlled", Role: "comment",
				Status: "measured", ProseWords: 20, Findings: 1, ByRule: map[string]int{"a.rule": 1, "b.rule": 0}},
			{SourceID: generation.ControlledID("run-1", "", "generated/run-1/r2g.md"), Path: "generated/run-1/r2g.md", GroupID: "B",
				Partition: "training", Cohort: "controlled", Role: "comment",
				Status: "measured", ProseWords: 20, Findings: 1, ByRule: map[string]int{"a.rule": 1, "b.rule": 0}},
		},
		Units: []corpus.UnitFindings{
			{UnitID: "u-h1-p", SourceID: "h1", Cohort: "historical", Kind: "paragraph", Findings: []corpus.FindingRecord{{RuleID: "a.rule"}}},
			{UnitID: "u-h2-p", SourceID: "h2", Cohort: "historical", Kind: "paragraph", Findings: []corpus.FindingRecord{}},
		}}
	for i := range art.Documents {
		if art.Documents[i].Status == "" {
			art.Documents[i].Status = "measured"
		}
	}
	return records, tasks, art
}

func TestPairedTablesReportChangesPerArm(t *testing.T) {
	c := qt.New(t)
	records, tasks, art := pairedFixture()
	tables, err := patterns.AnalyzePaired(t.Context(), records, tasks, []corpus.FindingsArtifact{art}, classes())
	c.Assert(err, qt.IsNil)
	c.Assert(tables.Version, qt.Equals, patterns.PairedVersion)
	c.Assert(tables.Role, qt.Equals, "comment")
	c.Assert(tables.Arms, qt.DeepEquals, []patterns.ArmCoverage{
		{Arm: patterns.Arm{Operation: "generate", Prompt: "neutral"}, Responses: 2, Complete: 2, Measured: 2, Components: 2},
		{Arm: patterns.Arm{Operation: "polish", Prompt: "neutral"}, Responses: 2, Complete: 1, Measured: 1, Components: 1},
	})
	// a.rule admits documentation only in the fixture classes, so b.rule is
	// the one rule with the comment role.
	c.Assert(tables.Rules, qt.HasLen, 1)
	b := tables.Rules[0]
	c.Assert(b.RuleID, qt.Equals, "b.rule")
	c.Assert(b.H0Documents, qt.Equals, 3)
	c.Assert(b.H0WithFindng, qt.Equals, 0)
	generate := b.Arms[0]
	c.Assert(generate.Pairs, qt.Equals, 2)
	c.Assert(generate.ResponseWithFinding, qt.Equals, 0)
	c.Assert(*generate.PairedChange.Value, qt.Equals, 0.0)
	c.Assert(generate.ResponseZeroUpper, qt.IsNotNil)
	// A class that admits comments sees the a.rule changes.
	admitting := classes()
	admitting.Rules[0].Roles = []string{"comment"}
	tables, err = patterns.AnalyzePaired(t.Context(), records, tasks, []corpus.FindingsArtifact{art}, admitting)
	c.Assert(err, qt.IsNil)
	a := tables.Rules[0]
	c.Assert(a.RuleID, qt.Equals, "a.rule")
	c.Assert(a.H0WithFindng, qt.Equals, 2)
	generate, polish := a.Arms[0], a.Arms[1]
	c.Assert(generate.OriginalWithFinding, qt.Equals, 1)
	c.Assert(generate.ResponseWithFinding, qt.Equals, 1)
	c.Assert(*generate.PairedChange.Value, qt.Equals, 0.0)
	c.Assert(*generate.DifferenceFromH0.Value < 0, qt.IsTrue)
	c.Assert(polish.Pairs, qt.Equals, 1)
	c.Assert(polish.OriginalWithFinding, qt.Equals, 1)
	c.Assert(polish.ResponseWithFinding, qt.Equals, 1)
	c.Assert(polish.PairedChange.Status, qt.Equals, "fewer_than_two_components")
	// A record naming an unknown task fails; a missing response document is coverage.
	broken := records
	broken.Records = append([]generation.Record(nil), records.Records...)
	broken.Records[0].TaskID = "t9"
	_, err = patterns.AnalyzePaired(t.Context(), broken, tasks, []corpus.FindingsArtifact{art}, classes())
	c.Assert(err, qt.IsNotNil)
	partial := art
	partial.Documents = art.Documents[:5]
	tables, err = patterns.AnalyzePaired(t.Context(), records, tasks, []corpus.FindingsArtifact{partial}, classes())
	c.Assert(err, qt.IsNil)
	c.Assert(tables.Arms[0].MissingResponses, qt.Equals, 1)
}
