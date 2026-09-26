package unswell_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

//go:embed research/reviews/2026-09-26-evidence-claims/cases.json
var verificationAssuranceCases []byte

func TestVerificationAssuranceConstructions(t *testing.T) {
	var cases []struct {
		ID    string `json:"id"`
		Text  string `json:"text"`
		Match bool   `json:"match"`
	}
	c := qt.New(t)
	c.Assert(json.Unmarshal(verificationAssuranceCases, &cases), qt.IsNil)
	engine := singleRuleEngine(t, "filler.unscoped-assurance", "", "")
	for _, row := range cases {
		t.Run(row.ID, func(t *testing.T) {
			c := qt.New(t)
			// Keep a fully protected candidate a valid scan without turning its
			// code into prose. The neutral paragraph is separated from the claim.
			text := "The client opens a connection.\n\n" + row.Text
			result, err := engine.Analyze(t.Context(), document.Source{
				Name: "claim.md", Format: document.Markdown, Bytes: []byte(text),
			})
			c.Assert(err, qt.IsNil)
			count := 0
			if row.Match {
				count = 1
			}
			c.Assert(result.Findings, qt.HasLen, count, qt.Commentf("%s", row.Text))
			for _, finding := range result.Findings {
				span := finding.Primary.Span
				c.Assert(text[span.Start:span.End], qt.Equals, finding.Primary.Snippet)
			}
		})
	}
}

func TestVerificationAssuranceSourceAndAllowance(t *testing.T) {
	const id = "filler.unscoped-assurance"
	const claim = "Reliability is **ensured** by thorough testing"
	text := "Résumé.\r\n\r\n" + claim + ".\r\n"
	result := singleRuleResult(t, id, text, "", "")
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, claim)
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, claim))
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}
