package feature_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp/english"
)

func TestExtractedSourceSegments(t *testing.T) {
	for _, row := range []struct {
		name   string
		format document.Format
		text   string
		words  int
		parts  []string
	}{
		{"entities", document.Markdown, "\ufeffA **caf&#233;** and tea. `hidden prose`\r\n", 4,
			[]string{"A", "caf&#233;", "and", "tea"}},
		{"escaped-string", document.Go, "package p\nconst label = \"A caf\\u00e9 and tea.\"\n", 4,
			[]string{"A", "caf\\u00e9", "and", "tea"}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "sample", Format: row.format, Bytes: []byte(row.text)},
				extract.Options{Policy: extract.Policy{Contexts: []string{"paragraph", "string"}}})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			provider, err := english.New()
			c.Assert(err, qt.IsNil)
			identity := testIdentity()
			identity.NLP, identity.Source = provider.Identity(), doc.Hash
			doc.Blocks[0].Sentences, err = provider.Analyze(t.Context(), doc.Blocks[0].MappedText, identity.Capabilities)
			c.Assert(err, qt.IsNil)
			m, err := feature.Measure(t.Context(), doc.Blocks[0], identity, testLimits())
			c.Assert(err, qt.IsNil)
			c.Assert(m.Counts().Words, qt.Equals, row.words)
			var parts []string
			for _, span := range m.Spans() {
				parts = append(parts, row.text[span.Start:span.End])
			}
			c.Assert(parts, qt.DeepEquals, row.parts)
			c.Assert(strings.Join(parts, " "), qt.Not(qt.Contains), "hidden")
		})
	}
}
