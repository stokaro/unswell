package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestAbstractActionCarrier(t *testing.T) {
	for _, text := range []string{
		"One final piece of solver logic allows merging two edges into one when they have both returned the same cache key.",
		"A final piece of renderer code enables formatting dates in the selected timezone.",
		"Another piece of parser logic allows combining two trees when their roots match.",
		"An additional part of retry logic supports reusing recorded responses after the connection returns.",
		"One final bit of compiler code enables folding constant expressions.",
	} {
		t.Run(text, func(t *testing.T) {
			result := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c := qt.New(t)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(finding.Message, qt.Contains, "unnamed part of logic")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "conditions and limits")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "capability is not an obligation")
		})
	}
}

func TestAbstractActionCarrierSourceAndPolicy(t *testing.T) {
	id := "filler.instruction-scaffolding"
	text := "One final piece of solver logic allows merging two café records when their `key` values match."
	markdown := "\ufeff" + strings.ReplaceAll(text, "solver logic", "solver **logic**") + "\r\n"
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte(markdown)},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			result, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
			c := qt.New(t)
			c.Assert(err, qt.IsNil)
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(string(source.Bytes[finding.Primary.Span.Start:finding.Primary.Span.End]), qt.Equals, finding.Primary.Snippet)
			c.Assert(finding.Primary.Snippet, qt.Contains, "two café records when")
			c.Assert(finding.Primary.Snippet, qt.Contains, "`key` values match")
		})
	}
	plain := strings.ReplaceAll(text, "`key`", "key")
	qt.New(t).Assert(singleRuleResult(t, id, plain, "", windowTerm(id, strings.TrimSuffix(plain, "."))).Findings, qt.HasLen, 0)
	qt.New(t).Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	long := "One final piece of solver logic allows merging " + strings.Repeat("large ", 49) + "trees."
	qt.New(t).Assert(singleRuleResult(t, id, long, "", "").Findings, qt.HasLen, 0)
}

func TestAbstractActionCarrierControls(t *testing.T) {
	for _, text := range []string{
		"The final merge handler allows merging two edges when their keys match.",
		"The solver can merge two edges when their keys match.",
		"The final phase of solver logic allows merging two edges when their keys match.",
		"One piece of solver logic allows merging two edges when their keys match.",
		"The third part of solver logic allows merging two edges when their keys match.",
		"One final piece of Solver logic allows merging two edges when their keys match.",
		"One final piece of `solver` logic allows merging two edges when their keys match.",
		"One final piece of solver logic `allows merging` two edges when their keys match.",
		"One final piece of solver logic allows `merging` two edges when their keys match.",
		"One final piece of solver logic refuses merging two edges when their keys differ.",
		"One final piece of solver logic allows merging.",
		"One final piece of solver logic allows merging only when both keys match.",
		"One final piece of solver logic allows merging without losing constraints.",
		"One final piece of solver logic does not allow merging two edges.",
		"The manual says one final piece of solver logic allows merging two edges.",
		"The manual says: \"One final piece of solver logic allows merging two edges.\"",
		"Does one final piece of solver logic allow merging two edges?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
