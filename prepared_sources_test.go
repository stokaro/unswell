package unswell_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestPreparedCollectionUsesFilePoliciesAndPreservesPartialEvidence(t *testing.T) {
	c := qt.New(t)
	options := preparedOptions()
	options.Jobs = 2
	options.Config = []byte("version: 1\nextraction: {contexts: [comment, string]}\noverrides:\n" +
		"  - files: [private.go]\n    extraction: {contexts: [comment]}\n")
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	code := []byte("package sample\n// The client retries.\nvar Message = \"The cache expires.\"\n")
	sources := []document.Source{{Name: "public.go", Format: document.Go, Bytes: code},
		{Name: "private.go", Format: document.Go, Bytes: code}}
	result, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	a, b := result.PreparedFeatures.Sources[0], result.PreparedFeatures.Sources[1]
	c.Assert(a.Path, qt.Equals, "private.go")
	c.Assert(a.Units, qt.HasLen, 2)
	c.Assert(b.Units, qt.HasLen, 3)
	c.Assert(a.SourceHash, qt.Equals, b.SourceHash)
	c.Assert(a.PolicyHash, qt.Not(qt.Equals), b.PolicyHash)
	c.Assert(a.ExtractionPolicyHash, qt.Not(qt.Equals), b.ExtractionPolicyHash)
	c.Assert(a.PreparationHash, qt.Not(qt.Equals), b.PreparationHash)
	c.Assert(a.Units[0].Binding.TextSHA256, qt.Equals, b.Units[0].Binding.TextSHA256)
	c.Assert(a.Units[0].InputHash, qt.Not(qt.Equals), b.Units[0].InputHash)
	sources[0] = document.Source{Name: "broken.cs", Format: document.CSharp, Bytes: []byte("class Sample { string value = \"unfinished")}
	partial, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNotNil)
	c.Assert(partial.Status, qt.Equals, "incomplete")
	c.Assert(partial.Gate.Passed, qt.IsFalse)
	c.Assert(partial.PreparedFeatures.Sources, qt.HasLen, 1)
	c.Assert(partial.PreparedFeatures.Sources[0], qt.DeepEquals, a)
}

func TestPreparedCollectionDoesNotJoinIndependentCommentsOrStrings(t *testing.T) {
	c := qt.New(t)
	options := preparedOptions()
	options.Config = []byte("version: 1\nextraction: {contexts: [comment, string]}\n")
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "settings.yaml", Format: document.YAML,
		Bytes: []byte("# The client retries.\r\nfirst: \"The cache\\nexpires.\"\r\n# Keep the timeout.\r\nsecond: 'The queue drains.'\r\n")})
	c.Assert(err, qt.IsNil)
	units := result.PreparedFeatures.Sources[0].Units
	c.Assert(units, qt.HasLen, 6)
	contexts := make(map[string]bool)
	for _, unit := range units {
		contexts[unit.Binding.ContextSHA256] = true
	}
	c.Assert(contexts, qt.HasLen, 4)
	c.Assert(units[2].Binding.Kind, qt.Equals, "fragment")
	c.Assert(*units[2].Values[1].Number, qt.Equals, float64(3))
}

func TestPreparedCollectionRecordsNoTargetsForUnrequestedKinds(t *testing.T) {
	c := qt.New(t)
	options := preparedOptions()
	options.PreparedKinds = []string{"sentence"}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("# A heading\n\nWords with `code` between.\n")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Status, qt.Equals, "complete")
	c.Assert(result.PreparedFeatures.Sources, qt.HasLen, 1)
	c.Assert(result.PreparedFeatures.Sources[0].TargetCount, qt.Equals, 0)
	c.Assert(result.PreparedFeatures.Sources[0].Units, qt.HasLen, 0)
}
