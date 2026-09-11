package corpus_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestManifestRejectsInvalidAcquisition(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*corpus.Manifest)
	}{
		{"version", func(m *corpus.Manifest) { m.Version = "future" }},
		{"weights", func(m *corpus.Manifest) { m.Weights.Training++ }},
		{"empty partition", func(m *corpus.Manifest) { m.Weights.Calibration = 0 }},
		{"unit kind", func(m *corpus.Manifest) { m.UnitKinds = []string{"document"} }},
		{"duplicate source", func(m *corpus.Manifest) { m.Sources = append(m.Sources, m.Sources[0]) }},
		{"duplicate author", func(m *corpus.Manifest) { m.Sources[0].Authors = []string{"a", "a"} }},
		{"language evidence", func(m *corpus.Manifest) { m.Sources[0].AuthorLanguage = "en" }},
		{"invalid metadata", func(m *corpus.Manifest) { m.Sources[0].Origin.GenerationRecord = "\x00" }},
		{"unsupported prose", func(m *corpus.Manifest) { m.Sources[0].ProseLanguage = "unknown" }},
		{"permission", func(m *corpus.Manifest) { m.Sources[0].Rights.AllowedUses = []string{"training"} }},
		{"notice", func(m *corpus.Manifest) { m.Sources[0].Notices = nil }},
		{"conflicting notice", func(m *corpus.Manifest) { m.Sources[1].Notices[0].Bytes++ }},
		{"origin evidence", func(m *corpus.Manifest) { m.Sources[0].Origin.Evidence = "" }},
		{"generation evidence", func(m *corpus.Manifest) { m.Sources[0].Origin.Label = "generated" }},
		{"unknown context", func(m *corpus.Manifest) { m.Policy.Contexts = []string{"invented"} }},
		{"invalid role region", func(m *corpus.Manifest) {
			m.Sources[0].Roles = []corpus.RoleRegion{{Span: document.Span{Start: 10, End: 1}, Role: "comment"}}
		}},
		{"snapshot cohort", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("2019-06-30", "corroborated", "pre-llm") }},
		{"snapshot confidence", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("2019-06-30", "guessed", "historical") }},
		{"snapshot date form", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("30/06/2019", "corroborated", "historical") }},
		{"snapshot calendar", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("2019-02-30", "corroborated", "historical") }},
		{"undated corroboration", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("", "vcs_only", "contemporary") }},
		{"undated historical", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("", "unknown", "historical") }},
		{"undated period", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("", "unknown", "historical-2016") }},
		{"unknown period", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("2014-06-30", "corroborated", "historical-2014") }},
		{"snapshot evidence", func(m *corpus.Manifest) {
			m.Sources[0].Snapshot = snapshot("2019-06-30", "corroborated", "historical")
			m.Sources[0].Snapshot.Evidence = " "
		}},
		{"controlled human origin", func(m *corpus.Manifest) { m.Sources[0].Snapshot = snapshot("2026-09-10", "corroborated", "controlled") }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			m, _ := sample()
			row.edit(&m)
			_, err := corpus.LoadManifest(t.Context(), encoded(c, m))
			c.Assert(err, qt.IsNotNil)
			_, err = corpus.MakePlan(t.Context(), m)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestContextOverridesAndExceptions(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	m.Policy = extract.Policy{Languages: map[document.Format]extract.LanguagePolicy{
		document.Go: {Contexts: []string{}},
	}, Exceptions: []extract.Exception{{ID: "reviewed-docs", Paths: []string{"readme.md"},
		Kinds: []string{"comment"}, Reason: "Reserved source fixture."}}}
	// A comment-only exception is valid without optional formats or symbols.
	loaded, err := corpus.LoadManifest(t.Context(), encoded(c, m))
	c.Assert(err, qt.IsNil)
	p, err := corpus.MakePlan(t.Context(), loaded)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	c.Assert(a.Sources[1].Units, qt.Equals, 0)
	c.Assert(a.Sources[1].EmptyReason, qt.Equals, "no_eligible_requested_units")
	for _, unit := range a.Units {
		c.Assert(unit.SourceID, qt.Equals, "d0")
	}
	m.Policy.Languages = nil
	m.Policy.Exceptions[0].Paths, m.Policy.Exceptions[0].Kinds = []string{"sample.go"}, []string{"string"}
	p, err = corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	c.Assert(a.Sources[1].Units > 0, qt.IsTrue)
	for _, unit := range a.Units {
		c.Assert(unit.Unit.Text, qt.Not(qt.Contains), "cannot be decoded")
	}
}

func TestPartialRolesAndContextLimits(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	m.Sources[1].Roles = []corpus.RoleRegion{{Span: document.Span{Start: 3, End: 10}, Role: "doc_comment"}}
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	_, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.ErrorMatches, ".*role region cuts through.*")
	m, files = sample()
	files["readme.md"] = []byte(strings.Repeat("word ", 14000))
	m.Sources[0].SHA256, m.Sources[0].Bytes = hash(files["readme.md"]), len(files["readme.md"])
	p, err = corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	_, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.ErrorMatches, ".*context exceeds.*")
}

func TestEmptyAndUnlistedInputs(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	m.Policy.Contexts = []string{}
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	_, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.ErrorMatches, ".*no eligible corpus units.*")
	files["unlisted"] = []byte("Extra source")
	_, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.ErrorMatches, ".*exactly match.*")
	delete(files, "unlisted")
	delete(files, "LICENSE")
	_, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.ErrorMatches, ".*exactly match.*")
}

func TestCandidatesUseExistingAnnotationContract(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	m.Sources[0].Origin = annotation.Origin{Label: "generated", Scope: "document", Evidence: "Test assertion",
		GenerationRecord: "Test assertion must not become a unit label"}
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	units := make([]annotation.Unit, 0, len(a.Units))
	for _, candidate := range a.Units {
		c.Assert(candidate.Unit.Origin.Label, qt.Equals, "unknown")
		units = append(units, candidate.Unit)
	}
	data, err := os.ReadFile("../testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	var round map[string]json.RawMessage
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	round["units"] = encoded(c, units)
	round["judgments"], round["adjudications"] = []byte("[]"), []byte("[]")
	_, err = annotation.Load(t.Context(), encoded(c, round))
	c.Assert(err, qt.IsNil)
	// Changes to the caller's manifest or source buffers cannot rewrite an artifact.
	m.Sources[0].Purpose = "Changed"
	files["readme.md"][0] = '!'
	c.Assert(corpus.ValidatePlan(t.Context(), a.Plan), qt.IsNil)
	c.Assert(a.Plan.Manifest.Sources[0].Purpose, qt.Equals, "Explain cache behavior")
}

func FuzzLoadArtifact(f *testing.F) {
	f.Add([]byte(`{"version":"unswell-corpus-v1","status":"unlabeled_candidates"}`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = corpus.LoadArtifact(t.Context(), data)
	})
}

func TestCollectionLimitsBeforeCorpusDecoding(t *testing.T) {
	c := qt.New(t)
	objects := []byte(`{"sources":[` + strings.Repeat("{},", corpus.MaxSources) + `{}]}`)
	_, err := corpus.LoadManifest(t.Context(), objects)
	c.Assert(err, qt.ErrorMatches, `research JSON array "sources" exceeds 10000 entries`)
	_, err = corpus.LoadPlan(t.Context(), []byte(`{"manifest":`+string(objects)+`}`))
	c.Assert(err, qt.ErrorMatches, `research JSON array "sources" exceeds 10000 entries`)
	units := []byte(`{"Units":[` + strings.Repeat("{},", corpus.MaxUnits) + `{}]}`)
	_, err = corpus.LoadArtifact(t.Context(), units)
	c.Assert(err, qt.ErrorMatches, `research JSON array "Units" exceeds 10000 entries`)
}

func snapshot(date, confidence, cohort string) *corpus.Snapshot {
	return &corpus.Snapshot{Date: date, Confidence: confidence, Evidence: "Fixture dating evidence", Cohort: cohort}
}

func TestSnapshotCohortsReachCandidates(t *testing.T) {
	c := qt.New(t)
	m, files := sample()
	// The dated historical periods are cohorts of their own.
	for _, cohort := range []string{"historical-2012", "historical-2016", "historical-2018"} {
		m.Sources[0].Snapshot = snapshot("2012-06-30", "corroborated", cohort)
		_, err := corpus.LoadManifest(t.Context(), encoded(c, m))
		c.Assert(err, qt.IsNil, qt.Commentf("%s", cohort))
	}
	m.Sources[0].Snapshot = snapshot("2019-06-30", "corroborated", "historical")
	m.Sources[1].Snapshot = snapshot("", "unknown", "contemporary")
	m.Sources[0].Role = "unknown"
	loaded, err := corpus.LoadManifest(t.Context(), encoded(c, m))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded.Sources[0].Snapshot, qt.DeepEquals, m.Sources[0].Snapshot)
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err := corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	cohorts := map[string]string{}
	for _, candidate := range a.Units {
		cohorts[candidate.SourceID] = candidate.Cohort
		if candidate.SourceID == "d0" {
			c.Assert(candidate.Unit.Role, qt.Equals, "unknown")
		}
		// A snapshot dates bytes; it never becomes a unit origin label.
		c.Assert(candidate.Unit.Origin.Label, qt.Equals, "unknown")
	}
	c.Assert(cohorts, qt.DeepEquals, map[string]string{"d0": "historical", "d1": "contemporary"})
	// A generated source may join the controlled cohort; a source without a
	// snapshot keeps an empty cohort in its candidates.
	m.Sources[1].Origin = annotation.Origin{Label: "generated", Scope: "document", Evidence: "Fixture",
		GenerationRecord: "Fixture generation record"}
	m.Sources[1].Snapshot = snapshot("2026-09-10", "corroborated", "controlled")
	m.Sources[0].Snapshot = nil
	p, err = corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	a, err = corpus.Build(t.Context(), p, files)
	c.Assert(err, qt.IsNil)
	cohorts = map[string]string{}
	for _, candidate := range a.Units {
		cohorts[candidate.SourceID] = candidate.Cohort
	}
	c.Assert(cohorts, qt.DeepEquals, map[string]string{"d0": "", "d1": "controlled"})
	c.Assert(a.Plan.Manifest.Sources[0].Snapshot, qt.IsNil)
}
