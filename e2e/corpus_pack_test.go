package e2e_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

// The tutorial artifact is simulated, so every pack it produces stays
// experimental. This test asserts the conversion, not a qualified model.
func TestCorpusPackConvertsAFittedArtifact(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	artifact := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	trained := researchCommand(t, binary, []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--feature", "prose-words", "--calibration", "isotonic", "--allow-simulation"}, artifact, 0)

	args := []string{"pack", "--id", "tutorial-pack", "--min-words", "12"}
	output := researchCommand(t, binary, args, trained, 0)
	c.Assert(researchCommand(t, binary, args, trained, 0), qt.DeepEquals, output)
	var pack struct {
		Version        string `json:"version"`
		SHA256         string `json:"sha256"`
		ID             string `json:"id"`
		DeclaredStatus string `json:"declared_status"`
		HumanCorpus    string `json:"human_corpus"`
		Task           string `json:"task"`
		Kind           string `json:"kind"`
		Estimator      string `json:"estimator"`
		Limits         struct {
			MinWords int `json:"min_words"`
		} `json:"limits"`
		Calibration struct {
			Algorithm string `json:"algorithm"`
		} `json:"calibration"`
	}
	c.Assert(json.Unmarshal(output, &pack), qt.IsNil)
	c.Assert(pack.Version, qt.Equals, "unswell-probability-pack-v1")
	c.Assert(pack.ID, qt.Equals, "tutorial-pack")
	c.Assert(pack.DeclaredStatus, qt.Equals, "experimental")
	c.Assert(pack.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(pack.Task, qt.Equals, "editorial_needs_revision")
	c.Assert(pack.Kind, qt.Equals, "paragraph")
	c.Assert(pack.Estimator, qt.Equals, "logistic")
	c.Assert(pack.Limits.MinWords, qt.Equals, 12)
	c.Assert(pack.Calibration.Algorithm, qt.Equals, "isotonic")
	c.Assert(pack.SHA256, qt.HasLen, 64)

	for _, row := range []struct{ name, value string }{
		{"--min-words", "0"}, {"--task", "unknown"}, {"--accepted", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed := slices.Clone(args)
			if row.value == "" {
				changed = append(changed, row.name)
			} else {
				changed = append(changed, row.name, row.value)
			}
			c.Assert(researchCommand(t, binary, changed, trained, 2), qt.HasLen, 0)
		})
	}
	// An artifact fitted for the editorial task cannot become an origin pack.
	c.Assert(researchCommand(t, binary, append(slices.Clone(args), "--task", "origin_endpoint"), trained, 2), qt.HasLen, 0)

	// An origin pack comes from an artifact fitted on provenance labels, and it
	// keeps the same contract in the other channel.
	originManifest, originRoot := originFixture(t)
	originPlan := researchCommand(t, binary, []string{"plan"}, originManifest, 0)
	originArtifact := researchCommand(t, binary, []string{"extract", "--root", originRoot}, originPlan, 0)
	originTrained := researchCommand(t, binary, []string{"train", "--root", originRoot, "--labels", "provenance",
		"--kind", "paragraph", "--feature", "prose-words", "--calibration", "isotonic"}, originArtifact, 0)
	origin := researchCommand(t, binary, append(slices.Clone(args), "--task", "origin_endpoint"), originTrained, 0)
	c.Assert(json.Unmarshal(origin, &pack), qt.IsNil)
	c.Assert(pack.Task, qt.Equals, "origin_endpoint")
	c.Assert(pack.DeclaredStatus, qt.Equals, "experimental")
	c.Assert(pack.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(researchCommand(t, binary, args, originTrained, 2), qt.HasLen, 0)
	checkPackLoadsInTheEngine(t, output, origin)
}

// checkPackLoadsInTheEngine closes the loop: a pack this command produced is
// accepted by the CLI, in the channel its task names and in no other.
func checkPackLoadsInTheEngine(t *testing.T, revision, origin []byte) {
	t.Helper()
	c := qt.New(t)
	cli := buildCLI(t)
	workspace := t.TempDir()
	write := func(name string, data []byte) string {
		path := filepath.Join(workspace, name)
		c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
		return path
	}
	revisionPack := write("revision-pack.json", revision)
	originPack := write("origin-pack.json", origin)
	write("draft.md", []byte("# Notes\n\nThe client retries after a transport failure and records the attempt.\n"))
	policy := write(".unswell.yaml", []byte("version: 1\nextends: [builtin:strict-v1]\n"+
		"calibration:\n  model: pack\n  accept_experimental: true\n"+
		"origin:\n  model: pack\n  accept_experimental: true\n"))
	report := filepath.Join(workspace, "result.json")
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	// #nosec G204 -- Fixed local binary and test-owned temporary paths.
	command := exec.CommandContext(ctx, cli, "check", ".", "--config", policy,
		"--model", revisionPack, "--origin-model", originPack, "--report", "json:"+report)
	command.Dir = workspace
	output, err := command.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("check: %s", output))
	var result struct {
		Manifest struct {
			Probability *struct{ PackID, Task string } `json:"probability"`
			Origin      *struct{ PackID, Task string } `json:"origin"`
		} `json:"manifest"`
		Assessments []struct {
			Scope             string   `json:"scope"`
			SlopProbability   *float64 `json:"slop_probability"`
			ProbabilityStatus string   `json:"probability_status"`
			OriginStatus      string   `json:"origin_status"`
		} `json:"assessments"`
	}
	// #nosec G304 -- The report path is built from this test's temporary directory.
	saved, err := os.ReadFile(report)
	c.Assert(err, qt.IsNil)
	c.Assert(json.Unmarshal(saved, &result), qt.IsNil)
	c.Assert(result.Manifest.Probability, qt.IsNotNil)
	c.Assert(result.Manifest.Probability.Task, qt.Equals, "editorial_needs_revision")
	c.Assert(result.Manifest.Origin, qt.IsNotNil)
	c.Assert(result.Manifest.Origin.Task, qt.Equals, "origin_endpoint")
	statuses := map[string]int{}
	for _, assessment := range result.Assessments {
		statuses[assessment.ProbabilityStatus]++
		if assessment.Scope == "paragraph" {
			c.Assert(assessment.OriginStatus, qt.Not(qt.Equals), "")
		}
	}
	c.Assert(len(statuses) > 0, qt.IsTrue)
}
