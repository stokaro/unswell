package unswell_test

import (
	"fmt"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/probability"
)

// acceptedPolicy gates on an accepted pack. Experimental packs cannot gate, so
// this policy deliberately omits accept_experimental.
const acceptedPolicy = "version: 1\ncalibration:\n  model: pack\ngate:\n  probability:\n    fail_at: 0.4\n"

const requirePolicy = "version: 1\ncalibration:\n  model: pack\n" +
	"gate:\n  probability:\n    fail_at: 0.9\n    require: true\n"

// acceptedPack declares the acceptance the gate requires. The declaration is a
// fixture: this test asserts engine behavior, not a qualified model.
func acceptedPack(c *qt.C, edit func(*probability.File)) []byte {
	c.Helper()
	return packFixture(c, "sentence", func(file *probability.File) {
		file.DeclaredStatus, file.HumanCorpus = "accepted", "qualified"
		file.Evaluation = "fixture evaluation record"
		if edit != nil {
			edit(file)
		}
	})
}

func gateCodes(result unswell.RunResult) []string {
	codes := []string{}
	for _, reason := range result.Gate.Reasons {
		codes = append(codes, reason.Code)
	}
	return codes
}

func TestProbabilityGateFailsOnACalibratedEstimate(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(acceptedPolicy), Model: acceptedPack(c, nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(gateCodes(result), qt.Contains, "gate.sentence-probability")
	index := slices.IndexFunc(result.Findings, func(f unswell.Finding) bool { return f.RuleID == "gate.sentence-probability" })
	c.Assert(index >= 0, qt.IsTrue)
	finding := result.Findings[index]
	c.Assert(finding.Derived, qt.IsTrue)
	c.Assert(finding.Severity, qt.Equals, "error")
	c.Assert(finding.Gate, qt.Equals, "forbid")
	c.Assert(finding.Message, qt.Matches, `sentence revision probability .* reaches the configured threshold of 0.4\.`)
	c.Assert(finding.Evidence.Metrics, qt.HasLen, 1)
	c.Assert(finding.Evidence.Metrics[0].Name, qt.Equals, "revision-probability")
	c.Assert(finding.Evidence.Metrics[0].Onset, qt.Equals, 0.4)
	c.Assert(finding.Primary.Span, qt.Not(qt.Equals), unswell.Location{}.Span)
}

func TestProbabilityGateKeepsExpectedAbstentionsPassing(t *testing.T) {
	c := qt.New(t)
	// fail_at 0.9 is above every fixture estimate, so only abstentions can fail.
	engine, err := unswell.New(unswell.Options{Config: []byte(requirePolicy), Model: acceptedPack(c, nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	statuses := packStatuses(result.Assessments, "sentence")
	c.Assert(statuses, qt.Contains, probability.StatusInsufficientEvidence)
	c.Assert(packStatuses(result.Assessments, "paragraph"), qt.DeepEquals,
		[]string{probability.StatusUnsupportedUnit, probability.StatusUnsupportedUnit})
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Gate.Reasons, qt.HasLen, 0)
}

func TestProbabilityGateFailsOnAnUnexplainedAbsence(t *testing.T) {
	c := qt.New(t)
	incompatible := func(f *probability.File) { f.Contract.PreparationHash = fmt.Sprintf("%064d", 1) }
	engine, err := unswell.New(unswell.Options{Config: []byte(requirePolicy), Model: acceptedPack(c, incompatible)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	index := slices.IndexFunc(result.Findings,
		func(f unswell.Finding) bool { return f.RuleID == "gate.sentence-probability-unavailable" })
	c.Assert(index >= 0, qt.IsTrue)
	c.Assert(result.Findings[index].Message, qt.Contains, probability.StatusIncompatible)
	c.Assert(result.Findings[index].Evidence.Metrics, qt.HasLen, 0)
	// The same incompatibility is an operational failure under on_incompatible: fail.
	strict := "version: 1\ncalibration:\n  model: pack\n  on_incompatible: fail\n" +
		"gate:\n  probability:\n    fail_at: 0.9\n    require: true\n"
	failing, err := unswell.New(unswell.Options{Config: []byte(strict), Model: acceptedPack(c, incompatible)})
	c.Assert(err, qt.IsNil)
	incomplete, err := failing.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNotNil)
	c.Assert(incomplete.Status, qt.Equals, "incomplete")
	c.Assert(incomplete.Gate.Passed, qt.IsFalse)
}

func TestProbabilityGateFailsOnAScoreOutsideCalibration(t *testing.T) {
	c := qt.New(t)
	// Narrow knots leave every fixture score outside the fitted range.
	narrow := func(f *probability.File) {
		f.Calibration.Scores, f.Calibration.Responses = []float64{-0.001, 0}, []float64{0.4, 0.5}
	}
	engine, err := unswell.New(unswell.Options{Config: []byte(requirePolicy), Model: acceptedPack(c, narrow)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(packStatuses(result.Assessments, "sentence"), qt.Contains, probability.StatusCalibrationRange)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	index := slices.IndexFunc(result.Findings,
		func(f unswell.Finding) bool { return f.RuleID == "gate.sentence-probability-unavailable" })
	c.Assert(index >= 0, qt.IsTrue)
	c.Assert(result.Findings[index].Message, qt.Contains, probability.StatusCalibrationRange)
}

func TestProbabilityGateRequiresAnAcceptedPack(t *testing.T) {
	for _, row := range []struct{ name, policy, message string }{
		{"no model", "version: 1\ngate:\n  probability:\n    fail_at: 0.5\n", ".*requires calibration.model: pack.*"},
		{"experimental", "version: 1\ncalibration:\n  model: pack\n  accept_experimental: true\n" +
			"gate:\n  probability:\n    fail_at: 0.5\n", ".*requires an accepted pack.*"},
		{"zero threshold", "version: 1\ncalibration:\n  model: pack\ngate:\n  probability:\n    fail_at: 0\n",
			".*fail_at must be greater than 0 and at most 1.*"},
		{"above one", "version: 1\ncalibration:\n  model: pack\ngate:\n  probability:\n    fail_at: 1.5\n",
			".*fail_at must be greater than 0 and at most 1.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(row.policy), Model: acceptedPack(c, nil)})
			c.Assert(err, qt.ErrorMatches, row.message)
			c.Assert(engine, qt.IsNil)
		})
	}
}

func TestProbabilityGateRefusesAnExperimentalPack(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(acceptedPolicy), Model: packFixture(c, "sentence", nil)})
	c.Assert(err, qt.ErrorMatches, ".*declares experimental status.*")
	c.Assert(engine, qt.IsNil)
}

func TestProbabilityGateKeepsNoGateAndTheIndexSeparate(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(acceptedPolicy), NoGate: true,
		Model: acceptedPack(c, nil)})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), packSource())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(gateCodes(result), qt.Contains, "gate.sentence-probability")
	for _, assessment := range result.Assessments {
		c.Assert(assessment.EffectiveSlopScore < 80, qt.IsTrue)
	}
}
