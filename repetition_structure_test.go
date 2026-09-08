package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestParagraphOverlapClustersAndWindows(t *testing.T) {
	c := qt.New(t)
	second := strings.Replace(overlapParagraph, "opens", "creates", 1)
	third := strings.Replace(overlapParagraph, "opens", "starts", 1)
	text := overlapParagraph + "\n\n" + second + "\n\n" + third
	result := repetitionResult(t, "repetition.paragraph-overlap", text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Related, qt.HasLen, 2)
	c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, float64(3))
	c.Assert(finding.Evidence.Metrics[1].Value, qt.Equals, float64(3))
	c.Assert(finding.Primary.Snippet, qt.Equals, overlapParagraph)
	c.Assert(finding.Related[0].Snippet, qt.Equals, second)
	c.Assert(finding.Related[1].Snippet, qt.Equals, third)
	clean := repetitionResult(t, "repetition.paragraph-overlap", text+"\n\nThe reader checks the configuration.", "", "")
	c.Assert(clean.Findings, qt.DeepEquals, result.Findings)
	c.Assert(clean.Assessments[:len(result.Assessments)], qt.DeepEquals, result.Assessments)
	c.Assert(repetitionResult(t, "repetition.paragraph-overlap", text, "{allowed_occurrences: 3}", "").Findings, qt.HasLen, 0)
	outside := overlapParagraph + "\n\n" + strings.Repeat("The reader waits.\n\n", 8) + second
	c.Assert(repetitionResult(t, "repetition.paragraph-overlap", outside, "", "").Findings, qt.HasLen, 0)
}

func TestHeadingEchoUsesSelectedStructure(t *testing.T) {
	const title = "A practical approach to the delivery process"
	for _, row := range []struct {
		name, text, extra string
		want              int
	}{
		{"ATX", "# " + title + "\n\n" + title + ".", "", 1},
		{"setext", title + "\n===\n\n" + title + ".", "", 1},
		{"excluded heading", "# " + title + "\n\n" + title + ".", "extraction: {contexts: [paragraph]}\n", 0},
		{"later paragraph", "# " + title + "\n\nThe reader checks the configuration.\n\n" + title + ".", "", 0},
		{"code", "# " + title + "\n\n```go\nconst Value = 1\n```\n\n" + title + ".", "", 0},
		{"definition", "# Retry budget\n\nThe retry budget limits the total time spent reconnecting after a transport failure.", "", 0},
		{"changed modality", "# The client may retry requests after a transport failure\n\n" +
			"The client must retry requests after a transport failure.", "", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := repetitionResult(t, "repetition.heading-echo", row.text, "", row.extra)
			c.Assert(result.Findings, qt.HasLen, row.want)
			if row.want > 0 {
				c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, title+".")
				c.Assert(result.Findings[0].Related[0].Snippet, qt.Equals, title)
			}
		})
	}
}

func TestSummaryEchoRequiresExplicitSelectedScope(t *testing.T) {
	second := strings.Replace(overlapParagraph, "opens", "creates", 1)
	for _, row := range []struct {
		name, middle, parameters, extra string
		want                            int
	}{
		{"summary", "## Summary", "", "", 1},
		{"case folding", "## SUMMARY", "", "", 1},
		{"nested section", "## Summary\n\n### Delivery", "", "", 1},
		{"setext", "Summary\n-------", "", "", 1},
		{"configured title", "## Recap", "{phrases: [recap]}", "", 1},
		{"title is not a prefix", "## Summary of installation", "", "", 0},
		{"outside scope", "## Summary\n\nThe reader checks the configuration.\n\n## Details", "", "", 0},
		{"excluded heading", "## Summary", "", "extraction: {contexts: [paragraph]}\n", 0},
		{"plain paragraph", "Summary", "", "", 0},
		{"unconfigured title", "## Recap", "", "", 0},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			result := repetitionResult(t, "repetition.summary-echo", overlapParagraph+"\n\n"+row.middle+"\n\n"+second,
				row.parameters, row.extra)
			c.Assert(result.Findings, qt.HasLen, row.want)
			if row.want > 0 {
				c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, second)
				c.Assert(result.Findings[0].Related[0].Snippet, qt.Equals, overlapParagraph)
			}
		})
	}
	c := qt.New(t)
	future := "## Summary\n\n" + second + "\n\n## Details\n\n" + overlapParagraph
	c.Assert(repetitionResult(t, "repetition.summary-echo", future, "", "").Findings, qt.HasLen, 0)
}

func TestOverlapTermsAndProtectedContent(t *testing.T) {
	c := qt.New(t)
	text := overlapParagraph + "\n\n" + strings.Replace(overlapParagraph, "opens", "creates", 1)
	code := strings.ReplaceAll(text, "server", "`server`")
	c.Assert(repetitionResult(t, "repetition.paragraph-overlap", code, "", "").Findings, qt.HasLen, 0)
	terms := "vocabulary:\n  terms: [a connection to the server and sends the request with its credentials]\n" +
		"  term_exemptions: [repetition.paragraph-overlap]\n"
	c.Assert(repetitionResult(t, "repetition.paragraph-overlap", text, "", terms).Findings, qt.HasLen, 0)
}
