package extract_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func TestGlobalAndLanguageContexts(t *testing.T) {
	cases := []struct {
		name                   string
		global, override, want []string
	}{
		{"default", nil, nil, []string{"comment", "string"}},
		{"global comments", []string{"comment"}, nil, []string{"comment"}},
		{"global strings", []string{"string"}, nil, []string{"string"}},
		{"replace global", []string{"comment"}, []string{"string"}, []string{"string"}},
		{"disable language", []string{"comment", "string"}, []string{}, []string{}},
		{"disable globally", []string{}, nil, []string{}},
		{"override disabled default", []string{}, []string{"comment"}, []string{"comment"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			policy := extract.Policy{Contexts: tc.global}
			if tc.override != nil {
				policy.Languages = map[document.Format]extract.LanguagePolicy{document.CSharp: {Contexts: tc.override}}
			}
			source := "// Read the manual.\nclass Sample { string message = \"Runtime prose.\"; }"
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.cs", Format: document.CSharp, Bytes: []byte(source)},
				extract.Options{Policy: policy})
			c.Assert(err, qt.IsNil)
			kinds := []string{}
			for _, block := range doc.Blocks {
				kinds = append(kinds, block.Kind)
			}
			c.Assert(kinds, qt.DeepEquals, tc.want)
			for _, exclusion := range doc.Excluded {
				c.Assert(strings.HasPrefix(exclusion.Reason, "config:context-disabled:"), qt.IsTrue)
				c.Assert(exclusion.Span.Valid(len(source)), qt.IsTrue)
			}
		})
	}
}

func TestContextOverrideStaysWithinLanguage(t *testing.T) {
	c := qt.New(t)
	policy := extract.Policy{Contexts: []string{"comment"}, Languages: map[document.Format]extract.LanguagePolicy{
		document.CSharp: {Contexts: []string{"string"}},
	}}
	source := "# Read the manual.\nmessage: Runtime prose.\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.yaml", Format: document.YAML, Bytes: []byte(source)},
		extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 1)
	c.Assert(doc.Blocks[0].Kind, qt.Equals, "comment")
}

func TestMarkdownContextSelection(t *testing.T) {
	c := qt.New(t)
	policy := extract.Policy{Contexts: []string{"comment"}, Languages: map[document.Format]extract.LanguagePolicy{
		document.Markdown: {Contexts: []string{"heading", "table-cell"}},
	}}
	source := "# Read the manual\n\nBody prose.\n\n- List prose.\n\n| Column |\n| --- |\n| Cell prose. |\n"
	doc, err := extract.Parse(t.Context(), document.Source{Name: "sample.md", Format: document.Markdown, Bytes: []byte(source)},
		extract.Options{Policy: policy})
	c.Assert(err, qt.IsNil)
	c.Assert(doc.Blocks, qt.HasLen, 3)
	c.Assert(doc.Blocks[0].Kind, qt.Equals, "heading")
	c.Assert(doc.Blocks[1].Kind, qt.Equals, "table-cell")
	c.Assert(doc.Blocks[2].Kind, qt.Equals, "table-cell")
	c.Assert(doc.Blocks[2].ID, qt.Equals, 2)
	c.Assert(doc.Excluded, qt.HasLen, 2)
}

func TestDisabledContextsDoNotAcceptInvalidSyntax(t *testing.T) {
	c := qt.New(t)
	_, err := extract.Parse(t.Context(), document.Source{
		Name: "broken.cs", Format: document.CSharp, Bytes: []byte("class Sample { string value = \"unfinished"),
	}, extract.Options{Policy: extract.Policy{Contexts: []string{}}})
	c.Assert(err, qt.IsNotNil)
}
