package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestSupportedActionUsefulnessPurpose(t *testing.T) {
	for _, text := range []string{
		"It can also be used for associating messages with the vertex that can be helpful for tracing purposes.",
		"The client can be used to associate messages with vertices, which may be useful for tracing purposes.",
		"The worker has the ability to attach messages to records, which could be helpful for detailed request tracing purposes.",
		"The parser has the capability to correlate records, which might be useful for debugging purposes.",
		"The reader may be used for indexing files that can be useful for search purposes.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			result := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(result.Findings, qt.HasLen, 1)
			finding := result.Findings[0]
			c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(finding.Message, qt.Contains, "support and helpfulness clauses")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "relative clause's antecedent")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "modality of both clauses")
			c.Assert(finding.Evidence.Suggestion, qt.Contains, "possible benefit is not a guarantee")
		})
	}
}

func TestSupportedActionUsefulnessPurposeControls(t *testing.T) {
	for _, text := range []string{
		"The client can associate messages with vertices, which may be useful for tracing purposes.",
		"Associating messages with vertices can be helpful for tracing purposes.",
		"Message correlation is helpful for tracing purposes.",
		"The parser can be used to correlate records, which may be useful when a trace is missing.",
		"The parser can be used to correlate records, which may be useful for identifying duplicates.",
		"The parser can be used to correlate records, which may be useful for two purposes.",
		"The parser can be used to correlate records, which may be useful for several purposes.",
		"The parser can be used to correlate records, which may be useful for many purposes.",
		"The parser can be used to correlate records, which may be useful for tracing purposes only.",
		"The parser can be used to correlate records, which may be useful for tracing purposes if a request is slow.",
		"The parser can be used to correlate records, which may be useful for tracing purposes and may consume memory.",
		"The parser can be used to correlate records, which is occasionally useful for tracing purposes.",
		"The parser can be used to correlate records, which is useful for tracing purposes.",
		"The parser can be used to correlate records, which may be `useful` for tracing purposes.",
		"The parser can be used to correlate records, which may be useful `for tracing purposes`.",
		"The parser `can be used to` correlate records, which may be useful for tracing purposes.",
		"The parser can be used to correlate records, `which may be useful for tracing purposes`.",
		"The parser cannot be used to correlate records, which may be useful for tracing purposes.",
		"The parser can be used to correlate records, which may not be useful for tracing purposes.",
		"The parser can be used only by administrators to correlate records, which may be useful for tracing purposes.",
		"The manual says the parser can be used to correlate records, which may be useful for tracing purposes.",
		"\"The parser can be used to correlate records, which may be useful for tracing purposes.\"",
		"Can the parser be used to correlate records, which may be useful for tracing purposes?",
		"The parser can be used to correlate records; this may be useful for tracing purposes.",
		"The parser can be used to correlate records. This may be useful for tracing purposes.",
	} {
		t.Run(text, func(t *testing.T) {
			// Some controls retain the existing support-only warning. They must
			// not receive the specific diagnosis of both complete constructions.
			result := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			for _, finding := range result.Findings {
				qt.New(t).Assert(finding.Message, qt.Not(qt.Contains), "support and helpfulness clauses")
			}
		})
	}
}

func TestSupportedActionUsefulnessSourceAndPolicy(t *testing.T) {
	id := "filler.instruction-scaffolding"
	text := "The client can be used to associate café messages with vertices that may be helpful for request tracing purposes."
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("\ufeff" + strings.ReplaceAll(text, "café", "**café**") + "\r\n")},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			finding := instructionSourceFinding(t, source)
			qt.New(t).Assert(finding.Message, qt.Contains, "support and helpfulness clauses")
			qt.New(t).Assert(finding.Primary.Snippet, qt.Contains, "request tracing purposes")
		})
	}
	qt.New(t).Assert(singleRuleResult(t, id, text, "", windowTerm(id, strings.TrimSuffix(text, "."))).Findings, qt.HasLen, 0)
	qt.New(t).Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	long := "The client can be used to associate " + strings.Repeat("additional ", 49) +
		"messages with vertices that may be helpful for tracing purposes."
	qt.New(t).Assert(singleRuleResult(t, id, long, "", "").Findings, qt.HasLen, 0)
}

func TestSupportedActionUsefulnessMixedExplanation(t *testing.T) {
	text := "The client can be used to associate messages with vertices that may be helpful for tracing purposes. " +
		"The reader has the ability to query the index."
	result := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
	c := qt.New(t)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].Evidence.Occurrences, qt.HasLen, 2)
	c.Assert(result.Findings[0].Message, qt.Equals, "State the supported action directly; preserve its conditions and optionality.")
}

func TestSupportedActionUsefulnessKeepsMethod(t *testing.T) {
	first := "It can be used for associating messages with vertices that may be helpful for tracing purposes."
	method := "This is done by using the correlation option."
	for _, gap := range []string{" ", "\n\n"} {
		result := singleRuleResult(t, "filler.instruction-scaffolding", first+gap+method, "", "")
		c := qt.New(t)
		c.Assert(result.Findings, qt.HasLen, 1)
		finding := result.Findings[0]
		c.Assert(finding.Message, qt.Contains, "support and helpfulness clauses")
		c.Assert(finding.Related, qt.HasLen, 1)
		c.Assert(finding.Related[0].Snippet, qt.Equals, strings.TrimSuffix(method, "."))
		c.Assert(finding.Related[0].Span.Start, qt.Equals, len(first+gap))
		c.Assert(finding.Evidence.Suggestion, qt.Contains, "methods")
	}
}
