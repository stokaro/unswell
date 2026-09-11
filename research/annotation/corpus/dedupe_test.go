package corpus_test

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// copyEdit adds a second repository whose document repeats the readme's
// first paragraph, so one text appears twice inside the historical cohort.
func copyEdit(m *corpus.Manifest, files map[string][]byte) {
	addOther(m, files, "# Other\n\nThe **cache** may retry &amp; wait.\n\nOnly this repository says this.\n")
}

func TestDedupeKeepsTheFirstOccurrenceAcrossCohorts(t *testing.T) {
	c := qt.New(t)
	earlier := cohortArtifact(t, "historical", "2019-06-30", copyEdit)
	later := cohortArtifact(t, "contemporary", "2024-01-15", laterEdit)
	builder, err := corpus.NewDedupe("paragraph", []string{"historical", "contemporary"})
	c.Assert(err, qt.IsNil)
	// Artifacts may arrive in any order; the stated order decides precedence.
	c.Assert(builder.Add(t.Context(), later), qt.IsNil)
	c.Assert(builder.Add(t.Context(), earlier), qt.IsNil)
	selection, err := builder.Selection()
	c.Assert(err, qt.IsNil)
	c.Assert(selection.Version, qt.Equals, corpus.SelectionVersion)
	c.Assert(selection.Order, qt.DeepEquals, []string{"historical", "contemporary"})
	c.Assert(selection.Cohorts, qt.HasLen, 2)
	historical, contemporary := selection.Cohorts[0], selection.Cohorts[1]
	c.Assert(historical.Cohort, qt.Equals, "historical")
	c.Assert(historical.RepeatedEarlier, qt.Equals, 0)
	// The copied paragraph repeats inside the historical cohort once; the
	// first occurrence by repository name is the one in other-project.
	c.Assert(historical.RepeatedWithin, qt.Equals, 1)
	c.Assert(historical.Units, qt.Equals, historical.Kept+1)
	other := paragraphs(earlier, "other-project")
	main := paragraphs(earlier, "test-project")
	c.Assert(selection.Kept, qt.Contains, other["The cache may retry & wait."])
	c.Assert(selection.Kept, qt.Not(qt.Contains), main["The cache may retry & wait."])
	c.Assert(contemporary.Cohort, qt.Equals, "contemporary")
	c.Assert(contemporary.RepeatedWithin, qt.Equals, 0)
	// Contemporary text repeated from either historical repository is not new;
	// the paragraph added later and the other repository's new text are kept.
	c.Assert(contemporary.Kept, qt.Equals, 2)
	c.Assert(contemporary.RepeatedEarlier, qt.Equals, contemporary.Units-2)
	c.Assert(selection.Kept, qt.Contains, paragraphs(later, "test-project")["A later snapshot adds this paragraph."])
	c.Assert(selection.Kept, qt.Not(qt.Contains), paragraphs(later, "test-project")["The cache may retry & wait."])
	c.Assert(selection.Kept, qt.HasLen, historical.Kept+contemporary.Kept)
	// The selection survives a strict round trip; broken counts and keys do not.
	data, err := json.Marshal(selection)
	c.Assert(err, qt.IsNil)
	loaded, err := corpus.LoadUnitSelection(t.Context(), data)
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, selection)
	for _, edit := range []func(*corpus.UnitSelection){
		func(s *corpus.UnitSelection) { s.Cohorts[0].Kept++ },
		func(s *corpus.UnitSelection) { s.Kept = append(s.Kept, "a#a") },
		func(s *corpus.UnitSelection) { s.Order = []string{"historical", "historical"} },
		func(s *corpus.UnitSelection) { s.Cohorts[1].Cohort = "natural" },
		func(s *corpus.UnitSelection) { s.Version = "other" },
		func(s *corpus.UnitSelection) { s.Kind = "" },
	} {
		edited := loaded
		edited.Cohorts = append([]corpus.CohortSelection(nil), loaded.Cohorts...)
		edited.Kept = append([]string(nil), loaded.Kept...)
		edit(&edited)
		data, err := json.Marshal(edited)
		c.Assert(err, qt.IsNil)
		_, err = corpus.LoadUnitSelection(t.Context(), data)
		c.Assert(err, qt.IsNotNil)
	}
}

func TestDedupeRejectsUnknownCohortsAndEmptySelections(t *testing.T) {
	c := qt.New(t)
	earlier := cohortArtifact(t, "historical", "2019-06-30", nil)
	_, err := corpus.NewDedupe("", []string{"historical"})
	c.Assert(err, qt.IsNotNil)
	_, err = corpus.NewDedupe("paragraph", nil)
	c.Assert(err, qt.IsNotNil)
	_, err = corpus.NewDedupe("paragraph", []string{"historical", "historical"})
	c.Assert(err, qt.IsNotNil)
	builder, err := corpus.NewDedupe("paragraph", []string{"contemporary"})
	c.Assert(err, qt.IsNil)
	// A cohort outside the order, a wrong version, a repeated unit key, and a
	// canceled context are refused.
	c.Assert(builder.Add(t.Context(), earlier), qt.IsNotNil)
	twice, err := corpus.NewDedupe("paragraph", []string{"historical"})
	c.Assert(err, qt.IsNil)
	c.Assert(twice.Add(t.Context(), earlier), qt.IsNil)
	c.Assert(twice.Add(t.Context(), earlier), qt.IsNotNil)
	wrong := earlier
	wrong.Version = "other"
	c.Assert(builder.Add(t.Context(), wrong), qt.IsNotNil)
	_, err = builder.Selection()
	c.Assert(err, qt.IsNotNil)
	builder, err = corpus.NewDedupe("fragment", []string{"historical"})
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(builder.Add(ctx, earlier), qt.IsNotNil)
	c.Assert(builder.Add(t.Context(), earlier), qt.IsNil)
	selection, err := builder.Selection()
	c.Assert(err, qt.IsNil)
	c.Assert(selection.Kind, qt.Equals, "fragment")
	c.Assert(selection.Cohorts, qt.HasLen, 1)
	c.Assert(selection.Cohorts[0].Kept > 0, qt.IsTrue)
}
