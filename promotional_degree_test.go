package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestPromotionalDegreeRequiresCriteria(t *testing.T) {
	for _, row := range []struct{ text, matched string }{
		{"- Blazingly fast", "Blazingly fast"},
		{"## Blindingly quick", "Blindingly quick"},
		{"The client is phenomenally responsive.", "phenomenally responsive"},
		{"The library is astonishingly portable.", "astonishingly portable"},
		{"The interface is remarkably intuitive.", "remarkably intuitive"},
		{"The engine is heavily optimized for deduplication and caching.", "heavily optimized"},
		{"The solver is highly optimized for concurrent requests.", "highly optimized"},
		{"A project file can be perfectly adoptable in front of a database that is not.", "perfectly adoptable"},
		{"We provide an easy to use client package.", "easy to use client"},
		{"An effortless-to-use library is available.", "effortless-to-use library"},
		{"This will be the biggest win!", "This will be the biggest win"},
		{"Use the builder for this purpose, this will be the biggest win!", "this will be the biggest win"},
		{"That might be the greatest benefit.", "That might be the greatest benefit"},
		{"It is the best improvement.", "It is the best improvement"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.unscoped-assurance", row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.RuleVersion, qt.Equals, "10")
			c.Assert(f.Primary.Snippet, qt.Equals, row.matched)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.matched))
			c.Assert(f.Evidence.Suggestion, qt.Contains, "Verify")
			c.Assert(f.Evidence.Suggestion, qt.Contains, "uncertainty")
		})
	}
}

func TestPromotionalDegreePreservesTechnicalMeanings(t *testing.T) {
	for _, text := range []string{
		"The service is highly available.", "The database is strongly consistent.",
		"The result is perfectly square.", "The module is strongly typed.",
		"The slots are perfectly aligned.", "The values are highly correlated.",
		"The service is heavily loaded.", "The library is perfectly compatible with the protocol.",
		"The client is not blazingly fast.", "No client is blazingly fast.",
		"Is the client blazingly fast?", "The solver is never heavily optimized.",
		"The README claims the client is blazingly fast.",
		"The README says: the client is blazingly fast.",
		"The slogan is \"Blazingly fast\".", "The client starts.\n\n- `Blazingly fast`", "## `Blazingly` fast",
		"The client is `blazingly fast`.", "The client is blazingly `fast`.",
		"The client is blazingly fast when its cache is warm.",
		"The solver is heavily optimized because allocations dominate the measured workload.",
		"The client is blazingly fast: the benchmark reports its throughput.",
		"The client is blazingly fast, with a median latency of 2 milliseconds.",
		"The client is blazingly fast. The benchmark reports its latency.",
		"The benchmark reports throughput. The engine is heavily optimized.",
		"The solver is heavily optimized by reusing the work graph.",
		"The solver is heavily optimized through allocation elimination.",
		"The client is blazingly fast compared with the previous implementation.",
		"A file is perfectly adoptable if all declared fields are supported.",
		"This will be the biggest win in the measured workload.",
		"This is not the biggest win.", "Will this be the biggest win?",
		"This will be the biggest win: the comparison follows.",
		"This will be the biggest win. The benchmark reports the comparison.",
		"The biggest win was recorded by the tournament scoreboard.",
		"If you need fewer steps, use the easy to use client package.",
		"The package is easy to use because the adapter supplies its configuration.",
		"The easy to use `client` is available.",
		"## The configuration is intentionally explicit",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.unscoped-assurance", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestPromotionalDegreeOriginalMappingsAndPolicy(t *testing.T) {
	id := "filler.unscoped-assurance"
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("\ufeffA café client is **blazingly fast**.\r\n")},
		{Name: "guide.mdx", Format: document.MDX, Bytes: []byte("export const demo = 'Blazingly fast';\n\n## Blazingly fast\n")},
		{Name: "client.go", Format: document.Go, Bytes: []byte("package p\n// The client is blazingly fast.\nfunc Run() {}\n")},
		{Name: "client.py", Format: document.Python, Bytes: []byte("'''The client is blazingly fast.'''\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			c := qt.New(t)
			r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.LastIndex(strings.ToLower(string(source.Bytes)), "blazingly"))
		})
	}
	text := "The client is blazingly fast. The tool is unbelievably flexible."
	c := qt.New(t)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 2, saturation_occurrences: 3}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, "The client is blazingly fast.", "", windowTerm(id, "blazingly fast")).Findings, qt.HasLen, 0)
	code := "The client requires a key.\n\n```text\nThe client is blazingly fast.\n```"
	c.Assert(singleRuleResult(t, id, code, "", "").Findings, qt.HasLen, 0)
}

func TestPromotionalDegreePreservesExistingQualityDiagnosis(t *testing.T) {
	for _, text := range []string{
		"It is easy to use the existing settings.",
		"The tool is incredibly flexible.",
		"The service is unbelievably reliable.",
		"Caddy is also ridiculously extensible, with a powerful plugin system.",
	} {
		t.Run(text, func(t *testing.T) {
			result := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(result.Findings[0].Message, qt.Equals, "State the behavior, scope or evidence behind this quality claim.")
		})
	}
}

func TestMixedPromotionalDegreeRetainsVerificationBeforeEditing(t *testing.T) {
	for _, text := range []string{
		"The client is blazingly fast. This will be the biggest win.",
		"The parser is intentionally explicit. The client is blazingly fast.",
		"The client is blazingly fast. Reliability is ensured by rigorous testing.",
	} {
		t.Run(text, func(t *testing.T) {
			result := singleRuleResult(t, "filler.unscoped-assurance", text, "", "")
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, 1)
			f := result.Findings[0]
			c.Assert(f.Related, qt.HasLen, 1)
			c.Assert(f.Message, qt.Equals, "State the behavior, scope or evidence behind this quality claim.")
			c.Assert(f.Evidence.Suggestion, qt.Contains, "Verify their basis before removing or weakening any claim")
			c.Assert(f.Evidence.Suggestion, qt.Contains, "established commitments")
		})
	}
}
