package e2e_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestCompressionReferenceBankCommand(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	fixture := "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	artifact := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	directory := t.TempDir()
	ids := compressionSeedIDs(t, artifact)
	round := writeCompressionRound(t, directory, ids)
	selection := map[string]any{"version": "unswell-compression-bank-v1", "kind": "paragraph", "allow_simulation": true,
		"compression": map[string]any{"level": 0, "max_input_bytes": 1048576}, "cohorts": []map[string]any{
			{"id": "human", "origin": "human", "unit_ids": []string{ids[0]}},
			{"id": "generated", "origin": "generated", "unit_ids": []string{ids[1]}},
			{"id": "mixed", "origin": "mixed", "unit_ids": []string{ids[1], ids[0]}},
		}}
	path := writeCompressionJSON(t, directory, "selection.json", selection)
	args := []string{"reference-bank", "--root", fixture + "sources", "--round", round, "--selection", path}
	output := researchCommand(t, binary, args, artifact, 0)
	checkCompressionBankGolden(t, output)
	c.Assert(researchCommand(t, binary, args, artifact, 0), qt.DeepEquals, output)
	selection["allow_simulation"] = false
	writeCompressionJSON(t, directory, "selection.json", selection)
	c.Assert(researchCommand(t, binary, args, artifact, 2), qt.HasLen, 0)
	c.Assert(researchCommand(t, binary, []string{"reference-bank"}, artifact, 2), qt.HasLen, 0)
}

func compressionSeedIDs(t *testing.T, artifact []byte) []string {
	t.Helper()
	c := qt.New(t)
	var candidates struct {
		Units []struct {
			Partition string
			Unit      struct{ ID, Kind string }
		}
	}
	c.Assert(json.Unmarshal(artifact, &candidates), qt.IsNil)
	var ids []string
	for _, candidate := range candidates.Units {
		if candidate.Partition == "training" && candidate.Unit.Kind == "paragraph" {
			ids = append(ids, candidate.Unit.ID)
		}
	}
	c.Assert(ids, qt.HasLen, 2)
	return ids
}

func writeCompressionRound(t *testing.T, directory string, ids []string) string {
	t.Helper()
	var round map[string]any
	decodeFile(t, "../research/annotation/training/testdata/round.json", &round)
	for _, value := range round["units"].([]any) {
		unit := value.(map[string]any)
		index := slices.Index(ids, unit["id"].(string))
		if index < 0 {
			continue
		}
		origin := unit["origin"].(map[string]any)
		origin["label"], origin["scope"] = []string{"human", "generated"}[index], "unit"
		origin["evidence"] = "Simulated endpoint claim for this tutorial command test."
		origin["generation_record"] = "fixture:simulated-provenance"
	}
	return writeCompressionJSON(t, directory, "round.json", round)
}

func writeCompressionJSON(t *testing.T, directory, name string, value any) string {
	t.Helper()
	c := qt.New(t)
	data, err := json.Marshal(value)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(directory, name)
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	return path
}

func checkCompressionBankGolden(t *testing.T, output []byte) {
	t.Helper()
	c := qt.New(t)
	var bank struct {
		Basis                  string                       `json:"basis"`
		HumanCorpus            string                       `json:"human_corpus"`
		RemainingTrainingUnits int                          `json:"remaining_training_units"`
		ReservedGroups         []struct{ Sources []string } `json:"reserved_groups"`
		Cohorts                []struct {
			ID        string `json:"id"`
			Origin    string `json:"origin"`
			Reference string `json:"reference"`
			Units     []struct {
				SourceID string `json:"source_id"`
				Start    int    `json:"start"`
				End      int    `json:"end"`
			} `json:"units"`
		} `json:"cohorts"`
	}
	c.Assert(json.Unmarshal(output, &bank), qt.IsNil)
	var sources []string
	for _, group := range bank.ReservedGroups {
		sources = append(sources, group.Sources...)
	}
	slices.Sort(sources)
	projection := map[string]any{"basis": bank.Basis, "human_corpus": bank.HumanCorpus, "cohorts": bank.Cohorts,
		"reserved_sources": sources, "remaining_training_units": bank.RemainingTrainingUnits}
	actual, err := json.MarshalIndent(projection, "", "  ")
	c.Assert(err, qt.IsNil)
	expected, err := os.ReadFile("corpusdata/compression-bank.golden.json")
	c.Assert(err, qt.IsNil)
	c.Assert(append(actual, '\n'), qt.DeepEquals, expected)
}
