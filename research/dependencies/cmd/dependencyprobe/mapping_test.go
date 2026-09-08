package main

import (
	"testing"

	gdoc "github.com/bioshock/gospacy/v3/doc"
	"github.com/bioshock/gospacy/v3/vocab"
	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
)

func mappingExample() *gdoc.Doc {
	v := vocab.NewVocab()
	doc := gdoc.NewDoc(v, "Café retries.")
	doc.Tokens = []gdoc.Token{
		{Text: "Café", Idx: 0, Head: 1, Dep: v.StringStore().Add("nsubj"), SentStart: 1},
		{Text: "retries", Idx: 5, Head: 1, Dep: v.StringStore().Add("ROOT")},
		{Text: ".", Idx: 12, Head: 1, Dep: v.StringStore().Add("punct")},
	}
	return doc
}

func TestMappedDependencyTokens(t *testing.T) {
	c := qt.New(t)
	source, err := extract.Parse(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("Caf&eacute; **retries**.")}, extract.Options{})
	c.Assert(err, qt.IsNil)
	sentences, err := mapSentences(t.Context(), source.Blocks[0].MappedText, 0, mappingExample())
	c.Assert(err, qt.IsNil)
	c.Assert(sentences, qt.HasLen, 1)
	c.Assert(sentences[0].Tokens[0].Spans, qt.DeepEquals, []document.Span{{Start: 0, End: 11}})
	c.Assert(sentences[0].Tokens[1].Spans, qt.DeepEquals, []document.Span{{Start: 14, End: 21}})
	c.Assert(sentences[0].Tokens[1].Start, qt.Equals, 6)
	c.Assert(sentences[0].Dependencies.Arcs, qt.DeepEquals, []document.DependencyArc{
		{Head: 1, Relation: "nsubj"}, {Head: -1, Relation: "ROOT"}, {Head: 1, Relation: "punct"},
	})
}

func TestMappingRejectsIncompleteParserOutput(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*gdoc.Doc)
		err    string
	}{
		{"no tokens", func(d *gdoc.Doc) { d.Tokens = nil }, ".*sentence start"},
		{"no first boundary", func(d *gdoc.Doc) { d.Tokens[0].SentStart = 0 }, ".*sentence start"},
		{"unknown boundary", func(d *gdoc.Doc) { d.Tokens[1].SentStart = -1 }, ".*unknown sentence boundary"},
		{"negative rune index", func(d *gdoc.Doc) { d.Tokens[1].Idx = -1 }, ".*invalid rune offset.*"},
		{"excessive rune index", func(d *gdoc.Doc) { d.Tokens[1].Idx = 90 }, ".*invalid rune offset.*"},
		{"excessive token length", func(d *gdoc.Doc) { d.Tokens[2].Text = "too long" }, ".*outside the mapped block"},
		{"cross sentence head", func(d *gdoc.Doc) { d.Tokens[0].Head = 5 }, ".*head outside its sentence"},
		{"root without label", func(d *gdoc.Doc) { d.Tokens[1].Dep = 0 }, ".*invalid head.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			source, err := extract.Parse(t.Context(), document.Source{Name: "guide.txt", Format: document.Plain,
				Bytes: []byte("Café retries.")}, extract.Options{})
			c.Assert(err, qt.IsNil)
			doc := mappingExample()
			row.change(doc)
			_, err = mapSentences(t.Context(), source.Blocks[0].MappedText, 0, doc)
			c.Assert(err, qt.ErrorMatches, row.err)
		})
	}
}
