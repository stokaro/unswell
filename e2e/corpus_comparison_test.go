package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
)

func comparisonPrediction(t *testing.T, binary, directory string, args []string, corpus []byte, trainArgs []string, rules bool) []byte {
	t.Helper()
	c := qt.New(t)
	var plan map[string]any
	decodeFile(t, filepath.Join(directory, "plan.json"), &plan)
	plan["id"], plan["threshold"] = "scripted-comparator-v1", 1.0
	if !rules {
		modelPath := filepath.Join(directory, "comparator-model.json")
		data := researchCommand(t, binary, append(trainArgs, "--feature", "noun-token-ratio"), corpus, 0)
		var fitted struct {
			SHA256 string `json:"sha256"`
		}
		c.Assert(json.Unmarshal(data, &fitted), qt.IsNil)
		plan["model_sha256"] = fitted.SHA256
		c.Assert(os.WriteFile(modelPath, data, 0o600), qt.IsNil)
		defer func() { c.Assert(os.Remove(modelPath), qt.IsNil) }()
		args = append(args, "--model", modelPath)
	}
	data, err := json.Marshal(plan)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(directory, "comparator-plan.json")
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	return researchCommand(t, binary, append(args, "--plan", path), corpus, 0)
}

func checkCorpusComparison(t *testing.T, binary, directory string, predictions, comparator []byte) {
	t.Helper()
	c := qt.New(t)
	var left, right struct {
		SHA256 string `json:"sha256"`
		Plan   struct {
			ProtocolSHA256 string `json:"protocol_sha256"`
		} `json:"plan"`
	}
	c.Assert(json.Unmarshal(predictions, &left), qt.IsNil)
	c.Assert(json.Unmarshal(comparator, &right), qt.IsNil)
	plan := map[string]any{"version": "unswell-research-comparison-v1", "id": "scripted-paired-e2e-v1",
		"protocol_sha256": left.Plan.ProtocolSHA256, "candidate_predictions_sha256": left.SHA256,
		"comparator_predictions_sha256": right.SHA256}
	data, err := json.Marshal(plan)
	c.Assert(err, qt.IsNil)
	planPath, comparatorPath := filepath.Join(directory, "comparison.json"), filepath.Join(directory, "comparator.json")
	c.Assert(os.WriteFile(planPath, data, 0o600), qt.IsNil)
	c.Assert(os.WriteFile(comparatorPath, comparator, 0o600), qt.IsNil)
	args := []string{"compare", "--plan", planPath, "--protocol", filepath.Join(directory, "protocol.md"),
		"--corpus", filepath.Join(directory, "corpus.json"), "--comparator", comparatorPath,
		"--round", "../research/annotation/training/testdata/round.json"}
	c.Assert(researchCommand(t, binary, args, predictions, 2), qt.HasLen, 0)
	args = append(args, "--allow-simulation")
	output := researchCommand(t, binary, args, predictions, 0)
	c.Assert(researchCommand(t, binary, args, predictions, 0), qt.DeepEquals, output)
	checkComparisonOutput(t, output)
	c.Assert(researchCommand(t, binary, append(args, "--root", directory), predictions, 2), qt.HasLen, 0)
	c.Assert(researchCommand(t, binary, append(args, "--model", "missing.json"), predictions, 2), qt.HasLen, 0)
	c.Assert(os.WriteFile(comparatorPath, predictions, 0o600), qt.IsNil)
	c.Assert(researchCommand(t, binary, args, predictions, 2), qt.HasLen, 0)
	c.Assert(os.WriteFile(comparatorPath, comparator, 0o600), qt.IsNil)
	checkComparisonLabelSeparation(t, binary, args, directory, predictions)
}

func checkComparisonOutput(t *testing.T, data []byte) {
	t.Helper()
	c := qt.New(t)
	var result struct {
		Status            string `json:"status"`
		Basis             string `json:"basis"`
		ProbabilityStatus string `json:"probability_status"`
		Summary           struct {
			Replicates    int `json:"replicates"`
			CommonCovered struct {
				Metrics []struct {
					ID    string `json:"id"`
					Micro struct {
						Difference struct {
							Interval struct {
								Status string   `json:"status"`
								Upper  *float64 `json:"upper"`
							} `json:"interval"`
						} `json:"difference"`
					} `json:"micro"`
				} `json:"metrics"`
			} `json:"common_covered"`
		} `json:"summary"`
	}
	c.Assert(json.Unmarshal(data, &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "experimental_metrics")
	c.Assert(result.Basis, qt.Equals, "simulation")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(result.Summary.Replicates, qt.Equals, 10000)
	c.Assert(result.Summary.CommonCovered.Metrics, qt.HasLen, 9)
	for _, metric := range result.Summary.CommonCovered.Metrics {
		c.Assert(metric.Micro.Difference.Interval.Status, qt.Equals, "insufficient_evidence")
		c.Assert(metric.Micro.Difference.Interval.Upper, qt.IsNil)
	}
}

func checkComparisonLabelSeparation(t *testing.T, binary string, args []string, directory string, predictions []byte) {
	t.Helper()
	c := qt.New(t)
	var round map[string]any
	decodeFile(t, "../research/annotation/training/testdata/round.json", &round)
	round["judgments"] = []any{}
	data, err := json.Marshal(round)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(directory, "empty-judgments.json")
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	output := researchCommand(t, binary, append(args, "--round", path), predictions, 0)
	var result struct {
		Excluded map[string]int `json:"excluded_labels"`
		Summary  struct {
			FullFlow struct {
				Candidate struct {
					Micro struct {
						Coverage *float64 `json:"coverage"`
					} `json:"micro"`
				} `json:"candidate"`
			} `json:"full_flow"`
		} `json:"summary"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Excluded, qt.DeepEquals, map[string]int{"decision/missing_judgments": 1})
	c.Assert(result.Summary.FullFlow.Candidate.Micro.Coverage, qt.IsNil)
}
