package unswell_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
)

func TestActionScaffolding(t *testing.T) {
	for _, text := range []string{
		"This package allows additional configuration of features to be enabled by build tags.",
		"The clause allows specifying additional labels to be attached to the alert.",
		"The interface enables configuration to be stored in a file.",
		"User management can be done by directly using the connection API.",
		"Configuration of the service may be performed by using the console.",
		"Deployment is accomplished by running the installer.",
	} {
		t.Run(text, func(t *testing.T) {
			c := qt.New(t)
			r := singleRuleResult(t, "filler.instruction-scaffolding", text, "", "")
			c.Assert(r.Findings, qt.HasLen, 1)
			f := r.Findings[0]
			c.Assert(f.Primary.Snippet, qt.Equals, text[f.Primary.Span.Start:f.Primary.Span.End])
			c.Assert(strings.Contains(f.Evidence.Suggestion, "actors"), qt.IsTrue)
		})
	}
}

func TestActionScaffoldingControls(t *testing.T) {
	for _, text := range []string{
		"The interface allows users to configure the service.",
		"The interface allows requests to be canceled.",
		"The interface allows messages to be received.",
		"The interface allows configuration files to be uploaded.",
		"The console enables administrators to configure the service.",
		"Configuration can be changed by an administrator.",
		"The assignment is done by the next worker.",
		"The component is used to configure the service.",
		"Configuration can be done only by using the console.",
		"Configuration cannot be done by using the console.",
		"Configuration is not performed by using the console.",
		"Configuration can be done by using the console if the worker is stopped.",
		"Configuration can be done by using the console without permission.",
		"The interface allows configuration to be enabled only after approval.",
		"The interface denies the ability to configure the service.",
		"The agreement is performed by singing the contract aloud.",
		"The option `allows configuration to be enabled` is reserved.",
		"The interface `allows` configuration to be enabled.",
		"The interface allows configuration to be `enabled`.",
		"The manual says \"Configuration can be done by using the console\".",
		"Can configuration be done by using the console?",
	} {
		t.Run(text, func(t *testing.T) {
			qt.New(t).Assert(singleRuleResult(t, "filler.instruction-scaffolding", text, "", "").Findings, qt.HasLen, 0)
		})
	}
}

func TestActionScaffoldingMappingAndPolicy(t *testing.T) {
	c := qt.New(t)
	id := "filler.instruction-scaffolding"
	text := "Configuration can be done by using the café console."
	for _, source := range []document.Source{
		{Name: "guide.md", Format: document.Markdown, Bytes: []byte("\ufeffConfiguration can be done by using the **café** console.\r\n")},
		{Name: "guide.go", Format: document.Go, Bytes: []byte("package example\n// " + text + "\nfunc Example() {}\n")},
		{Name: "guide.py", Format: document.Python, Bytes: []byte("message = '" + text + "'\n")},
	} {
		r, err := singleRuleEngine(t, id, "", "").Analyze(t.Context(), source)
		c.Assert(err, qt.IsNil)
		c.Assert(r.Findings, qt.HasLen, 1)
		f := r.Findings[0]
		c.Assert(string(source.Bytes[f.Primary.Span.Start:f.Primary.Span.End]), qt.Equals, f.Primary.Snippet)
	}
	c.Assert(singleRuleResult(t, id, text, "{allowed_occurrences: 1, saturation_occurrences: 2}", "").Findings, qt.HasLen, 0)
	c.Assert(singleRuleResult(t, id, text, "", windowTerm(id, strings.TrimSuffix(text, "."))).Findings, qt.HasLen, 0)
	long := "The interface allows " + strings.Repeat("additional ", 49) + "configuration to be enabled."
	c.Assert(singleRuleResult(t, id, long, "", "").Findings, qt.HasLen, 0)
}
