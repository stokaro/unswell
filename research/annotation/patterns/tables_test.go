package patterns_test

import (
	"context"
	"math"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/patterns"
	"github.com/stokaro/unswell/rule"
)

func classes() corpus.RuleClasses {
	return corpus.RuleClasses{Format: corpus.RuleClassesVersion, Revision: 1, DecidedOn: "2026-09-10",
		Protocol: "unswell-llm-patterns-v1", Scope: "test",
		Classes: map[string]string{"general_style": "g", "llm_associated_candidate": "c"},
		Rules: []corpus.RuleClass{
			{RuleID: "a.rule", Class: "llm_associated_candidate", Roles: []string{"documentation"}, Reason: "r"},
			{RuleID: "b.rule", Class: "general_style", Roles: []string{"documentation", "comment"}, Reason: "r"},
		}}
}

func policy() corpus.PolicyIdentity {
	return corpus.PolicyIdentity{ConfigHash: "config", RulesetHash: "rules", ScoringProfile: "technical-v1",
		Rules: []rule.Descriptor{{ID: "a.rule"}, {ID: "b.rule"}}}
}

func doc(id, group, cohort, role string, words, a, b int) corpus.DocumentFindings {
	return corpus.DocumentFindings{SourceID: id, Path: id + ".md", GroupID: group, Partition: "training", Cohort: cohort,
		Role: role, ProseWords: words, Findings: a + b, ByRule: map[string]int{"a.rule": a, "b.rule": b}}
}

// The reference example: three components. Historical documents in A (two,
// one with a.rule), B (one, none), and C (one, with a.rule) give prevalence
// 2/4. The controlled document in A carries a.rule, so its prevalence is 1.
// b.rule never fires, and one document has no cohort.
func fixture() corpus.FindingsArtifact {
	art := corpus.FindingsArtifact{Version: corpus.FindingsVersion, HumanCorpus: "not_qualified", SHA256: "abc", Policy: policy(),
		Documents: []corpus.DocumentFindings{
			doc("h1", "A", "historical", "documentation", 100, 1, 0),
			doc("h2", "A", "historical", "documentation", 300, 0, 0),
			doc("h3", "B", "historical", "documentation", 200, 0, 0),
			doc("h4", "C", "historical", "documentation", 400, 2, 0),
			doc("c1", "A", "controlled", "documentation", 50, 1, 0),
			doc("s1", "B", "historical", "string", 10, 5, 5),
			doc("u1", "D", "", "documentation", 10, 1, 1),
		},
		Units: []corpus.UnitFindings{
			{UnitID: "u-h1-p", SourceID: "h1", Cohort: "historical", Kind: "paragraph", Findings: []corpus.FindingRecord{{RuleID: "a.rule"}}},
			{UnitID: "u-h1-s", SourceID: "h1", Cohort: "historical", Kind: "sentence", Findings: []corpus.FindingRecord{{RuleID: "a.rule"}}},
			{UnitID: "u-h2-p", SourceID: "h2", Cohort: "historical", Kind: "paragraph", Findings: []corpus.FindingRecord{}},
			{UnitID: "u-c1-p", SourceID: "c1", Cohort: "controlled", Kind: "paragraph", Findings: []corpus.FindingRecord{{RuleID: "a.rule"}}},
		}}
	return art
}

func TestTablesReproduceTheReferenceExample(t *testing.T) {
	c := qt.New(t)
	tables, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture()}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(tables.Version, qt.Equals, patterns.Version)
	c.Assert(tables.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(tables.Baseline, qt.Equals, "historical")
	c.Assert(tables.Unassigned, qt.Equals, 1)
	c.Assert(tables.Inputs, qt.HasLen, 1)
	c.Assert(tables.Cohorts, qt.HasLen, 2)
	summary := map[string]patterns.CohortSummary{}
	for _, item := range tables.Cohorts {
		summary[item.Cohort] = item
	}
	// The string document counts in the cohort summary and in no prose rule.
	c.Assert(summary["historical"].Documents, qt.Equals, 5)
	c.Assert(summary["historical"].Components, qt.Equals, 3)
	c.Assert(summary["historical"].Findings, qt.Equals, 13)
	c.Assert(summary["historical"].Units, qt.Equals, 3)
	c.Assert(*summary["historical"].PerThousandWords, qt.Equals, 1000*13.0/1010)
	c.Assert(tables.Rules, qt.HasLen, 2)
	a := tables.Rules[0]
	c.Assert(a.RuleID, qt.Equals, "a.rule")
	c.Assert(a.Class, qt.Equals, "llm_associated_candidate")
	rows := map[string]patterns.RuleCohort{}
	for _, row := range a.Cohorts {
		rows[row.Cohort] = row
	}
	historical := rows["historical"]
	c.Assert(historical.Documents, qt.Equals, 4)
	c.Assert(historical.Components, qt.Equals, 3)
	c.Assert(historical.DocumentsWithFinding, qt.Equals, 2)
	c.Assert(historical.Findings, qt.Equals, 3)
	c.Assert(*historical.Prevalence.Value, qt.Equals, 0.5)
	c.Assert(historical.Prevalence.Status, qt.Equals, "cluster_percentile")
	c.Assert(*historical.Prevalence.Lower <= 0.5 && 0.5 <= *historical.Prevalence.Upper, qt.IsTrue)
	c.Assert(*historical.Prevalence.Lower >= 0 && *historical.Prevalence.Upper <= 1, qt.IsTrue)
	c.Assert(historical.Prevalence.Replicates > 9000, qt.IsTrue)
	c.Assert(*historical.PerThousandWords, qt.Equals, 3.0)
	c.Assert(historical.ZeroUpperBound, qt.IsNil)
	c.Assert(historical.Units, qt.DeepEquals, []patterns.KindPrevalence{
		{Kind: "paragraph", Units: 2, WithFinding: 1, Prevalence: 0.5}, {Kind: "sentence", Units: 1, WithFinding: 1, Prevalence: 1}})
	controlled := rows["controlled"]
	c.Assert(controlled.Documents, qt.Equals, 1)
	c.Assert(controlled.Components, qt.Equals, 1)
	c.Assert(*controlled.Prevalence.Value, qt.Equals, 1.0)
	c.Assert(controlled.Prevalence.Status, qt.Equals, "fewer_than_two_components")
	c.Assert(a.Contrasts, qt.HasLen, 1)
	c.Assert(a.Contrasts[0].Cohort, qt.Equals, "controlled")
	c.Assert(*a.Contrasts[0].Difference.Value, qt.Equals, 0.5)
	c.Assert(a.Contrasts[0].Difference.Status, qt.Equals, "fewer_than_two_components")
	c.Assert(*a.Contrasts[0].Ratio, qt.Equals, 2.0)
	c.Assert(a.Contrasts[0].RatioStatus, qt.Equals, "defined")
	// b.rule never fires: the zero bound over three components is 1-0.025^(1/3).
	b := tables.Rules[1]
	rows = map[string]patterns.RuleCohort{}
	for _, row := range b.Cohorts {
		rows[row.Cohort] = row
	}
	c.Assert(rows["historical"].Findings, qt.Equals, 0)
	c.Assert(*rows["historical"].Prevalence.Value, qt.Equals, 0.0)
	c.Assert(math.Abs(*rows["historical"].ZeroUpperBound-(1-math.Pow(0.025, 1.0/3))) < 1e-12, qt.IsTrue)
	c.Assert(b.Contrasts[0].Ratio, qt.IsNil)
	c.Assert(b.Contrasts[0].RatioStatus, qt.Equals, "undefined")
	// The same inputs give the same tables.
	again, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture()}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, tables)
}

func TestTablesRejectInconsistentInputs(t *testing.T) {
	c := qt.New(t)
	_, err := patterns.Analyze(t.Context(), nil, classes(), patterns.Options{})
	c.Assert(err, qt.IsNotNil)
	other := fixture()
	other.Policy.RulesetHash = "changed"
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture(), other}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNotNil)
	unknown := fixture()
	unknown.Documents[0].ByRule["z.rule"] = 1
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unknown}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNotNil)
	partial := classes()
	partial.Rules = partial.Rules[:1]
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture()}, partial, patterns.Options{})
	c.Assert(err, qt.IsNotNil)
	qualified := fixture()
	qualified.HumanCorpus = "qualified"
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{qualified}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNotNil)
	// Two artifacts under one policy join; a missing baseline leaves no contrast.
	tables, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture(), fixture()}, classes(),
		patterns.Options{Baseline: "natural"})
	c.Assert(err, qt.IsNil)
	c.Assert(tables.Inputs, qt.HasLen, 2)
	c.Assert(tables.Rules[0].Contrasts, qt.HasLen, 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = patterns.Analyze(ctx, []corpus.FindingsArtifact{fixture()}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNotNil)
}
