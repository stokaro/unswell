package nlp

import (
	"context"
	"crypto/sha256"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/internal/textutil"
)

func preparePieces(ctx context.Context, block document.Block, provider Provider, options UnitOptions) ([]PreparedUnit, error) {
	binding, err := bindUnitBlock(block)
	if err != nil {
		return nil, err
	}
	result := []PreparedUnit{}
	kind := enclosingUnitKind(block)
	start, tokens := 0, 0
	for part := range strings.SplitSeq(block.Text, "\x00") {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		left, right := textutil.TrimSpaceBounds(part)
		text := part[left:right]
		if text != "" {
			piece := document.MappedText{Text: text, Map: block.Map[start+left : start+left+len(text)]}
			units, visits, err := preparePiece(ctx, block, piece, kind, provider, options, binding)
			if err != nil {
				return nil, err
			}
			tokens += visits
			if tokens > options.Limits.MaxTokens || len(result)+len(units) > options.Limits.MaxUnits {
				return nil, fmt.Errorf("unit preparation exceeds total token or target limit")
			}
			result = append(result, units...)
		}
		start += len(part) + 1
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func preparePiece(ctx context.Context, block document.Block, piece document.MappedText, kind string,
	provider Provider, options UnitOptions, binding UnitBinding,
) ([]PreparedUnit, int, error) {
	if len(piece.Text) > options.Limits.MaxContextBytes {
		return nil, 0, fmt.Errorf("eligible context exceeds %d bytes", options.Limits.MaxContextBytes)
	}
	piece.Text, piece.Map = strings.Clone(piece.Text), slices.Clone(piece.Map)
	sentences, err := provider.Analyze(ctx, piece, slices.Clone(options.Capabilities))
	if err != nil {
		return nil, 0, err
	}
	visits, err := validateUnitSentences(ctx, piece, sentences, options)
	if err != nil {
		return nil, 0, err
	}
	units, err := selectPieceUnits(ctx, block, piece, kind, sentences, options, binding)
	return units, visits, err
}

func selectPieceUnits(ctx context.Context, block document.Block, piece document.MappedText, kind string,
	sentences []document.Sentence, options UnitOptions, binding UnitBinding,
) ([]PreparedUnit, error) {
	binding.ContextSHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(piece.Text)))
	binding.ContextSpans = piece.Spans(0, len(piece.Text))
	if len(binding.ContextSpans) < 1 || len(binding.ContextSpans) > options.Limits.MaxSegments {
		return nil, fmt.Errorf("eligible context exceeds source segment limit")
	}
	result := []PreparedUnit{}
	words := 0
	for _, sentence := range sentences {
		words += sentence.Words
		if !selectSentenceUnit(kind, options.Kinds, sentence.Words) {
			continue
		}
		start, end := sentence.Tokens[0].Start, sentence.Tokens[len(sentence.Tokens)-1].End
		unit := makeUnit(block, piece, "sentence", start, end, []document.Sentence{sentence}, binding)
		result = append(result, unit)
		if len(result) > options.Limits.MaxUnits {
			return nil, fmt.Errorf("unit preparation exceeds target limit")
		}
	}
	if slices.Contains(options.Kinds, kind) && words > 0 {
		unit := makeUnit(block, piece, kind, 0, len(piece.Text), sentences, binding)
		result = append(result, unit)
	}
	return result, ctx.Err()
}

func makeUnit(parent document.Block, context document.MappedText, kind string, start, end int,
	sentences []document.Sentence, binding UnitBinding,
) PreparedUnit {
	block := document.Block{ID: parent.ID, Kind: parent.Kind, Context: slices.Clone(parent.Context),
		MappedText: document.MappedText{Text: context.Text[start:end], Map: slices.Clone(context.Map[start:end])},
		Sentences:  cloneUnitSentences(sentences, start)}
	block.Span = document.Bounds(block.Spans(0, len(block.Text)))
	for i := range block.Sentences {
		block.Sentences[i].ID, block.Sentences[i].BlockID = i, block.ID
		block.Words += block.Sentences[i].Words
	}
	binding.Kind = kind
	binding.TextSHA256 = fmt.Sprintf("%x", sha256.Sum256([]byte(block.Text)))
	binding.Segments = block.Spans(0, len(block.Text))
	return PreparedUnit{block: block, context: context.Text, binding: binding}
}

func selectSentenceUnit(kind string, kinds []string, words int) bool {
	return kind == "paragraph" && slices.Contains(kinds, "sentence") && words > 0
}

func enclosingUnitKind(block document.Block) string {
	if !strings.ContainsRune(block.Text, 0) && slices.Contains([]string{"paragraph", "comment"}, block.Kind) {
		return "paragraph"
	}
	return "fragment"
}
