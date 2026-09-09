package unswell_test

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestFeatureBindingsPreserveCompleteMappedBlocks(t *testing.T) {
	for _, row := range []struct {
		name, source, text string
		segments           []document.Span
	}{
		{"protected", "One **two** three `hidden words`.\r\n", "One two three  \x00 .",
			[]document.Span{{Start: 0, End: 4}, {Start: 6, End: 9}, {Start: 11, End: 33}}},
		{"Unicode", "\ufeffCafé **😀** works &amp; stays.\r\n", "Café 😀 works & stays.",
			[]document.Span{{Start: 3, End: 9}, {Start: 11, End: 15}, {Start: 17, End: 36}}},
		{"leading whitespace", "  Keep both conditions.  \n", "  Keep both conditions.",
			[]document.Span{{Start: 0, End: 23}}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Features: []string{"prose-words"}})
			c.Assert(err, qt.IsNil)
			source := document.Source{Name: "sample.md", Format: document.Markdown, Bytes: []byte(row.source)}
			result, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			units := result.Features.Sources[0].Units
			c.Assert(units, qt.HasLen, 1)
			binding := units[0].Binding
			c.Assert(binding, qt.IsNotNil)
			c.Assert(binding.Contract, qt.Equals, unswell.FeatureBlockBindingContract)
			c.Assert(binding.TextSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(row.text))))
			c.Assert(binding.Segments, qt.DeepEquals, row.segments)
			c.Assert(binding.TrimmedSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(strings.TrimSpace(row.text)))))
			if row.name == "leading whitespace" {
				c.Assert(binding.TrimmedSegments, qt.DeepEquals, []document.Span{{Start: 2, End: 23}})
			} else {
				c.Assert(binding.TrimmedSegments, qt.DeepEquals, row.segments)
			}
			c.Assert(binding.Segments, qt.Not(qt.DeepEquals), units[0].Segments)
		})
	}
}

func TestFeatureBindingSegmentsConsumeTheCollectionBudget(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{Features: []string{"prose-words"},
		Config: []byte("version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 5}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "sample.md", Format: document.Markdown,
		Bytes: []byte("a**b**c**d**e**f**g**h**i**j**k")})
	c.Assert(err, qt.ErrorMatches, "collected block mappings exceed max_candidates")
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Features.Sources, qt.HasLen, 0)
}
