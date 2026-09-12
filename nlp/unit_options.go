package nlp

import (
	"context"
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

func validateUnitOptions(provider Provider, options UnitOptions) error {
	if !validUnitLimits(options.Limits) {
		return fmt.Errorf("invalid unit preparation limits")
	}
	if len(options.Kinds) < 1 || len(options.Kinds) > 3 {
		return fmt.Errorf("unit preparation requires 1 through 3 target kinds")
	}
	for i, kind := range options.Kinds {
		if !slices.Contains([]string{"sentence", "paragraph", "fragment"}, kind) || slices.Contains(options.Kinds[:i], kind) {
			return fmt.Errorf("unknown or duplicate unit kind %q", kind)
		}
	}
	return validateUnitCapabilities(provider, options.Capabilities)
}

func validUnitLimits(l UnitLimits) bool {
	return boundedUnitLimit(l.MaxBytes, 16<<20) && boundedUnitLimit(l.MaxContextBytes, 65536) &&
		boundedUnitLimit(l.MaxUnits, 100000) && boundedUnitLimit(l.MaxTokens, 1000000) && boundedUnitLimit(l.MaxSegments, 4096)
}

func boundedUnitLimit(value, maximum int) bool { return value >= 1 && value <= maximum }

func validateUnitCapabilities(provider Provider, capabilities []Capability) error {
	if provider == nil || len(capabilities) > 5 || !slices.Contains(capabilities, Tokens) || !slices.Contains(capabilities, Sentences) {
		return fmt.Errorf("unit preparation requires a provider, tokens, and sentences")
	}
	supported := provider.Identity().Capabilities
	for i, capability := range capabilities {
		if !slices.Contains([]Capability{Tokens, Sentences, POS, Chunks, Dependencies}, capability) ||
			slices.Contains(capabilities[:i], capability) || !slices.Contains(supported, capability) {
			return fmt.Errorf("unavailable or duplicate unit capability %q", capability)
		}
	}
	return nil
}

func validateUnitBlock(ctx context.Context, block document.Block, limits UnitLimits) error {
	if !validUnitBlockShape(block, limits) {
		return fmt.Errorf("invalid mapped unit block or byte limit")
	}
	bytes := 0
	for _, label := range block.Context {
		bytes += len(label)
		if !utf8.ValidString(label) || bytes > limits.MaxBytes {
			return fmt.Errorf("unit grammar context exceeds byte limit or is invalid UTF-8")
		}
	}
	return validateUnitMap(ctx, block)
}

func validateUnitMap(ctx context.Context, block document.Block) error {
	for i, span := range block.Map {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !validUnitOrigin(span, block.Span, block.Text[i] == 0) {
			return fmt.Errorf("unit source map exceeds block bounds or has an empty prose span")
		}
		if i > 0 && span != block.Map[i-1] && span.Start < block.Map[i-1].End {
			return fmt.Errorf("unit source map is not ordered")
		}
	}
	return validateUnitRunes(block.MappedText)
}

func validUnitBlockShape(block document.Block, limits UnitLimits) bool {
	return block.ID >= 0 && block.Kind != "" && len(block.Kind) <= 128 && len(block.Context) <= 256 &&
		len(block.Text) <= limits.MaxBytes && len(block.Text) == len(block.Map) && utf8.ValidString(block.Text)
}

func validUnitOrigin(span, bounds document.Span, protected bool) bool {
	return bounds.Start >= 0 && bounds.End >= bounds.Start && span.Start >= bounds.Start && span.End <= bounds.End &&
		(span.End > span.Start || (protected && span.Start == span.End))
}

// validateUnitRunes checks the origins of an encoded rune. Every byte of one
// rune points at the same origin, or at a contiguous run of origins. A Go
// string literal can spell one rune with several escape sequences. The
// literal "\xef\xbb\xbf" spells U+FEFF that way. Each sequence is its own
// source span. A gap or a reversal between the spans would place evidence
// outside the rune.
func validateUnitRunes(mapped document.MappedText) error {
	for start, r := range mapped.Text {
		for i := 1; i < utf8.RuneLen(r); i++ {
			previous, current := mapped.Map[start+i-1], mapped.Map[start+i]
			if current != previous && current.Start != previous.End {
				return fmt.Errorf("unit source map splits an encoded rune")
			}
		}
	}
	return nil
}
