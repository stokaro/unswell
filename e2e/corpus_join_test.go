package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

func TestCorpusJoinsExactFeaturesWithoutInventingLabels(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	manifest, err := os.ReadFile("../research/annotation/corpus/testdata/ptah-manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	root := "../research/annotation/corpus/testdata/ptah"
	artifact := researchCommand(t, binary, []string{"extract", "--root", root}, plan, 0)
	round := corpusJoinRound(t, artifact)
	path := filepath.Join(t.TempDir(), "round.json")
	c.Assert(os.WriteFile(path, round, 0o600), qt.IsNil)
	args := []string{"join", "--root", root, "--round", path, "--feature", "prose-words"}
	output := researchCommand(t, binary, args, artifact, 0)
	checkJoinedCorpus(t, output)
	var tampered map[string]any
	c.Assert(json.Unmarshal(round, &tampered), qt.IsNil)
	unit := tampered["units"].([]any)[0].(map[string]any)
	unit["source"].(map[string]any)["repository_id"] = "different-repository"
	changed, err := json.Marshal(tampered)
	c.Assert(err, qt.IsNil)
	c.Assert(os.WriteFile(path, changed, 0o600), qt.IsNil)
	c.Assert(researchCommand(t, binary, args, artifact, 2), qt.HasLen, 0)
}

func corpusJoinRound(t *testing.T, artifact []byte) []byte {
	t.Helper()
	c := qt.New(t)
	var candidates struct {
		Units []struct {
			Unit json.RawMessage `json:"unit"`
		} `json:"units"`
	}
	c.Assert(json.Unmarshal(artifact, &candidates), qt.IsNil)
	c.Assert(candidates.Units, qt.HasLen, 378)
	data, err := os.ReadFile("../research/annotation/testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	var round map[string]json.RawMessage
	c.Assert(json.Unmarshal(data, &round), qt.IsNil)
	var units []json.RawMessage
	for _, candidate := range candidates.Units[:5] {
		units = append(units, candidate.Unit)
	}
	round["units"], err = json.Marshal(units)
	c.Assert(err, qt.IsNil)
	round["judgments"], round["adjudications"] = json.RawMessage(`[]`), json.RawMessage(`[]`)
	encoded, err := json.Marshal(round)
	c.Assert(err, qt.IsNil)
	return encoded
}

func checkJoinedCorpus(t *testing.T, output []byte) {
	t.Helper()
	c := qt.New(t)
	var result struct {
		Status      string `json:"status"`
		HumanCorpus string `json:"human_corpus"`
		Decisions   struct {
			Basis string `json:"basis"`
			Units []struct {
				Reason string  `json:"reason"`
				Label  *string `json:"label"`
			} `json:"units"`
		} `json:"decisions"`
		Features unswell.PreparedFeatureCollection `json:"features"`
		Bindings []struct {
			UnitID           string `json:"unit_id"`
			Path             string `json:"path"`
			Partition        string `json:"partition"`
			FeatureInputHash string `json:"feature_input_hash"`
		} `json:"bindings"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Status, qt.Equals, "verified_targets_with_measured_features")
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.Decisions.Basis, qt.Equals, "simulation")
	c.Assert(result.Decisions.Units, qt.HasLen, 5)
	for _, decision := range result.Decisions.Units {
		c.Assert(decision.Reason, qt.Equals, "missing_judgments")
		c.Assert(decision.Label, qt.IsNil)
	}
	c.Assert(result.Bindings, qt.HasLen, 378)
	c.Assert(result.Features.Sources, qt.HasLen, 8)
	measurements := make(map[string]unswell.PreparedFeatureUnit)
	for _, source := range result.Features.Sources {
		for _, unit := range source.Units {
			measurements[source.Path+unit.InputHash] = unit
		}
	}
	var expected []corpusExpectedUnit
	decodeFile(t, "corpusdata/ptah-units.golden.json", &expected)
	for i, binding := range result.Bindings {
		c.Assert(binding.UnitID, qt.Equals, expected[i].ID)
		c.Assert(binding.Partition, qt.Equals, "development")
		unit, exists := measurements[binding.Path+binding.FeatureInputHash]
		c.Assert(exists, qt.IsTrue)
		c.Assert(unit.Binding.Kind, qt.Equals, expected[i].Kind)
		c.Assert(unit.Binding.Segments, qt.DeepEquals, expected[i].Segments)
		if i < 2 {
			c.Assert(*unit.Values[0].Number, qt.Equals, []float64{1, 10}[i])
		}
	}
}
