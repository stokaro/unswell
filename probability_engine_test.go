package unswell_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/probability"
)

const packPolicy = "version: 1\ncalibration:\n  model: pack\n  accept_experimental: true\n"

const packProse = "Retry failed requests when the transport reports a timeout.\n\n" +
	"Short text here. Another sentence follows it now.\n"

func packSource() document.Source {
	return document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(packProse)}
}

// packFixture builds a loadable pack for the effective policy of a model-free
// engine. Its numerical parameters are fixtures without editorial meaning.
func packFixture(c *qt.C, kind string, edit func(*probability.File)) []byte {
	c.Helper()
	file := probability.File{
		Version: probability.Version, ID: "engine-fixture", DeclaredStatus: "experimental", HumanCorpus: "not_qualified",
		Task: probability.Task, Rubric: "fixture-rubric-v1", Kind: kind, Contract: packFixtureContract(c, kind),
		Limits: probability.Limits{MinWords: 5}, Estimator: "logistic",
		Logistic: &probability.Logistic{Means: []float64{8, 0.9}, Scales: []float64{4, 0.2},
			Weights: []float64{0.2, -0.1}, Intercept: 0},
		Calibration: probability.Calibration{Algorithm: "isotonic", Scores: []float64{-9, 0, 9},
			Responses: []float64{0.05, 0.5, 0.95}},
	}
	if edit != nil {
		edit(&file)
	}
	digest, err := probability.Digest(file)
	c.Assert(err, qt.IsNil)
	file.SHA256 = digest
	data, err := json.Marshal(file)
	c.Assert(err, qt.IsNil)
	return data
}

func packFixtureContract(c *qt.C, kind string) probability.Contract {
	c.Helper()
	catalog, err := feature.UnitCatalog(kind)
	c.Assert(err, qt.IsNil)
	columns := make([]feature.Descriptor, 0, 2)
	for _, id := range []string{"prose-words", "type-token-ratio"} {
		index := slices.IndexFunc(catalog, func(d feature.Descriptor) bool { return d.ID == id })
		c.Assert(index >= 0, qt.IsTrue)
		columns = append(columns, catalog[index])
	}
	encoded, err := json.Marshal(columns)
	c.Assert(err, qt.IsNil)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	return probability.Contract{FeatureContract: feature.UnitContract, UnitContract: nlp.UnitContract, Columns: columns,
		ColumnsSHA256: fmt.Sprintf("%x", sha256.Sum256(encoded)), NLP: provider.Identity(),
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences}, PreparationHash: packPreparationHash(c)}
}

// packPreparationHash reproduces the engine's effective preparation identity
// from its public policy, without assuming a fixed default extraction policy.
func packPreparationHash(c *qt.C) string {
	c.Helper()
	probe, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	policy, err := probe.PolicyForFile("")
	c.Assert(err, qt.IsNil)
	extraction, err := json.Marshal(policy.Extraction)
	c.Assert(err, qt.IsNil)
	preparation, err := nlp.PreparationHash(fmt.Sprintf("%x", sha256.Sum256(extraction)), policy.Analysis.IncludeQuotes, false)
	c.Assert(err, qt.IsNil)
	return preparation
}

func packStatuses(assessments []unswell.Assessment, scope string) []string {
	result := []string{}
	for _, assessment := range assessments {
		if assessment.Scope == scope {
			result = append(result, assessment.ProbabilityStatus)
		}
	}
	return result
}

func TestProbabilityEstimatesOnlyTheQualifiedUnitKind(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy), Model: packFixture(c, "sentence", nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(packStatuses(result.Assessments, "sentence"), qt.DeepEquals,
		[]string{probability.StatusAvailable, probability.StatusInsufficientEvidence, probability.StatusAvailable})
	c.Assert(packStatuses(result.Assessments, "paragraph"), qt.DeepEquals,
		[]string{probability.StatusUnsupportedUnit, probability.StatusUnsupportedUnit})
	for _, assessment := range result.Assessments {
		if assessment.ProbabilityStatus == probability.StatusAvailable {
			c.Assert(*assessment.SlopProbability >= 0 && *assessment.SlopProbability <= 1, qt.IsTrue)
			c.Assert(assessment.ProbabilityDetail, qt.Equals, "")
			continue
		}
		c.Assert(assessment.SlopProbability, qt.IsNil)
	}
	model := result.Manifest.Probability
	c.Assert(model, qt.IsNotNil)
	c.Assert(model.PackID, qt.Equals, "engine-fixture")
	c.Assert(model.DeclaredStatus, qt.Equals, "experimental")
	c.Assert(model.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(model.Kind, qt.Equals, "sentence")
	c.Assert(model.MinWords, qt.Equals, 5)
	c.Assert(model.SHA256, qt.HasLen, 64)
}

func TestProbabilityLeavesTheIndexAndGateUnchanged(t *testing.T) {
	c := qt.New(t)
	plain, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	expected, err := plain.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy), Model: packFixture(c, "paragraph", nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate, qt.DeepEquals, expected.Gate)
	c.Assert(result.Findings, qt.DeepEquals, expected.Findings)
	c.Assert(result.Assessments, qt.HasLen, len(expected.Assessments))
	for i, assessment := range result.Assessments {
		reference := expected.Assessments[i]
		c.Assert(assessment.EffectiveSlopScore, qt.Equals, reference.EffectiveSlopScore)
		c.Assert(assessment.Span, qt.Equals, reference.Span)
		c.Assert(reference.ProbabilityStatus, qt.Equals, probability.StatusUnavailable)
	}
	c.Assert(packStatuses(result.Assessments, "paragraph"), qt.DeepEquals,
		[]string{probability.StatusAvailable, probability.StatusAvailable})
	c.Assert(expected.Manifest.Probability, qt.IsNil)
}

func TestProbabilityRequiresAnExplicitPackAndPolicy(t *testing.T) {
	c := qt.New(t)
	valid := packFixture(c, "sentence", nil)
	parsing := packFixture(c, "sentence", func(f *probability.File) {
		f.Contract.Capabilities = append(f.Contract.Capabilities, nlp.Dependencies)
	})
	for _, row := range []struct {
		name, policy, message string
		model                 []byte
	}{
		{"unrequested pack", "version: 1\n", ".*requires calibration.model: pack.*", valid},
		{"corrupt pack", packPolicy, ".*probability pack.*", []byte("{}")},
		{"experimental pack", "version: 1\ncalibration:\n  model: pack\n", ".*declares experimental status.*", valid},
		{"unavailable capability", packPolicy, ".*requires unavailable capability dependencies.*", parsing},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(row.policy), Model: row.model})
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(engine, qt.IsNil)
		})
	}
}

// Configuration inspection stays available without a pack, so an unanalyzable
// policy is diagnosable; analysis itself cannot silently abstain.
func TestProbabilityAnalysisRequiresTheRequestedPack(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy)})
	c.Assert(err, qt.IsNil)
	policy, err := engine.PolicyForFile("")
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Calibration.Model, qt.Equals, "pack")
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.ErrorMatches, ".*requires an explicit model pack.*")
	c.Assert(result.Status, qt.Equals, "incomplete")
	c.Assert(result.Assessments, qt.HasLen, 0)
}

func TestProbabilityAbstainsOrFailsOnIncompatibleSources(t *testing.T) {
	c := qt.New(t)
	incompatible := func(f *probability.File) { f.Contract.PreparationHash = fmt.Sprintf("%064d", 1) }
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy), Model: packFixture(c, "sentence", incompatible)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusIncompatible)
		c.Assert(assessment.SlopProbability, qt.IsNil)
	}
	strict := "version: 1\ncalibration:\n  model: pack\n  accept_experimental: true\n  on_incompatible: fail\n"
	failing, err := unswell.New(unswell.Options{Config: []byte(strict), Model: packFixture(c, "sentence", incompatible)})
	c.Assert(err, qt.IsNil)
	failed, err := failing.Analyze(t.Context(), packSource())
	c.Assert(err, qt.ErrorMatches, ".*expects a different extraction and preparation policy.*")
	c.Assert(failed.Status, qt.Equals, "incomplete")
	c.Assert(failed.Gate.Passed, qt.IsFalse)
}

func TestProbabilityPackChangesAcceptedDebtCompatibility(t *testing.T) {
	c := qt.New(t)
	plain, err := unswell.New(unswell.Options{CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	expected, err := plain.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy), CollectBaseline: true,
		Model: packFixture(c, "sentence", nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.BaselineSnapshot.Compatibility.ModelHash, qt.Not(qt.Equals),
		expected.BaselineSnapshot.Compatibility.ModelHash)
	c.Assert(result.BaselineSnapshot.Compatibility.ModelHash, qt.HasLen, 64)
}
