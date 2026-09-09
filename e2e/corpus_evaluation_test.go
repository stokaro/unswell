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

func TestCorpusFrozenPredictionAndEvaluation(t *testing.T) {
	binary := buildResearchTool(t, "corpus")
	for _, row := range []struct {
		name, feature, context string
		rules                  bool
		positive               bool
	}{
		{"prepared", "prose-words", "prepared_piece", false, true},
		{"rule activations", "activation/policy.banned-phrases", "source_document", true, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			fixture := "../research/annotation/training/testdata/"
			manifest, err := os.ReadFile(fixture + "manifest.json")
			c.Assert(err, qt.IsNil)
			plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
			corpus := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
			args := []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
				"--kind", "paragraph", "--feature", row.feature, "--calibration", "isotonic", "--allow-simulation"}
			if row.rules {
				args = append(args, "--rule-config", fixture+"rules.yaml")
			}
			model := researchCommand(t, binary, args, corpus, 0)
			directory := t.TempDir()
			predictArgs := frozenPredictionFiles(t, directory, corpus, model, row.context)
			if row.rules {
				predictArgs = append(predictArgs, "--rule-config", fixture+"rules.yaml")
			}
			predictions := researchCommand(t, binary, predictArgs, corpus, 0)
			c.Assert(researchCommand(t, binary, predictArgs, corpus, 0), qt.DeepEquals, predictions)
			c.Assert(researchCommand(t, binary, append(predictArgs, "--round", fixture+"round.json"), corpus, 2), qt.HasLen, 0)
			comparator := comparisonPrediction(t, binary, directory, predictArgs, corpus, args, row.rules)
			c.Assert(os.Remove(filepath.Join(directory, "model.json")), qt.IsNil)
			checkCorpusComparison(t, binary, directory, predictions, comparator)
			evaluateArgs := []string{"evaluate", "--corpus", filepath.Join(directory, "corpus.json"), "--round", fixture + "round.json"}
			c.Assert(researchCommand(t, binary, evaluateArgs, predictions, 2), qt.HasLen, 0)
			evaluateArgs = append(evaluateArgs, "--allow-simulation")
			output := researchCommand(t, binary, evaluateArgs, predictions, 0)
			checkEvaluation(t, output, row.positive)
			c.Assert(researchCommand(t, binary, evaluateArgs, predictions, 0), qt.DeepEquals, output)
			checkChangedEvaluationLabels(t, binary, directory, evaluateArgs, predictions, output)
		})
	}
}

func frozenPredictionFiles(t *testing.T, directory string, corpus, model []byte, context string) []string {
	t.Helper()
	c := qt.New(t)
	var artifact struct {
		SHA256 string `json:"sha256"`
	}
	var fitted struct {
		SHA256 string `json:"sha256"`
	}
	c.Assert(json.Unmarshal(corpus, &artifact), qt.IsNil)
	c.Assert(json.Unmarshal(model, &fitted), qt.IsNil)
	protocol := []byte("Scripted test of frozen prediction and independent evaluation. No scientific qualification.\n")
	plan := map[string]any{"version": "unswell-research-predictions-v1", "id": "scripted-e2e-v1",
		"protocol_sha256": fmt.Sprintf("%x", sha256.Sum256(protocol)), "model_sha256": fitted.SHA256,
		"corpus_sha256": artifact.SHA256, "partition": "final_test", "context": context,
		"response": "logistic", "threshold": 0.5}
	encoded, err := json.Marshal(plan)
	c.Assert(err, qt.IsNil)
	for name, data := range map[string][]byte{"plan.json": encoded, "corpus.json": corpus, "model.json": model, "protocol.md": protocol} {
		c.Assert(os.WriteFile(filepath.Join(directory, name), data, 0o600), qt.IsNil)
	}
	return []string{"predict", "--root", "../research/annotation/training/testdata/sources",
		"--model", filepath.Join(directory, "model.json"), "--plan", filepath.Join(directory, "plan.json"),
		"--protocol", filepath.Join(directory, "protocol.md")}
}

func checkEvaluation(t *testing.T, output []byte, positive bool) {
	t.Helper()
	c := qt.New(t)
	var result struct {
		Status            string  `json:"status"`
		Basis             string  `json:"basis"`
		ProbabilityStatus string  `json:"probability_status"`
		Candidates        int     `json:"candidates"`
		TrainingConstant  float64 `json:"training_constant"`
		Summary           struct {
			Micro struct {
				Counts struct {
					Eligible int `json:"eligible"`
					Covered  int `json:"covered"`
					TP       int `json:"true_positive"`
					FN       int `json:"false_negative"`
				} `json:"counts"`
				FPR   *float64 `json:"false_positive_rate"`
				Brier *float64 `json:"brier"`
			} `json:"micro"`
		} `json:"summary"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "experimental_metrics")
	c.Assert(result.Basis, qt.Equals, "simulation")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(result.Candidates, qt.Equals, 1)
	c.Assert(result.TrainingConstant, qt.Equals, 0.5)
	c.Assert(result.Summary.Micro.Counts.Eligible, qt.Equals, 1)
	c.Assert(result.Summary.Micro.Counts.Covered, qt.Equals, 1)
	c.Assert(result.Summary.Micro.Counts.TP == 1, qt.Equals, positive)
	c.Assert(result.Summary.Micro.Counts.FN == 1, qt.Equals, !positive)
	c.Assert(result.Summary.Micro.Brier, qt.IsNotNil)
	c.Assert(result.Summary.Micro.FPR, qt.IsNil)
}

func checkChangedEvaluationLabels(t *testing.T, binary, directory string, args []string, predictions, before []byte) {
	t.Helper()
	c := qt.New(t)
	data, err := os.ReadFile("../research/annotation/training/testdata/round.json")
	c.Assert(err, qt.IsNil)
	var round map[string]any
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	judgments, ok := round["judgments"].([]any)
	c.Assert(ok, qt.IsTrue)
	for _, value := range judgments {
		judgment, ok := value.(map[string]any)
		c.Assert(ok, qt.IsTrue)
		if judgment["unit_id"] == "u000012" {
			judgment["label"], judgment["categories"] = "acceptable", []any{}
		}
	}
	encoded, err := json.Marshal(round)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(directory, "changed-labels.json")
	c.Assert(os.WriteFile(path, encoded, 0o600), qt.IsNil)
	after := researchCommand(t, binary, append(args, "--round", path), predictions, 0)
	c.Assert(after, qt.Not(qt.DeepEquals), before)
	checkMissingEvaluationLabels(t, binary, args, predictions, round, path)
}

func checkMissingEvaluationLabels(t *testing.T, binary string, args []string, predictions []byte, round map[string]any, path string) {
	t.Helper()
	c := qt.New(t)
	judgments, ok := round["judgments"].([]any)
	c.Assert(ok, qt.IsTrue)
	kept := []any{}
	for _, value := range judgments {
		judgment, ok := value.(map[string]any)
		c.Assert(ok, qt.IsTrue)
		if judgment["unit_id"] != "u000012" {
			kept = append(kept, value)
		}
	}
	round["judgments"] = kept
	encoded, err := json.Marshal(round)
	c.Assert(err, qt.IsNil)
	c.Assert(os.WriteFile(path, encoded, 0o600), qt.IsNil)
	output := researchCommand(t, binary, append(args, "--round", path), predictions, 0)
	var result struct {
		Excluded map[string]int `json:"excluded_labels"`
		Summary  struct {
			Micro struct {
				Brier    *float64 `json:"brier"`
				Coverage *float64 `json:"coverage"`
			} `json:"micro"`
		} `json:"summary"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Excluded, qt.DeepEquals, map[string]int{"decision/missing_judgments": 1})
	c.Assert(result.Summary.Micro.Brier, qt.IsNil)
	c.Assert(result.Summary.Micro.Coverage, qt.IsNil)
}
