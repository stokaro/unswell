package unswell_test

import (
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/probability"
)

const originPolicy = "version: 1\norigin:\n  model: pack\n  accept_experimental: true\n"

// originPack declares the separate origin target. Its parameters are fixtures:
// this test asserts engine behavior, not that any origin claim is true.
func originPack(c *qt.C, edit func(*probability.File)) []byte {
	c.Helper()
	return packFixture(c, "sentence", func(file *probability.File) {
		file.Task, file.ID = probability.TaskOrigin, "origin-fixture"
		file.Rubric = "fixture-origin-classes-v1"
		if edit != nil {
			edit(file)
		}
	})
}

func originStatuses(assessments []unswell.Assessment, scope string) []string {
	result := []string{}
	for _, assessment := range assessments {
		if assessment.Scope == scope {
			result = append(result, assessment.OriginStatus)
		}
	}
	return result
}

func TestOriginChannelIsOffByDefault(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Origin, qt.IsNil)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.OriginStatus, qt.Equals, "")
		c.Assert(assessment.OriginEstimate, qt.IsNil)
		c.Assert(assessment.OriginDetail, qt.Equals, "")
	}
}

func TestOriginChannelReportsSeparatelyFromTheIndexAndGate(t *testing.T) {
	c := qt.New(t)
	plain, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	expected, err := plain.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Config: []byte(originPolicy), OriginModel: originPack(c, nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate, qt.DeepEquals, expected.Gate)
	c.Assert(result.Findings, qt.DeepEquals, expected.Findings)
	c.Assert(result.Assessments, qt.HasLen, len(expected.Assessments))
	for i, assessment := range result.Assessments {
		reference := expected.Assessments[i]
		c.Assert(assessment.EffectiveSlopScore, qt.Equals, reference.EffectiveSlopScore)
		c.Assert(assessment.SlopProbability, qt.IsNil)
		c.Assert(assessment.ProbabilityStatus, qt.Equals, probability.StatusUnavailable)
	}
	c.Assert(originStatuses(result.Assessments, "sentence"), qt.DeepEquals,
		[]string{probability.StatusAvailable, probability.StatusInsufficientEvidence, probability.StatusAvailable})
	c.Assert(originStatuses(result.Assessments, "paragraph"), qt.DeepEquals,
		[]string{probability.StatusUnsupportedUnit, probability.StatusUnsupportedUnit})
	model := result.Manifest.Origin
	c.Assert(model, qt.IsNotNil)
	c.Assert(model.Task, qt.Equals, probability.TaskOrigin)
	c.Assert(model.PackID, qt.Equals, "origin-fixture")
	c.Assert(result.Manifest.Probability, qt.IsNil)
}

func TestOriginChannelKeepsTheTwoTargetsApart(t *testing.T) {
	c := qt.New(t)
	for _, row := range []struct{ name, policy, message string }{
		{"revision pack in the origin channel", originPolicy, ".*origin.model expects a origin_endpoint pack.*"},
		{"origin pack unrequested", "version: 1\n", ".*requires origin.model: pack.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			model := originPack(c, nil)
			if row.name == "revision pack in the origin channel" {
				model = packFixture(c, "sentence", nil)
			}
			engine, err := unswell.New(unswell.Options{Config: []byte(row.policy), OriginModel: model})
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(engine, qt.IsNil)
		})
	}
	// The revision channel refuses an origin pack for the same reason.
	engine, err := unswell.New(unswell.Options{Config: []byte(packPolicy), Model: originPack(c, nil)})
	c.Assert(err, qt.ErrorMatches, ".*calibration.model expects a editorial_needs_revision pack.*")
	c.Assert(engine, qt.IsNil)
}

func TestOriginChannelRequiresItsPackAndAbstainsOnIncompatibility(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(originPolicy)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.ErrorMatches, ".*origin.model: pack requires an explicit model pack.*")
	c.Assert(result.Status, qt.Equals, "incomplete")

	incompatible := func(f *probability.File) { f.Contract.PreparationHash = fmt.Sprintf("%064d", 1) }
	abstaining, err := unswell.New(unswell.Options{Config: []byte(originPolicy),
		OriginModel: originPack(c, incompatible)})
	c.Assert(err, qt.IsNil)
	abstained, err := abstaining.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(abstained.Gate.Passed, qt.IsTrue)
	for _, assessment := range abstained.Assessments {
		c.Assert(assessment.OriginStatus, qt.Equals, probability.StatusIncompatible)
		c.Assert(assessment.OriginEstimate, qt.IsNil)
	}
	strict := "version: 1\norigin:\n  model: pack\n  accept_experimental: true\n  on_incompatible: fail\n"
	failing, err := unswell.New(unswell.Options{Config: []byte(strict), OriginModel: originPack(c, incompatible)})
	c.Assert(err, qt.IsNil)
	incomplete, err := failing.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNotNil)
	c.Assert(incomplete.Status, qt.Equals, "incomplete")
}

func TestOriginChannelCannotFailAGate(t *testing.T) {
	c := qt.New(t)
	// A probability gate belongs to the revision channel; an origin policy has
	// no threshold at all, so the only way it could decide a build is a bug.
	both := "version: 1\ncalibration:\n  model: pack\ngate:\n  probability:\n    fail_at: 0.9\n    require: true\n" +
		"origin:\n  model: pack\n  accept_experimental: true\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(both), Model: acceptedPack(c, nil),
		OriginModel: originPack(c, nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Gate.Reasons, qt.HasLen, 0)
	for _, assessment := range result.Assessments {
		if assessment.Scope != "sentence" {
			continue
		}
		c.Assert(assessment.OriginStatus, qt.Not(qt.Equals), "")
	}
	c.Assert(result.Manifest.Origin, qt.IsNotNil)
	c.Assert(result.Manifest.Probability, qt.IsNotNil)
}
