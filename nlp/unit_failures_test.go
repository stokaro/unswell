package nlp_test

import (
	"context"
	"fmt"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
)

type alteredUnitProvider struct {
	nlp.Provider
	change func([]document.Sentence)
	before func() error
	calls  int
}

func (p *alteredUnitProvider) Analyze(ctx context.Context, mapped document.MappedText,
	capabilities []nlp.Capability,
) ([]document.Sentence, error) {
	p.calls++
	if p.before != nil {
		if err := p.before(); err != nil {
			return nil, err
		}
	}
	sentences, err := p.Provider.Analyze(ctx, mapped, capabilities)
	if err == nil && p.change != nil {
		p.change(sentences)
	}
	return sentences, err
}

func TestUnitPreparationCallsProviderOncePerContext(t *testing.T) {
	c := qt.New(t)
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	provider := &alteredUnitProvider{Provider: base}
	units, err := nlp.PrepareUnits(t.Context(), unitBlock("The cache may retry. The cache cannot wait.", "comment"), provider, unitOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(units, qt.HasLen, 3)
	c.Assert(provider.calls, qt.Equals, 1)
	_, err = nlp.PrepareUnits(t.Context(), unitBlock("Keep\x00 unchanged.", "comment"), provider, unitOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(provider.calls, qt.Equals, 3)
}

func TestUnitPreparationNeverAnalyzesExcludedProse(t *testing.T) {
	c := qt.New(t)
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	provider := &alteredUnitProvider{Provider: base}
	block := unitBlock("Excluded prose retains its source mapping.", "paragraph")
	block.Excluded = true
	units, err := nlp.PrepareUnits(t.Context(), block, provider, unitOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(units, qt.HasLen, 0)
	c.Assert(provider.calls, qt.Equals, 0)
	block.Map = nil
	_, err = nlp.PrepareUnits(t.Context(), block, provider, unitOptions())
	c.Assert(err, qt.IsNotNil)
	c.Assert(provider.calls, qt.Equals, 0)
}

func TestUnitPreparationRejectsMalformedProvider(t *testing.T) {
	changes := map[string]func([]document.Sentence){
		"text":           func(s []document.Sentence) { s[0].Text = "Other text." },
		"word_count":     func(s []document.Sentence) { s[0].Words++ },
		"missing_pos":    func(s []document.Sentence) { s[0].Tokens[0].Tag = "" },
		"missing_normal": func(s []document.Sentence) { s[0].Tokens[0].Normal = "" },
		"word_flag":      func(s []document.Sentence) { s[0].Tokens[0].Word = false },
		"token_text":     func(s []document.Sentence) { s[0].Tokens[0].Text = "Other" },
		"token_span":     func(s []document.Sentence) { s[0].Tokens[0].Spans[0].End++ },
		"sentence_span":  func(s []document.Sentence) { s[0].Spans[0].End++ },
		"missing_tokens": func(s []document.Sentence) { s[0].Tokens = nil },
		"chunk": func(s []document.Sentence) {
			s[0].Chunks = []document.Chunk{{Kind: "NP", FirstToken: 0, EndToken: 999}}
		},
		"overlap":    func(s []document.Sentence) { s[1] = s[0] },
		"omit_prose": func(s []document.Sentence) { s[0] = s[1] },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			base, err := english.New()
			c.Assert(err, qt.IsNil)
			provider := &alteredUnitProvider{Provider: base, change: change}
			units, err := nlp.PrepareUnits(t.Context(), unitBlock("The cache may retry. The cache cannot wait.", "comment"), provider, unitOptions())
			c.Assert(err, qt.IsNotNil)
			c.Assert(units, qt.IsNil)
		})
	}
}

func TestUnitPreparationRejectsLimitsAndOptions(t *testing.T) {
	changes := map[string]func(*nlp.UnitOptions){
		"bytes":                func(o *nlp.UnitOptions) { o.Limits.MaxBytes = 3 },
		"context":              func(o *nlp.UnitOptions) { o.Limits.MaxContextBytes = 3 },
		"tokens":               func(o *nlp.UnitOptions) { o.Limits.MaxTokens = 1 },
		"units":                func(o *nlp.UnitOptions) { o.Limits.MaxUnits = 1 },
		"invalid_limit":        func(o *nlp.UnitOptions) { o.Limits.MaxTokens = 0 },
		"kind":                 func(o *nlp.UnitOptions) { o.Kinds = []string{"document"} },
		"empty_kinds":          func(o *nlp.UnitOptions) { o.Kinds = nil },
		"duplicate_kind":       func(o *nlp.UnitOptions) { o.Kinds = []string{"sentence", "sentence"} },
		"missing_tokens":       func(o *nlp.UnitOptions) { o.Capabilities = []nlp.Capability{nlp.Sentences} },
		"capability":           func(o *nlp.UnitOptions) { o.Capabilities = append(o.Capabilities, nlp.Dependencies) },
		"duplicate_capability": func(o *nlp.UnitOptions) { o.Capabilities = append(o.Capabilities, nlp.Tokens) },
	}
	for name, change := range changes {
		t.Run(name, func(t *testing.T) {
			c := qt.New(t)
			provider, err := english.New()
			c.Assert(err, qt.IsNil)
			options := unitOptions()
			change(&options)
			units, err := nlp.PrepareUnits(t.Context(), unitBlock("The cache may retry. The cache cannot wait.", "comment"), provider, options)
			c.Assert(err, qt.IsNotNil)
			c.Assert(units, qt.IsNil)
		})
	}
}

func TestUnitPreparationRejectsMappingsAndCancellation(t *testing.T) {
	c := qt.New(t)
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	block := unitBlock("The cache may retry.", "paragraph")
	for _, change := range []func(*document.Block){
		func(b *document.Block) { b.Map = nil },
		func(b *document.Block) { b.Map[0] = document.Span{} },
		func(b *document.Block) { b.Map[2] = b.Map[0] },
		func(b *document.Block) { b.Text = "\xff" },
	} {
		copy := unitBlock(block.Text, block.Kind)
		change(&copy)
		_, err := nlp.PrepareUnits(t.Context(), copy, base, unitOptions())
		c.Assert(err, qt.IsNotNil)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = nlp.PrepareUnits(ctx, block, base, unitOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	ctx, cancel = context.WithCancel(t.Context())
	defer cancel()
	provider := &alteredUnitProvider{Provider: base, before: func() error { cancel(); return ctx.Err() }}
	_, err = nlp.PrepareUnits(ctx, block, provider, unitOptions())
	c.Assert(err, qt.ErrorIs, context.Canceled)
	provider.before = func() error { return fmt.Errorf("injected provider failure") }
	_, err = nlp.PrepareUnits(t.Context(), block, provider, unitOptions())
	c.Assert(err, qt.ErrorMatches, "injected provider failure")
}

// escapedRuneBlock maps one rune onto several adjacent source spans, the way
// the Go extractor maps a string literal that spells a rune as separate
// escape sequences.
func escapedRuneBlock(text string, width int) document.Block {
	block := document.Block{ID: 9, Kind: "string", Context: []string{"Cache"},
		Span: document.Span{Start: 0, End: width * len(text)}, MappedText: document.MappedText{Text: text}}
	for i := range len(text) {
		block.Map = append(block.Map, document.Span{Start: width * i, End: width * (i + 1)})
	}
	return block
}

func TestUnitPreparationKeepsRunesSpelledBySeveralEscapes(t *testing.T) {
	c := qt.New(t)
	base, err := english.New()
	c.Assert(err, qt.IsNil)
	block := escapedRuneBlock("The \ufeffcache may retry.", 4)
	units, err := nlp.PrepareUnits(t.Context(), block, base, unitOptions())
	c.Assert(err, qt.IsNil)
	c.Assert(len(units) > 0, qt.IsTrue)
	gapped := escapedRuneBlock(block.Text, 4)
	gapped.Map[5] = document.Span{Start: 21, End: 24}
	_, err = nlp.PrepareUnits(t.Context(), gapped, base, unitOptions())
	c.Assert(err, qt.ErrorMatches, "unit source map splits an encoded rune")
	reversed := escapedRuneBlock(block.Text, 4)
	reversed.Map[5], reversed.Map[6] = reversed.Map[6], reversed.Map[5]
	_, err = nlp.PrepareUnits(t.Context(), reversed, base, unitOptions())
	c.Assert(err, qt.IsNotNil)
}
