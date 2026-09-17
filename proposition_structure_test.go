package unswell_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestGerundPropositionIdentity(t *testing.T) {
	for _, row := range []struct {
		text string
		want int
	}{
		{"Typing both flags is still typing both.", 1},
		{"Reading both files is still reading both files.", 1},
		{"Reading both files is still reading the cached file.", 0},
		{"Reading both files is still reading one snapshot.", 0},
		{"Reading both files is not reading both.", 0},
		{"Reading both valid files is still reading both.", 0},
		{"Reading both files is still reading both if the cache is warm.", 0},
		{"The register `R` is still the register `r`.", 0},
		{"Reading `both` files is still reading `both` files.", 0},
		{"\"Typing both flags is still typing both.\"", 0},
	} {
		qt.New(t).Assert(singleRuleResult(t, "repetition.definition-echo", row.text, "", "").Findings,
			qt.HasLen, row.want, qt.Commentf("%s", row.text))
	}
}

func TestCircularReasonJudgments(t *testing.T) {
	const prefix = "The reasons for why this is are quite complex, and they are complex because "
	for _, row := range []struct {
		tail  string
		parts int
	}{
		{"the optimizations that ripgrep uses to implement fast search are complex.", 3},
		{"the design is complex.", 3},
		{"the algorithm contains multiple interdependent state transitions.", 2},
		{"the optimizations are not complex.", 2},
		{"the optimizations are complex when the cache is full.", 2},
		{"the optimizations in three modules are complex.", 2},
		{"the implementation is more complex.", 2},
		{"the `optimizations` are complex.", 2},
	} {
		r := singleRuleResult(t, "repetition.explanatory-restart", prefix+row.tail, "", "")
		c := qt.New(t)
		c.Assert(r.Findings, qt.HasLen, 1)
		c.Assert(len(r.Findings[0].Related)+1, qt.Equals, row.parts, qt.Commentf("%s", row.tail))
		if row.parts == 3 {
			c.Assert(r.Findings[0].Related[1].Snippet, qt.Equals, strings.TrimSuffix(row.tail, "."))
		}
	}
}

func TestRestrictionEvidenceMappingAndPolicy(t *testing.T) {
	id := "repetition.repeated-claim"
	text := "The client accepts only signed requests. That is, the client never accepts requests that are not signed."
	for _, source := range []document.Source{
		{Name: "page.md", Format: document.Markdown, Bytes: []byte("\ufeff" + strings.ReplaceAll(text, "signed", "**signed**") + "\r\n")},
		{Name: "page.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "page.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		c := qt.New(t)
		r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(f.Related, qt.HasLen, 1)
		c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
		c.Assert(string(source.Bytes[f.Related[0].Span.Start:f.Related[0].Span.End]), qt.Equals, f.Related[0].Snippet)
	}
	c := qt.New(t)
	c.Assert(singleRuleResult(t, id, text, "{window_sentences: 1}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{min_words: 30}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, "The client accepts only signed requests")).Findings, qt.HasLen, 0)
	result := singleRuleResult(t, id, strings.Repeat(text+"\n\n", 50), "", "analysis: {max_candidates: 100}\n")
	c.Assert(result.Abstentions, qt.HasLen, 1)
	c.Assert(result.Findings, qt.HasLen, 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := singleRuleEngine(t, id, "", "").Analyze(ctx, document.Source{Name: "page.md", Format: document.Markdown, Bytes: []byte(text)})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestRestrictionBudgetSkipsUnmarkedPairs(t *testing.T) {
	c := qt.New(t)
	const paragraph = "The engine uses `open file` to read the selected source and preserve its original contents. " +
		"The writer uses `close file` to release the selected source and preserve its recorded metadata.\n\n"
	text := strings.Repeat(paragraph, 100) +
		"The client accepts only signed requests. That is, the client never accepts requests that are not signed."
	result := singleRuleResult(t, "repetition.repeated-claim", text, "", "analysis: {max_candidates: 10000}\n")
	c.Assert(result.Abstentions, qt.HasLen, 0)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Metrics[0].Name, qt.Equals, "restriction-restatements")
}
