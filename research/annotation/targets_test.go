package annotation_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

func TestMatchTargetsRejectsChangedBindings(t *testing.T) {
	for _, row := range []struct {
		name string
		edit func(*annotation.Unit)
	}{
		{"id", func(u *annotation.Unit) { u.ID = "other" }},
		{"text", func(u *annotation.Unit) { u.Text += "." }},
		{"context", func(u *annotation.Unit) { u.Context += "." }},
		{"kind", func(u *annotation.Unit) { u.Kind = "fragment" }},
		{"role", func(u *annotation.Unit) { u.Role = "error_message" }},
		{"document", func(u *annotation.Unit) { u.Source.DocumentID += "-other" }},
		{"repository", func(u *annotation.Unit) { u.Source.RepositoryID += "-other" }},
		{"author", func(u *annotation.Unit) { u.Source.AuthorGroup += "-other" }},
		{"template", func(u *annotation.Unit) { u.Source.TemplateID += "-other" }},
		{"group", func(u *annotation.Unit) { u.Source.RelatedGroup += "-other" }},
		{"reference", func(u *annotation.Unit) { u.Source.Reference += "-other" }},
		{"source hash", func(u *annotation.Unit) { u.Source.SHA256 = "changed" }},
		{"source size", func(u *annotation.Unit) { u.Source.Bytes++ }},
		{"source format", func(u *annotation.Unit) { u.Source.Language = "go" }},
		{"prose language", func(u *annotation.Unit) { u.Source.ProseLanguage = "ru" }},
		{"segments", func(u *annotation.Unit) { u.Source.Segments[0].End-- }},
		{"extraction", func(u *annotation.Unit) { u.Extraction.Identity += "-other" }},
		{"policy", func(u *annotation.Unit) { u.Extraction.PolicySHA256 = "changed" }},
		{"context policy", func(u *annotation.Unit) { u.Extraction.ContextPolicy += "-other" }},
		{"license", func(u *annotation.Unit) { u.Rights.License = "changed" }},
		{"rights evidence", func(u *annotation.Unit) { u.Rights.Evidence += "-other" }},
		{"allowed uses", func(u *annotation.Unit) { u.Rights.AllowedUses = []string{"annotation"} }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			data := fixture(c)
			round, err := annotation.Load(t.Context(), encode(c, data))
			c.Assert(err, qt.IsNil)
			c.Assert(round.MatchTargets(t.Context(), data.Units), qt.IsNil)
			row.edit(&data.Units[0])
			c.Assert(round.MatchTargets(t.Context(), data.Units), qt.ErrorMatches, ".*does not match.*")
		})
	}
}

func TestMatchTargetsOriginSetsAndLimits(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	data.Units[0].Origin = annotation.Origin{Label: "unknown", Scope: "unit", Evidence: "Independent curation."}
	slices.Reverse(data.Units[0].Rights.AllowedUses)
	slices.Reverse(data.Units)
	c.Assert(round.MatchTargets(t.Context(), data.Units), qt.IsNil)
	c.Assert(round.MatchTargets(t.Context(), nil), qt.IsNotNil)
	c.Assert(round.MatchTargets(t.Context(), make([]annotation.Unit, 10001)), qt.IsNotNil)
	duplicate := append(slices.Clone(data.Units), data.Units[0])
	c.Assert(round.MatchTargets(t.Context(), duplicate), qt.ErrorMatches, ".*unique.*")
	data.Units[0].ID = ""
	c.Assert(round.MatchTargets(t.Context(), data.Units), qt.ErrorMatches, ".*nonempty.*")
	var absent *annotation.Round
	c.Assert(absent.MatchTargets(t.Context(), data.Units), qt.IsNotNil)
	c.Assert((&annotation.Round{}).MatchTargets(t.Context(), data.Units), qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(round.MatchTargets(ctx, data.Units), qt.ErrorIs, context.Canceled)
}
