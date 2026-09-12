// Package english adapts Prose's tokenizer, Punkt segmenter, and POS tagger.
// It adds a shallow chunker; it does not implement lemmas or dependencies.
package english

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/jdkato/prose/v3/segment"
	"github.com/jdkato/prose/v3/tag"
	"github.com/jdkato/prose/v3/tokenize"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

// Provider shares immutable upstream models and keeps per-call work separate.
type Provider struct {
	tokenizer *tokenize.Tokenizer
	segmenter *segment.Segmenter
	tagger    func() (*tag.Tagger, error)
}

// New loads the sentence model. POS is loaded only when requested.
func New() (*Provider, error) {
	segmenter, err := segment.New()
	if err != nil {
		return nil, fmt.Errorf("load English sentence model: %w", err)
	}
	return &Provider{
		tokenizer: tokenize.New(),
		segmenter: segmenter,
		tagger:    sync.OnceValues(func() (*tag.Tagger, error) { return tag.New() }),
	}, nil
}

// Identity returns an independent capability slice and pinned model metadata.
func (p *Provider) Identity() nlp.Identity {
	return nlp.Identity{
		Name: "builtin-en", Version: "prose-v3.2.1/chunks-v1/boundaries-v1", Model: "prose/aptagmodel/en.bin", License: "MIT",
		ModelHash:         "209282c71733883be5c08af24301cc8917079c9deb7721b89a8dd118d071a780",
		SentenceModelHash: "2498632eaf8c3d0480c074e0331057ea9b20f1306523db06ff742689269a21a8",
		Capabilities:      []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks},
	}
}

// Analyze segments one block, preserving the source map for all output tokens.
// Punkt segments are repaired at a period that ends a dotted identifier or
// version and precedes a sentence opener; see repairBoundaries.
func (p *Provider) Analyze(ctx context.Context, mapped document.MappedText, required []nlp.Capability) ([]document.Sentence, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	for _, capability := range required {
		if !slices.Contains(p.Identity().Capabilities, capability) {
			return nil, fmt.Errorf("unavailable capability %q", capability)
		}
	}
	if len(mapped.Text) != len(mapped.Map) {
		return nil, fmt.Errorf("invalid source map")
	}
	withPOS := slices.Contains(required, nlp.POS) || slices.Contains(required, nlp.Chunks)
	result := make([]document.Sentence, 0)
	for _, sent := range p.segmenter.Segment(mapped.Text) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, part := range repairBoundaries(sent, p.tokenizer.Tokenize(sent.Text)) {
			out, err := p.sentence(mapped, part.sentence, part.tokens, withPOS)
			if err != nil {
				return nil, err
			}
			result = append(result, out)
		}
	}
	return result, ctx.Err()
}

// sentence builds one output sentence from its already tokenized text. Tags
// and chunks are computed on the repaired sentence, not the Punkt segment.
func (p *Provider) sentence(
	mapped document.MappedText, sent segment.Sentence, tokens []tokenize.Token, withPOS bool,
) (document.Sentence, error) {
	spans := mapped.Spans(sent.Start, sent.End())
	result := document.Sentence{
		Text:   sent.Text,
		Span:   document.Bounds(spans),
		Spans:  spans,
		Tokens: []document.Token{},
		Chunks: []document.Chunk{},
	}
	if withPOS {
		tagger, err := p.tagger()
		if err != nil {
			return result, fmt.Errorf("load POS model: %w", err)
		}
		tagger.TagTokens(tokens)
	}
	for _, tok := range tokens {
		start, end := sent.Start+tok.Start, sent.Start+tok.End()
		protected := strings.ContainsRune(tok.Text, 0)
		word := !protected && document.IsWord(tok.Text)
		result.Tokens = append(result.Tokens, document.Token{
			Text: tok.Text, Normal: document.Normalize(tok.Text), Tag: tok.Tag,
			Start: start, End: end, Spans: mapped.Spans(start, end), Word: word, Protected: protected,
		})
		if word {
			result.Words++
		}
	}
	if withPOS {
		result.Chunks = chunks(result.Tokens)
	}
	return result, nil
}

func chunks(tokens []document.Token) []document.Chunk {
	result := make([]document.Chunk, 0)
	for i := 0; i < len(tokens); {
		kind := chunkKind(tokens[i])
		if kind == "" {
			i++
			continue
		}
		end := i + 1
		for end < len(tokens) && chunkKind(tokens[end]) == kind {
			end++
		}
		if kind == "PP" && end < len(tokens) && chunkKind(tokens[end]) == "NP" {
			for end < len(tokens) && chunkKind(tokens[end]) == "NP" {
				end++
			}
		}
		result = append(result, document.Chunk{Kind: kind, FirstToken: i, EndToken: end})
		i = end
	}
	return result
}

func chunkKind(token document.Token) string {
	if token.Protected {
		return ""
	}
	tag := token.Tag
	if hasTagPrefix(tag, []string{"NN", "JJ"}) || slices.Contains([]string{"DT", "PRP$", "CD"}, tag) {
		return "NP"
	}
	if hasTagPrefix(tag, []string{"VB", "RB"}) || tag == "MD" {
		return "VP"
	}
	if tag == "IN" || tag == "TO" {
		return "PP"
	}
	return ""
}

func hasTagPrefix(tag string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(tag, prefix) {
			return true
		}
	}
	return false
}
