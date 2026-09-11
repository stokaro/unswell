package generation_test

import (
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/generation"
)

// fixtureArtifact holds comment paragraphs of two repositories in one file
// each; the file text is the comment followed by a declaration.
func fixtureArtifact(cohort string, count int, repository string) (corpus.Artifact, map[string][]byte) {
	files := map[string][]byte{}
	art := corpus.Artifact{Version: corpus.Version, Status: "unlabeled_candidates"}
	for i := range count {
		text := fmt.Sprintf("Handler%d serves the request for item %d and returns the stored value to the caller.", i, i)
		source := "// " + text + "\nfunc Handler" + fmt.Sprint(i) + "(w Writer, r *Request) error {\n}\n"
		path := fmt.Sprintf("h%d.go", i)
		files[path] = []byte(source)
		art.Units = append(art.Units, corpus.Candidate{SourceID: fmt.Sprintf("s-%s-%d", repository, i), GroupID: "g-" + repository,
			Partition: "training", Cohort: cohort, Words: len(strings.Fields(text)),
			Unit: annotation.Unit{ID: fmt.Sprintf("u%06d", i), Kind: "paragraph", Role: "comment", Text: text,
				Source: annotation.Source{DocumentID: repository + "/" + path, RepositoryID: repository, Language: document.Go,
					Segments: []document.Span{{Start: 3, End: 3 + len(text)}}}}})
	}
	return art, files
}

func plan(ids map[string]string) corpus.DatasetPlan {
	p := corpus.DatasetPlan{Version: corpus.DatasetVersion}
	for id, partition := range ids {
		p.Sources = append(p.Sources, corpus.DatasetSource{ID: id, Partition: partition})
	}
	return p
}

func TestSamplerDrawsStratifiedDeterministicTasks(t *testing.T) {
	c := qt.New(t)
	goArt, goFiles := fixtureArtifact("historical", 10, "org/go-lib")
	rustArt, rustFiles := fixtureArtifact("historical", 4, "org/rust-lib")
	ids := map[string]string{}
	for _, unit := range append(goArt.Units, rustArt.Units...) {
		ids[unit.SourceID] = "training"
	}
	ids["s-org/go-lib-9"] = "final_test"
	options := generation.Options{Protocol: "unswell-llm-patterns-v1", Seed: "seed-a", Cohort: "historical",
		Partitions: []string{"training", "development"}, Roles: []string{"comment"}, Count: 6,
		Ecosystems: map[string]string{"org/go-lib": "go", "org/rust-lib": "rust"}}
	draw := func(seed string) generation.Tasks {
		options.Seed = seed
		sampler, err := generation.NewSampler(options, plan(ids))
		c.Assert(err, qt.IsNil)
		c.Assert(sampler.Add(t.Context(), goArt, func(p string) ([]byte, error) { return goFiles[p], nil }), qt.IsNil)
		c.Assert(sampler.Add(t.Context(), rustArt, func(p string) ([]byte, error) { return rustFiles[p], nil }), qt.IsNil)
		tasks, err := sampler.Sample()
		c.Assert(err, qt.IsNil)
		return tasks
	}
	tasks := draw("seed-a")
	c.Assert(tasks.Version, qt.Equals, generation.TasksVersion)
	c.Assert(tasks.Tasks, qt.HasLen, 6)
	// Nine Go units are eligible (one sits in the confirmation partition) and
	// four Rust units; each stratum keeps one task and the rest go by share.
	c.Assert(tasks.Strata, qt.DeepEquals, []generation.Stratum{{Ecosystem: "go", Role: "comment", Eligible: 9, Selected: 4},
		{Ecosystem: "rust", Role: "comment", Eligible: 4, Selected: 2}})
	for _, task := range tasks.Tasks {
		c.Assert(task.SourceID, qt.Not(qt.Equals), "s-org/go-lib-9")
		c.Assert(task.FactSheet.Signature, qt.Matches, `func Handler\d+\(w Writer, r \*Request\) error`)
		c.Assert(task.FactSheet.Parameters, qt.Equals, "w Writer, r *Request")
		c.Assert(task.Words >= generation.MinWords, qt.IsTrue)
		c.Assert(task.TextSHA256, qt.HasLen, 64)
	}
	c.Assert(draw("seed-a"), qt.DeepEquals, tasks)
	other := draw("seed-b")
	c.Assert(other.Strata, qt.DeepEquals, tasks.Strata)
	c.Assert(other.Tasks, qt.Not(qt.DeepEquals), tasks.Tasks)
}

func TestSamplerRefusesBadOptionsAndEmptyPools(t *testing.T) {
	c := qt.New(t)
	art, files := fixtureArtifact("contemporary", 2, "org/go-lib")
	ids := map[string]string{"s-org/go-lib-0": "training", "s-org/go-lib-1": "training"}
	good := generation.Options{Protocol: "p", Seed: "s", Cohort: "historical", Partitions: []string{"training"},
		Roles: []string{"comment"}, Count: 1, Ecosystems: map[string]string{"org/go-lib": "go"}}
	for _, edit := range []func(*generation.Options){
		func(o *generation.Options) { o.Seed = "" }, func(o *generation.Options) { o.Count = 0 },
		func(o *generation.Options) { o.Count = generation.MaxTasks + 1 }, func(o *generation.Options) { o.Roles = nil },
	} {
		options := good
		edit(&options)
		_, err := generation.NewSampler(options, plan(ids))
		c.Assert(err, qt.IsNotNil)
	}
	sampler, err := generation.NewSampler(good, plan(ids))
	c.Assert(err, qt.IsNil)
	// The cohort differs, so nothing is eligible.
	c.Assert(sampler.Add(t.Context(), art, func(p string) ([]byte, error) { return files[p], nil }), qt.IsNil)
	_, err = sampler.Sample()
	c.Assert(err, qt.IsNotNil)
	wrong := art
	wrong.Version = "other"
	c.Assert(sampler.Add(t.Context(), wrong, nil), qt.IsNotNil)
	// A file the reader cannot supply is an error, not a silent skip.
	art.Units[0].Cohort = "historical"
	failing := func(string) ([]byte, error) { return nil, fmt.Errorf("gone") }
	c.Assert(sampler.Add(t.Context(), art, failing), qt.ErrorMatches, "h0.go: gone")
}

// A task set of an earlier run leaves the pool, so a later draw under the
// same seed takes new tasks and counts what it left out.
func TestSamplerExcludesEarlierTasks(t *testing.T) {
	c := qt.New(t)
	goArt, goFiles := fixtureArtifact("historical", 10, "org/go-lib")
	ids := map[string]string{}
	for _, unit := range goArt.Units {
		ids[unit.SourceID] = "training"
	}
	options := generation.Options{Protocol: "unswell-llm-patterns-v1", Seed: "seed-a", Cohort: "historical",
		Partitions: []string{"training"}, Roles: []string{"comment"}, Count: 4, Ecosystems: map[string]string{"org/go-lib": "go"}}
	draw := func(excluded map[string]bool) generation.Tasks {
		options.Excluded = excluded
		sampler, err := generation.NewSampler(options, plan(ids))
		c.Assert(err, qt.IsNil)
		c.Assert(sampler.Add(t.Context(), goArt, func(p string) ([]byte, error) { return goFiles[p], nil }), qt.IsNil)
		tasks, err := sampler.Sample()
		c.Assert(err, qt.IsNil)
		return tasks
	}
	first := draw(nil)
	c.Assert(first.Tasks, qt.HasLen, 4)
	c.Assert(first.Excluded, qt.Equals, 0)
	excluded := map[string]bool{}
	for _, task := range first.Tasks {
		excluded[task.ID] = true
	}
	second := draw(excluded)
	c.Assert(second.Tasks, qt.HasLen, 4)
	c.Assert(second.Excluded, qt.Equals, 4)
	c.Assert(second.Strata[0].Eligible, qt.Equals, 6)
	for _, task := range second.Tasks {
		c.Assert(excluded[task.ID], qt.IsFalse)
	}
}
