package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCorpusLexicalBaselineCommands(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	candidates := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	base := []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--calibration", "isotonic", "--allow-simulation"}
	args := append(append([]string{}, base...), "--lexical", "--lexical-max-features", "64")
	model := researchCommand(t, binary, args, candidates, 0)
	c.Assert(researchCommand(t, binary, args, candidates, 0), qt.DeepEquals, model)
	checkLexicalCommandRejections(t, binary, args, base, candidates)
	directory := t.TempDir()
	predict := frozenPredictionFiles(t, directory, candidates, model, "prepared_piece")
	predictions := researchCommand(t, binary, predict, candidates, 0)
	c.Assert(researchCommand(t, binary, predict, candidates, 0), qt.DeepEquals, predictions)
	preparedArgs := append(append([]string{}, base...), "--feature", "prose-words")
	comparator := comparisonPrediction(t, binary, directory, predict, candidates, preparedArgs, false)
	c.Assert(os.Remove(filepath.Join(directory, "model.json")), qt.IsNil)
	checkCorpusComparison(t, binary, directory, predictions, comparator)
	evaluate := []string{"evaluate", "--corpus", filepath.Join(directory, "corpus.json"),
		"--round", fixture + "round.json", "--allow-simulation"}
	output := researchCommand(t, binary, evaluate, predictions, 0)
	var result struct {
		Status            string `json:"status"`
		ProbabilityStatus string `json:"probability_status"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "experimental_metrics")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	checkChangedEvaluationLabels(t, binary, directory, evaluate, predictions, output)
}

func checkLexicalCommandRejections(t *testing.T, binary string, args, base []string, candidates []byte) {
	t.Helper()
	c := qt.New(t)
	for _, extra := range [][]string{
		{"--feature", "prose-words"}, {"--rule-config", "missing.yaml"},
		{"--lexical-max-features", "129"}, {"--lexical-word-min", "4"},
	} {
		invalid := append(append([]string{}, args...), extra...)
		c.Assert(researchCommand(t, binary, invalid, candidates, 2), qt.HasLen, 0)
	}
	missing := append(append([]string{}, base...), "--feature", "prose-words", "--lexical-char-min", "2")
	c.Assert(researchCommand(t, binary, missing, candidates, 2), qt.HasLen, 0)
}
