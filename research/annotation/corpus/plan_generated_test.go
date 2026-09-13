package corpus_test

import (
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// A model sometimes answers two unrelated prompts with the same words. Joining
// the two documents by their bytes would merge the repositories they answer
// for, and a component that spans two pinned repositories has no partition.
// Human sources keep the content join, because identical bytes there mean a
// vendored copy of one file.
func TestIdenticalGeneratedTextDoesNotJoinRepositories(t *testing.T) {
	c := qt.New(t)
	manifest, _ := sample()
	shared := manifest.Sources[0]

	first := shared
	first.ID, first.Path = "g1", "generated/run/a.md"
	first.Repository, first.Document = "example/one", "example/one/"+first.Path
	first.Reference = "fixture:" + first.Path
	first.GenerationTasks = []string{"t1"}
	first.Partition = "training"

	second := first
	second.ID, second.Path = "g2", "generated/run/b.md"
	second.Repository, second.Document = "example/two", "example/two/"+second.Path
	second.Reference = "fixture:" + second.Path
	second.GenerationTasks = []string{"t2"}
	second.Partition = "development"

	manifest.Sources = append(manifest.Sources, first, second)
	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil, qt.Commentf("two identical responses must not share a component"))

	group := func(id string) string {
		for _, g := range plan.Groups {
			if slices.Contains(g.Sources, id) {
				return g.ID
			}
		}
		return ""
	}
	c.Assert(group("g1"), qt.Not(qt.Equals), "")
	c.Assert(group("g1"), qt.Not(qt.Equals), group("g2"))
}
