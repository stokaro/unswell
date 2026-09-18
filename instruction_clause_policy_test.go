package unswell_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestInstructionClauseMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.instruction-scaffolding"
	text := "Now that we have a café client, we need to send the request."
	for _, source := range []document.Source{
		{Name: "a.md", Format: document.Markdown, Bytes: []byte("\ufeff" + strings.ReplaceAll(text, "café", "**café**") + "\r\n")},
		{Name: "a.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "a.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(f.RuleVersion, qt.Equals, "11")
		c.Assert(f.Primary.Snippet, qt.Equals, string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]))
		c.Assert(f.Evidence.Suggestion, qt.Contains, "prerequisites")
	}
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, strings.TrimSuffix(text, "."))).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	r := singleRuleResult(t, id, strings.Repeat(text+"\n\n", 80), "", "analysis: {max_candidates: 100}\n")
	c.Assert(r.Abstentions, qt.HasLen, 1)
	c.Assert(r.Findings, qt.HasLen, 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := singleRuleEngine(t, id, "", "").Analyze(ctx, document.Source{Name: "a.md", Bytes: []byte(text)})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestInstructionClauseScope(t *testing.T) {
	for _, text := range []string{
		"The adapter has the ability to query the index: a failed query returns an error.",
		"The adapter has the ability to query the index: the client says it is ready.",
	} {
		r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
		qt.New(t).Assert(r.Findings, qt.HasLen, 1)
		qt.New(t).Assert(r.Findings[0].Primary.Snippet, qt.Equals, "The adapter has the ability to query the index")
	}
	qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding",
		"The report says: the adapter has the ability to query the index.", "", "").Findings, qt.HasLen, 0)
}
