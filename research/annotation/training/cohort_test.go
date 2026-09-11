package training_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/probability"
	"github.com/stokaro/unswell/research/annotation"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/training"
)

// cohortFixture builds a corpus of contemporary and historical sources whose
// groups fall in every partition.
func cohortFixture(c *qt.C) (corpus.Artifact, map[string][]byte) {
	c.Helper()
	files := map[string][]byte{"LICENSE": []byte("Test-owned source and notice fixture.\n")}
	digest := func(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
	notice := corpus.Notice{Path: "LICENSE", SHA256: digest(files["LICENSE"]), Bytes: len(files["LICENSE"])}
	manifest := corpus.Manifest{Version: corpus.Version, ID: "cohort-fixture", Seed: "cohort-seed",
		Weights: corpus.Weights{Training: 6000, Development: 1500, Calibration: 1500, FinalTest: 1000},
		Policy:  extract.Policy{}, UnitKinds: []string{"paragraph"}}
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
	for i, text := range texts {
		name := fmt.Sprintf("doc%02d.md", i)
		files[name] = []byte("# Cache\n\n" + text + "\n\nThis paragraph explains the behavior of the cache in plain words for readers.\n")
		snapshot := corpus.Snapshot{Date: "2012-12-28", Confidence: "vcs_only", Evidence: "commit dated 2012-12-28", Cohort: "historical-2012"}
		if i%2 == 0 {
			snapshot = corpus.Snapshot{Date: "2026-06-19", Confidence: "corroborated", Evidence: "release dated 2026-06-19", Cohort: "contemporary"}
		}
		manifest.Sources = append(manifest.Sources, corpus.Source{ID: fmt.Sprintf("s%02d", i), Path: name,
			SHA256: digest(files[name]), Bytes: len(files[name]), Format: document.Markdown, ProseLanguage: "en",
			Repository: fmt.Sprintf("project-%02d", i), Document: fmt.Sprintf("project-%02d/%s", i, name),
			Reference: "fixture:" + name, Topic: "cache", Purpose: "Explain cache behavior", Role: "documentation",
			Origin: annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Dated snapshot; no unit-level authorship record."},
			Rights: annotation.Rights{License: "test-fixture", Evidence: "Test-owned bytes",
				AllowedUses: []string{"annotation", "training", "evaluation"}},
			Notices: []corpus.Notice{notice}, Snapshot: &snapshot})
	}
	plan, err := corpus.MakePlan(c.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(c.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	return artifact, files
}

// The cohort task fits and restores like the origin task, its classes carry
// the cohort labels, and its artifact never becomes a pack.
func TestRunDecisionsFitsTheCohortTask(t *testing.T) {
	c := qt.New(t)
	artifact, files := cohortFixture(c)
	decisions, err := corpus.CohortDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	options := fittingOptions()
	options.AllowSimulation = false

	fitted, err := training.RunDecisions(c.Context(), artifact, decisions, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Identity.Task, qt.Equals, annotation.TaskCohort)
	c.Assert(fitted.Identity.Rubric, qt.Equals, annotation.CohortRubric)
	c.Assert(fitted.Partitions[0].Classes["contemporary_snapshot"] > 0, qt.IsTrue)
	c.Assert(fitted.Partitions[0].Classes["historical_snapshot"] > 0, qt.IsTrue)

	encoded, err := json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	restored, err := training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(restored.Identity.Task, qt.Equals, annotation.TaskCohort)

	for _, task := range []string{probability.Task, probability.TaskOrigin} {
		_, err = training.BuildPack(c.Context(), fitted, training.PackOptions{ID: "cohort-pack", MinWords: 8, Task: task})
		c.Assert(err, qt.ErrorMatches, "the artifact was fitted for cohort_membership, not .*")
	}
}
