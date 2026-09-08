package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

const suppressionPolicy = `version: 1
extends: [builtin:custom]
rules:
  policy.banned-phrases:
    enabled: true
    parameters: {phrases: [robust]}
    score: {weight: 20, cap: 20}
`

func TestRepeatedIdenticalMatchesHaveIndependentSuppressionTraces(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy)})
	c.Assert(err, qt.IsNil)
	text := "<!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. -->\n\n" +
		"The robust client retries. The robust client retries."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 2)
	c.Assert(result.Findings[0].ID, qt.Not(qt.Equals), result.Findings[1].ID)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Findings[1].Suppressed, qt.IsFalse)
	c.Assert(result.Gate.Reasons, qt.HasLen, 1)
	c.Assert(result.Gate.Reasons[0].FindingID, qt.Equals, result.Findings[1].ID)
	c.Assert(result.Assessments[0].EffectiveSlopScore, qt.Equals, float64(20))
}

func TestSuppressionPreservesRawScoresAndOtherFindings(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy), IncludeSource: true})
	c.Assert(err, qt.IsNil)
	source := "<!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. -->\n\n" +
		"The robust estimator works. The robust client retries."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(source)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 2)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Findings[1].Suppressed, qt.IsFalse)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Gate.Reasons, qt.HasLen, 1)
	c.Assert(result.Suppressions, qt.HasLen, 1)
	c.Assert(result.Suppressions[0].Status, qt.Equals, "used")
	c.Assert(result.Suppressions[0].Reason, qt.Equals, "Required contract wording.")
	c.Assert(result.Suppressions[0].Targets[0].Scope, qt.Equals, "sentence")
	c.Assert(result.Findings[0].SuppressionIDs, qt.DeepEquals, []string{result.Suppressions[0].ID})
	c.Assert(result.Documents[0].Source, qt.Equals, source)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.SlopScore, qt.Equals, float64(20))
		if assessment.Scope == "sentence" && assessment.UnitID == 0 {
			c.Assert(assessment.EffectiveSlopScore, qt.Equals, float64(0))
			c.Assert(assessment.EffectiveContributions[0].Reason, qt.Equals, "source-suppression")
		} else {
			c.Assert(assessment.EffectiveSlopScore, qt.Equals, float64(20))
		}
	}
}

func TestInlineDirectiveTargetsTheFollowingSentence(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy)})
	c.Assert(err, qt.IsNil)
	text := "A clean sentence. <!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. --> " +
		"The **robust** client starts."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsTrue)
}

func TestSuppressionRegionsAndGoCommentTargets(t *testing.T) {
	cases := []struct {
		name, text           string
		format               document.Format
		findings, suppressed int
	}{
		{"region", "<!-- unswell-disable policy.banned-phrases -- Required contract wording. -->\n\n" +
			"The robust estimator works.\n\nThe robust client retries.\n\n" +
			"<!-- unswell-enable policy.banned-phrases -->\n", document.Markdown, 2, 2},
		{"go-comment", "package sample\n// unswell-disable-next-block policy.banned-phrases -- Required contract wording.\n" +
			"// The robust estimator works.\nvar message = \"The robust client retries.\"\n", document.Go, 2, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy)})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: tc.name, Format: tc.format, Bytes: []byte(tc.text)})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, tc.findings)
			count := 0
			for _, finding := range result.Findings {
				if finding.Suppressed {
					count++
				}
			}
			c.Assert(count, qt.Equals, tc.suppressed)
			c.Assert(result.Gate.Passed, qt.Equals, tc.findings == tc.suppressed)
		})
	}
}

func TestInvalidSuppressionsCannotPass(t *testing.T) {
	for _, directive := range []string{
		"unswell-disable-next-block policy.banned-phases -- Required contract wording.",
		"unswell-disable-next-blok policy.banned-phrases -- Required contract wording.",
		"unswell-disable-next-block policy.banned-phrases",
		"unswell-disable-next-block policy.banned-phrases -- TODO",
		"unswell-disable-next-block all -- Required contract wording.",
		"unswell-disable policy.banned-phrases -- Required contract wording.",
		"unswell-enable policy.banned-phrases",
		"unswell-disable-file policy.banned-phrases -- Required contract wording.",
	} {
		t.Run(directive, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy), NoGate: true})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
				Bytes: []byte("<!-- " + directive + " -->\n\nThe robust client starts.")})
			c.Assert(err, qt.IsNotNil)
			c.Assert(result.Status, qt.Equals, "incomplete")
			c.Assert(result.Gate.Passed, qt.IsFalse)
		})
	}
}

func TestUnusedAndPartialSuppressionPreservesEvidence(t *testing.T) {
	c := qt.New(t)
	policy := suppressionPolicy + "  repetition.exact-sentence: {enabled: true, parameters: {min_words: 1}}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	source := "<!-- unswell-disable-next-block repetition.exact-sentence -- Required repeated definition. -->\n\n" +
		"The robust client starts.\n\nThe robust client starts."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(source)})
	c.Assert(err, qt.ErrorMatches, ".*unused suppression.*")
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 3)
	for _, finding := range result.Findings {
		c.Assert(finding.Suppressed, qt.IsFalse)
	}
	c.Assert(result.Suppressions[0].Status, qt.Equals, "unused")
	source = strings.ReplaceAll(source, "robust", "small")
	_, err = engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(source)})
	c.Assert(err, qt.ErrorMatches, ".*unused suppression.*")
}

func TestDisjointPermissionsCoverOneRepetitionCluster(t *testing.T) {
	c := qt.New(t)
	policy := "version: 1\nextends: [builtin:custom]\nrules:\n" +
		"  repetition.exact-sentence: {enabled: true, parameters: {min_words: 1}}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	block := "<!-- unswell-disable-next-block repetition.exact-sentence -- Required repeated definition. -->\n\n" +
		"The client starts.\n\n"
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(block + block)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	c.Assert(result.Findings[0].SuppressionIDs, qt.HasLen, 2)
	c.Assert(result.Suppressions, qt.HasLen, 2)
	for _, record := range result.Suppressions {
		c.Assert(record.Status, qt.Equals, "used")
		c.Assert(record.FindingIDs, qt.DeepEquals, []string{result.Findings[0].ID})
	}
	c.Assert(result.Gate.Passed, qt.IsTrue)
}

func TestNestedRegionsPreserveOtherRuleGates(t *testing.T) {
	c := qt.New(t)
	policy := suppressionPolicy + "  repetition.exact-sentence: {enabled: true, parameters: {min_words: 1}}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	text := "<!-- unswell-disable repetition.exact-sentence -- Required repeated definition. -->\n\n" +
		"The robust client starts.\n\n" +
		"<!-- unswell-disable policy.banned-phrases -- Required contract wording. -->\n\n" +
		"The robust client starts.\n\n" +
		"<!-- unswell-enable policy.banned-phrases -->\n\n<!-- unswell-enable repetition.exact-sentence -->\n"
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 3)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Gate.Reasons, qt.HasLen, 1)
	c.Assert(result.Suppressions, qt.HasLen, 2)
	for _, record := range result.Suppressions {
		c.Assert(record.Status, qt.Equals, "used")
		c.Assert(record.End, qt.IsNotNil)
	}
}

func TestInlineDirectiveDoesNotPermitEarlierProseInItsSentence(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Config: []byte(suppressionPolicy)})
	c.Assert(err, qt.IsNil)
	text := "The <!-- unswell-disable-next-sentence policy.banned-phrases -- Required contract wording. --> " +
		"robust client starts. The robust client retries."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 2)
	c.Assert(result.Findings[0].Suppressed, qt.IsFalse)
	c.Assert(result.Findings[1].Suppressed, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsFalse)
}

func TestOptionalReasonAndUnusedPoliciesKeepAuditRecords(t *testing.T) {
	c := qt.New(t)
	policy := suppressionPolicy + "suppressions: {require_reason: false, reject_unused: false}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	for _, word := range []string{"small", "robust"} {
		text := "<!-- unswell-disable-next-block policy.banned-phrases -->\n\nThe " + word + " client starts."
		result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
		c.Assert(err, qt.IsNil)
		c.Assert(result.Suppressions, qt.HasLen, 1)
		c.Assert(result.Suppressions[0].Reason, qt.Equals, "")
		c.Assert(result.Gate.Passed, qt.IsTrue)
		if word == "small" {
			c.Assert(result.Suppressions[0].Status, qt.Equals, "unused")
		} else {
			c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
		}
	}
}

func TestSuppressionRecomputesPreviouslyMaskedEvidence(t *testing.T) {
	c := qt.New(t)
	policy := strings.ReplaceAll(suppressionPolicy, "weight: 20, cap: 20", "weight: 30, cap: 30") +
		"  filler.wordy-phrase:\n    enabled: true\n    parameters: {phrases: [robust]}\n    score: {weight: 20, cap: 20}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	text := "<!-- unswell-disable-next-block policy.banned-phrases -- Required contract wording. -->\n\nThe robust client starts."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 2)
	for _, assessment := range result.Assessments {
		c.Assert(assessment.SlopScore, qt.Equals, float64(30))
		c.Assert(assessment.EffectiveSlopScore, qt.Equals, float64(20))
		c.Assert(assessment.Contributions[1].Effective, qt.Equals, float64(0))
		c.Assert(assessment.EffectiveContributions[0].RuleID, qt.Equals, "filler.wordy-phrase")
		c.Assert(assessment.EffectiveContributions[0].Effective, qt.Equals, float64(20))
	}
}

func TestDisabledCommentProseDoesNotHideMalformedDirectives(t *testing.T) {
	c := qt.New(t)
	policy := suppressionPolicy + "extraction: {contexts: [string]}\n"
	engine, err := unswell.New(unswell.Options{Config: []byte(policy)})
	c.Assert(err, qt.IsNil)
	for _, format := range []document.Format{document.Go, document.JavaScript} {
		text := "// unswell-disable-next-blok policy.banned-phrases -- Required contract wording.\n" +
			"var message = \"The client starts.\"\n"
		if format == document.Go {
			text = "package sample\n" + text
		}
		result, err := engine.Analyze(t.Context(), document.Source{Name: "sample", Format: format, Bytes: []byte(text)})
		c.Assert(err, qt.ErrorMatches, "(?s).*unknown suppression command.*")
		c.Assert(result.Manifest.Complete, qt.IsFalse)
		c.Assert(result.Gate.Passed, qt.IsFalse)
	}
}
