package unswell_test

import (
	_ "embed"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

//go:embed research/reviews/2026-09-26-outcome-announcements/cases.json
var outcomeCases []byte

//go:embed research/reviews/2026-09-26-outcome-announcements/additional-cases.json
var additionalOutcomeCases []byte

func TestOutcomeAnnouncementConstructions(t *testing.T) {
	engine := singleRuleEngine(t, "filler.document-justification", "", "")
	for _, row := range []struct {
		name string
		data []byte
	}{
		{"recovered", outcomeCases},
		{"additional", additionalOutcomeCases},
	} {
		t.Run(row.name, func(t *testing.T) {
			checkOutcomeFixture(t, engine, row.data)
		})
	}
}

func checkOutcomeFixture(t *testing.T, engine *unswell.Engine, data []byte) {
	t.Helper()
	var fixture struct {
		Cases []struct {
			Text  string `json:"text"`
			Match bool   `json:"match"`
		} `json:"cases"`
	}
	qt.New(t).Assert(json.Unmarshal(data, &fixture), qt.IsNil)
	for _, row := range fixture.Cases {
		t.Run(row.Text, func(t *testing.T) {
			c := qt.New(t)
			result, err := engine.Analyze(t.Context(), document.Source{
				Name: "guide.md", Format: document.Markdown, Bytes: []byte(row.Text),
			})
			c.Assert(err, qt.IsNil)
			count := 0
			if row.Match {
				count = 1
			}
			c.Assert(result.Findings, qt.HasLen, count)
			for _, finding := range result.Findings {
				c.Assert(finding.Related, qt.HasLen, 2)
				for _, loc := range append(finding.Related, finding.Primary) {
					c.Assert(row.Text[loc.Span.Start:loc.Span.End], qt.Equals, loc.Snippet)
				}
			}
		})
	}
}

func TestOutcomeAnnouncementSourceAndAllowance(t *testing.T) {
	const intro = "There are several possible outcomes"
	const text = intro + ". If the café cache matches, return the result. If the cache misses, run the operation."
	for _, row := range []struct {
		name, prefix, suffix string
		format               document.Format
	}{
		{"guide.mdx", "\ufeff", "\r\n", document.MDX},
		{"guide.go", "package example\n// ", "\nfunc Example() {}\n", document.Go},
		{"guide.py", "message = '", "'\n", document.Python},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source := []byte(row.prefix + text + row.suffix)
			r, err := singleRuleEngine(t, "filler.document-justification", "", "").Analyze(t.Context(),
				document.Source{Name: row.name, Format: row.format, Bytes: source})
			c.Assert(err, qt.IsNil)
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Related, qt.HasLen, 2)
			c.Assert(f.Primary.Snippet, qt.Equals, intro)
			for _, loc := range append(f.Related, f.Primary) {
				c.Assert(string(source[loc.Span.Start:loc.Span.End]), qt.Equals, loc.Snippet)
			}
		})
	}
	qt.New(t).Assert(singleRuleResult(t, "filler.document-justification", text,
		"{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
}

func TestOutcomeAnnouncementDoesNotTreatPluralNounsAsPredicates(t *testing.T) {
	const text = "There are several possible outcomes. If the cache entries, return the result. " +
		"If the server responses, run the operation."
	qt.New(t).Assert(singleRuleResult(t, "filler.document-justification", text, "", "").Findings, qt.HasLen, 0)
}
