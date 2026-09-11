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

func hashBytes(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// originFixture builds a corpus of generated and historical sources whose
// groups fall in every partition, so a fit has both classes in training and
// rows to calibrate on.
func originFixture(c *qt.C) (corpus.Artifact, map[string][]byte) {
	c.Helper()
	files := map[string][]byte{"LICENSE": []byte("Test-owned source and notice fixture.\n")}
	notice := corpus.Notice{Path: "LICENSE", SHA256: hashBytes(files["LICENSE"]), Bytes: len(files["LICENSE"])}
	manifest := corpus.Manifest{Version: corpus.Version, ID: "origin-fixture", Seed: "origin-seed",
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
		generated := i%2 == 0
		name := fmt.Sprintf("doc%02d.md", i)
		files[name] = []byte("# Cache\n\n" + text + "\n\nThis paragraph explains the behavior of the cache in plain words for readers.\n")
		origin := annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Dated snapshot; no unit-level authorship record."}
		snapshot := corpus.Snapshot{Date: "2012-12-28", Confidence: "vcs_only", Evidence: "commit dated 2012-12-28", Cohort: "historical-2012"}
		if generated {
			origin = annotation.Origin{Label: "generated", Scope: "document", Evidence: fmt.Sprintf("Response r%d of run pilot", i),
				GenerationRecord: fmt.Sprintf("records.json#r%d", i)}
			snapshot = corpus.Snapshot{Date: "2026-09-11", Confidence: "corroborated",
				Evidence: "Generation record dates the response", Cohort: "controlled"}
		}
		manifest.Sources = append(manifest.Sources, corpus.Source{ID: fmt.Sprintf("s%02d", i), Path: name,
			SHA256: hashBytes(files[name]), Bytes: len(files[name]), Format: document.Markdown, ProseLanguage: "en",
			Repository: fmt.Sprintf("project-%02d", i), Document: fmt.Sprintf("project-%02d/%s", i, name),
			Reference: "fixture:" + name, Topic: "cache", Purpose: "Explain cache behavior", Role: "documentation",
			Origin: origin, Rights: annotation.Rights{License: "test-fixture", Evidence: "Test-owned bytes",
				AllowedUses: []string{"annotation", "training", "evaluation"}},
			Notices: []corpus.Notice{notice}, Snapshot: &snapshot})
	}
	plan, err := corpus.MakePlan(c.Context(), manifest)
	c.Assert(err, qt.IsNil)
	artifact, err := corpus.Build(c.Context(), plan, files)
	c.Assert(err, qt.IsNil)
	return artifact, files
}

func TestRunDecisionsFitsTheOriginTask(t *testing.T) {
	c := qt.New(t)
	artifact, files := originFixture(c)
	decisions, err := corpus.OriginDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	options := fittingOptions()
	options.AllowSimulation = false

	fitted, err := training.RunDecisions(c.Context(), artifact, decisions, files, options)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Basis, qt.Equals, "declared_provenance")
	c.Assert(fitted.Identity.Task, qt.Equals, annotation.TaskOrigin)
	c.Assert(fitted.Identity.Rubric, qt.Equals, annotation.OriginRubric)
	c.Assert(fitted.RoundSHA256, qt.Equals, artifact.SHA256)
	c.Assert(fitted.HumanCorpus, qt.Equals, "not_qualified")
	fittedTraining := fitted.Partitions[0]
	c.Assert(fittedTraining.Classes["endpoint_generated"] > 0, qt.IsTrue)
	c.Assert(fittedTraining.Classes["human_snapshot"] > 0, qt.IsTrue)
	_, hasEditorial := fittedTraining.Classes["needs_revision"]
	c.Assert(hasEditorial, qt.IsFalse)

	encoded, err := json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	restored, err := training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)
	c.Assert(restored.Identity.Task, qt.Equals, annotation.TaskOrigin)

	// A restored artifact whose task disagrees with its rubric is refused.
	mismatched := fitted
	mismatched.Identity.Task = annotation.TaskEditorial
	mismatched.SHA256 = ""
	encoded, err = json.Marshal(mismatched)
	c.Assert(err, qt.IsNil)
	_, err = training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNotNil)

	pack, err := training.BuildPack(c.Context(), fitted, training.PackOptions{ID: "origin-pack", MinWords: 8, Task: probability.TaskOrigin})
	c.Assert(err, qt.IsNil)
	c.Assert(pack.Task, qt.Equals, probability.TaskOrigin)
	_, err = training.BuildPack(c.Context(), fitted, training.PackOptions{ID: "origin-pack", MinWords: 8, Task: probability.Task})
	c.Assert(err, qt.ErrorMatches, "the artifact was fitted for origin_endpoint, not editorial_needs_revision")
}

func TestRunDecisionsRefusesForeignDecisions(t *testing.T) {
	c := qt.New(t)
	artifact, files := originFixture(c)
	decisions, err := corpus.OriginDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	decisions.RoundSHA256 = "0"
	_, err = training.RunDecisions(c.Context(), artifact, decisions, files, fittingOptions())
	c.Assert(err, qt.ErrorMatches, "decisions bind a different candidate artifact")
}
