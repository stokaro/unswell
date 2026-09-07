package suppress

import (
	"context"

	"github.com/stokaro/unswell/document"
)

type resolvedBlock struct {
	kind      string
	words     int
	paragraph Target
	sentences []Target
}

// Resolve prose boundaries once. Protected comments can occur inside a raw
// sentence contour; they must not make the following prose precede its directive.
func (p *Plan) resolveBlocks(ctx context.Context, blocks []document.Block) error {
	for _, block := range blocks {
		if err := p.work(ctx); err != nil {
			return err
		}
		resolved := resolvedBlock{kind: block.Kind, words: block.Words,
			paragraph: Target{Scope: "paragraph", ID: block.ID, Span: block.Span, prose: block.Span}}
		for _, sentence := range block.Sentences {
			if err := p.work(ctx); err != nil {
				return err
			}
			if sentence.Words > 0 {
				resolved.sentences = append(resolved.sentences, sentenceTarget(sentence))
			}
		}
		if len(resolved.sentences) > 0 {
			resolved.paragraph.prose = document.Span{
				Start: resolved.sentences[0].prose.Start, End: resolved.sentences[len(resolved.sentences)-1].prose.End,
			}
		}
		p.blocks = append(p.blocks, resolved)
	}
	return nil
}

func sentenceTarget(sentence document.Sentence) Target {
	result := Target{Scope: "sentence", ID: sentence.ID, Span: sentence.Span, prose: sentence.Span}
	first := true
	for _, token := range sentence.Tokens {
		if token.Protected || len(token.Spans) == 0 {
			continue
		}
		bounds := document.Bounds(token.Spans)
		if first {
			result.prose.Start = bounds.Start
			first = false
		}
		result.prose.End = bounds.End
	}
	return result
}
