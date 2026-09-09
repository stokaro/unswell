package nlp

import (
	"context"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

func validateUnitSentences(ctx context.Context, mapped document.MappedText, sentences []document.Sentence,
	options UnitOptions,
) (int, error) {
	if len(sentences) > options.Limits.MaxTokens {
		return 0, fmt.Errorf("unit sentences exceed token limit")
	}
	visits, end := 0, 0
	for _, sentence := range sentences {
		visits += len(sentence.Tokens)
		if visits > options.Limits.MaxTokens || len(sentence.Tokens) == 0 {
			return 0, fmt.Errorf("unit sentence has no tokens or exceeds token limit")
		}
		if err := validateUnitSentenceAt(ctx, mapped, sentence, end); err != nil {
			return 0, err
		}
		end = sentence.Tokens[len(sentence.Tokens)-1].End
		if err := validateUnitRepresentations(sentence, options.Capabilities, len(mapped.Text)); err != nil {
			return 0, err
		}
	}
	if dependencyGapHasProse(mapped.Text[end:]) {
		return 0, fmt.Errorf("unit sentences omit prose")
	}
	if slices.Contains(options.Capabilities, Dependencies) {
		if err := ValidateDependencies(ctx, mapped, sentences); err != nil {
			return 0, err
		}
	}
	return visits, ctx.Err()
}

func validateUnitSentenceAt(ctx context.Context, mapped document.MappedText, sentence document.Sentence, end int) error {
	if err := validateDependencyTokens(ctx, mapped, sentence); err != nil {
		return fmt.Errorf("invalid unit sentence: %w", err)
	}
	start := sentence.Tokens[0].Start
	if start < end || dependencyGapHasProse(mapped.Text[end:start]) {
		return fmt.Errorf("unit sentences overlap or omit prose")
	}
	return nil
}

func validateUnitRepresentations(sentence document.Sentence, capabilities []Capability, bytes int) error {
	words := 0
	withPOS := slices.Contains(capabilities, POS) || slices.Contains(capabilities, Chunks)
	for _, token := range sentence.Tokens {
		if err := validateUnitWord(token, withPOS, bytes); err != nil {
			return err
		}
		if token.Word {
			words++
		}
	}
	if sentence.Words != words || len(sentence.Chunks) > len(sentence.Tokens) {
		return fmt.Errorf("unit word count or chunk count does not match tokens")
	}
	for _, chunk := range sentence.Chunks {
		if !validUnitChunk(chunk, len(sentence.Tokens)) {
			return fmt.Errorf("unit chunk has invalid bounds or kind")
		}
	}
	return nil
}

func validateUnitWord(token document.Token, withPOS bool, bytes int) error {
	if token.Word != document.IsWord(token.Text) || len(token.Normal) > bytes*4 || len(token.Tag) > 128 ||
		!utf8.ValidString(token.Normal) || !utf8.ValidString(token.Tag) {
		return fmt.Errorf("unit token has an invalid word, normalization, or tag representation")
	}
	if token.Word && (token.Normal == "" || (withPOS && token.Tag == "")) {
		return fmt.Errorf("unit word is missing normalization or requested POS")
	}
	return nil
}

func validUnitChunk(chunk document.Chunk, count int) bool {
	return chunk.FirstToken >= 0 && chunk.EndToken > chunk.FirstToken && chunk.EndToken <= count &&
		slices.Contains([]string{"NP", "VP", "PP"}, chunk.Kind)
}
