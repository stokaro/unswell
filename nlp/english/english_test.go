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

// Punkt drops the sentence break after a dotted identifier or version; the
// provider restores it when a capitalized word or a code span follows. Real
// abbreviations, lowercase continuations, and plain versions keep Punkt's result.
func TestRepairedSentenceBoundaries(t *testing.T) {
	provider, err := english.New()
	qt.New(t).Assert(err, qt.IsNil)
	for _, row := range []struct {
		name, source string
		want         []string
	}{
		{"dotted identifier", "The handler is another chi.Router. As a result, the chain ends.",
			[]string{"The handler is another chi.Router.", "As a result, the chain ends."}},
		{"code span opener", "The handler is another chi.Router. `Mount` attaches the routes.",
			[]string{"The handler is another chi.Router.", "`Mount` attaches the routes."}},
		{"version", "The handler uses v1.2. Then it stops.", []string{"The handler uses v1.2.", "Then it stops."}},
		{"file name", "The file is main.go. (It compiles.)", []string{"The file is main.go.", "(It compiles.)"}},
		{"repeated repair", "The parent is chi.Router. `Mount` attaches. It works, etc. Then it stops.",
			[]string{"The parent is chi.Router.", "`Mount` attaches.", "It works, etc.", "Then it stops."}},
		{"abbreviation before code", "Use the helper, e.g. `WithTimeout`. The client waits.",
			[]string{"Use the helper, e.g. `WithTimeout`.", "The client waits."}},
		{"abbreviation before name", "The U.S. Army uses it. It works.", []string{"The U.S. Army uses it.", "It works."}},
		{"lowercase continuation", "The file is a.b. c follows.", []string{"The file is a.b. c follows."}},
		{"version without period", "Install version 1.2.3 first. Then restart the client.",
			[]string{"Install version 1.2.3 first.", "Then restart the client."}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(),
				document.Source{Name: "a.md", Format: document.Markdown, Bytes: []byte(row.source)}, extract.Options{})
			c.Assert(err, qt.IsNil)
			c.Assert(doc.Blocks, qt.HasLen, 1)
			sentences, err := provider.Analyze(t.Context(), doc.Blocks[0].MappedText,
				[]nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks})
			c.Assert(err, qt.IsNil)
			var got []string
			for _, sentence := range sentences {
				got = append(got, row.source[sentence.Span.Start:sentence.Span.End])
				for _, token := range sentence.Tokens {
					if token.Protected {
						continue
					}
					c.Assert(row.source[token.Spans[0].Start:token.Spans[0].End], qt.Equals, token.Text)
					c.Assert(token.Tag, qt.Not(qt.Equals), "")
				}
			}
			c.Assert(got, qt.DeepEquals, row.want)
		})
	}
}
