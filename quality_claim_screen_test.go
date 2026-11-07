package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestSourceLocalQualityClaimCandidates(t *testing.T) {
	for _, row := range []struct{ text, matched string }{
		{"APIs are polished and future-proofed.", "polished and future-proofed"},
		{"PromQL is a powerful and flexible query language to leverage this dimensionality.",
			"powerful and flexible query language"},
		{"The most comprehensive feature set", "most comprehensive feature set"},
		{"containerd provides reasonable defaults for generating OCI runtime specs.", "reasonable defaults"},
		{"Correct sRGB blending will have an important effect on font rendering quality.", "important effect"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.unscoped-assurance", row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			finding := r.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Contains, row.matched)
			c.Assert(row.text[finding.Primary.Span.Start:finding.Primary.Span.End], qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Evidence.Suggestion, qt.Not(qt.Equals), "")
		})
	}
}

func TestSourceLocalQualityClaimControls(t *testing.T) {
	for _, text := range []string{
		"The query language supports label matching, joins, and aggregations.",
		"The flexible layout lets users resize columns and choose a grid.",
		"The API supports a new feature flag combination.",
		"The most comprehensive feature set among the three measured libraries is listed in Table 2.",
		"Defaults are reasonable for this test because they keep the cache below 2 GiB.",
		"RLS omission is important: the generated schema cannot describe field or row access.",
		"If there are important fixes to backport, open a pull request.",
		"The role is safe to remove only after its grants have been revoked.",
		"The phrase \"powerful and flexible query language\" is a quotation from the release note.",
	} {
		t.Run(text, func(t *testing.T) {
			result := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
			qt.New(t).Assert(result.Findings, qt.HasLen, 0, qt.Commentf("%s", strings.TrimSpace(text)))
		})
	}
}
