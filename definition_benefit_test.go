package unswell_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

type definitionCase struct {
	ID           string `json:"id"`
	Text         string `json:"text"`
	Expected     bool   `json:"expected_definition_benefit_restatement"`
	Propositions []struct {
		Start int    `json:"start"`
		End   int    `json:"end"`
		Quote string `json:"quote"`
	} `json:"propositions"`
}

func definitionCases(t *testing.T) []definitionCase {
	t.Helper()
	c := qt.New(t)
	raw, err := os.ReadFile("testdata/definition-benefit/controls.json")
	c.Assert(err, qt.IsNil)
	var fixture struct {
		Cases []definitionCase `json:"cases"`
	}
	c.Assert(json.Unmarshal(raw, &fixture), qt.IsNil)
	return fixture.Cases
}

func definitionFindings(result unswell.RunResult) []unswell.Finding {
	var findings []unswell.Finding
	for _, finding := range result.Findings {
		if finding.Evidence.Metrics[0].Name == "definition-benefit-restatements" {
			findings = append(findings, finding)
		}
	}
	return findings
}

func TestDefinitionBenefitControls(t *testing.T) {
	for _, row := range definitionCases(t) {
		t.Run(row.ID, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "repetition.repeated-claim", row.Text, "", "")
			findings := definitionFindings(result)
			want := 0
			if row.Expected {
				want = 1
			}
			c.Assert(findings, qt.HasLen, want)
			if want == 0 {
				return
			}
			f := findings[0]
			c.Assert(f.Related, qt.HasLen, 1)
			c.Assert(f.Evidence.Kind, qt.Equals, "heuristic")
			c.Assert(f.Primary.Span.Start, qt.Equals, row.Propositions[0].Start)
			c.Assert(f.Related[0].Span.Start, qt.Equals, row.Propositions[1].Start)
			c.Assert(f.Primary.Snippet, qt.Equals, strings.TrimSuffix(row.Propositions[0].Quote, "."))
			c.Assert(f.Related[0].Snippet, qt.Equals, strings.TrimSuffix(row.Propositions[1].Quote, "."))
		})
	}
}

func TestDefinitionBenefitRejectsUnresolvedBridge(t *testing.T) {
	text := definitionCases(t)[0].Text
	for _, row := range []struct{ name, old, replacement string }{
		{"different source", "Add it to `docs/site/src/glossary.ts`", "Add it to `other/glossary.ts`"},
		{"source case", "Add it to `docs/site/src/glossary.ts`", "Add it to `docs/site/src/Glossary.ts`"},
		{"no alias", "The list above renders from that map", "The list above renders from another map"},
		{"intervening alias", "Then link here from the pages that use it.", "A separate map stores local definitions."},
		{"other section", "## Adding a term", "## Client-specific meanings"},
		{"qualified benefit", "The definition stays in the map", "When overrides are disabled, the definition stays in the map"},
		{"changed modality", "which is what keeps", "which may keep"},
		{"changed meaning", "something else on the next", "a different spelling on the next"},
		{"bridge redefinition", "transcribed, and", "transcribed, but that map uses different definitions, and"},
		{"addition qualifier", "pins the fact.", "pins the fact, unless an override is active."},
		{"qualified bridge", "The list above renders from that map", "When overrides are disabled, the list above renders from that map"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(strings.Count(text, row.old), qt.Equals, 1)
			changed := strings.Replace(text, row.old, row.replacement, 1)
			c.Assert(definitionFindings(singleRuleResult(t, "repetition.repeated-claim", changed, "", "")), qt.HasLen, 0)
		})
	}
}

func TestDefinitionBenefitWindowAndMapping(t *testing.T) {
	text := definitionCases(t)[1].Text
	c := qt.New(t)
	c.Assert(definitionFindings(singleRuleResult(t, "repetition.repeated-claim", text, "{window_sentences: 1}", "")), qt.HasLen, 0)
	text = "\ufeff" + strings.ReplaceAll(text, "\n", "\r\n")
	text = strings.Replace(text, "same meaning", "same **meaning**", 1)
	r := definitionFindings(singleRuleResult(t, "repetition.repeated-claim", text, "", ""))
	c.Assert(r, qt.HasLen, 1)
	c.Assert(r[0].Primary.Span.Start, qt.Equals, len("\ufeff"))
	c.Assert(r[0].Primary.Snippet, qt.Contains, "same **meaning**")
	c.Assert(r[0].Related[0].Span.Start, qt.Equals, strings.Index(text, "Keeping"))
	text = strings.Replace(text, "Definitions in", "[Definitions](https://example.test/other) in", 1)
	c.Assert(definitionFindings(singleRuleResult(t, "repetition.repeated-claim", text, "", "")), qt.HasLen, 0)
}

func TestDefinitionBenefitGroupsAllLocations(t *testing.T) {
	text := definitionCases(t)[1].Text + "\n\nKeeping definitions in `terms.json` prevents different pages " +
		"from assigning different meanings to the same term."
	c := qt.New(t)
	findings := definitionFindings(singleRuleResult(t, "repetition.repeated-claim", text, "", ""))
	c.Assert(findings, qt.HasLen, 1)
	c.Assert(findings[0].Related, qt.HasLen, 2)
	c.Assert(findings[0].Related[1].Span.Start, qt.Equals, strings.LastIndex(text, "Keeping"))
	findings = definitionFindings(singleRuleResult(t, "repetition.repeated-claim", text+"\n\n## Other\n\nThe client exits.", "", ""))
	c.Assert(findings, qt.HasLen, 1)
	c.Assert(findings[0].Related, qt.HasLen, 2)
}

func TestDefinitionBenefitPreservesBoundaries(t *testing.T) {
	text := definitionCases(t)[1].Text
	for _, separator := range []string{
		"\n\n## Other deployment\n\n", "\n\n> A quoted explanation.\n\n",
		"\n\n```text\nIndependent definitions.\n```\n\n", "\n\n- An independent instruction.\n\n",
	} {
		changed := strings.Replace(text, "\n\n", separator, 1)
		qt.New(t).Assert(definitionFindings(singleRuleResult(t, "repetition.repeated-claim", changed, "", "")), qt.HasLen, 0)
	}
}
