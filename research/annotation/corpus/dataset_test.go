package corpus_test

import (
	"context"
	"encoding/json"
	"maps"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// datasetFixture splits the sample sources into two shards that share a
// repository key, so one global group spans both shards.
func datasetFixture(c *qt.C) (corpus.Dataset, map[string][]byte, map[string]corpus.Manifest) {
	c.Helper()
	m, _ := sample()
	first, second := m, m
	first.ID, second.ID = "shard-a", "shard-b"
	first.Sources = []corpus.Source{m.Sources[0]}
	second.Sources = []corpus.Source{m.Sources[1]}
	shards := map[string][]byte{"shards/a.json": encodedIndent(c, first), "shards/b.json": encodedIndent(c, second)}
	dataset := corpus.Dataset{Version: corpus.DatasetVersion, ID: "dataset-test", Seed: m.Seed, Weights: m.Weights,
		Policy: m.Policy, UnitKinds: m.UnitKinds, RuleClassesSHA256: hash([]byte("rule classes"))}
	for _, path := range []string{"shards/b.json", "shards/a.json"} {
		dataset.Shards = append(dataset.Shards, corpus.Notice{Path: path, SHA256: hash(shards[path]), Bytes: len(shards[path])})
	}
	return dataset, shards, map[string]corpus.Manifest{"shards/a.json": first, "shards/b.json": second}
}

func encodedIndent(c *qt.C, value any) []byte {
	c.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	c.Assert(err, qt.IsNil)
	return append(data, '\n')
}

func TestDatasetPlanSpansShardsAndPinsThem(t *testing.T) {
	c := qt.New(t)
	dataset, shards, manifests := datasetFixture(c)
	loaded, err := corpus.LoadDataset(t.Context(), encoded(c, dataset))
	c.Assert(err, qt.IsNil)
	plan, err := corpus.MakeDatasetPlan(t.Context(), loaded, shards, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(plan.Version, qt.Equals, corpus.DatasetVersion)
	c.Assert(plan.Dataset.Shards[0].Path, qt.Equals, "shards/a.json")
	c.Assert(plan.Groups, qt.HasLen, 1)
	c.Assert(plan.Groups[0].Shards, qt.DeepEquals, []string{"shards/a.json", "shards/b.json"})
	c.Assert(plan.Groups[0].Sources, qt.Equals, 2)
	c.Assert(plan.Sources, qt.HasLen, 2)
	c.Assert(plan.Sources[0].Partition, qt.Equals, plan.Sources[1].Partition)
	c.Assert(plan.Sources[0].Group, qt.Equals, plan.Groups[0].ID)
	// The global group equals the group a single manifest with both sources would form.
	whole, _ := sample()
	single, err := corpus.MakePlan(t.Context(), whole)
	c.Assert(err, qt.IsNil)
	c.Assert(single.Groups, qt.HasLen, 1)
	c.Assert(single.Groups[0].ID, qt.Equals, plan.Groups[0].ID)
	c.Assert(single.Groups[0].Partition, qt.Equals, plan.Groups[0].Partition)
	// Pinned shards reproduce their digests, and a per-shard plan inherits the assignment.
	pinned, err := corpus.PinShards(t.Context(), plan, shards, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(corpus.VerifyPinnedShards(t.Context(), plan, pinned), qt.IsNil)
	for _, shard := range plan.Shards {
		manifest, err := corpus.LoadManifest(t.Context(), pinned[shard.Path])
		c.Assert(err, qt.IsNil)
		c.Assert(manifest.ID, qt.Equals, manifests[shard.Path].ID)
		local, err := corpus.MakePlan(t.Context(), manifest)
		c.Assert(err, qt.IsNil)
		c.Assert(local.Groups, qt.HasLen, 1)
		c.Assert(local.Groups[0].Pinned, qt.IsTrue)
		c.Assert(local.Groups[0].Partition, qt.Equals, plan.Groups[0].Partition)
	}
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), plan, shards, nil), qt.IsNil)
	// Shard order in the dataset does not change the plan.
	reordered := dataset
	reordered.Shards = []corpus.Notice{dataset.Shards[1], dataset.Shards[0]}
	again, err := corpus.MakeDatasetPlan(t.Context(), reordered, shards, nil)
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, plan)
}

func TestDatasetRejectsInconsistentShards(t *testing.T) {
	c := qt.New(t)
	dataset, shards, manifests := datasetFixture(c)
	edit := func(path string, change func(*corpus.Manifest)) map[string][]byte {
		m := manifests[path]
		m.Sources = append([]corpus.Source{}, m.Sources...)
		change(&m)
		result := maps.Clone(shards)
		result[path] = encodedIndent(c, m)
		return result
	}
	withShards := func(files map[string][]byte) corpus.Dataset {
		d := dataset
		d.Shards = nil
		for _, path := range []string{"shards/a.json", "shards/b.json"} {
			d.Shards = append(d.Shards, corpus.Notice{Path: path, SHA256: hash(files[path]), Bytes: len(files[path])})
		}
		return d
	}
	for _, row := range []struct {
		name  string
		files map[string][]byte
	}{
		{"header seed", edit("shards/a.json", func(m *corpus.Manifest) { m.Seed = "other" })},
		{"header kinds", edit("shards/b.json", func(m *corpus.Manifest) { m.UnitKinds = []string{"sentence"} })},
		{"duplicate source", edit("shards/b.json", func(m *corpus.Manifest) {
			m.Sources = []corpus.Source{manifests["shards/a.json"].Sources[0]}
		})},
		{"conflicting pins", edit("shards/a.json", func(m *corpus.Manifest) { m.Sources[0].Partition = "final_test" })},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			files := row.files
			if row.name == "conflicting pins" {
				files = edit("shards/b.json", func(m *corpus.Manifest) { m.Sources[0].Partition = "training" })
				files["shards/a.json"] = row.files["shards/a.json"]
			}
			_, err := corpus.MakeDatasetPlan(t.Context(), withShards(files), files, nil)
			c.Assert(err, qt.IsNotNil)
		})
	}
	// Declared shard bytes must match, and a tampered plan or pinned copy fails verification.
	plan, err := corpus.MakeDatasetPlan(t.Context(), dataset, shards, nil)
	c.Assert(err, qt.IsNil)
	missing := map[string][]byte{"shards/a.json": shards["shards/a.json"]}
	_, err = corpus.MakeDatasetPlan(t.Context(), dataset, missing, nil)
	c.Assert(err, qt.IsNotNil)
	tampered := plan
	tampered.Sources = append([]corpus.DatasetSource{}, plan.Sources...)
	tampered.Sources[0].Partition = "training"
	if plan.Sources[0].Partition == "training" {
		tampered.Sources[0].Partition = "development"
	}
	c.Assert(corpus.VerifyDatasetPlan(t.Context(), tampered, shards, nil), qt.IsNotNil)
	_, err = corpus.PinShards(t.Context(), tampered, shards, nil)
	c.Assert(err, qt.IsNotNil)
	pinned, err := corpus.PinShards(t.Context(), plan, shards, nil)
	c.Assert(err, qt.IsNil)
	pinned["shards/a.json"] = []byte(strings.Replace(string(pinned["shards/a.json"]), `"partition": "`, `"partition": "x`, 1))
	c.Assert(corpus.VerifyPinnedShards(t.Context(), plan, pinned), qt.IsNotNil)
	for _, bad := range []func(*corpus.Dataset){
		func(d *corpus.Dataset) { d.Version = "future" },
		func(d *corpus.Dataset) { d.RuleClassesSHA256 = "abc" },
		func(d *corpus.Dataset) { d.Shards = nil },
		func(d *corpus.Dataset) { d.Shards = append(d.Shards, d.Shards[0]) },
		func(d *corpus.Dataset) { d.Weights.Training = 0 },
	} {
		d := dataset
		d.Shards = append([]corpus.Notice{}, dataset.Shards...)
		bad(&d)
		_, err := corpus.LoadDataset(t.Context(), encoded(c, d))
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.MakeDatasetPlan(ctx, dataset, shards, nil)
	c.Assert(err, qt.IsNotNil)
}
