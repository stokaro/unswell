package nlp_test

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"testing"
	"unicode/utf8"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

func unitOptions() nlp.UnitOptions {
	return nlp.UnitOptions{Kinds: []string{"sentence", "paragraph", "fragment"},
		Capabilities: []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks},
		Limits:       nlp.UnitLimits{MaxBytes: 1 << 20, MaxContextBytes: 65536, MaxTokens: 10000, MaxUnits: 1000, MaxSegments: 1024}}
}

func unitBlock(text, kind string) document.Block {
	block := document.Block{ID: 7, Kind: kind, Context: []string{"Cache"},
		Span: document.Span{Start: 10, End: 10 + len(text)}, MappedText: document.MappedText{Text: text}}
	for i, r := range text {
		for range utf8.RuneLen(r) {
			block.Map = append(block.Map, document.Span{Start: 10 + i, End: 10 + i + utf8.RuneLen(r)})
		}
	}
	return block
}

func prepared(t *testing.T, block document.Block, options nlp.UnitOptions) []nlp.PreparedUnit {
	t.Helper()
	c := qt.New(t)
	provider, err := english.New()
	c.Assert(err, qt.IsNil)
	units, err := nlp.PrepareUnits(t.Context(), block, provider, options)
	c.Assert(err, qt.IsNil)
	return units
}

func TestPreparedSentenceAndParagraph(t *testing.T) {
	c := qt.New(t)
	block := unitBlock(" \tThe cache may retry.\r\nThe cache cannot wait.  ", "comment")
	units := prepared(t, block, unitOptions())
	c.Assert(units, qt.HasLen, 3)
	for i, kind := range []string{"sentence", "sentence", "paragraph"} {
		binding, target := units[i].Binding(), units[i].Block()
		c.Assert(binding.Contract, qt.Equals, nlp.UnitContract)
		c.Assert(binding.Kind, qt.Equals, kind)
		c.Assert(binding.BlockKind, qt.Equals, "comment")
		c.Assert(binding.BlockID, qt.Equals, 7)
		c.Assert(binding.TextSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(target.Text))))
		c.Assert(binding.ContextSHA256, qt.Equals, units[2].Binding().TextSHA256)
		c.Assert(binding.Segments, qt.DeepEquals, target.Spans(0, len(target.Text)))
		c.Assert(target.Sentences[0].Tokens[0].Start, qt.Equals, 0)
		c.Assert(units[i].Context(), qt.Equals, "The cache may retry.\r\nThe cache cannot wait.")
	}
	c.Assert(units[0].Block().Words, qt.Equals, 4)
	c.Assert(units[2].Block().Words, qt.Equals, 8)
	c.Assert(units[0].Binding().ContextSHA256, qt.Not(qt.Equals), units[0].Binding().GrammarSHA256)
	c.Assert(units[0].Binding().Segments, qt.Not(qt.DeepEquals), units[1].Binding().Segments)
}

func TestPreparedFragmentsAndEmptySelection(t *testing.T) {
	c := qt.New(t)
	for _, kind := range []string{"paragraph", "comment", "string", "heading"} {
		units := prepared(t, unitBlock("Keep\x00 unchanged.", kind), unitOptions())
		c.Assert(units, qt.HasLen, 2)
		c.Assert(units[0].Block().Text, qt.Equals, "Keep")
		c.Assert(units[1].Context(), qt.Equals, "unchanged.")
		for _, unit := range units {
			c.Assert(unit.Binding().Kind, qt.Equals, "fragment")
			c.Assert(unit.Context(), qt.Equals, unit.Block().Text)
		}
	}
	options := unitOptions()
	options.Kinds = []string{"sentence"}
	c.Assert(prepared(t, unitBlock("A string message.", "string"), options), qt.HasLen, 0)
	c.Assert(prepared(t, unitBlock(" \t\x00\r\n", "comment"), options), qt.HasLen, 0)
}

func TestPreparedMappingThroughExtractors(t *testing.T) {
	for _, test := range []struct {
		format       document.Format
		source, text string
	}{
		{document.Markdown, "\ufeffThe **cache** may retry &amp; wait.\r\n", "The cache may retry & wait."},
		{document.Go, "package p\nconst Message = \"Cache \\u0065ntry cannot wait.\"\n", "Cache entry cannot wait."},
		{document.YAML, "message: \"Café \\u0065ntry cannot wait.\"\n", "Café entry cannot wait."},
	} {
		t.Run(string(test.format), func(t *testing.T) {
			c := qt.New(t)
			doc, err := extract.Parse(t.Context(), document.Source{Name: "input", Format: test.format, Bytes: []byte(test.source)},
				extract.Options{MaxBytes: 65536, MaxBlocks: 100})
			c.Assert(err, qt.IsNil)
			unit, block, found := findPreparedText(t, doc, test.text)
			c.Assert(found, qt.IsTrue)
			binding := unit.Binding()
			for _, span := range binding.Segments {
				c.Assert(span.Valid(len(test.source)), qt.IsTrue)
			}
			c.Assert(binding.Segments, qt.DeepEquals, block.Spans(0, len(block.Text)))
		})
	}
}

func TestPreparedIdentityAndOwnership(t *testing.T) {
	c := qt.New(t)
	block := unitBlock("The cache may retry. The cache may retry.", "paragraph")
	units := prepared(t, block, unitOptions())
	first, second := units[0].Binding(), units[1].Binding()
	c.Assert(first.TextSHA256, qt.Equals, second.TextSHA256)
	c.Assert(first.Segments, qt.Not(qt.DeepEquals), second.Segments)
	block.Map[0], block.Context[0] = document.Span{}, "changed"
	view := units[0].Block()
	view.Map[0], view.Sentences[0].Tokens[0].Spans[0] = document.Span{}, document.Span{}
	view.Sentences[0].Tokens[0].Normal = "changed"
	first.ContextSpans[0] = document.Span{}
	identity := units[0].Identity()
	identity.Capabilities[0] = "changed"
	c.Assert(units[0].Block().Map[0].Start, qt.Equals, 10)
	c.Assert(units[0].Block().Sentences[0].Tokens[0].Normal, qt.Equals, "the")
	c.Assert(units[0].Block().Context, qt.DeepEquals, []string{"Cache"})
	c.Assert(units[0].Binding().ContextSpans[0].Start, qt.Equals, 10)
	c.Assert(units[0].Identity().Capabilities, qt.Contains, nlp.Tokens)
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 20 {
				view := units[0].Block()
				view.Sentences[0].Tokens[0].Normal = "owned"
				_ = units[0].Binding()
			}
		})
	}
	group.Wait()
	c.Assert(units[0].Block().Sentences[0].Tokens[0].Normal, qt.Equals, "the")
}

func findPreparedText(t *testing.T, doc document.Document, text string) (nlp.PreparedUnit, document.Block, bool) {
	t.Helper()
	for _, block := range doc.Blocks {
		for _, unit := range prepared(t, block, unitOptions()) {
			if unit.Block().Text == text {
				return unit, block, true
			}
		}
	}
	return nlp.PreparedUnit{}, document.Block{}, false
}
