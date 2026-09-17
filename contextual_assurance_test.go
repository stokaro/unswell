package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestContextualEvaluationFrames(t *testing.T) {
	for _, row := range []struct{ id, text, matched string }{
		{"filler.evaluative-closure", "That is the whole design: the two verbs render through templates.", "That is the whole design"},
		{"filler.evaluative-closure", "The scoping is the whole point, and it runs in both directions.", "The scoping is the whole point"},
		{"filler.evaluative-closure", "The server answers, which is exactly the question.", "which is exactly the question"},
		{"filler.evaluative-closure", "The result survives, which is the case worth having.", "which is the case worth having"},
		{"filler.evaluative-closure", "This was precisely the point.", "This was precisely the point"},
		{"filler.evaluative-closure", "The check succeeds, and that is measured rather than assumed.", "that is measured rather than assumed"},
		{"filler.evaluative-closure", "The result agrees; it is verified rather than asserted.", "it is verified rather than asserted"},
		{"filler.evaluative-closure", "Run the check, and it is worth seeing which:", "it is worth seeing which"},
		{"filler.document-justification", "Approval is where the column earns its place:", "Approval is where the column earns its place"},
		{"filler.document-justification", "Recovery is where this diagram earned its place.", "Recovery is where this diagram earned its place"},
		{"filler.unscoped-assurance", "Ptah config parsing is intentionally explicit.", "Ptah config parsing is intentionally explicit"},
		{"filler.unscoped-assurance", "The retry policy is deliberately simple.", "The retry policy is deliberately simple"},
		{"filler.unscoped-assurance", "Our defaults are purposely safe.", "Our defaults are purposely safe"},
		{"filler.unscoped-assurance", "That is the right answer nearly always, and the wrong one where storage differs.",
			"That is the right answer nearly always"},
		{"filler.unscoped-assurance", "This is the best choice almost always.", "This is the best choice almost always"},
		{"filler.unscoped-assurance", "Keeping them apart is the fastest way to understand everything else here.",
			"Keeping them apart is the fastest way to understand everything else here"},
		{"filler.unscoped-assurance", "Reading the guide is the easiest way to learn everything.",
			"Reading the guide is the easiest way to learn everything"},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, row.matched)
			c.Assert(f.Primary.Span.Start, qt.Equals, strings.Index(row.text, row.matched))
			c.Assert(f.Evidence.Suggestion, qt.Not(qt.Equals), "")
		})
	}
}

func TestContextualEvaluationKeepsTechnicalClaims(t *testing.T) {
	controls := map[string][]string{
		"filler.evaluative-closure": {
			"That is not the whole design.", "Is that the whole design?",
			"That is the whole design of the 3-stage pipeline.",
			"The scoping is the whole point of the isolation test.",
			"That is the whole point when the cache is disabled.",
			"That is the whole point, if the test fails.",
			"That is measured rather than assumed from the configuration.",
			"Latency is measured rather than assumed.",
			"That is the `whole design`.", "That `is` the whole design.",
			"The metric answers exactly the question posed by the benchmark.",
			"This is precisely the point where the transaction commits.",
			"It is worth seeing which client sent the request.",
		},
		"filler.document-justification": {
			"Approval is where the column earns its place in the audit record.",
			"Approval is not where the column earns its place.",
			"This column lists which approvals remain valid.",
			"Approval is where the `column earns its place`.",
			"The column is outside the table so the fields fit the viewport.",
			"Approval is where the database records the decision.",
		},
		"filler.unscoped-assurance": {
			"The parsing is not intentionally explicit.", "Is parsing intentionally explicit?",
			"No parser is intentionally simple.", "The parser is intentionally `explicit`.",
			"The parser `is intentionally` explicit.",
			"Parsing is intentionally explicit about accepted keys.",
			"Parsing is deliberately simple because only two keys are supported.",
			"Parsing is deliberately simple when extensions are disabled.",
			"The 3 settings are intentionally explicit.",
			"This is the right answer nearly always in the measured workload.",
			"This is the right answer nearly always, according to the benchmark.",
			"This is the right answer when the stored type matches the API.",
			"This is almost always true because the parser rejects negative lengths.",
			"Reading the guide is the fastest way to understand this benchmark.",
			"This is the fastest way to learn everything if only two keys exist.",
			"The client is deliberately explicit: it sets the Accept header.",
		},
	}
	for id, texts := range controls {
		for _, text := range texts {
			t.Run(id+"/"+text, func(t *testing.T) {
				qt.New(t).Assert(singleRuleResult(t, id, text, "", "").Findings, qt.HasLen, 0)
			})
		}
	}
}

func TestAssuranceSourceMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.unscoped-assurance"
	text := "\ufeffThe café parser is **intentionally** explicit.\r\n"
	r := singleRuleResult(t, id, text, "", "")
	c.Assert(r.Findings, qt.HasLen, 1)
	c.Assert(r.Findings[0].Primary.Snippet, qt.Equals, "The café parser is **intentionally** explicit")
	c.Assert(r.Findings[0].Primary.Span.Start, qt.Equals, len("\ufeff"))
	phrase := "The parser is intentionally explicit"
	c.Assert(singleRuleResult(t, id, phrase+".", "", windowTerm(id, phrase)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, "The parser rejects unknown keys.\n\n```text\n"+phrase+".\n```", "", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, phrase+".", "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	checkWindowAbsence(t, id, []windowAbsenceCase{{"The client closes.", "", "", ""},
		{"# The parser is intentionally explicit", "", "", "unsupported_unit"}})
}

func TestAssuranceInSourceContexts(t *testing.T) {
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("- The parser is intentionally explicit.")},
		{Name: "source.go", Format: document.Go, Bytes: []byte("package p\n// The parser is intentionally explicit.\nconst x = 1\n")},
		{Name: "source.py", Format: document.Python, Bytes: []byte("text = 'The parser is intentionally explicit.'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			c := qt.New(t)
			result, err := singleRuleEngine(t, "filler.unscoped-assurance", "", "").Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Primary.Snippet, qt.Equals, "The parser is intentionally explicit")
		})
	}
}
