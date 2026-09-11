package training_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/llmdet"
	"github.com/stokaro/unswell/research/annotation/training"
)

// toyPack writes a pack of one model whose tokenizer knows the letters and
// the space and whose table retains a few contexts of the fixture's prose.
func toyPack(c *qt.C, directory string) string {
	c.Helper()
	vocabulary := map[string]int{}
	letters := "Ġ a b c d e h i l n o p r s t u w y f g k m v x z q j T A C W E L R O I B S F K M N P D G H J U V X Y Z Q Ċ"
	marks := ". , - ' 1 2 3 4 5 6 7 8 9 0 ( ) : ; \" ? ! / _ ` # * [ ] { } < > = + & % $ @ ~ ^ | \\"
	symbols := strings.Split(letters+" "+marks, " ")
	for i, symbol := range symbols {
		vocabulary[symbol] = i
	}
	vocabulary["Ġt"], vocabulary["Ġth"], vocabulary["Ġthe"] = len(vocabulary), len(vocabulary)+1, len(vocabulary)+2
	merges := "#version: 0.2\nĠ t\nĠt h\nĠth e\n"
	encodedVocabulary, err := json.Marshal(vocabulary)
	c.Assert(err, qt.IsNil)
	the := vocabulary["Ġthe"]
	rows := []llmdet.TableRow{
		{Context: []int{the}, Continuations: []int{vocabulary["Ġ"], vocabulary["a"]}, Probabilities: []float64{0.25, 0.125}},
		{Context: []int{vocabulary["e"]}, Continuations: []int{vocabulary["Ġ"]}, Probabilities: []float64{0.5}},
		{Context: []int{vocabulary["Ġ"], vocabulary["c"]}, Continuations: []int{vocabulary["a"]}, Probabilities: []float64{0.5}},
	}
	table, err := llmdet.EncodeTable("toy", 256, llmdet.TableSource{File: "toy.npz", SHA256: strings.Repeat("b", 64)}, rows)
	c.Assert(err, qt.IsNil)
	digest := func(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
	c.Assert(os.WriteFile(filepath.Join(directory, "vocab.json"), encodedVocabulary, 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(directory, "merges.txt"), []byte(merges), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(directory, "toy.table"), table, 0o600), qt.IsNil)
	manifest := training.LLMDetPack{Version: training.LLMDetPackVersion, Models: []training.LLMDetModel{{Model: "toy",
		Table: "toy.table", TableSHA256: digest(table), Vocabulary: "vocab.json", VocabularySHA256: digest(encodedVocabulary),
		Merges: "merges.txt", MergesSHA256: digest([]byte(merges))}}}
	data, err := json.Marshal(manifest)
	c.Assert(err, qt.IsNil)
	path := filepath.Join(directory, "pack.json")
	c.Assert(os.WriteFile(path, data, 0o600), qt.IsNil)
	return path
}

// The LLMDet baseline tokenizes every target, measures it against each
// table, records a perplexity and a coverage per model with the reference's
// abstention reasons, restores, and predicts with the same pack.
func TestRunLLMDetDecisionsFitsProxyColumns(t *testing.T) {
	c := qt.New(t)
	artifact, files := cohortFixture(c)
	decisions, err := corpus.CohortDecisions(c.Context(), artifact)
	c.Assert(err, qt.IsNil)
	pack, err := training.LoadLLMDetPack(c.Context(), toyPack(c, t.TempDir()))
	c.Assert(err, qt.IsNil)
	c.Assert(pack.ModelNames(), qt.DeepEquals, []string{"toy"})
	options := fittingOptions()
	options.AllowSimulation, options.Features, options.MissingFeatures = false, nil, "exclude"

	fitted, err := training.RunLLMDetDecisions(c.Context(), artifact, decisions, files, options, pack)
	c.Assert(err, qt.IsNil)
	c.Assert(fitted.Identity.FeatureSource, qt.Equals, "llmdet_tables")
	c.Assert(fitted.Identity.LLMDetPackHash, qt.Equals, pack.SHA256)
	c.Assert(fitted.Options.Features, qt.DeepEquals, []string{"llmdet.context-coverage/toy", "llmdet.proxy-perplexity/toy"})
	c.Assert(len(fitted.Partitions[0].Rows) > 0, qt.IsTrue)
	for _, partition := range fitted.Partitions[:1] {
		for key, count := range partition.Unavailable {
			c.Assert(strings.HasPrefix(key, "llmdet."), qt.IsTrue, qt.Commentf("%s", key))
			c.Assert(count > 0, qt.IsTrue)
		}
	}

	encoded, err := json.Marshal(fitted)
	c.Assert(err, qt.IsNil)
	_, err = training.Load(c.Context(), encoded)
	c.Assert(err, qt.IsNil)

	threshold := 0.5
	plan := training.PredictionPlan{Version: training.PredictionVersion, ID: "llmdet-trial", ProtocolSHA256: protocolDigest(),
		ModelSHA256: fitted.SHA256, CorpusSHA256: artifact.SHA256, Partition: "final_test",
		Context: "prepared_piece", Response: "isotonic", Threshold: &threshold}
	predictions, err := training.PredictWith(c.Context(), artifact, files, fitted, plan, training.PredictionResources{LLMDet: &pack})
	c.Assert(err, qt.IsNil)
	c.Assert(len(predictions.Rows) > 0, qt.IsTrue)
	_, err = training.Predict(c.Context(), artifact, files, fitted, plan, nil)
	c.Assert(err, qt.ErrorMatches, "LLMDet prediction requires the fitted table pack")

	joint := options
	joint.Features = []string{"prose-words"}
	combined, err := training.RunLLMDetDecisions(c.Context(), artifact, decisions, files, joint, pack)
	c.Assert(err, qt.IsNil)
	c.Assert(combined.Options.Features, qt.DeepEquals, []string{"llmdet.context-coverage/toy", "llmdet.proxy-perplexity/toy", "prose-words"})

	other := options
	other.MissingFeatures = "zero"
	_, err = training.RunLLMDetDecisions(c.Context(), artifact, decisions, files, other, pack)
	c.Assert(err, qt.ErrorMatches, "the zero missing-feature policy applies to rule activations only")
	broken := pack
	broken.SHA256 = strings.Repeat("0", 64)
	plan.ModelSHA256 = fitted.SHA256
	_, err = training.PredictWith(c.Context(), artifact, files, fitted, plan, training.PredictionResources{LLMDet: &broken})
	c.Assert(err, qt.ErrorMatches, "LLMDet prediction requires the fitted table pack")
}

// A manifest with a wrong digest, an absolute path, or a repeated model is
// refused before any table is read.
func TestLoadLLMDetPackChecksEveryFile(t *testing.T) {
	c := qt.New(t)
	directory := t.TempDir()
	path := toyPack(c, directory)
	var manifest map[string]any
	data, err := os.ReadFile(path) // #nosec G304 -- The test reads its own fixture.
	c.Assert(err, qt.IsNil)
	c.Assert(json.Unmarshal(data, &manifest), qt.IsNil)
	models := manifest["models"].([]any)
	model := models[0].(map[string]any)
	for _, row := range []struct {
		name    string
		change  func()
		message string
	}{
		{"digest", func() { model["table_sha256"] = strings.Repeat("1", 64) }, "LLMDet model toy: file \"toy.table\" does not match its digest"},
		{"absolute", func() { model["table"] = "/toy.table" }, "LLMDet model toy: pack file \"/toy.table\" must be a local relative path"},
		{"escape", func() { model["table"] = "../toy.table" }, "LLMDet model toy: pack file \"../toy.table\" must be a local relative path"},
		{"repeated", func() { manifest["models"] = []any{model, model} }, "LLMDet pack models require distinct names"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(json.Unmarshal(data, &manifest), qt.IsNil)
			models = manifest["models"].([]any)
			model = models[0].(map[string]any)
			row.change()
			changed, err := json.Marshal(manifest)
			c.Assert(err, qt.IsNil)
			c.Assert(os.WriteFile(path, changed, 0o600), qt.IsNil)
			_, err = training.LoadLLMDetPack(c.Context(), path)
			c.Assert(err, qt.ErrorMatches, row.message)
		})
	}
}
