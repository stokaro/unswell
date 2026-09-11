package e2e_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

// originFixture writes a corpus of generated and historical sources whose
// groups fall in every partition, so provenance labels give a fit both
// classes in training and rows to calibrate on. It returns the manifest bytes
// and the source root. The manifest is spelled as JSON here because the
// research module is not imported by these tests.
func originFixture(t *testing.T) ([]byte, string) {
	t.Helper()
	c := qt.New(t)
	root := t.TempDir()
	write := func(name string, data []byte) {
		c.Assert(os.WriteFile(filepath.Join(root, name), data, 0o600), qt.IsNil)
	}
	hash := func(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
	license := []byte("Test-owned source and notice fixture.\n")
	write("LICENSE", license)
	notice := map[string]any{"path": "LICENSE", "sha256": hash(license), "bytes": len(license)}
	texts := []string{
		"The cache retries a request after a short wait and records the delay for the caller.",
		"A retry keeps the same identifier, so the log of one request stays readable across attempts.",
		"Callers see one answer per request, and the cache never returns a partial entry.",
		"When the wait ends without an answer, the cache reports the failure and keeps the entry empty.",
		"Each entry carries the time of its last write, which the reader compares before use.",
		"An expired entry is read once more only when the writer has confirmed the new value.",
		"The reader holds a lock for the length of one comparison and releases it before the copy.",
		"A writer that fails midway leaves the previous value in place and reports the error.",
		"Configuration names the wait in milliseconds, and the default is short on purpose.",
		"Long waits hide slow writers, so the log records every wait above the default.",
		"The cache stores at most one value per key and rejects a second writer for the same key.",
		"Rejected writers return at once with the key and the time of the conflicting write.",
	}
	sources := make([]map[string]any, 0, len(texts))
	for i, text := range texts {
		name := fmt.Sprintf("doc%02d.md", i)
		data := []byte("# Cache\n\n" + text + "\n\nThis paragraph explains the behavior of the cache in plain words for readers.\n")
		write(name, data)
		origin := map[string]any{"label": "unknown", "scope": "repository",
			"evidence": "Dated snapshot; no unit-level authorship record.", "generation_record": ""}
		snapshot := map[string]any{"date": "2012-12-28", "confidence": "vcs_only",
			"evidence": "commit dated 2012-12-28", "cohort": "historical-2012"}
		if i%2 == 0 {
			origin = map[string]any{"label": "generated", "scope": "document",
				"evidence": fmt.Sprintf("Response r%d of run pilot", i), "generation_record": fmt.Sprintf("records.json#r%d", i)}
			snapshot = map[string]any{"date": "2026-09-11", "confidence": "corroborated",
				"evidence": "Generation record dates the response", "cohort": "controlled"}
		}
		sources = append(sources, map[string]any{
			"id": fmt.Sprintf("s%02d", i), "path": name, "sha256": hash(data), "bytes": len(data), "format": "markdown",
			"prose_language": "en", "repository": fmt.Sprintf("project-%02d", i),
			"document": fmt.Sprintf("project-%02d/%s", i, name), "reference": "fixture:" + name,
			"topic": "cache", "purpose": "Explain cache behavior", "role": "documentation", "origin": origin,
			"rights": map[string]any{"license": "test-fixture", "evidence": "Test-owned bytes",
				"allowed_uses": []string{"annotation", "training", "evaluation"}},
			"notices": []map[string]any{notice}, "snapshot": snapshot,
			"authors": nil, "templates": nil, "related": nil, "generation_tasks": nil,
			"author_language": "", "author_language_basis": "", "role_regions": nil, "partition": "",
		})
	}
	manifest := map[string]any{"version": "unswell-corpus-v1", "id": "origin-fixture", "seed": "origin-seed",
		"weights":           map[string]int{"training": 6000, "development": 1500, "calibration": 1500, "final_test": 1000},
		"extraction_policy": map[string]any{},
		"unit_kinds":        []string{"paragraph"}, "sources": sources}
	encoded, err := json.Marshal(manifest)
	c.Assert(err, qt.IsNil)
	return encoded, root
}
