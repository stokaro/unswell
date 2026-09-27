package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestAdjacentWordsKeepSourceRangesAroundOperands(t *testing.T) {
	engine := singleRuleEngine(t, "repetition.adjacent-word", "", "")
	for _, row := range []struct {
		name   string
		format document.Format
		text   string
	}{
		{"guide.md", document.Markdown, "\ufeff# Café\r\n\r\nThe AND and OR parser returns %s %s when\r\nwhen values are 1 1."},
		{"message.go", document.Go, "package example\r\nvar message = \"unit %s has unavailable %s %s and requires requires a parser\"\r\n"},
		{"table.md", document.Markdown, "| Values | Explanation |\r\n| --- | --- |\r\n| 1 1 | The scanner requires requires a parser. |\r\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := engine.Analyze(t.Context(), document.Source{Name: row.name, Format: row.format, Bytes: []byte(row.text)})
			c.Assert(err, qt.IsNil)
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			word := "requires"
			if row.name == "guide.md" {
				word = "when"
			}
			c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(row.text, word))
			c.Assert(finding.Primary.Span.End, qt.Equals, strings.Index(row.text, word)+len(word))
			c.Assert(finding.Related, qt.HasLen, 1)
			c.Assert(finding.Related[0].Span.Start, qt.Equals, strings.LastIndex(row.text, word))
			c.Assert(finding.Related[0].Span.End, qt.Equals, strings.LastIndex(row.text, word)+len(word))
		})
	}
}
