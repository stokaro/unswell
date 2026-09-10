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
			{SourceID: "f1", Path: "f1.md", GroupID: "E", Partition: "training", Cohort: "historical",
				Role: "documentation", Status: "failed", Error: "budget exceeded", ByRule: map[string]int{}},
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
	c.Assert(tables.Unit, qt.Equals, "document")
	c.Assert(tables.FirstAppearance, qt.IsNil)
	c.Assert(tables.Unassigned, qt.Equals, 1)
	c.Assert(tables.Inputs, qt.HasLen, 1)
	c.Assert(tables.Cohorts, qt.HasLen, 2)
	summary := map[string]patterns.CohortSummary{}
	for _, item := range tables.Cohorts {
		summary[item.Cohort] = item
	}
	// The string document counts in the cohort summary and in no prose rule;
	// the failed document is a coverage gap outside every count.
	c.Assert(summary["historical"].Documents, qt.Equals, 5)
	c.Assert(summary["historical"].FailedDocuments, qt.Equals, 1)
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
	c.Assert(historical.Counted, qt.Equals, 4)
	c.Assert(historical.CountedWithFinding, qt.Equals, 2)
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
}

// b.rule never fires: the zero bound over three components is 1-0.025^(1/3),
// the ratio stays undefined, and the same inputs give the same tables.
func TestTablesReportZeroCountsAndStayDeterministic(t *testing.T) {
	c := qt.New(t)
	tables, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{fixture()}, classes(), patterns.Options{})
	c.Assert(err, qt.IsNil)
	b := tables.Rules[1]
	c.Assert(b.RuleID, qt.Equals, "b.rule")
	rows := map[string]patterns.RuleCohort{}
	for _, row := range b.Cohorts {
		rows[row.Cohort] = row
	}
	c.Assert(rows["historical"].Findings, qt.Equals, 0)
	c.Assert(*rows["historical"].Prevalence.Value, qt.Equals, 0.0)
	c.Assert(math.Abs(*rows["historical"].ZeroUpperBound-(1-math.Pow(0.025, 1.0/3))) < 1e-12, qt.IsTrue)
	c.Assert(b.Contrasts[0].Ratio, qt.IsNil)
	c.Assert(b.Contrasts[0].RatioStatus, qt.Equals, "undefined")
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

// unitFixture gives every unit a word count and adds paragraphs so that the
// historical cohort spans two components: h1 (one a.rule finding), h2 (none),
// and h4 (two). The controlled document has one paragraph with a finding and
// one without. A paragraph of the failed document stays unmeasured.
func unitFixture() corpus.FindingsArtifact {
	art := fixture()
	for i := range art.Units {
		art.Units[i].Words = 10 * (i + 1)
	}
	art.Units = append(art.Units,
		corpus.UnitFindings{UnitID: "u-h4-p", SourceID: "h4", Cohort: "historical", Kind: "paragraph", Words: 50,
			Findings: []corpus.FindingRecord{{RuleID: "a.rule"}, {RuleID: "a.rule"}}},
		corpus.UnitFindings{UnitID: "u-c1-q", SourceID: "c1", Cohort: "controlled", Kind: "paragraph", Words: 60,
			Findings: []corpus.FindingRecord{}},
		corpus.UnitFindings{UnitID: "u-f1-p", SourceID: "f1", Cohort: "historical", Kind: "paragraph", Words: 5,
			Unmeasured: true, Findings: []corpus.FindingRecord{}})
	return art
}

func filter() *corpus.AppearanceFilter {
	return &corpus.AppearanceFilter{Version: corpus.AppearanceVersion, Kind: "paragraph", Earlier: "historical",
		Later: "controlled", EarlierUnits: 3, LaterUnits: 2, RepeatedUnits: 1, NewUnits: 1, New: []string{"c1#u-c1-q"},
		Repositories: []corpus.RepositoryAppearance{}, LaterOnly: []string{}}
}

func rowsOf(table patterns.RuleTable) map[string]patterns.RuleCohort {
	rows := map[string]patterns.RuleCohort{}
	for _, row := range table.Cohorts {
		rows[row.Cohort] = row
	}
	return rows
}

func TestUnitTablesCountRetainedParagraphs(t *testing.T) {
	c := qt.New(t)
	tables, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph"})
	c.Assert(err, qt.IsNil)
	c.Assert(tables.Unit, qt.Equals, "paragraph")
	c.Assert(tables.FirstAppearance, qt.IsNil)
	summary := map[string]patterns.CohortSummary{}
	for _, item := range tables.Cohorts {
		summary[item.Cohort] = item
	}
	// h3 and the string document hold no paragraph and count nowhere; the
	// failed document's paragraph stays a coverage gap.
	c.Assert(summary["historical"], qt.DeepEquals, patterns.CohortSummary{Cohort: "historical", Documents: 3,
		FailedDocuments: 1, Units: 3, Words: 90, Components: 2, Findings: 3, DocumentsWithFinding: 2,
		PerThousandWords: summary["historical"].PerThousandWords})
	c.Assert(*summary["historical"].PerThousandWords, qt.Equals, 1000*3.0/90)
	c.Assert(summary["controlled"].Units, qt.Equals, 2)
	c.Assert(summary["controlled"].Words, qt.Equals, 100)
	rows := rowsOf(tables.Rules[0])
	historical := rows["historical"]
	c.Assert(historical.Documents, qt.Equals, 3)
	c.Assert(historical.DocumentsWithFinding, qt.Equals, 2)
	c.Assert(historical.Components, qt.Equals, 2)
	c.Assert(historical.Counted, qt.Equals, 3)
	c.Assert(historical.CountedWithFinding, qt.Equals, 2)
	c.Assert(historical.Words, qt.Equals, 90)
	c.Assert(historical.Findings, qt.Equals, 3)
	c.Assert(*historical.Prevalence.Value, qt.Equals, 2.0/3)
	c.Assert(historical.Prevalence.Status, qt.Equals, "cluster_percentile")
	c.Assert(historical.Units, qt.DeepEquals, []patterns.KindPrevalence{{Kind: "paragraph", Units: 3, WithFinding: 2, Prevalence: 2.0 / 3}})
	controlled := rows["controlled"]
	c.Assert(controlled.Counted, qt.Equals, 2)
	c.Assert(controlled.CountedWithFinding, qt.Equals, 1)
	c.Assert(*controlled.Prevalence.Value, qt.Equals, 0.5)
	contrast := tables.Rules[0].Contrasts[0]
	c.Assert(math.Abs(*contrast.Difference.Value-(0.5-2.0/3)) < 1e-12, qt.IsTrue)
	c.Assert(*contrast.Ratio, qt.Equals, 0.75)
	// b.rule admits comments too but never fires: the zero bound spans the
	// two historical components that hold paragraphs.
	b := rowsOf(tables.Rules[1])["historical"]
	c.Assert(b.Counted, qt.Equals, 3)
	c.Assert(math.Abs(*b.ZeroUpperBound-(1-math.Pow(0.025, 1.0/2))) < 1e-12, qt.IsTrue)
	// A kind nobody extracted leaves every row empty rather than failing.
	empty, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "fragment"})
	c.Assert(err, qt.IsNil)
	c.Assert(rowsOf(empty.Rules[0])["historical"].Counted, qt.Equals, 0)
	c.Assert(rowsOf(empty.Rules[0])["historical"].Prevalence.Status, qt.Equals, "nothing_counted")
}

func TestUnitTablesApplyTheFirstAppearanceFilter(t *testing.T) {
	c := qt.New(t)
	tables, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph", FirstAppearance: filter()})
	c.Assert(err, qt.IsNil)
	c.Assert(tables.FirstAppearance, qt.DeepEquals, &patterns.FirstAppearance{Earlier: "historical", Later: "controlled",
		Kind: "paragraph", Kept: 1, Excluded: 1, LaterOnly: []string{}})
	summary := map[string]patterns.CohortSummary{}
	for _, item := range tables.Cohorts {
		summary[item.Cohort] = item
	}
	// The earlier cohort is untouched; the later one keeps the listed paragraph only.
	c.Assert(summary["historical"].Units, qt.Equals, 3)
	c.Assert(summary["historical"].Excluded, qt.Equals, 0)
	c.Assert(summary["controlled"].Units, qt.Equals, 1)
	c.Assert(summary["controlled"].Excluded, qt.Equals, 1)
	c.Assert(summary["controlled"].Words, qt.Equals, 60)
	controlled := rowsOf(tables.Rules[0])["controlled"]
	c.Assert(controlled.Counted, qt.Equals, 1)
	c.Assert(controlled.CountedWithFinding, qt.Equals, 0)
	c.Assert(*controlled.Prevalence.Value, qt.Equals, 0.0)
	c.Assert(rowsOf(tables.Rules[0])["historical"].Counted, qt.Equals, 3)
	c.Assert(tables.Rules[0].Contrasts[0].RatioStatus, qt.Equals, "undefined")
	// The filter must match the analysis kind, name present cohorts, keep the
	// baseline as the earlier side, and list units the inputs hold.
	wrongKind := filter()
	wrongKind.Kind = "sentence"
	for _, options := range []patterns.Options{
		{FirstAppearance: filter()},
		{UnitKind: "paragraph", FirstAppearance: wrongKind},
		{UnitKind: "paragraph", Baseline: "controlled", FirstAppearance: filter()},
	} {
		_, err := patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(), options)
		c.Assert(err, qt.IsNotNil, qt.Commentf("%+v", options))
	}
	absent := filter()
	absent.Later = "natural"
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph", FirstAppearance: absent})
	c.Assert(err, qt.IsNotNil)
	unknown := filter()
	unknown.New = []string{"c1#missing"}
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph", FirstAppearance: unknown})
	c.Assert(err, qt.IsNotNil)
	duplicated := filter()
	duplicated.New = []string{"c1#u-c1-q", "c1#u-c1-q"}
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph", FirstAppearance: duplicated})
	c.Assert(err, qt.IsNotNil)
	// A listed unit found in two artifacts is not one observation either.
	_, err = patterns.Analyze(t.Context(), []corpus.FindingsArtifact{unitFixture(), unitFixture()}, classes(),
		patterns.Options{UnitKind: "paragraph", FirstAppearance: filter()})
	c.Assert(err, qt.IsNotNil)
}
