package e2e_test

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCorpusTrainingUsesSeparateFrozenPartitions(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	artifact := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	args := []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--feature", "prose-words", "--calibration", "isotonic"}
	c.Assert(researchCommand(t, binary, args, artifact, 2), qt.HasLen, 0)
	args = append(args, "--allow-simulation")
	output := researchCommand(t, binary, args, artifact, 0)
	checkTrainingResult(t, output)
	c.Assert(researchCommand(t, binary, args, artifact, 0), qt.DeepEquals, output)
	for _, row := range []struct{ name, value string }{
		{"--kind", "sentence"}, {"--feature", "activation/syntax.noun-stack"},
		{"--max-operations", "1"}, {"--calibration", "unknown"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			changed := append(slices.Clone(args), row.name, row.value)
			c.Assert(researchCommand(t, binary, changed, artifact, 2), qt.HasLen, 0)
		})
	}
}

func checkTrainingResult(t *testing.T, output []byte) {
	t.Helper()
	checkFittedResult(t, output, 6, 3)
}

func checkFittedResult(t *testing.T, output []byte, mean, scale float64) {
	t.Helper()
	c := qt.New(t)
	var result struct {
		Status            string                            `json:"status"`
		Basis             string                            `json:"basis"`
		HumanCorpus       string                            `json:"human_corpus"`
		ProbabilityStatus string                            `json:"probability_status"`
		Logistic          struct{ Means, Scales []float64 } `json:"logistic"`
		Calibration       struct {
			Samples   int       `json:"samples"`
			Responses []float64 `json:"responses"`
		} `json:"calibration"`
		Partitions []struct {
			Name     string            `json:"name"`
			Rows     []json.RawMessage `json:"rows"`
			Classes  map[string]int    `json:"classes"`
			Excluded map[string]int    `json:"excluded"`
		} `json:"partitions"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "experimental_numerical_fit")
	c.Assert(result.Basis, qt.Equals, "simulation")
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.ProbabilityStatus, qt.Equals, "unavailable_unqualified_model")
	c.Assert(result.Logistic.Means, qt.DeepEquals, []float64{mean})
	c.Assert(result.Logistic.Scales, qt.HasLen, 1)
	c.Assert(math.Abs(result.Logistic.Scales[0]-scale) < 1e-12, qt.IsTrue)
	c.Assert(result.Calibration.Samples, qt.Equals, 2)
	c.Assert(result.Calibration.Responses, qt.DeepEquals, []float64{0, 1})
	c.Assert(result.Partitions, qt.HasLen, 4)
	for _, partition := range result.Partitions {
		if partition.Name == "training" || partition.Name == "calibration" {
			c.Assert(partition.Rows, qt.HasLen, 2)
			c.Assert(partition.Classes, qt.DeepEquals, map[string]int{"acceptable": 1, "needs_revision": 1})
		} else {
			c.Assert(partition.Rows, qt.HasLen, 0)
			c.Assert(partition.Classes, qt.HasLen, 0)
			c.Assert(partition.Excluded["reserved_partition"], qt.Equals, 2)
		}
	}
	c.Assert(string(output), qt.Not(qt.Contains), "It is important to note")
}
