package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestExactSentenceRepeatsRetainOpaqueOperands(t *testing.T) {
	const first = "It is an\nenvironment variable rather than a flag because `ptah-compat` registers exactly\n" +
		"the flags the Atlas community CLI registers."
	const second = "It is an environment variable rather than a flag because `ptah-compat`\n" +
		"registers exactly the flags the Atlas community CLI registers."
	c := qt.New(t)
	text := "## Roles and grants\n\n" + first + "\n\n" + second
	result := singleRuleResult(t, "repetition.exact-sentence", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.RuleVersion, qt.Equals, "3")
	c.Assert(finding.Evidence.Kind, qt.Equals, "exact")
	c.Assert(finding.Primary.Snippet, qt.Equals, first)
	c.Assert(finding.Related, qt.HasLen, 1)
	c.Assert(finding.Related[0].Snippet, qt.Equals, second)
	c.Assert(text[finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, first)
	c.Assert(text[finding.Related[0].Span.Start:finding.Related[0].Span.End], qt.Equals, second)
	c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, 2.0)
}

func TestExactSentenceOpaqueIdentityKeepsDistinctValues(t *testing.T) {
	const prefix = "The client reads the value from "
	const suffix = " before it sends the request to the configured server."
	for _, row := range []struct {
		name, left, right string
		want              int
	}{
		{"same identifier", "`LEFT`", "`LEFT`", 1},
		{"different identifier", "`LEFT`", "`RIGHT`", 0},
		{"case sensitive identifier", "`LEFT`", "`left`", 0},
		{"different number", "`20`", "`30`", 0},
		{"significant code whitespace", "`a  b`", "`a b`", 0},
		{"same URI", "https://example.test/a", "https://example.test/a", 1},
		{"different URI", "https://example.test/a", "https://example.test/b", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			text := prefix + row.left + suffix + "\n\n" + prefix + row.right + suffix
			result := singleRuleResult(t, "repetition.exact-sentence", text, "", "")
			c.Assert(result.Findings, qt.HasLen, row.want)
		})
	}
}

func TestExactSentenceOpaqueOperandsKeepIndependentScopes(t *testing.T) {
	const sentence = "The client reads the value from `LEFT` before it sends the request to the configured server."
	for _, text := range []string{
		"## First\n\n" + sentence + "\n\n## Second\n\n" + sentence,
		"## Notes\n\n" + sentence + "\n\n## Notes\n\n" + sentence,
		"| API | Behavior |\n| --- | --- |\n| A | " + sentence + " |\n| B | " + sentence + " |",
		"If enabled, " + sentence + "\n\nIf disabled, " + sentence,
		"The client reports status.\n\n```text\n" + sentence + "\n" + sentence + "\n```",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, "repetition.exact-sentence", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestExactSentenceOpaqueRangesPreserveUnicodeAndCRLF(t *testing.T) {
	const first = "The Café client reads `résumé` before it sends the request to the configured server."
	const second = "The Café client reads `résumé` before it sends\r\nthe request to the configured server."
	text := "\ufeff# Client\r\n\r\n" + first + "\r\n\r\n" + second
	c := qt.New(t)
	result := singleRuleResult(t, "repetition.exact-sentence", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, first))
	c.Assert(finding.Primary.Snippet, qt.Equals, first)
	c.Assert(finding.Related, qt.HasLen, 1)
	c.Assert(finding.Related[0].Span.Start, qt.Equals, strings.Index(text, second))
	c.Assert(finding.Related[0].Snippet, qt.Equals, second)
}

func TestExactSentenceOpaqueIdentityAbstainsWithinBudget(t *testing.T) {
	c := qt.New(t)
	engine := singleRuleEngine(t, "repetition.exact-sentence", "", "analysis: {max_candidates: 500}\n")
	sentence := "The client reads the value from `" + strings.Repeat("x", 1000) +
		"` before it sends the request to the configured server."
	const plain = "The client opens a connection to the server and sends the request with its credentials."
	result, err := engine.Analyze(t.Context(), document.Source{
		Name: "guide.md", Format: document.Markdown,
		Bytes: []byte(plain + "\n\n" + plain + "\n\n" + sentence + "\n\n" + sentence),
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Findings, qt.HasLen, 0)
	assertBudgetAbstention(t, result, "guide.md", "repetition.exact-sentence", "max_candidates")
}
