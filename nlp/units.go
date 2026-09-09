package nlp

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/document"
)

// UnitContract identifies corpus-compatible target and surrounding prose selection.
const UnitContract = "eligible-piece-v1"

// UnitLimits bounds one block's preparation, including all its eligible pieces.
type UnitLimits struct {
	MaxBytes, MaxContextBytes, MaxUnits, MaxTokens, MaxSegments int
}

// UnitOptions selects target kinds and actual NLP representations explicitly.
// Kinds accepts sentence, paragraph, and fragment. Tokens and Sentences are required.
type UnitOptions struct {
	Kinds        []string
	Capabilities []Capability
	Limits       UnitLimits
}

// UnitBinding identifies complete extracted target/context ranges, not token spans.
// ContextSHA256 hashes prose; GrammarSHA256 hashes the JSON grammar scope array.
type UnitBinding struct {
	Contract      string          `json:"contract"`
	Kind          string          `json:"kind"`
	BlockID       int             `json:"block_id"`
	BlockKind     string          `json:"block_kind"`
	TextSHA256    string          `json:"text_sha256"`
	ContextSHA256 string          `json:"context_sha256"`
	GrammarSHA256 string          `json:"grammar_sha256"`
	Segments      []document.Span `json:"segments"`
	ContextSpans  []document.Span `json:"context_segments"`
}

// PreparedUnit owns one target and its enclosing prose. Its zero value is invalid.
// Accessors return detached data and support concurrent use.
type PreparedUnit struct {
	block        document.Block
	context      string
	binding      UnitBinding
	provider     Identity
	capabilities []Capability
}

// Binding returns the complete target/context identity without source text.
func (u PreparedUnit) Binding() UnitBinding {
	b := u.binding
	b.Segments = slices.Clone(b.Segments)
	b.ContextSpans = slices.Clone(b.ContextSpans)
	return b
}

// Block returns detached NLP data with token offsets relative to target Text.
// Kind remains the original extraction kind; Binding.Kind is the target scope.
func (u PreparedUnit) Block() document.Block { return cloneUnitBlock(u.block) }

// Context returns the exact extracted enclosing prose used during preparation.
func (u PreparedUnit) Context() string { return u.context }

// Identity returns the provider metadata captured during preparation.
func (u PreparedUnit) Identity() Identity {
	i := u.provider
	i.Capabilities = slices.Clone(i.Capabilities)
	return i
}

// Capabilities returns the representations actually requested for this unit.
func (u PreparedUnit) Capabilities() []Capability { return slices.Clone(u.capabilities) }

// PrepareUnits selects targets from one extracted block and analyzes each piece
// once. It preserves protected boundaries and returns no partial result on error.
// The provider is trusted Go code and must honor its existing contract and limits.
func PrepareUnits(ctx context.Context, block document.Block, provider Provider, options UnitOptions) ([]PreparedUnit, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validateUnitOptions(provider, options); err != nil {
		return nil, err
	}
	if err := validateUnitBlock(ctx, block, options.Limits); err != nil {
		return nil, err
	}
	units, err := preparePieces(ctx, block, provider, options)
	if err != nil {
		return nil, err
	}
	identity := provider.Identity()
	identity.Capabilities = slices.Clone(identity.Capabilities)
	capabilities := slices.Clone(options.Capabilities)
	for i := range units {
		units[i].provider, units[i].capabilities = identity, capabilities
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return units, nil
}

func bindUnitContext(block document.Block, context document.MappedText) (UnitBinding, error) {
	grammar, err := json.Marshal(block.Context)
	if err != nil {
		return UnitBinding{}, err
	}
	return UnitBinding{Contract: UnitContract, BlockID: block.ID, BlockKind: block.Kind,
		ContextSHA256: fmt.Sprintf("%x", sha256.Sum256([]byte(context.Text))),
		GrammarSHA256: fmt.Sprintf("%x", sha256.Sum256(grammar)),
		ContextSpans:  context.Spans(0, len(context.Text))}, nil
}

func cloneUnitBlock(block document.Block) document.Block {
	block.Context = slices.Clone(block.Context)
	block.Map = slices.Clone(block.Map)
	block.List = nil // List aggregation is not an input to these isolated targets.
	block.Sentences = cloneUnitSentences(block.Sentences, 0)
	return block
}

func cloneUnitSentences(input []document.Sentence, offset int) []document.Sentence {
	result := slices.Clone(input)
	for i := range result {
		s := &result[i]
		s.Spans = slices.Clone(s.Spans)
		s.Chunks = slices.Clone(s.Chunks)
		if s.Dependencies != nil {
			s.Dependencies = &document.DependencyTree{Arcs: slices.Clone(s.Dependencies.Arcs)}
		}
		s.Tokens = slices.Clone(s.Tokens)
		for j := range s.Tokens {
			token := &s.Tokens[j]
			token.Start -= offset
			token.End -= offset
			token.Spans = slices.Clone(token.Spans)
		}
	}
	return result
}
