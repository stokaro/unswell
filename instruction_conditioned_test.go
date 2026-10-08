package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestConditionedMethodInstruction(t *testing.T) {
	for _, pair := range []struct{ first, second string }{
		{"After vertexes have been loaded to the job, it is safe to request a result from an edge pointing to a previously loaded vertex.",
			"To do this `build(ctx, Edge) (CachedResult, error)` method is called on the static scheduler instance associated with the solver."},
		{"Once the client is ready, it is possible to fetch the records.",
			"To do this, the `Fetch` method is called."},
		{"When the queue is initialized, it is safe to read the messages.",
			"To do that, the `Read` function is invoked with the queue handle."},
		{"After the records have been loaded, it is possible to query the index.",
			"To do this, the Query method can be called on the client."},
		{"Once the request has been prepared, it is safe to send the packet.",
			"To do that, the Send function may be invoked from the client."},
	} {
		t.Run(pair.first, func(t *testing.T) {
			for _, gap := range []string{" ", "\n\n"} {
				result := singleRuleResult(t, "filler.instruction-scaffolding", pair.first+gap+pair.second, "", "")
				c := qt.New(t)
				c.Assert(result.Findings, qt.HasLen, 1)
				finding := result.Findings[0]
				c.Assert(finding.Primary.Snippet, qt.Equals, strings.TrimSuffix(pair.first, "."))
				c.Assert(finding.Related, qt.HasLen, 1)
				c.Assert(finding.Related[0].Snippet, qt.Equals, strings.TrimSuffix(pair.second, "."))
				c.Assert(finding.Evidence.Metrics, qt.HasLen, 1)
				c.Assert(finding.Evidence.Metrics[0].Value, qt.Equals, float64(1))
				c.Assert(finding.Message, qt.Contains, "separate method-invocation announcement")
				c.Assert(finding.Evidence.Suggestion, qt.Contains, "safety or possibility qualifier")
				c.Assert(finding.Evidence.Suggestion, qt.Contains, "do not turn permission or capability into an obligation")
			}
		})
	}
}

func TestConditionedInstructionMappingAndPolicy(t *testing.T) {
	id := "filler.instruction-scaffolding"
	first := "After the café records have been loaded, it is safe to query the index."
	second := "To do this, the `Query` method is called on the client."
	text := first + " " + second
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("\ufeff" + strings.ReplaceAll(text, "café", "**café**") + "\r\n")},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		t.Run(source.Name, func(t *testing.T) {
			finding := instructionSourceFinding(t, source)
			c := qt.New(t)
			c.Assert(finding.Related, qt.HasLen, 1)
			location := finding.Related[0]
			c.Assert(string(source.Bytes[location.Span.Start:location.Span.End]), qt.Equals, location.Snippet)
			c.Assert(finding.Primary.Snippet, qt.Contains, "café")
		})
	}
	plain := strings.ReplaceAll(second, "`Query`", "Query")
	qt.New(t).Assert(singleRuleResult(t, id, first+" "+plain, "", windowTerm(id, strings.TrimSuffix(plain, "."))).Findings, qt.HasLen, 0)
	qt.New(t).Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	long := strings.Replace(first, "records", strings.Repeat("large ", 49)+"records", 1)
	qt.New(t).Assert(singleRuleResult(t, id, long+" "+second, "", "").Findings, qt.HasLen, 0)
}

func TestConditionedInstructionBoundaries(t *testing.T) {
	first := "After the client is ready, it is safe to fetch the records."
	second := "To do this, the `Fetch` method is called."
	for _, gap := range []string{
		"\n\n## Another operation\n\n", "\n\n```text\nother action\n```\n\n",
		" The worker starts. ", "\n\n> Another operation.\n\n", "\n\n- ",
	} {
		t.Run(gap, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", first+gap+second, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestConditionedInstructionControls(t *testing.T) {
	first := "After the client is ready, it is safe to fetch the records."
	second := "To do this, the `Fetch` method is called."
	for _, text := range []string{
		first, second,
		"After the client is ready, call the `Fetch` method to fetch records.",
		"After the client is ready, it fetches the records. " + second,
		"It is safe to fetch the records. " + second,
		"After startup, it is safe to fetch the records. " + second,
		"After the client is ready, it is unsafe to fetch the records. " + second,
		"After the client is ready, it is not safe to fetch the records. " + second,
		"After the client is ready, it is safe to fetch only cached records. " + second,
		"After the client is ready, it is safe to fetch the records without authentication. " + second,
		"After the client is ready, it is safe to request permission. " + second,
		"After the client is ready, it is safe to fetch the records unless they are stale. " + second,
		"After the client is ready, `it is safe to` fetch the records. " + second,
		"After the client is ready, it is safe to `fetch` the records. " + second,
		first + " To do this, a method is called.",
		first + " To do this, the `method` object is called.",
		first + " To do this, the `Fetch` method is called automatically.",
		first + " To do this, the `Fetch` method is called by the worker.",
		first + " To do this, the `Fetch` method is called on the client after startup.",
		first + " To do this, the `Fetch` method is called on the client and the worker starts.",
		first + " To do this, the `Fetch` method is not called.",
		first + " To do this, the `Fetch` method is called only when requested.",
		first + " To do this, the `Fetch` method `is called`.",
		first + " To do this, the worker calls `Fetch`.",
		first + " The `Fetch` method is called.",
		first + " To do this, the client opens the connection.",
		"The manual says: " + first + " " + second,
		"\"" + first + " " + second + "\"",
		strings.TrimSuffix(first, ".") + "; it also starts the worker. " + second,
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}
