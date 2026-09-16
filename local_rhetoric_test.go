package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestLocalRhetoricGeneralizes(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"filler.document-justification", "This guide exists to make the deployment process visible."},
		{"filler.document-justification", "The following sections exist because the boundary needs stating."},
		{"filler.document-justification", "The reference intentionally doesn't repeat the installation steps."},
		{"filler.document-justification", "The reference intentionally does not duplicate those instructions."},
		{"filler.document-justification", "The examples live elsewhere, and this guide purposely did not reproduce them."},
		{"filler.evaluative-closure", "Requests remain separate, which is the purpose of isolating them."},
		{"filler.evaluative-closure", "The token expires — which is exactly what limiting it was for."},
		{"filler.evaluative-closure", "The command refuses; that is an honest response."},
		{"filler.evaluative-closure", "The operation failed, and this was the honest outcome."},
		{"repetition.definition-echo", "A retry budget is a retry budget, not a completion guarantee."},
		{"repetition.definition-echo", "The defaults are defaults, not a recovery strategy."},
		{"syntax.slogan-contrast", "The tool verifies rather than assumes."},
		{"syntax.slogan-contrast", "The pipeline proves, rather than asserts."},
	} {
		t.Run(row.text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, row.id, row.text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			c.Assert(result.Findings[0].Evidence.Suggestion, qt.Not(qt.Equals), "")
		})
	}
}

func TestLocalRhetoricControls(t *testing.T) {
	for _, row := range []struct{ id, text string }{
		{"filler.document-justification", "This page exists in the cache until the TTL expires."},
		{"filler.document-justification", "This page does not repeat the installation steps."},
		{"filler.document-justification", "This page deliberately does not restate?"},
		{"filler.document-justification", "The flag exists so callers can disable retries."},
		{"filler.document-justification", "This `page exists` so readers can find the operator."},
		{"filler.document-justification", "This page `exists` so readers can find the operator."},
		{"filler.evaluative-closure", "The wrapper starts a separate batch, which is what `EXEC` is for."},
		{"filler.evaluative-closure", "This is not the honest answer."},
		{"filler.evaluative-closure", "Is that the honest answer?"},
		{"filler.evaluative-closure", "That is the `honest answer`."},
		{"filler.evaluative-closure", "The operator gave an honest answer."},
		{"filler.evaluative-closure", "This guard prevents honest mistakes."},
		{"filler.evaluative-closure", "The line reports the delta, which is what makes it a stop condition."},
		{"repetition.definition-echo", "A retry budget is a shared retry budget, not a completion guarantee."},
		{"repetition.definition-echo", "A byte is a byte."},
		{"repetition.definition-echo", "A byte is a byte on both platforms, not a character."},
		{"repetition.definition-echo", "The default is not a default, not a fallback."},
		{"repetition.definition-echo", "`Ready` is `Ready`, not an error code."},
		{"syntax.slogan-contrast", "The tool verifies the signature rather than trusting the filename."},
		{"syntax.slogan-contrast", "The client fails rather than guessing."},
		{"syntax.slogan-contrast", "The server does not check rather than trust."},
		{"syntax.slogan-contrast", "The tool verifies rather than assumes success."},
		{"syntax.slogan-contrast", "The tool `verifies` rather than assumes."},
		{"syntax.slogan-contrast", "The tool verifies rather than `assumes`."},
		{"syntax.slogan-contrast", "The tool never verifies rather than trusts."},
		{"repetition.definition-echo", "It is it, not a different value."},
	} {
		t.Run(row.text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, row.id, row.text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestLocalRhetoricMappingContextsAndExemption(t *testing.T) {
	id := "filler.evaluative-closure"
	text := "\ufeffThe café closes — which is the **honest** answer.\r\n"
	c := qt.New(t)
	result := singleRuleResult(t, id, text, "", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	finding := result.Findings[0]
	c.Assert(finding.Primary.Span.Start, qt.Equals, strings.Index(text, "which"))
	c.Assert(finding.Primary.Snippet, qt.Equals, "which is the **honest** answer")
	c.Assert(result.Assessments[0].SlopScore > 0, qt.IsTrue)
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("- The operation fails, which is the honest answer.")},
		{Name: "source.go", Format: document.Go, Bytes: []byte("package p\n// That is the honest answer.\nconst x = 1\n")},
		{Name: "source.py", Format: document.Python, Bytes: []byte("text = 'That is the honest answer.'\n")},
	} {
		result, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(result.Findings, qt.HasLen, 1, qt.Commentf("%s", source.Name))
	}
	phrase := "That is the honest answer"
	c.Assert(singleRuleResult(t, id, phrase+".", "", windowTerm(id, phrase)).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, "The connection closes.\n\n```text\n"+phrase+".\n```", "", "").Findings, qt.HasLen, 0)
}

func TestLocalRhetoricThresholdsAndBlocks(t *testing.T) {
	c := qt.New(t)
	id := "filler.document-justification"
	text := "This guide exists so the boundary is visible. This page exists so it can be found."
	result := singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 3}", "")
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Related, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Metrics[0].Value, qt.Equals, float64(2))
	separate := strings.Replace(text, ". ", ".\n\n", 1)
	c.Assert(singleRuleResult(t, id, separate, "{allowed_occurrences: 1, saturation_occurrences: 3}", "").Findings, qt.HasLen, 0)
	for _, ruleID := range []string{id, "filler.evaluative-closure", "repetition.definition-echo", "syntax.slogan-contrast"} {
		checkWindowAbsence(t, ruleID, []windowAbsenceCase{{"The connection closes.", "", "", ""},
			{"# The connection closes", "", "", "unsupported_unit"}})
	}
}
