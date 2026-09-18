package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestAdjacentWordsOutsideQuotations(t *testing.T) {
	for _, text := range []string{
		`The library was named "ImGui" when when the maintainer released it.`,
		`The maintainer must must release "ImGui" today.`,
		`The library was named 'ImGui' when when the maintainer released it.`,
		`The library was named “ImGui” when when the maintainer released it.`,
		`The library was named ‘ImGui’ when when the maintainer released it.`,
		`The users' client must must retry.`,
		`The users’ client must must retry.`,
		`The user's client can't retry retry here.`,
		"The library was named `\"ImGui` when when the maintainer released it.",
		`The maintainer must must check "an unfinished quotation.`,
		`The manual says "First sentence. See See the report. Last sentence." See See the notes.`,
		`The manual says “First sentence. See See the report. Last sentence.” See See the notes.`,
		`The manual says 'First sentence. See See the report. Last sentence.' See See the notes.`,
		`The manual says ‘First sentence. See See the report. Last sentence.’ See See the notes.`,
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "repetition.adjacent-word", text, "", "")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].RuleVersion, qt.Equals, "2")
			c.Assert(result.Findings[0].Related, qt.HasLen, 1)
		})
	}
}

func TestAdjacentWordsPennQuotesAndPolicy(t *testing.T) {
	c := qt.New(t)
	engine := singleRuleEngine(t, "repetition.adjacent-word", "", "")
	text := "The manual says ``First sentence. See See the report.'' See See the notes."
	result, err := engine.Analyze(t.Context(), document.Source{Name: "quote.txt", Format: document.Plain, Bytes: []byte(text)})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.LastIndex(text, "See See"))
	const draft = `The library was named "ImGui" when when the maintainer released it.`
	terms := "vocabulary:\n  terms: [when when]\n  term_exemptions: [repetition.adjacent-word]\n"
	c.Assert(singleRuleResult(t, "repetition.adjacent-word", draft, "", terms).Findings, qt.HasLen, 0)
	suppressed := "<!-- unswell-disable-next-block repetition.adjacent-word -- Required example wording. -->\n\n" + draft
	result = singleRuleResult(t, "repetition.adjacent-word", suppressed, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Suppressed, qt.IsTrue)
	limited := singleRuleResult(t, "repetition.adjacent-word", draft, "", "analysis: {max_candidates: 1}\n")
	assertBudgetAbstention(t, limited, "guide.md", "repetition.adjacent-word", "max_candidates")
}

func TestAdjacentWordsInsideQuotationsRemainProtected(t *testing.T) {
	for _, text := range []string{
		`The manual quotes "when when" as the invalid input.`,
		`The manual quotes 'when when' as the invalid input.`,
		`The manual quotes “when when” as the invalid input.`,
		`The manual quotes ‘when when’ as the invalid input.`,
		`The manual says "First sentence. See See the report. Last sentence."`,
		`The manual says 'First sentence. See See the report. Last sentence.'`,
		`The manual says ‘First sentence. See See the report. Last sentence.’`,
		`The manual says "The 'ImGui' client must must retry."`,
		`The manual says 'The "ImGui" client must must retry.'`,
		`The manual says 'The user's client must must retry.'`,
		`The manual says 'The users' client must must retry.'`,
		`The manual says ‘The users’ client must must retry.’`,
		`The manual says "First sentence. See See the report.`,
		"The manual says \"Ignore `\"` here. See See the report.\"",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "repetition.adjacent-word", text, "", "")
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 0)
		})
	}
}

func TestAdjacentQuotationScopeKeepsSourceSpansAndBlockBoundaries(t *testing.T) {
	c := qt.New(t)
	const text = "\ufeff# Café\r\n\r\nThe manual starts an unfinished \"quotation.\r\n\r\n" +
		"The library was named “ImGui” when\r\nwhen the maintainer released it."
	result := singleRuleResult(t, "repetition.adjacent-word", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, "when"))
	c.Assert(finding.Related[0].Span.Start, qt.Equals, strings.LastIndex(text, "when"))
	for _, span := range append(finding.Primary.Segments, finding.Related[0].Segments...) {
		c.Assert(text[span.Start:span.End], qt.Equals, "when")
	}
}
