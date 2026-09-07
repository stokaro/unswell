package english_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func TestTechnicalOffsetsAndChunks(t *testing.T) {
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	source := "Dr. Smith configures the client. The client opens connections to the server."
	doc, err := extract.Parse(
		t.Context(),
		document.Source{Name: "a.txt", Format: document.Plain, Bytes: []byte(source)},
		extract.Options{},
	)
	c.Assert(err, qt.IsNil)
	sentences, err := provider.Analyze(
		t.Context(),
		doc.Blocks[0].MappedText,
		[]nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks},
	)
	c.Assert(err, qt.IsNil)
	c.Assert(sentences, qt.HasLen, 2)
	for _, sentence := range sentences {
		c.Assert(len(sentence.Chunks) > 0, qt.IsTrue)
		for _, token := range sentence.Tokens {
			c.Assert(source[token.Spans[0].Start:token.Spans[0].End], qt.Equals, token.Text)
			c.Assert(token.Tag, qt.Not(qt.Equals), "")
		}
	}
	_, err = provider.Analyze(t.Context(), doc.Blocks[0].MappedText, []nlp.Capability{nlp.Dependencies})
	c.Assert(err, qt.ErrorMatches, `unavailable capability "dependencies"`)
}
