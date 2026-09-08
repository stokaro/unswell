package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

const fragmentedLists = "The release offers these benefits.\n\n- Clear reports\n- Useful summaries\n\n" +
	"Its output has these traits.\n\n- Simple navigation\n- Consistent terminology\n\n" +
	"The workflow has these qualities.\n\n- Quick checks\n- Direct feedback"

func TestShortListsUseGrammarOwnership(t *testing.T) {
	c := qt.New(t)
	text := "\ufeff" + strings.ReplaceAll(fragmentedLists, "\n", "\r\n")
	result := singleRuleResult(t, "format.list-fragmentation", text, "", "")
	c.Assert(surfaceMetric(t, result, "short-unordered-lists"), qt.Equals, float64(3))
	c.Assert(surfaceMetric(t, result, "short-list-items"), qt.Equals, float64(6))
	c.Assert(surfaceMetric(t, result, "list-prose-words"), qt.Equals, float64(12))
	c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "Clear reports")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, strings.Index(text, "Clear reports"))
	c.Assert(result.Findings[0].Related, qt.HasLen, 5)
}

func TestShortListsExcludeStructuredInstructions(t *testing.T) {
	for _, row := range []struct{ name, text, parameters, extra string }{
		{"one list", "- Clear reports\n- Useful summaries\n- Simple navigation\n- Quick checks", "", ""},
		{"ordered", strings.ReplaceAll(fragmentedLists, "- ", "1. "), "", ""},
		{"tasks", strings.ReplaceAll(fragmentedLists, "- ", "- [ ] "), "", ""},
		{"nested", strings.ReplaceAll(fragmentedLists, "- Useful summaries", "  - Useful summaries"), "", ""},
		{"full sentences", strings.ReplaceAll(fragmentedLists, "reports", "reports."), "", ""},
		{"procedure", strings.ReplaceAll(fragmentedLists, "Clear reports", "Read reports"), "", ""},
		{"reference", "# API Reference\n\n" + fragmentedLists, "", ""},
		{"identifier", strings.ReplaceAll(fragmentedLists, "Clear reports", "RetryStatus reports"), "", ""},
		{"code item", strings.ReplaceAll(fragmentedLists, "Clear reports", "`Clear reports`"), "", ""},
		{"multiple paragraphs", strings.ReplaceAll(fragmentedLists, "- Useful summaries", "\n  More details.\n\n- Useful summaries"), "", ""},
		{"excluded gap", strings.ReplaceAll(fragmentedLists, "Its output", "```go\nx()\n```\n\nIts output"), "", ""},
		{"heading boundary", strings.ReplaceAll(fragmentedLists, "Its output", "# Output\n\nIts output"), "", ""},
		{"window", fragmentedLists, "{window_blocks: 3}", ""},
		{"selected paragraphs", fragmentedLists, "", "extraction:\n  contexts: [paragraph]\n"},
		{"approved term", fragmentedLists, "", "vocabulary:\n  terms: [Clear reports]\n  term_exemptions: [format.list-fragmentation]\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(singleRuleResult(t, "format.list-fragmentation", row.text, row.parameters, row.extra).Findings, qt.HasLen, 0)
		})
	}
}
