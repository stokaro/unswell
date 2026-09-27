package unswell_test

import (
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/probability"
	"github.com/stokaro/unswell/rule"
)

func TestShippedOriginPackKeepsItsDeclaredPreparation(t *testing.T) {
	c := qt.New(t)
	pack, err := os.ReadFile("research/origin/packs/unswell-origin-lexical-v1.json")
	c.Assert(err, qt.IsNil)
	policy := "version: 1\nextraction:\n  contexts: [comment, heading, list-item, paragraph, string, table-cell]\n" +
		"origin:\n  model: pack\n  accept_experimental: true\n  on_incompatible: fail\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), OriginModel: pack})
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
		"# Retry behavior\n\nThe client retries a failed request after a short wait. " +
			"Each attempt keeps the original request identifier so callers can find its log entries " +
			"and compare the recorded transport errors.\n")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	available := 0
	for _, assessment := range result.Assessments {
		c.Assert(assessment.OriginStatus, qt.Not(qt.Equals), probability.StatusIncompatible)
		if assessment.OriginEstimate != nil {
			c.Assert(assessment.Scope, qt.Equals, "paragraph")
			c.Assert(assessment.OriginStatus, qt.Equals, probability.StatusAvailable)
			available++
		}
	}
	c.Assert(available, qt.Equals, 1)
}

func structuralPack(c *qt.C, structure bool, origin bool) []byte {
	c.Helper()
	prepared := packPreparation(c)
	return packFixture(c, "sentence", func(file *probability.File) {
		file.Contract.IncludeStructure = structure
		var err error
		file.Contract.PreparationHash, err = nlp.PreparationHash(prepared.ExtractionPolicyHash,
			file.Contract.IncludeQuotes, structure)
		c.Assert(err, qt.IsNil)
		if origin {
			file.Task, file.ID, file.Rubric = probability.TaskOrigin, "origin-fixture", "fixture-origin-classes-v1"
		}
	})
}

func TestProbabilityPackPreparationIsIndependentOfRuleStructure(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# Retry policy\r\n\r\nRetry failed requests when the transport reports a timeout.\r\n\r\n" +
			"> Quoted instructions must stay outside the selected prose.\r\n\r\n" +
			"Клиент повторяет запрос после временной ошибки подключения.\r\n")}
	policy := "version: 1\ncalibration:\n  model: pack\n  accept_experimental: true\n  on_incompatible: fail\n" +
		"origin:\n  model: pack\n  accept_experimental: true\n  on_incompatible: fail\n"
	plain, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	expected, err := plain.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	for _, structure := range []bool{false, true} {
		t.Run(map[bool]string{false: "unstructured revision", true: "structured revision"}[structure], func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(policy), CollectBaseline: true,
				PreparedFeatures: []string{"prose-words"}, PreparedKinds: []string{"sentence"},
				Model: structuralPack(c, structure, false), OriginModel: structuralPack(c, !structure, true)})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Status, qt.Equals, "complete")
			c.Assert(result.Findings, qt.DeepEquals, expected.Findings)
			c.Assert(result.Documents, qt.HasLen, 1)
			expectedDocument := expected.Documents[0]
			expectedDocument.ConfigHash = result.Documents[0].ConfigHash
			c.Assert(result.Documents[0], qt.DeepEquals, expectedDocument)
			c.Assert(result.Gate, qt.DeepEquals, expected.Gate)
			c.Assert(result.PreparedFeatures.Sources[0].IncludeStructure, qt.IsTrue)
			available := 0
			for _, assessment := range result.Assessments {
				c.Assert(assessment.ProbabilityStatus, qt.Not(qt.Equals), probability.StatusIncompatible)
				c.Assert(assessment.OriginStatus, qt.Not(qt.Equals), probability.StatusIncompatible)
				if assessment.Scope != "sentence" {
					continue
				}
				text := string(source.Bytes[assessment.Span.Start:assessment.Span.End])
				if text == "Retry policy" {
					c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusUnsupportedUnit)
					c.Assert(assessment.OriginStatus, qt.Equals, probability.StatusUnsupportedUnit)
					continue
				}
				c.Assert(text, qt.Equals, "Retry failed requests when the transport reports a timeout.")
				c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusAvailable)
				c.Assert(assessment.OriginStatus, qt.Equals, probability.StatusAvailable)
				c.Assert(assessment.SlopProbability, qt.DeepEquals, assessment.OriginEstimate)
				available++
			}
			c.Assert(available, qt.Equals, 1)
		})
	}
}

func TestProbabilityStructureDoesNotOverrideCallerQuotePolicy(t *testing.T) {
	c := qt.New(t)
	pack := structuralPack(c, false, false)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy +
		"analysis:\n  include_quotes: true\n"), Model: pack})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusIncompatible)
		c.Assert(assessment.SlopProbability, qt.IsNil)
	}
}

func TestStructuredProbabilityPackWorksWithoutStructuralRules(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy),
		Rules: []rule.Rule{}, Model: structuralPack(c, true, false)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(packStatuses(result.Assessments, "sentence"), qt.DeepEquals,
		[]string{probability.StatusAvailable, probability.StatusInsufficientEvidence, probability.StatusAvailable})
}
