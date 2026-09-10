package main

// White-box tests: The dataset commands write pinned shards through the process root and refuse to overwrite;
// package main exposes no importable API for that file handling or for its result types.

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// datasetFixture splits the Ptah manifest into two shards under a temporary root.
func datasetFixture(c *qt.C) (string, []byte) {
	c.Helper()
	var manifest corpus.Manifest
	c.Assert(json.Unmarshal(manifestBytes(c), &manifest), qt.IsNil)
	root := c.TB.(*testing.T).TempDir()
	c.Assert(os.MkdirAll(filepath.Join(root, "shards"), 0o750), qt.IsNil)
	dataset := corpus.Dataset{Version: corpus.DatasetVersion, ID: "ptah-dataset", Seed: manifest.Seed,
		Weights: manifest.Weights, Policy: manifest.Policy, UnitKinds: manifest.UnitKinds,
		RuleClassesSHA256: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"}
	half := len(manifest.Sources) / 2
	for i, sources := range [][]corpus.Source{manifest.Sources[:half], manifest.Sources[half:]} {
		shard := manifest
		shard.ID = manifest.ID + "-" + string(rune('a'+i))
		shard.Sources = sources
		data, err := json.MarshalIndent(shard, "", "  ")
		c.Assert(err, qt.IsNil)
		name := filepath.Join("shards", shard.ID+".json")
		c.Assert(os.WriteFile(filepath.Join(root, name), data, 0o600), qt.IsNil)
		dataset.Shards = append(dataset.Shards, corpus.Notice{Path: filepath.ToSlash(name), SHA256: sha(data), Bytes: len(data)})
	}
	encoded, err := json.Marshal(dataset)
	c.Assert(err, qt.IsNil)
	return root, encoded
}

func sha(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

func manifestBytes(c *qt.C) []byte {
	c.Helper()
	data, err := os.ReadFile("../../corpus/testdata/ptah-manifest.json")
	c.Assert(err, qt.IsNil)
	return data
}

func TestDatasetCommandsPlanPinAndVerify(t *testing.T) {
	c := qt.New(t)
	root, dataset := datasetFixture(c)
	var plan bytes.Buffer
	c.Assert(run(t.Context(), []string{"dataset", "plan", "--root", root}, bytes.NewReader(dataset), &plan), qt.IsNil)
	var decoded corpus.DatasetPlan
	c.Assert(json.Unmarshal(plan.Bytes(), &decoded), qt.IsNil)
	c.Assert(decoded.Shards, qt.HasLen, 2)
	c.Assert(decoded.Groups, qt.HasLen, 1)
	c.Assert(decoded.Groups[0].Partition, qt.Equals, "development")
	c.Assert(decoded.Sources, qt.HasLen, 8)
	output := t.TempDir()
	var pinned bytes.Buffer
	c.Assert(run(t.Context(), []string{"dataset", "pin", "--root", root, "--output", output},
		bytes.NewReader(plan.Bytes()), &pinned), qt.IsNil)
	var written DatasetPinResult
	c.Assert(json.Unmarshal(pinned.Bytes(), &written), qt.IsNil)
	c.Assert(written.Shards, qt.HasLen, 2)
	for _, shard := range written.Shards {
		// #nosec G304 -- The test reads back the file the command just wrote under its temporary directory.
		data, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(shard.Path)))
		c.Assert(err, qt.IsNil)
		c.Assert(sha(data), qt.Equals, shard.SHA256)
	}
	// A second pin refuses to overwrite, and verification accepts the pinned copies.
	c.Assert(run(t.Context(), []string{"dataset", "pin", "--root", root, "--output", output},
		bytes.NewReader(plan.Bytes()), &pinned), qt.IsNotNil)
	var verified bytes.Buffer
	c.Assert(run(t.Context(), []string{"dataset", "verify", "--root", root, "--pinned", output},
		bytes.NewReader(plan.Bytes()), &verified), qt.IsNil)
	var result DatasetVerification
	c.Assert(json.Unmarshal(verified.Bytes(), &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "dataset_plan_reproduced")
	c.Assert(result.Pinned, qt.IsTrue)
	c.Assert(result.Sources, qt.Equals, 8)
	// A tampered plan fails verification; bad arguments fail before any read.
	tampered := bytes.Replace(plan.Bytes(), []byte(`"partition": "development"`), []byte(`"partition": "training"`), 1)
	c.Assert(run(t.Context(), []string{"dataset", "verify", "--root", root}, bytes.NewReader(tampered), &verified), qt.IsNotNil)
	for _, args := range [][]string{{"dataset"}, {"dataset", "plan"}, {"dataset", "pin", "--root", root},
		{"dataset", "plan", "--root", root, "--output", output}, {"dataset", "other", "--root", root}} {
		c.Assert(run(t.Context(), args, bytes.NewReader(dataset), &verified), qt.IsNotNil, qt.Commentf("%v", args))
	}
}
