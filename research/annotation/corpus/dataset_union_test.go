package corpus_test

import (
	"path"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// unionFixture gives the two shards of the dataset fixture different cohorts
// and repositories, so a union has one checkout prefix per source.
func unionFixture(c *qt.C) (corpus.DatasetPlan, map[string][]byte, map[string][]byte) {
	c.Helper()
	_, files := sample()
	dataset, _, manifests := datasetFixture(c)
	first, second := manifests["shards/a.json"], manifests["shards/b.json"]
	first.Sources[0].Repository = "octo/historical"
	first.Sources[0].Snapshot = &corpus.Snapshot{Date: "2012-12-28", Confidence: "vcs_only",
		Evidence: "commit dated 2012-12-28", Cohort: "historical-2012"}
	second.Sources[0].Repository = "octo/contemporary"
	second.Sources[0].Snapshot = &corpus.Snapshot{Date: "2026-06-19", Confidence: "corroborated",
		Evidence: "release dated 2026-06-19", Cohort: "contemporary"}
	// A second contemporary source of the same checkout exercises the cap.
	third := first.Sources[0]
	third.ID, third.Path, third.Document = "d2", "docs/readme.md", "test-project/docs/readme.md"
	third.Repository, third.Snapshot = second.Sources[0].Repository, second.Sources[0].Snapshot
	second.Sources = append(second.Sources, third)
	shards := map[string][]byte{"shards/a.json": encodedIndent(c, first), "shards/b.json": encodedIndent(c, second)}
	dataset.Shards = nil
	for _, path := range []string{"shards/a.json", "shards/b.json"} {
		dataset.Shards = append(dataset.Shards, corpus.Notice{Path: path, SHA256: hash(shards[path]), Bytes: len(shards[path])})
	}
	plan, err := corpus.MakeDatasetPlan(c.Context(), dataset, shards, nil)
	c.Assert(err, qt.IsNil)
	pinned, err := corpus.PinShards(c.Context(), plan, shards, nil)
	c.Assert(err, qt.IsNil)
	return plan, pinned, files
}

func TestUnionManifestPrefixesCheckoutsAndKeepsPins(t *testing.T) {
	c := qt.New(t)
	plan, pinned, files := unionFixture(c)
	union, err := corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "union-test", UnitKinds: []string{"paragraph"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.ID, qt.Equals, "union-test")
	c.Assert(union.Seed, qt.Equals, plan.Dataset.Seed)
	c.Assert(union.UnitKinds, qt.DeepEquals, []string{"paragraph"})
	c.Assert(union.Sources, qt.HasLen, 3)
	c.Assert(union.Sources[0].Path, qt.Equals, "historical-2012/octo__historical/readme.md")
	c.Assert(union.Sources[0].Notices[0].Path, qt.Equals, "historical-2012/octo__historical/LICENSE")
	c.Assert(union.Sources[1].Path, qt.Equals, "contemporary/octo__contemporary/sample.go")
	partitions := map[string]string{}
	for _, source := range plan.Sources {
		partitions[source.ID] = source.Partition
	}
	for _, source := range union.Sources {
		c.Assert(source.Partition, qt.Equals, partitions[source.ID])
	}

	// The union extracts under one root whose layout names each checkout.
	prefixed := map[string][]byte{}
	for _, source := range union.Sources {
		prefixed[source.Path] = files[path.Base(source.Path)]
		for _, notice := range source.Notices {
			prefixed[notice.Path] = files[path.Base(notice.Path)]
		}
	}
	local, err := corpus.MakePlan(c.Context(), union)
	c.Assert(err, qt.IsNil)
	for _, group := range local.Groups {
		c.Assert(group.Pinned, qt.IsTrue)
	}
	artifact, err := corpus.Build(c.Context(), local, prefixed)
	c.Assert(err, qt.IsNil)
	cohorts := map[string]int{}
	for _, candidate := range artifact.Units {
		c.Assert(candidate.Unit.Kind, qt.Equals, "paragraph")
		cohorts[candidate.Cohort]++
	}
	c.Assert(cohorts["historical-2012"] > 0, qt.IsTrue)
	c.Assert(cohorts["contemporary"] > 0, qt.IsTrue)
}

func TestUnionManifestSelectsAndRefuses(t *testing.T) {
	c := qt.New(t)
	plan, pinned, _ := unionFixture(c)
	union, err := corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", Cohorts: []string{"contemporary"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 2)
	c.Assert(union.Sources[0].Repository, qt.Equals, "octo/contemporary")
	c.Assert(union.UnitKinds, qt.DeepEquals, plan.Dataset.UnitKinds)

	union, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", Repositories: []string{"octo/historical"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 1)
	union, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", Roles: []string{"documentation"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 3)
	_, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", Roles: []string{"comment"}})
	c.Assert(err, qt.ErrorMatches, "no source of the dataset matches the selection")
	union, err = corpus.UnionManifest(c.Context(), plan, pinned,
		corpus.UnionOptions{ID: "u", Roles: []string{"comment"}, EveryRoleCohorts: []string{"contemporary"}})
	c.Assert(err, qt.IsNil)
	c.Assert(len(union.Sources) > 0, qt.IsTrue)
	for _, source := range union.Sources {
		c.Assert(source.Snapshot.Cohort, qt.Equals, "contemporary")
	}

	// A capped checkout keeps its first source in ID order; an uncapped cohort keeps all.
	union, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", MaxPerCheckout: 1})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 2)
	c.Assert(union.Sources[0].ID, qt.Equals, "d1")
	c.Assert(union.Sources[1].ID, qt.Equals, "d0")
	union, err = corpus.UnionManifest(c.Context(), plan, pinned,
		corpus.UnionOptions{ID: "u", MaxPerCheckout: 1, UncappedCohorts: []string{"contemporary"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 3)
	_, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", MaxPerCheckout: -1})
	c.Assert(err, qt.ErrorMatches, "the per-checkout limit cannot be negative")
	smallest := plan.Sources[0]
	_ = smallest
	union, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", ExcludedSources: []string{"d0", "d2"}})
	c.Assert(err, qt.IsNil)
	c.Assert(union.Sources, qt.HasLen, 1)
	c.Assert(union.Sources[0].ID, qt.Equals, "d1")
	union, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", MaxSourceBytes: 120})
	c.Assert(err, qt.IsNil)
	for _, source := range union.Sources {
		c.Assert(source.Bytes <= 120, qt.IsTrue)
	}
	c.Assert(len(union.Sources) < 3, qt.IsTrue)

	_, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", Cohorts: []string{"natural"}})
	c.Assert(err, qt.ErrorMatches, "no source of the dataset matches the selection")
	_, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: "u", UnitKinds: []string{"fragment", "verse"}})
	c.Assert(err, qt.ErrorMatches, "unit kind verse is not in the dataset")
	_, err = corpus.UnionManifest(c.Context(), plan, pinned, corpus.UnionOptions{ID: ""})
	c.Assert(err, qt.IsNotNil)

	tampered := map[string][]byte{}
	for name, data := range pinned {
		tampered[name] = append([]byte(nil), data...)
	}
	tampered["shards/a.json"] = append(tampered["shards/a.json"], '\n')
	_, err = corpus.UnionManifest(c.Context(), plan, tampered, corpus.UnionOptions{ID: "u"})
	c.Assert(err, qt.ErrorMatches, "pinned shard shards/a.json does not match the plan")
	delete(tampered, "shards/b.json")
	_, err = corpus.UnionManifest(c.Context(), plan, tampered, corpus.UnionOptions{ID: "u"})
	c.Assert(err, qt.IsNotNil)
}
