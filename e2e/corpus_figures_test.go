package e2e_test

import (
	"encoding/xml"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
)

// A figure is only evidence when it is generated from the record it describes.
// This runs the whole path, so nothing between the metrics and the chart can be
// transcribed by hand.
func TestCorpusFiguresRenderSavedEvaluations(t *testing.T) {
	c := qt.New(t)
	binary := buildResearchTool(t, "corpus")
	evaluation := savedEvaluation(t, binary)
	for _, plot := range []string{"reliability", "risk-coverage"} {
		t.Run(plot, func(t *testing.T) {
			c := qt.New(t)
			svg := researchCommand(t, binary, []string{"figures", "--plot", plot}, evaluation, 0)
			c.Assert(xml.Unmarshal(svg, new(any)), qt.IsNil)
			c.Assert(string(svg), qt.Contains, "unswell-research-figure-v1")
			c.Assert(researchCommand(t, binary, []string{"figures", "--plot", plot}, evaluation, 0), qt.DeepEquals, svg)
		})
	}
	c.Assert(researchCommand(t, binary, []string{"figures", "--plot", "scatter"}, evaluation, 2), qt.HasLen, 0)
	c.Assert(researchCommand(t, binary, []string{"figures"}, evaluation, 2), qt.HasLen, 0)
	// A record from another evaluation version is refused rather than drawn.
	manifest, err := os.ReadFile("../research/annotation/training/testdata/manifest.json")
	c.Assert(err, qt.IsNil)
	corpus := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	c.Assert(researchCommand(t, binary, []string{"figures", "--plot", "reliability"}, corpus, 2), qt.HasLen, 0)
}

func savedEvaluation(t *testing.T, binary string) []byte {
	t.Helper()
	c := qt.New(t)
	const fixture = "../research/annotation/training/testdata/"
	manifest, err := os.ReadFile(fixture + "manifest.json")
	c.Assert(err, qt.IsNil)
	plan := researchCommand(t, binary, []string{"plan"}, manifest, 0)
	corpus := researchCommand(t, binary, []string{"extract", "--root", fixture + "sources"}, plan, 0)
	model := researchCommand(t, binary, []string{"train", "--root", fixture + "sources", "--round", fixture + "round.json",
		"--kind", "paragraph", "--feature", "prose-words", "--calibration", "isotonic", "--allow-simulation"}, corpus, 0)
	directory := t.TempDir()
	arguments := frozenPredictionFiles(t, directory, corpus, model, "prepared_piece")
	predictions := researchCommand(t, binary, arguments, corpus, 0)
	return researchCommand(t, binary, []string{"evaluate", "--corpus", directory + "/corpus.json",
		"--round", fixture + "round.json", "--allow-simulation"}, predictions, 0)
}
