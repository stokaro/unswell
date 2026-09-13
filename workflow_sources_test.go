package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestWorkflowPolicyChangesPreparationIdentity(t *testing.T) {
	c := qt.New(t)
	options := preparedOptions()
	options.CollectBaseline = true
	options.Config = []byte("version: 1\nextraction: {github_actions: shell}\noverrides:\n" +
		"  - files: [.github/workflows/strings.yml]\n    extraction: {github_actions: strings}\n")
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	text := []byte("jobs:\n  check:\n    steps:\n      - shell: bash\n        run: echo \"The client must not retry before 30 seconds.\"\n")
	result, err := engine.AnalyzeAll(t.Context(), []document.Source{
		{Name: ".github/workflows/shell.yml", Format: document.YAML, Bytes: text},
		{Name: ".github/workflows/strings.yml", Format: document.YAML, Bytes: text},
	})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	a, b := result.PreparedFeatures.Sources[0], result.PreparedFeatures.Sources[1]
	c.Assert(a.SourceHash, qt.Equals, b.SourceHash)
	c.Assert(a.ExtractionPolicyHash, qt.Not(qt.Equals), b.ExtractionPolicyHash)
	c.Assert(a.PreparationHash, qt.Not(qt.Equals), b.PreparationHash)
	c.Assert(a.PolicyHash, qt.Not(qt.Equals), b.PolicyHash)
	c.Assert(a.Units[len(a.Units)-1].Binding.GrammarSHA256, qt.Not(qt.Equals), b.Units[len(b.Units)-1].Binding.GrammarSHA256)
}
