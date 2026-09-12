package corpus_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// pinFile encodes a pin file naming the given repositories.
func pinFile(c *qt.C, decidedOn string, repositories map[string]string) []byte {
	c.Helper()
	return encoded(c, map[string]any{"format": corpus.PartitionPinsVersion, "decided_on": decidedOn,
		"protocol": "unswell-llm-patterns-v1", "repositories": repositories})
}

func loadPins(c *qt.C, data []byte) *corpus.PartitionPins {
	c.Helper()
	pins, err := corpus.LoadPartitionPins(c.Context(), data)
	c.Assert(err, qt.IsNil)
	return &pins
}

// editedFixture rebuilds the dataset fixture after a change to its two
// shard manifests, so the dataset declares the edited shard bytes.
func editedFixture(c *qt.C, change func(a, b *corpus.Manifest)) (corpus.Dataset, map[string][]byte) {
	c.Helper()
	dataset, _, manifests := datasetFixture(c)
	a, b := manifests["shards/a.json"], manifests["shards/b.json"]
	a.Sources, b.Sources = append([]corpus.Source{}, a.Sources...), append([]corpus.Source{}, b.Sources...)
	change(&a, &b)
	shards := map[string][]byte{"shards/a.json": encodedIndent(c, a), "shards/b.json": encodedIndent(c, b)}
	dataset.Shards = nil
	for _, path := range []string{"shards/a.json", "shards/b.json"} {
		dataset.Shards = append(dataset.Shards, corpus.Notice{Path: path, SHA256: hash(shards[path]), Bytes: len(shards[path])})
	}
	return dataset, shards
}

func TestPartitionPinsMoveWholeComponents(t *testing.T) {
	c := qt.New(t)
	dataset, shards, _ := datasetFixture(c)
	unpinned, err := corpus.MakeDatasetPlan(t.Context(), dataset, shards, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(unpinned.Groups[0].Pinned, qt.IsFalse)
	target := "final_test"
	if unpinned.Groups[0].Partition == target {
		target = "training"
	}
	file := pinFile(c, "2026-09-12", map[string]string{"test-project": target})
	pins := loadPins(c, file)
	c.Assert(pins.SHA256, qt.Equals, hash(file))
	plan, err := corpus.MakeDatasetPlan(t.Context(), dataset, shards, pins)
	c.Assert(err, qt.IsNil)
	c.Assert(plan.Groups, qt.HasLen, 1)
	c.Assert(plan.Groups[0].Partition, qt.Equals, target)
	c.Assert(plan.Groups[0].Pinned, qt.IsTrue)
	c.Assert(plan.Groups[0].ID, qt.Equals, unpinned.Groups[0].ID)
	c.Assert(plan.PartitionPinsSHA256, qt.Equals, hash(file))
	c.Assert(plan.PinnedRepositories, qt.Equals, 1)
	c.Assert(plan.UnmatchedPins, qt.IsNil)
	c.Assert(plan.DatasetSHA256, qt.Equals, unpinned.DatasetSHA256)
	for _, source := range plan.Sources {
		c.Assert(source.Partition, qt.Equals, target)
	}
	// The pinned shards carry the pin, so a local plan inherits it.
	pinned, err := corpus.PinShards(t.Context(), plan, shards, pins)
	c.Assert(err, qt.IsNil)
	for _, shard := range plan.Shards {
		manifest, err := corpus.LoadManifest(t.Context(), pinned[shard.Path])
		c.Assert(err, qt.IsNil)
		c.Assert(manifest.Sources[0].Partition, qt.Equals, target)
	}
	// Verification needs the same file, and a loaded copy of the plan reproduces.
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), plan, shards, pins), qt.IsNil)
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), plan, shards, nil), qt.ErrorMatches, ".*made with partition pins.*")
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), unpinned, shards, pins), qt.ErrorMatches, ".*made without partition pins.*")
	other := loadPins(c, pinFile(c, "2026-09-13", map[string]string{"test-project": target}))
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), plan, shards, other), qt.ErrorMatches, ".*differ from the plan.*")
	loaded, err := corpus.LoadDatasetPlan(t.Context(), encoded(c, plan))
	c.Assert(err, qt.IsNil)
	c.Assert(loaded, qt.DeepEquals, plan)
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), loaded, shards, pins), qt.IsNil)
}

func TestPartitionPinsReportUnmatchedAndRejectConflicts(t *testing.T) {
	c := qt.New(t)
	dataset, shards, _ := datasetFixture(c)
	pins := loadPins(c, pinFile(c, "2026-09-12", map[string]string{"test-project": "training", "absent/repo": "development"}))
	plan, err := corpus.MakeDatasetPlan(t.Context(), dataset, shards, pins)
	c.Assert(err, qt.IsNil)
	c.Assert(plan.PinnedRepositories, qt.Equals, 1)
	c.Assert(plan.UnmatchedPins, qt.DeepEquals, []string{"absent/repo"})
	c.Assert(plan.Groups[0].Partition, qt.Equals, "training")
	// A source that already carries another pin names both values.
	dataset, shards = editedFixture(c, func(a, _ *corpus.Manifest) { a.Sources[0].Partition = "final_test" })
	_, err = corpus.MakeDatasetPlan(t.Context(), dataset, shards, pins)
	c.Assert(err, qt.ErrorMatches, "source d0 pins partition final_test while the partition file pins test-project to training")
	// Two repositories of one component pinned apart name both sources.
	dataset, shards = editedFixture(c, func(a, b *corpus.Manifest) {
		b.Sources[0].Repository = "second-project"
		a.Sources[0].Authors, b.Sources[0].Authors = []string{"author:a"}, []string{"author:a"}
	})
	apart := loadPins(c, pinFile(c, "2026-09-12", map[string]string{"test-project": "training", "second-project": "development"}))
	_, err = corpus.MakeDatasetPlan(t.Context(), dataset, shards, apart)
	c.Assert(err, qt.ErrorMatches, "connected sources d0 and d1 have conflicting partition pins: training and development")
	// Pins that did not come from a file carry no digest.
	unhashed := &corpus.PartitionPins{Repositories: map[string]string{"test-project": "training"}}
	_, err = corpus.MakeDatasetPlan(t.Context(), dataset, shards, unhashed)
	c.Assert(err, qt.ErrorMatches, ".*no file digest.*")
}

func pinJSON(format, protocol, repositories string) string {
	return `{"format":"` + format + `","decided_on":"2026-09-12","protocol":"` + protocol + `","repositories":` + repositories + `}`
}

func TestLoadPartitionPinsRejectsBadFiles(t *testing.T) {
	good := pinJSON(corpus.PartitionPinsVersion, "p", `{"a/b":"training"}`)
	for _, row := range []struct{ name, input string }{
		{"format", pinJSON("other", "p", `{"a/b":"training"}`)},
		{"partition", pinJSON(corpus.PartitionPinsVersion, "p", `{"a/b":"test"}`)},
		{"empty", pinJSON(corpus.PartitionPinsVersion, "p", `{}`)},
		{"protocol", pinJSON(corpus.PartitionPinsVersion, " ", `{"a/b":"training"}`)},
		{"unknown field", strings.TrimSuffix(good, "}") + `,"x":1}`},
		{"duplicate key", pinJSON(corpus.PartitionPinsVersion, "p", `{"a/b":"training","a/b":"training"}`)},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := corpus.LoadPartitionPins(t.Context(), []byte(row.input))
			c.Assert(err, qt.IsNotNil)
		})
	}
	c := qt.New(t)
	pins, err := corpus.LoadPartitionPins(t.Context(), []byte(good))
	c.Assert(err, qt.IsNil)
	c.Assert(pins.Repositories, qt.DeepEquals, map[string]string{"a/b": "training"})
	c.Assert(pins.SHA256, qt.Equals, hash([]byte(good)))
}
