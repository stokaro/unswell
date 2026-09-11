package corpus_test

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

// cohortArtifact builds the sample corpus under one snapshot cohort, after an
// optional edit of the manifest and files.
func cohortArtifact(t *testing.T, cohort, date string, edit func(*corpus.Manifest, map[string][]byte)) corpus.Artifact {
	t.Helper()
	c := qt.New(t)
	m, files := sample()
	for i := range m.Sources {
		m.Sources[i].Snapshot = snapshot(date, "corroborated", cohort)
	}
	if edit != nil {
		edit(&m, files)
	}
	// Each cohort is its own snapshot, so its source IDs differ from the
	// other cohort's as they would under one dataset plan.
	for i := range m.Sources {
		m.Sources[i].ID = cohort[:1] + m.Sources[i].ID
	}
	plan, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(t.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	return artifact
}

// laterEdit keeps the first readme paragraph, replaces the second, and adds a
// source from a repository the earlier cohort never saw.
func laterEdit(m *corpus.Manifest, files map[string][]byte) {
	files["readme.md"] = []byte("# Cache\n\nThe **cache** may retry &amp; wait.\n\nA later snapshot adds this paragraph.\n")
	m.Sources[0].SHA256, m.Sources[0].Bytes = hash(files["readme.md"]), len(files["readme.md"])
	addOther(m, files, "# Other\n\nThis repository has no earlier snapshot in the dataset.\n")
}

// addOther adds a document from a second repository with the given text.
func addOther(m *corpus.Manifest, files map[string][]byte, content string) {
	files["other.md"] = []byte(content)
	template := m.Sources[0]
	m.Sources = append(m.Sources, corpus.Source{ID: "d2", Path: "other.md", SHA256: hash(files["other.md"]),
		Bytes: len(files["other.md"]), Format: document.Markdown, ProseLanguage: "en", Repository: "other-project",
		Document: "other-project/other.md", Reference: "fixture:other.md", Topic: "cache", Purpose: "Explain cache behavior",
		Role: "documentation", Origin: template.Origin, Rights: template.Rights, Notices: template.Notices,
		Snapshot: template.Snapshot})
}

func paragraphs(artifact corpus.Artifact, repository string) map[string]string {
	keys := map[string]string{}
	for _, candidate := range artifact.Units {
		if candidate.Unit.Kind == "paragraph" && candidate.Unit.Source.RepositoryID == repository {
			keys[candidate.Unit.Text] = corpus.UnitKey(candidate.SourceID, candidate.Unit.ID)
		}
	}
	return keys
}

func TestFirstAppearanceKeepsOnlyNewUnitsOfKnownRepositories(t *testing.T) {
	c := qt.New(t)
	earlier := cohortArtifact(t, "historical", "2019-06-30", nil)
	later := cohortArtifact(t, "contemporary", "2024-01-15", laterEdit)
	builder, err := corpus.NewAppearance("paragraph")
	c.Assert(err, qt.IsNil)
	c.Assert(builder.AddEarlier(t.Context(), earlier), qt.IsNil)
	c.Assert(builder.AddLater(t.Context(), later), qt.IsNil)
	filter, err := builder.Filter()
	c.Assert(err, qt.IsNil)
	c.Assert(filter.Version, qt.Equals, corpus.AppearanceVersion)
	c.Assert(filter.Kind, qt.Equals, "paragraph")
	c.Assert(filter.Earlier, qt.Equals, "historical")
	c.Assert(filter.Later, qt.Equals, "contemporary")
	c.Assert(filter.EarlierUnits, qt.Equals, len(paragraphs(earlier, "test-project")))
	c.Assert(filter.LaterUnits, qt.Equals, filter.RepeatedUnits+filter.LaterOnlyUnits+filter.NewUnits)
	c.Assert(filter.NewUnits, qt.Equals, len(filter.New))
	known := paragraphs(later, "test-project")
	repeated, found := known["The cache may retry & wait."]
	c.Assert(found, qt.IsTrue)
	fresh, found := known["A later snapshot adds this paragraph."]
	c.Assert(found, qt.IsTrue)
	c.Assert(filter.New, qt.DeepEquals, []string{fresh})
	c.Assert(filter.New, qt.Not(qt.Contains), repeated)
	c.Assert(filter.RepeatedUnits > 0, qt.IsTrue)
	c.Assert(filter.LaterOnly, qt.DeepEquals, []string{"other-project"})
	c.Assert(filter.LaterOnlyUnits, qt.Equals, len(paragraphs(later, "other-project")))
	c.Assert(filter.LaterOnlyUnits > 0, qt.IsTrue)
	c.Assert(filter.Repositories, qt.HasLen, 2)
	c.Assert(filter.Repositories[0].Repository, qt.Equals, "other-project")
	c.Assert(filter.Repositories[0].Earlier, qt.Equals, 0)
	c.Assert(filter.Repositories[1], qt.DeepEquals, corpus.RepositoryAppearance{Repository: "test-project",
		Earlier: filter.EarlierUnits, Later: len(known), Repeated: filter.RepeatedUnits, New: 1})
	// The filter survives a strict round trip; edited counts and unsorted keys do not.
	data, err := json.Marshal(filter)
	c.Assert(err, qt.IsNil)
	loaded, err := corpus.LoadAppearanceFilter(t.Context(), data)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, filter)
	for _, edit := range []func(*corpus.AppearanceFilter){
		func(f *corpus.AppearanceFilter) { f.NewUnits++ },
		func(f *corpus.AppearanceFilter) { f.RepeatedUnits-- },
		func(f *corpus.AppearanceFilter) { f.New = append(f.New, "a#a") },
		func(f *corpus.AppearanceFilter) { f.Version = "other" },
		func(f *corpus.AppearanceFilter) { f.Kind = "" },
		func(f *corpus.AppearanceFilter) { f.Later = f.Earlier },
	} {
		edited := loaded
		edited.New = append([]string(nil), loaded.New...)
		edit(&edited)
		data, err := json.Marshal(edited)
		c.Assert(err, qt.IsNil)
		_, err = corpus.LoadAppearanceFilter(t.Context(), data)
		c.Assert(err, qt.IsNotNil)
	}
}

func TestFirstAppearanceRejectsOrderAndCohortMistakes(t *testing.T) {
	c := qt.New(t)
	earlier := cohortArtifact(t, "historical", "2019-06-30", nil)
	later := cohortArtifact(t, "contemporary", "2024-01-15", laterEdit)
	_, err := corpus.NewAppearance(" ")
	c.Assert(err, qt.IsNotNil)
	builder, err := corpus.NewAppearance("paragraph")
	c.Assert(err, qt.IsNil)
	// Later before earlier, and a filter before both sides are present.
	c.Assert(builder.AddLater(t.Context(), later), qt.IsNotNil)
	_, err = builder.Filter()
	c.Assert(err, qt.IsNotNil)
	c.Assert(builder.AddEarlier(t.Context(), earlier), qt.IsNil)
	_, err = builder.Filter()
	c.Assert(err, qt.IsNotNil)
	// The same cohort on both sides, an artifact without cohorts, a mixed
	// artifact, and a wrong version are refused.
	c.Assert(builder.AddLater(t.Context(), earlier), qt.IsNotNil)
	blank := later
	blank.Units = append([]corpus.Candidate(nil), later.Units...)
	blank.Units[0].Cohort = ""
	c.Assert(builder.AddLater(t.Context(), blank), qt.IsNotNil)
	mixed := later
	mixed.Units = append([]corpus.Candidate(nil), later.Units...)
	mixed.Units[0].Cohort = "natural"
	c.Assert(builder.AddLater(t.Context(), mixed), qt.IsNotNil)
	wrong := later
	wrong.Version = "other"
	c.Assert(builder.AddLater(t.Context(), wrong), qt.IsNotNil)
	c.Assert(builder.AddLater(t.Context(), later), qt.IsNil)
	// Earlier artifacts cannot follow later ones, and a canceled context stops the walk.
	c.Assert(builder.AddEarlier(t.Context(), earlier), qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(builder.AddLater(ctx, later), qt.IsNotNil)
	filter, err := builder.Filter()
	c.Assert(err, qt.IsNil)
	c.Assert(filter.NewUnits, qt.Equals, 1)
}
