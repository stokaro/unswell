package corpus_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// A banned phrase gives a deterministic finding wherever "client" occurs.
const findingsPolicy = "version: 1\nextends: [builtin:custom]\nrules:\n" +
	"  policy.banned-phrases: {enabled: true, parameters: {phrases: [client]}}\n"

func findingsFixture(c *qt.C) (corpus.Artifact, map[string][]byte) {
	c.Helper()
	m, files := sample()
	files["notes.md"] = []byte("# Notes\n\nCertainly! The client opens connections.\n\nThe cache stores entries.\n")
	m.Sources = append(m.Sources, corpus.Source{ID: "d2", Path: "notes.md", SHA256: hash(files["notes.md"]),
		Bytes: len(files["notes.md"]), Format: document.Markdown, ProseLanguage: "en", Repository: "notes-project",
		Document: "notes-project/notes.md", Reference: "fixture:notes.md", Topic: "cache", Purpose: "Explain the client",
		Role: "documentation", Origin: m.Sources[0].Origin, Rights: m.Sources[0].Rights, Notices: m.Sources[0].Notices,
		Snapshot: &corpus.Snapshot{Date: "2019-06-30", Confidence: "corroborated", Evidence: "Fixture", Cohort: "historical"}})
	p, err := corpus.MakePlan(t(c), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t(c), p, files)
	c.Assert(err, qt.IsNil)
	return a, files
}

func t(c *qt.C) context.Context { return c.TB.(*testing.T).Context() }

func TestFindingsBindToContainingCandidates(t *testing.T) {
	c := qt.New(t)
	a, files := findingsFixture(c)
	result, err := corpus.MeasureFindings(t.Context(), a, files, []byte(findingsPolicy))
	c.Assert(err, qt.IsNil)
	c.Assert(result.Version, qt.Equals, corpus.FindingsVersion)
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.Verification.Status, qt.Equals, "source_and_candidates_reproduced")
	c.Assert(result.Policy.ConfigHash, qt.Not(qt.Equals), "")
	c.Assert(len(result.Policy.Rules) > 0, qt.IsTrue)
	c.Assert(result.SHA256, qt.HasLen, 64)
	c.Assert(result.Documents, qt.HasLen, 3)
	c.Assert(result.Units, qt.HasLen, len(a.Units))
	checkDocumentFindings(c, result.Documents)
	checkUnitFindings(c, a, result.Units)
}

func checkDocumentFindings(c *qt.C, documents []corpus.DocumentFindings) {
	c.Helper()
	docs := map[string]corpus.DocumentFindings{}
	for _, doc := range documents {
		docs[doc.SourceID] = doc
		total := 0
		for _, count := range doc.ByRule {
			total += count
		}
		c.Assert(total, qt.Equals, doc.Findings)
		c.Assert(doc.ProseWords > 0, qt.IsTrue)
		c.Assert(doc.GroupID, qt.Not(qt.Equals), "")
	}
	c.Assert(docs["d2"].Cohort, qt.Equals, "historical")
	c.Assert(docs["d2"].Findings > 0, qt.IsTrue)
	c.Assert(docs["d2"].ByRule, qt.DeepEquals, map[string]int{"policy.banned-phrases": docs["d2"].Findings})
	c.Assert(docs["d0"].Cohort, qt.Equals, "")
	c.Assert(docs["d0"].Findings, qt.Equals, 0)
}

// The violation lands in every nested candidate that contains it and in no
// candidate of a different source; the paragraph and its sentence agree.
func checkUnitFindings(c *qt.C, a corpus.Artifact, units []corpus.UnitFindings) {
	c.Helper()
	bound := 0
	for i, unit := range units {
		candidate := a.Units[i]
		c.Assert(unit.UnitID, qt.Equals, candidate.Unit.ID)
		c.Assert(unit.SourceID, qt.Equals, candidate.SourceID)
		c.Assert(unit.Words, qt.Equals, candidate.Words)
		text := candidate.Unit.Text
		if unit.SourceID == "d2" && strings.Contains(text, "client") {
			c.Assert(len(unit.Findings) > 0, qt.IsTrue, qt.Commentf("%s: %q", unit.Kind, text))
			c.Assert(unit.Cohort, qt.Equals, "historical")
		}
		if unit.SourceID == "d2" && strings.Contains(text, "stores entries") {
			c.Assert(unit.Findings, qt.HasLen, 0)
		}
		checkContained(c, candidate, unit.Findings)
		bound += len(unit.Findings)
	}
	c.Assert(bound > 0, qt.IsTrue)
}

// A finding is never attached to a candidate that does not contain it.
func checkContained(c *qt.C, candidate corpus.Candidate, findings []corpus.FindingRecord) {
	c.Helper()
	bounds := document.Bounds(candidate.Unit.Source.Segments)
	for _, finding := range findings {
		c.Assert(finding.RuleID, qt.Not(qt.Equals), "")
		c.Assert(finding.RuleVersion, qt.Not(qt.Equals), "")
		c.Assert(bounds.Start <= finding.Span.Start && finding.Span.End <= bounds.End, qt.IsTrue)
	}
}

func TestFindingsRequireFrozenExtractionAndReproduction(t *testing.T) {
	c := qt.New(t)
	a, files := findingsFixture(c)
	for _, policy := range []string{"", "version: 1\nunknown: true\n",
		findingsPolicy + "extraction: {contexts: [heading]}\n", findingsPolicy + "analysis: {include_quotes: true}\n"} {
		_, err := corpus.MeasureFindings(t.Context(), a, files, []byte(policy))
		c.Assert(err, qt.IsNotNil, qt.Commentf("%q", policy))
	}
	// A policy without enabled rules is accepted and finds nothing; the counts
	// stay explicit rather than absent.
	result, err := corpus.MeasureFindings(t.Context(), a, files, []byte("version: 1\nextends: [builtin:custom]\n"))
	c.Assert(err, qt.IsNil)
	for _, doc := range result.Documents {
		c.Assert(doc.Findings, qt.Equals, 0)
		c.Assert(doc.ByRule, qt.HasLen, 0)
	}
	// Tampered sources fail verification before any policy runs.
	files["notes.md"] = append(files["notes.md"], '!')
	_, err = corpus.MeasureFindings(t.Context(), a, files, []byte(findingsPolicy))
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.MeasureFindings(ctx, a, files, []byte(findingsPolicy))
	c.Assert(err, qt.IsNotNil)
	_ = annotation.Unit{}
}

// A document the engine cannot finish under the policy stays in the artifact
// as a failed coverage gap; the other documents are measured as usual.
func TestFindingsRecordOperationalFailuresPerDocument(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	var bomb strings.Builder
	bomb.WriteString("# Bomb\n\n")
	for i := range 2000 {
		bomb.WriteString("The cache retries the lookup number ")
		bomb.WriteString(strconv.Itoa(i))
		bomb.WriteString(" after the configured delay expires.\n\n")
	}
	files["bomb.md"] = []byte(bomb.String())
	m.Sources = append(m.Sources, corpus.Source{ID: "d9", Path: "bomb.md", SHA256: hash(files["bomb.md"]),
		Bytes: len(files["bomb.md"]), Format: document.Markdown, ProseLanguage: "en", Repository: "bomb-project",
		Document: "bomb-project/bomb.md", Reference: "fixture:bomb.md", Topic: "cache", Purpose: "Exhaust a rule budget",
		Role: "documentation", Origin: m.Sources[0].Origin, Rights: m.Sources[0].Rights, Notices: m.Sources[0].Notices})
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.near-sentence: {enabled: true}\n"
	result, err := corpus.MeasureFindings(t.Context(), a, files, []byte(policy))
	c.Assert(err, qt.IsNil)
	c.Assert(result.Failed, qt.Equals, 1)
	statuses := map[string]string{}
	for _, doc := range result.Documents {
		statuses[doc.SourceID] = doc.Status
		if doc.SourceID == "d9" {
			c.Assert(doc.Error, qt.Contains, "budget")
			c.Assert(doc.Findings, qt.Equals, 0)
		}
	}
	c.Assert(statuses, qt.DeepEquals, map[string]string{"d0": "measured", "d1": "measured", "d9": "failed"})
	for _, unit := range result.Units {
		c.Assert(unit.Unmeasured, qt.Equals, unit.SourceID == "d9")
	}
	c.Assert(result.Policy.ConfigHash, qt.Not(qt.Equals), "")
}

// An artifact whose only source fails still names the policy it ran under.
func TestFindingsKeepThePolicyIdentityWhenEverySourceFails(t *testing.T) {
	c := qt.New(t)
	files := map[string][]byte{"LICENSE": []byte("Test-owned source and notice fixture.\n")}
	var bomb strings.Builder
	bomb.WriteString("# Bomb\n\n")
	for i := range 2000 {
		bomb.WriteString("The cache retries the lookup number ")
		bomb.WriteString(strconv.Itoa(i))
		bomb.WriteString(" after the configured delay expires.\n\n")
	}
	files["bomb.md"] = []byte(bomb.String())
	notice := corpus.Notice{Path: "LICENSE", SHA256: hash(files["LICENSE"]), Bytes: len(files["LICENSE"])}
	m := corpus.Manifest{Version: corpus.Version, ID: "bomb-only", Seed: "frozen-seed",
		Weights:   corpus.Weights{Training: 6000, Development: 1500, Calibration: 1500, FinalTest: 1000},
		UnitKinds: []string{"paragraph", "sentence", "fragment"},
		Sources: []corpus.Source{{ID: "d9", Path: "bomb.md", SHA256: hash(files["bomb.md"]), Bytes: len(files["bomb.md"]),
			Format: document.Markdown, ProseLanguage: "en", Repository: "bomb-project", Document: "bomb-project/bomb.md",
			Reference: "fixture:bomb.md", Topic: "cache", Purpose: "Exhaust a rule budget", Role: "documentation",
			Origin:  annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Teaching fixture; origin is not a quality label."},
			Rights:  annotation.Rights{License: "test-fixture", Evidence: "Test-owned bytes", AllowedUses: []string{"annotation"}},
			Notices: []corpus.Notice{notice}}}}
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.near-sentence: {enabled: true}\n"
	result, err := corpus.MeasureFindings(t.Context(), a, files, []byte(policy))
	c.Assert(err, qt.IsNil)
	c.Assert(result.Failed, qt.Equals, 1)
	c.Assert(result.Documents, qt.HasLen, 1)
	c.Assert(result.Documents[0].Status, qt.Equals, "failed")
	c.Assert(result.Policy.ConfigHash, qt.Not(qt.Equals), "")
	c.Assert(len(result.Policy.Rules) > 0, qt.IsTrue)
}
