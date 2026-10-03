package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestVerificationAssuranceExplainsTheRecognizedClaim(t *testing.T) {
	for _, row := range []struct{ text, message, advice string }{
		{"Reliability is ensured by rigorous testing.", "evidence is presented as a quality guarantee", "Keep the checks, links"},
		{"Extensive reviews guarantee security.", "evidence is presented as a quality guarantee", "concrete criteria"},
		{"Nothing is taken on trust.", "unrestricted trust assurance", "Name the assumptions"},
		{"We accept nothing on faith.", "unrestricted trust assurance", "conditions, limitations"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "filler.unscoped-assurance", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Message, qt.Contains, row.message)
			c.Assert(finding.Evidence.Message, qt.Equals, finding.Message)
			c.Assert(finding.Evidence.Suggestion, qt.Contains, row.advice)
			c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimSuffix(row.text, "."))
		})
	}
}

func TestVerificationTrustAssuranceKeepsMechanismContext(t *testing.T) {
	c := qt.New(t)
	text := "The proxy's FATAL is kept, so nothing is taken on trust."
	result := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Primary.Snippet, qt.Equals, "nothing is taken on trust")
	c.Assert(finding.Message, qt.Equals,
		"This wording makes an unrestricted trust assurance; state the assumptions and limits of verification.")
	c.Assert(finding.Evidence.Suggestion, qt.Contains, "Preserve their conditions")
}

func TestVerificationAssuranceGroupedExplanations(t *testing.T) {
	const generalMessage = "State the behavior, scope or evidence behind this quality claim."
	for _, row := range []struct{ name, text, message string }{
		{"testing", "Reliability is ensured by rigorous testing. Security is guaranteed by automated testing.",
			"Test or review evidence is presented as a quality guarantee; distinguish what it establishes from the broader claim."},
		{"trust", "Nothing is taken on trust. We accept nothing on faith.",
			"This wording makes an unrestricted trust assurance; state the assumptions and limits of verification."},
		{"testing then trust", "Reliability is ensured by rigorous testing. Nothing is taken on trust.", generalMessage},
		{"trust then testing", "Nothing is taken on trust. Reliability is ensured by rigorous testing.", generalMessage},
		{"testing then other quality", "Reliability is ensured by rigorous testing. The configuration is intentionally explicit.",
			generalMessage},
		{"other quality then testing", "The configuration is intentionally explicit. Reliability is ensured by rigorous testing.",
			generalMessage},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "filler.unscoped-assurance", row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Message, qt.Equals, row.message)
			c.Assert(finding.Related, qt.HasLen, 1)
			c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, 2.0)
			if row.message == generalMessage {
				c.Assert(finding.Evidence.Message, qt.Equals, "")
				c.Assert(finding.Evidence.Suggestion, qt.Equals,
					"Keep the technical conditions. Replace a bare quality announcement, nearly-always-right judgment, "+
						"or universal learning promise with specific behavior.")
			}
		})
	}
}

func TestVerificationAssuranceKeepsWindowAllowances(t *testing.T) {
	const id = "filler.unscoped-assurance"
	const one = "Reliability is ensured by rigorous testing."
	const mixed = one + " Nothing is taken on trust."
	parameters := "{allowed_occurrences: 1, saturation_occurrences: 2}"
	c := qt.New(t)
	c.Assert(singleRuleResult(t, id, one, parameters, "").Findings, qt.HasLen, 0)
	result := singleRuleResult(t, id, mixed, parameters, "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Metrics[0].Value, qt.Equals, 2.0)
	c.Assert(result.Findings[0].Related, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Message, qt.Equals, "")
}

func TestVerificationExplanationsStayWithinTheirBlocks(t *testing.T) {
	text := "Reliability is ensured by rigorous testing.\n\nThe configuration is intentionally explicit."
	result := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 2)
	c.Assert(result.Findings[0].Message, qt.Contains, "quality guarantee")
	c.Assert(result.Findings[1].Message, qt.Equals, "State the behavior, scope or evidence behind this quality claim.")
	for _, finding := range result.Findings {
		c.Assert(finding.Related, qt.HasLen, 0)
	}
}

func TestVerificationExplanationSourceMapping(t *testing.T) {
	const id = "filler.unscoped-assurance"
	const plain = "Reliability is ensured by automated testing."
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
			"\ufeffRésumé.\r\n\r\nReliability is **ensured** by [**rigorous testing**](https://example.test/tests).\r\n")},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + plain + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + plain + "'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			c := qt.New(t)
			result, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Message, qt.Contains, "quality guarantee")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "links and measured results")
			c.Assert(string(source.Bytes[finding.Primary.Span.Start:finding.Primary.Span.End]), qt.Equals, finding.Primary.Snippet)
			for _, span := range finding.Primary.Segments {
				c.Assert(span.Start >= finding.Primary.Span.Start && span.End <= finding.Primary.Span.End, qt.IsTrue)
			}
		})
	}
}
