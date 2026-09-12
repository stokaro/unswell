package feature

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

func validateLimits(limits Limits) error {
	if limits.MaxTokens < 1 || limits.MaxTokens > 1<<30 || limits.MaxUniqueWords < 1 ||
		limits.MaxUniqueWords > limits.MaxTokens || limits.MaxBytes < 1 || limits.MaxBytes > 1<<30 ||
		limits.MaxBlocks < 1 || limits.MaxBlocks > 1<<30 {
		return fmt.Errorf("invalid feature limits")
	}
	return nil
}

func validateIdentity(identity Identity) error {
	for _, value := range []string{identity.NLP.Name, identity.NLP.Version, identity.Source,
		identity.Policy, identity.Vocabulary, identity.Preprocessing} {
		if strings.TrimSpace(value) == "" || len(value) > 1024 || !utf8.ValidString(value) {
			return fmt.Errorf("feature identity requires bounded, nonempty identifiers")
		}
	}

	for _, value := range []string{identity.NLP.Model, identity.NLP.ModelHash, identity.NLP.SentenceModelHash,
		identity.NLP.License, identity.NLP.DependencyScheme} {
		if len(value) > 1024 || !utf8.ValidString(value) {
			return fmt.Errorf("feature NLP identity exceeds metadata limits")
		}
	}
	return validateCapabilities(identity)
}

func validateCapabilities(identity Identity) error {
	if len(identity.Capabilities) > 7 || len(identity.NLP.Capabilities) > 7 {
		return fmt.Errorf("feature identity has too many capabilities")
	}
	for _, capability := range identity.NLP.Capabilities {
		if len(capability) > 128 {
			return fmt.Errorf("NLP capability identity exceeds metadata limits")
		}
	}
	for _, capability := range identity.Capabilities {
		if !slices.Contains([]nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS, nlp.Chunks,
			nlp.Dependencies, nlp.Lemmas, nlp.Entities}, capability) {
			return fmt.Errorf("unknown feature input capability %q", capability)
		}
		if !slices.Contains(identity.NLP.Capabilities, capability) {
			return fmt.Errorf("feature input capability %s is not supported by its NLP identity", capability)
		}
	}
	return nil
}

func validateInputs(ctx context.Context, block document.Block, identity Identity, limits Limits) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := validateLimits(limits); err != nil {
		return err
	}
	if err := validateIdentity(identity); err != nil {
		return err
	}
	if err := validateMappedBlock(block, limits); err != nil {
		return err
	}
	if block.Excluded {
		if len(block.Sentences) != 0 || block.Words != 0 {
			return fmt.Errorf("excluded feature block must not contain NLP data")
		}
		return nil
	}
	return validateTokens(ctx, block, identity, limits)
}

func validateMappedBlock(block document.Block, limits Limits) error {
	if block.ID < 0 || len(block.Kind) > 128 {
		return fmt.Errorf("feature block has an invalid identity")
	}
	if len(block.Text) > limits.MaxBytes || len(block.Sentences) > limits.MaxTokens || len(block.Context) > 256 {
		return fmt.Errorf("feature input exceeds text, sentence, or context limits")
	}
	if len(block.Map) != len(block.Text) || !utf8.ValidString(block.Text) {
		return fmt.Errorf("feature input has an invalid mapped text")
	}
	bytes := 0
	for _, part := range block.Context {
		bytes += len(part)
		if bytes > limits.MaxBytes {
			return fmt.Errorf("feature context exceeds max_bytes")
		}
	}
	return nil
}

func validateTokens(ctx context.Context, block document.Block, identity Identity, limits Limits) error {
	last := 0
	budget := tokenBudget{limits: limits}
	for _, sentence := range block.Sentences {
		for _, token := range sentence.Tokens {
			if err := ctx.Err(); err != nil {
				return err
			}
			if err := budget.add(token); err != nil {
				return err
			}
			if err := validateToken(block.MappedText, token, last, identity.Capabilities); err != nil {
				return err
			}
			last = token.End
		}
	}
	if slices.Contains(identity.Capabilities, nlp.Tokens) && proseGap(block.Text[last:]) {
		return fmt.Errorf("feature tokens omit extracted prose")
	}
	return nil
}

type tokenBudget struct {
	limits Limits
	visits int
	bytes  int64
}

func (b *tokenBudget) add(token document.Token) error {
	b.visits++
	b.bytes += int64(len(token.Text)) + int64(len(token.Normal)) + int64(len(token.Tag))
	if b.bytes > int64(b.limits.MaxBytes)*16 {
		return fmt.Errorf("feature token representations exceed byte budget")
	}
	if b.visits > b.limits.MaxTokens {
		return ErrTokenLimit
	}
	return nil
}

func validateToken(mapped document.MappedText, token document.Token, last int, capabilities []nlp.Capability) error {
	if token.Start < last || !(document.Span{Start: token.Start, End: token.End}).Valid(len(mapped.Text)) ||
		token.Text != mapped.Text[token.Start:token.End] {
		return fmt.Errorf("feature token has invalid bounds or text")
	}
	if proseGap(mapped.Text[last:token.Start]) {
		return fmt.Errorf("feature tokens omit extracted prose")
	}
	if err := validateSegments(mapped, token); err != nil {
		return err
	}
	return validateRepresentation(mapped.Text, token, capabilities)
}

func validateRepresentation(text string, token document.Token, capabilities []nlp.Capability) error {
	if len(token.Normal) > len(text)*4 || len(token.Tag) > 128 {
		return fmt.Errorf("feature token exceeds normalization or tag limits")
	}
	if !utf8.ValidString(token.Text) || !utf8.ValidString(token.Normal) || !utf8.ValidString(token.Tag) {
		return fmt.Errorf("feature token representation is not UTF-8")
	}
	return validateWord(token, capabilities)
}

func validateWord(token document.Token, capabilities []nlp.Capability) error {
	if !token.Word || token.Protected {
		return nil
	}
	if token.Normal == "" || !utf8.ValidString(token.Normal) || strings.ContainsRune(token.Text, 0) {
		return fmt.Errorf("feature word has invalid normalization")
	}
	if slices.Contains(capabilities, nlp.POS) && token.Tag == "" {
		return fmt.Errorf("feature input requested POS but an eligible word has no tag")
	}
	return nil
}

func validateSegments(mapped document.MappedText, token document.Token) error {
	if len(token.Spans) > token.End-token.Start || !slices.Equal(token.Spans, mapped.Spans(token.Start, token.End)) {
		return fmt.Errorf("feature token has invalid source segments")
	}
	for _, span := range token.Spans {
		if span.Start < 0 || span.End < span.Start {
			return fmt.Errorf("feature token has invalid original coordinates")
		}
	}
	return nil
}

func unitHash(block document.Block, identity Identity) (string, error) {
	identity.Capabilities = orderedCapabilities(identity.Capabilities)
	identity.NLP.Capabilities = orderedCapabilities(identity.NLP.Capabilities)
	// The map is represented by the token segments. Trees and chunks are not
	// inputs to these formulas and must not grow the serialized representation.
	tokens := make([][]document.Token, len(block.Sentences))
	for i, sentence := range block.Sentences {
		tokens[i] = sentence.Tokens
	}
	input := struct {
		Contract    string
		Definitions []Descriptor
		Identity    Identity
		ID          int
		Kind, Text  string
		Context     []string
		Span        document.Span
		Tokens      [][]document.Token
		Excluded    bool `json:",omitempty"`
	}{Contract, catalog(), identity, block.ID, block.Kind, block.Text, block.Context, block.Span, tokens, block.Excluded}
	hash := sha256.New()
	if err := json.NewEncoder(hash).Encode(input); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func orderedCapabilities(input []nlp.Capability) []nlp.Capability {
	result := slices.Clone(input)
	slices.Sort(result)
	return slices.Compact(result)
}

func proseGap(text string) bool {
	return strings.ContainsFunc(text, func(r rune) bool { return r != 0 && !unicode.IsSpace(r) })
}
