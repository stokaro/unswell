package unswell_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestInstructionProjection(t *testing.T) {
	for _, text := range []string{
		"The adapter has the ability to query a local index.",
		"The renderer also has the capability to format a date.",
		"The renderer has the ability to format a date.",
		"The function can be used to configure a label.",
		"Optionally, the function may be used to provide metadata.",
		"The function can be used to configure a label when rendering a chart.",
		"The map is intended to allow extra values to be passed to the template.",
		"The console is designed to enable users to configure the queue.",
		"The adapter has the ability to allow users to specify a header.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, strings.TrimSuffix(text, "."))
			c.Assert(f.Evidence.Suggestion, qt.Contains, "capability is not an obligation")
		})
	}
}

func TestInstructionProjectionControls(t *testing.T) {
	for _, text := range []string{
		"The interface allows users to configure the service.",
		"The client enables you to specify a header.",
		"The library allows the reader to parse an archive.",
		"The interface allows administrators to configure the service.",
		"The interface grants users access to configure the service.",
		"The interface allows users to configure the service only after approval.",
		"The interface does not allow users to configure the service.",
		"No interface allows users to configure the service.",
		"The interface allows users to lose data accidentally.",
		"The adapter can query a local index.",
		"The adapter is used to query a local index.",
		"The interface allows messages to be received.",
		"The configuration file is intended to be stored in this directory.",
		"The adapter has the ability to reconnect.",
		"The adapter has the ability to restart when the timer expires.",
		"The survey says the adapter has the ability to query a local index.",
		"The manual says \"The adapter has the ability to query a local index\".",
		"The adapter has `the ability to query` a local index.",
		"The adapter has the ability to `query` a local index.",
		"The adapter can be used to leak credentials.",
		"Can the adapter be used to query a local index?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestInstructionProjectionMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.instruction-scaffolding"
	text := "The adapter has the ability to query the café index."
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("\ufeffThe adapter has the ability to query the **café** index.\r\n")},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
	}
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, strings.TrimSuffix(text, "."))).Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, "The adapter has the ability to query "+strings.Repeat("local ", 49)+"indexes.", "", "").Findings,
		qt.HasLen, 0)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := singleRuleEngine(t, id, "", "").Analyze(ctx, document.Source{Name: "a.md", Bytes: []byte(text)})
	c.Assert(err, qt.ErrorIs, context.Canceled)
}

func TestInstructionProjectionAdjacentMethod(t *testing.T) {
	id := "filler.instruction-scaffolding"
	first := "It is possible to configure a label."
	second := "This is done by using the label option."
	for _, gap := range []string{" ", "\n\n", "\r\n\r\n"} {
		r := singleRuleResult(t, id, first+gap+second, "", "")
		c := qt.New(t)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(f.Primary.Snippet, qt.Equals, strings.TrimSuffix(first, "."))
		c.Assert(f.Related, qt.HasLen, 1)
		c.Assert(f.Related[0].Snippet, qt.Equals, strings.TrimSuffix(second, "."))
		c.Assert(f.Related[0].Span.Start, qt.Equals, len(first+gap))
	}
	for _, gap := range []string{
		"\n\n## Another topic\n\n", "\n\n```go\ncall()\n```\n\n", "\n\n- Another operation.\n\n",
		"\n\nThe server writes a checkpoint.\n\n", "\n\n<!-- omitted -->\n\n",
	} {
		r := singleRuleResult(t, id, first+gap+second, "", "")
		qt.New(t).Assert(r.Findings, qt.HasLen, 2)
		qt.New(t).Assert(r.Findings[0].Related, qt.HasLen, 0)
	}
}

func TestInstructionProjectionMethods(t *testing.T) {
	for _, text := range []string{
		"Configuration is carried out using the console.",
		"Deployment can be performed using the installer.",
		"Validation is accomplished through running the validator.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 1)
	}
	for _, text := range []string{
		"Configuration files are carried out using the console.",
		"Deployment artifacts are prepared using the compiler.",
		"Configuration is performed by the administrator.",
		"Configuration is performed using.",
	} {
		qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
	}
}
