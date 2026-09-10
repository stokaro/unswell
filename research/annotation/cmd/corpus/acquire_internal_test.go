package main

// White-box tests: The acquire command walks a checkout through the process root and skips
// version-control metadata; package main exposes no importable API for that walk.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
)

func TestAcquireCommandBuildsAShardFromACheckout(t *testing.T) {
	c := qt.New(t)
	checkout := t.TempDir()
	for name, content := range map[string]string{
		"LICENSE":          "MIT License\n",
		"README.md":        "# Fixture\n\nThe cache retries a failed lookup after the delay expires.\n",
		"docs/guide.md":    "# Guide\n\nThe cache retries a failed lookup after the delay expires.\n",
		"main.go":          "// Package main starts the cache.\npackage main\n",
		".git/HEAD":        "ref: refs/heads/main\n",
		"vendor/v/v.go":    "// vendored\npackage v\n",
		"docs/ja/guide.md": "# ガイド\n",
	} {
		full := filepath.Join(checkout, filepath.FromSlash(name))
		c.Assert(os.MkdirAll(filepath.Dir(full), 0o750), qt.IsNil)
		c.Assert(os.WriteFile(full, []byte(content), 0o600), qt.IsNil)
	}
	record := corpus.Acquisition{Version: corpus.AcquisitionVersion,
		Manifest: corpus.AcquisitionHeader{ID: "shard-fixture", Seed: "unswell-research-v1",
			Weights: corpus.Weights{Training: 5000, Development: 1500, Calibration: 1500, FinalTest: 2000},
			Policy:  extract.Policy{}, UnitKinds: []string{"sentence", "paragraph", "fragment"}},
		Repository: corpus.RepositoryRecord{Name: "example/fixture", Reference: "https://example.invalid/fixture/blob/abc",
			Commit: "abc", Topic: "fixtures", Purpose: "Exercise the command", Ecosystem: "go",
			Origin:   annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Historical snapshot."},
			Rights:   annotation.Rights{License: "MIT", Evidence: "LICENSE", AllowedUses: []string{"annotation"}},
			Notices:  []string{"LICENSE"},
			Snapshot: corpus.Snapshot{Date: "2019-06-30", Confidence: "corroborated", Evidence: "Registry", Cohort: "historical"}},
		Selection: corpus.SelectionRules{MaxSources: 100, ShardSources: 100, ShardBytes: 1 << 20, MaxSourceBytes: 4096,
			DocumentRoots:    []string{"docs"},
			SourceExtensions: []string{".go"}, ExcludedSegments: []string{"vendor"}, TranslationHints: []string{"ja"}}}
	encoded, err := json.Marshal(record)
	c.Assert(err, qt.IsNil)
	recordPath := filepath.Join(t.TempDir(), "record.json")
	c.Assert(os.WriteFile(recordPath, encoded, 0o600), qt.IsNil)
	var output bytes.Buffer
	c.Assert(run(t.Context(), []string{"acquire", "--root", checkout, "--record", recordPath}, nil, &output), qt.IsNil)
	var result corpus.AcquisitionResult
	c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
	c.Assert(result.Manifests, qt.HasLen, 1)
	paths := []string{}
	for _, source := range result.Manifests[0].Sources {
		paths = append(paths, source.Path)
	}
	c.Assert(paths, qt.DeepEquals, []string{"README.md", "docs/guide.md", "main.go"})
	reasons := map[string]string{}
	for _, item := range result.Excluded {
		reasons[item.Path] = item.Reason
	}
	c.Assert(reasons["vendor"], qt.Equals, "excluded_segment_directory")
	_, entered := reasons["vendor/v/v.go"]
	c.Assert(entered, qt.IsFalse)
	c.Assert(reasons["docs/ja/guide.md"], qt.Equals, "translation_hint")
	_, walked := reasons[".git/HEAD"]
	c.Assert(walked, qt.IsFalse)
	// The manifest the command wrote plans and extracts with the ordinary commands.
	manifest, err := json.Marshal(result.Manifests[0])
	c.Assert(err, qt.IsNil)
	var plan, artifact bytes.Buffer
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(manifest), &plan), qt.IsNil)
	c.Assert(run(t.Context(), []string{"extract", "--root", checkout}, &plan, &artifact), qt.IsNil)
	for _, args := range [][]string{{"acquire"}, {"acquire", "--root", checkout}, {"acquire", "--record", recordPath},
		{"acquire", "--root", checkout, "--record", recordPath, "extra"}} {
		c.Assert(run(t.Context(), args, nil, &output), qt.IsNotNil, qt.Commentf("%v", args))
	}
}
