package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCorpusForestUsesSharedTrainingAndEvaluation(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	corpus := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	base := []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--calibration", "isotonic", "--allow-simulation"}
	for _, row := range []struct {
		name, context string
		features      []string
	}{
		{"prepared", "prepared_piece", []string{"--feature", "prose-words"}},
		{"lexical", "prepared_piece", []string{"--lexical", "--lexical-max-features", "32"}},
		{"rules", "source_document", []string{"--feature", "activation/policy.banned-phrases", "--rule-config", fixture + "rules.yaml"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			args := append(slices.Clone(base), row.features...)
			checkForestTrial(t, binary, corpus, args, row.context)
		})
	}
	for _, flags := range [][]string{
		{"--estimator", "forest", "--l2", "1"},
		{"--estimator", "forest", "--tolerance", "1e-8"},
		{"--estimator", "logistic", "--forest-seed", "1"},
		{"--estimator", "forest", "--max-operations", "1"},
		{"--estimator", "unknown"},
	} {
		args := append(slices.Clone(base), "--feature", "prose-words")
		c.Assert(researchCommand(t, binary, append(args, flags...), corpus, 2), qt.HasLen, 0)
	}
}

func checkForestTrial(t *testing.T, binary string, corpus []byte, base []string, context string) {
	t.Helper()
	c := qt.New(t)
	args := append(slices.Clone(base), "--estimator", "forest", "--forest-trees", "3", "--forest-min-leaf", "1",
		"--forest-bootstrap=false", "--forest-seed", "7")
	fitted := researchCommand(t, binary, args, corpus, 0)
	c.Assert(researchCommand(t, binary, args, corpus, 0), qt.DeepEquals, fitted)
	directory := t.TempDir()
	predictArgs := frozenPredictionFiles(t, directory, corpus, fitted, context)
	if context == "source_document" {
		predictArgs = append(predictArgs, "--rule-config", "../research/annotation/training/testdata/rules.yaml")
	}
	predictions := researchCommand(t, binary, predictArgs, corpus, 0)
	var result struct {
		ProbabilityStatus string `json:"probability_status"`
		Rows              []struct {
			ForestResponse   *float64 `json:"forest_response"`
			LogisticResponse *float64 `json:"logistic_response"`
			LinearScore      *float64 `json:"linear_score"`
		} `json:"rows"`
	}
	c.Assert(json.Unmarshal(predictions, &result), qt.IsNil)
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(result.Rows, qt.HasLen, 1)
	c.Assert(result.Rows[0].ForestResponse, qt.IsNotNil)
	c.Assert(result.Rows[0].LogisticResponse, qt.IsNil)
	c.Assert(result.Rows[0].LinearScore, qt.IsNil)
	logistic := researchCommand(t, binary, base, corpus, 0)
	comparatorArgs := frozenPredictionFiles(t, directory, corpus, logistic, context)
	if context == "source_document" {
		comparatorArgs = append(comparatorArgs, "--rule-config", "../research/annotation/training/testdata/rules.yaml")
	}
	comparator := researchCommand(t, binary, comparatorArgs, corpus, 0)
	c.Assert(os.Remove(filepath.Join(directory, "model.json")), qt.IsNil)
	checkCorpusComparison(t, binary, directory, predictions, comparator)
	evaluate := []string{"evaluate", "--corpus", filepath.Join(directory, "corpus.json"),
		"--round", "../research/annotation/training/testdata/round.json", "--allow-simulation"}
	c.Assert(researchCommand(t, binary, evaluate, predictions, 0), qt.Not(qt.HasLen), 0)
}
