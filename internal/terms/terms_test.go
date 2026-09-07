package terms_test

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/terms"
	"github.com/stokaro/unswell/rule"
)

func TestExactTokensCaseAndProtectedBoundaries(t *testing.T) {
	c := qt.New(t)
	doc := document.Document{Blocks: []document.Block{{ID: 0, Sentences: []document.Sentence{
		{ID: 0, BlockID: 0, Tokens: []document.Token{
			{Text: "Control"}, {Text: "plane"}, {Text: "control"}, {Text: "planes"},
			{Text: "control"}, {Text: "hidden", Protected: true}, {Text: "plane"},
		}},
		{ID: 1, BlockID: 0, Tokens: []document.Token{{Text: "control"}}},
		{ID: 2, BlockID: 0, Tokens: []document.Token{{Text: "plane"}}},
	}}}}
	matcher, err := terms.Compile([]string{"control plane"}, false)
	c.Assert(err, qt.IsNil)
	found, err := matcher.Find(t.Context(), &doc, 100)
	c.Assert(err, qt.IsNil)
	c.Assert(found, qt.DeepEquals, []rule.TokenRange{{BlockID: 0, SentenceID: 0, Start: 0, End: 2}})
	matcher, err = terms.Compile([]string{"control plane"}, true)
	c.Assert(err, qt.IsNil)
	found, err = matcher.Find(t.Context(), &doc, 100)
	c.Assert(err, qt.IsNil)
	c.Assert(found, qt.HasLen, 0)
	c.Assert(doc.Blocks[0].Sentences[0].Tokens[0].Text, qt.Equals, "Control")
}

func TestTermValidationAndBoundedWork(t *testing.T) {
	c := qt.New(t)
	for _, values := range [][]string{
		{""}, {"   "}, {"---"}, {"bad\x00term"}, {"bad\xffterm"}, {strings.Repeat("word ", 33)},
		{"control plane", "Control  Plane"}, {strings.Repeat("a", 1001)}, make([]string, 10001),
	} {
		_, err := terms.Compile(values, false)
		c.Assert(err, qt.IsNotNil)
	}
	matcher, err := terms.Compile([]string{"control plane"}, false)
	c.Assert(err, qt.IsNil)
	doc := document.Document{Blocks: []document.Block{{Sentences: []document.Sentence{{
		Tokens: []document.Token{{Text: "control"}},
	}}}}}
	_, err = matcher.Find(t.Context(), &doc, 0)
	c.Assert(err, qt.ErrorMatches, ".*max_candidates")
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = matcher.Find(ctx, &document.Document{}, 100)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
