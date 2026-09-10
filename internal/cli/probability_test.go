package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/internal/cli"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/probability"
)

const modelConfig = "version: 1\nextends: [builtin:strict-v1]\ncalibration:\n  model: pack\n  accept_experimental: true\n"

// commandPack writes a loadable pack for the configuration under test. Its
// numerical parameters are fixtures and carry no editorial meaning.
func commandPack(c *qt.C, root, name, config, task string) string {
	c.Helper()
	catalog, err := feature.UnitCatalog("sentence")
	c.Assert(err, qt.IsNil)
	columns := []feature.Descriptor{}
	for _, id := range []string{"prose-words", "type-token-ratio"} {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		c.Assert(index >= 0, qt.IsTrue)
		columns = append(columns, catalog[index])
	}
	encoded, err := json.Marshal(columns)
	c.Assert(err, qt.IsNil)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	file := probability.File{Version: probability.Version, ID: "command-fixture", DeclaredStatus: "experimental",
		HumanCorpus: "not_qualified", Task: task, Rubric: "fixture-rubric-v1", Kind: "sentence",
		Contract: probability.Contract{FeatureContract: feature.UnitContract, UnitContract: nlp.UnitContract,
			Columns: columns, ColumnsSHA256: fmt.Sprintf("%x", sha256.Sum256(encoded)), NLP: provider.Identity(),
			Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, PreparationHash: commandPreparation(c, config)},
		Limits: probability.Limits{MinWords: 4}, Estimator: "logistic",
		Logistic: &probability.Logistic{Means: []float64{8, 0.9}, Scales: []float64{4, 0.2},
			Weights: []float64{0.2, -0.1}, Intercept: 0},
		Calibration: probability.Calibration{Algorithm: "isotonic", Scores: []float64{-9, 0, 9},
			Responses: []float64{0.05, 0.5, 0.95}}}
	digest, err := probability.Digest(file)
	c.Assert(err, qt.IsNil)
	file.SHA256 = digest
	data, err := json.Marshal(file)
	c.Assert(err, qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, name), data, 0o600), qt.IsNil)
	return name
}

// commandPreparation reads the effective policy without requesting a model, so
// the fixture matches whatever extraction the configuration under test selects.
func commandPreparation(c *qt.C, config string) string {
	c.Helper()
	engine, err := unswell.New(unswell.Options{Config: []byte(config)})
	c.Assert(err, qt.IsNil)
	policy, err := engine.PolicyForFile("")
	c.Assert(err, qt.IsNil)
	extraction, err := json.Marshal(policy.Extraction)
	c.Assert(err, qt.IsNil)
	preparation, err := nlp.PreparationHash(fmt.Sprintf("%x", sha256.Sum256(extraction)),
		policy.Analysis.IncludeQuotes, false)
	c.Assert(err, qt.IsNil)
	return preparation
}

func TestCheckReportsProbabilityFromAnExplicitPack(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"),
		[]byte("The client retries after a transport failure.\n"), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "model.yaml"), []byte(modelConfig), 0o600), qt.IsNil)
	pack := commandPack(c, root, "pack.json", modelConfig, probability.Task)
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	code := cli.Run(t.Context(), []string{"check", "draft.md", "--config", "model.yaml", "--model", pack, "--report", "json:-"},
		environment)
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr.String()))
	var result unswell.RunResult
	c.Assert(json.Unmarshal(out.Bytes(), &result), qt.IsNil)
	c.Assert(result.Manifest.Probability, qt.IsNotNil)
	c.Assert(result.Manifest.Probability.PackID, qt.Equals, "command-fixture")
	c.Assert(result.Manifest.Probability.DeclaredStatus, qt.Equals, "experimental")
	statuses := map[string]int{}
	for _, assessment := range result.Assessments {
		statuses[assessment.ProbabilityStatus]++
		if assessment.Scope == "sentence" {
			c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusAvailable)
			c.Assert(*assessment.SlopProbability >= 0 && *assessment.SlopProbability <= 1, qt.IsTrue)
		}
	}
	c.Assert(statuses[probability.StatusUnsupportedUnit], qt.Equals, 1)
	c.Assert(statuses[probability.StatusAvailable], qt.Equals, 1)
}

func TestCheckRejectsUnusableModelSelections(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), []byte("The client retries.\n"), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "model.yaml"), []byte(modelConfig), 0o600), qt.IsNil)
	c.Assert(os.WriteFile(filepath.Join(root, "broken.json"), []byte("{}"), 0o600), qt.IsNil)
	pack := commandPack(c, root, "pack.json", modelConfig, probability.Task)
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	for _, args := range [][]string{
		{"check", "draft.md", "--config", "model.yaml"},
		{"check", "draft.md", "--config", "model.yaml", "--model", "missing.json"},
		{"check", "draft.md", "--config", "model.yaml", "--model", "broken.json"},
		{"check", "draft.md", "--model", pack},
	} {
		out.Reset()
		stderr.Reset()
		c.Assert(cli.Run(t.Context(), args, environment), qt.Equals, 2, qt.Commentf("%v", args))
		c.Assert(stderr.String(), qt.Not(qt.Equals), "")
	}
}

func TestCheckReportsTheOriginChannelSeparately(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"),
		[]byte("The client retries after a transport failure.\n"), 0o600), qt.IsNil)
	policy := "version: 1\nextends: [builtin:strict-v1]\norigin:\n  model: pack\n  accept_experimental: true\n"
	c.Assert(os.WriteFile(filepath.Join(root, "origin.yaml"), []byte(policy), 0o600), qt.IsNil)
	pack := commandPack(c, root, "origin-pack.json", policy, "origin_endpoint")
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	code := cli.Run(t.Context(), []string{"check", "draft.md", "--config", "origin.yaml", "--origin-model", pack,
		"--report", "json:-"}, environment)
	c.Assert(code, qt.Equals, 0, qt.Commentf("%s", stderr.String()))
	var result unswell.RunResult
	c.Assert(json.Unmarshal(out.Bytes(), &result), qt.IsNil)
	c.Assert(result.Manifest.Origin, qt.IsNotNil)
	c.Assert(result.Manifest.Origin.Task, qt.Equals, "origin_endpoint")
	c.Assert(result.Manifest.Probability, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	estimated := 0
	for _, assessment := range result.Assessments {
		if assessment.OriginEstimate != nil {
			estimated++
			c.Assert(assessment.Scope, qt.Equals, "sentence")
			c.Assert(assessment.SlopProbability, qt.IsNil)
		}
	}
	c.Assert(estimated, qt.Equals, 1)
	out.Reset()
	stderr.Reset()
	c.Assert(cli.Run(t.Context(), []string{"check", "draft.md", "--origin-model", pack}, environment), qt.Equals, 2)
}

func TestDoctorReportsTheConfiguredCalibration(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	c.Assert(os.WriteFile(filepath.Join(root, "model.yaml"), []byte(modelConfig), 0o600), qt.IsNil)
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	c.Assert(cli.Run(t.Context(), []string{"doctor"}, environment), qt.Equals, 0)
	var plain struct {
		ProbabilityStatus string          `json:"probability_status"`
		CalibrationModel  json.RawMessage `json:"calibration_model"`
	}
	c.Assert(json.Unmarshal(out.Bytes(), &plain), qt.IsNil)
	c.Assert(plain.ProbabilityStatus, qt.Equals, probability.StatusUnavailable)
	c.Assert(string(plain.CalibrationModel), qt.Equals, "null")
	out.Reset()
	c.Assert(cli.Run(t.Context(), []string{"doctor", "--config", "model.yaml"}, environment), qt.Equals, 0)
	var configured struct {
		ProbabilityStatus string `json:"probability_status"`
		CalibrationModel  struct {
			Model              string `json:"model"`
			OnIncompatible     string `json:"on_incompatible"`
			AcceptExperimental bool   `json:"accept_experimental"`
		} `json:"calibration_model"`
	}
	c.Assert(json.Unmarshal(out.Bytes(), &configured), qt.IsNil)
	c.Assert(configured.ProbabilityStatus, qt.Equals, "decided_per_unit")
	c.Assert(configured.CalibrationModel.Model, qt.Equals, "pack")
	c.Assert(configured.CalibrationModel.OnIncompatible, qt.Equals, "unavailable")
	c.Assert(configured.CalibrationModel.AcceptExperimental, qt.IsTrue)
}
