package nlp

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
)

// ValidateDependencies checks that complete, ordered trees cover all extracted
// prose in one block. Only whitespace and protected NUL boundaries may lie
// outside sentences. It also applies ValidateDependencyTree to each sentence.
func ValidateDependencies(ctx context.Context, mapped document.MappedText, sentences []document.Sentence) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(mapped.Map) != len(mapped.Text) {
		return fmt.Errorf("dependency input has an invalid source map")
	}
	end := 0
	for _, sentence := range sentences {
		if err := ValidateDependencyTree(ctx, mapped, sentence); err != nil {
			return err
		}
		start := sentence.Tokens[0].Start
		if start < end || dependencyGapHasProse(mapped.Text[end:start]) {
			return fmt.Errorf("dependency sentences overlap or omit extracted prose")
		}
		end = sentence.Tokens[len(sentence.Tokens)-1].End
	}
	if dependencyGapHasProse(mapped.Text[end:]) {
		return fmt.Errorf("dependency sentences omit extracted prose")
	}
	return ctx.Err()
}

// ValidateDependencyTree checks a complete basic tree and its source mapping.
// Token Start/End values are byte offsets in mapped.Text. Sentence Text and Spans
// must cover its first through last token. Labels are opaque, nonempty strings
// without whitespace or control characters. This validates structure, not
// linguistic correctness. It neither changes the tree nor repairs invalid arcs.
func ValidateDependencyTree(ctx context.Context, mapped document.MappedText, sentence document.Sentence) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	count := len(sentence.Tokens)
	if sentence.Dependencies == nil {
		return fmt.Errorf("requested dependency tree is missing")
	}
	if count == 0 || len(sentence.Dependencies.Arcs) != count {
		return fmt.Errorf("dependency tree must have one arc per token in a nonempty sentence")
	}
	if len(mapped.Map) != len(mapped.Text) {
		return fmt.Errorf("dependency input has an invalid source map")
	}
	if err := validateDependencyTokens(ctx, mapped, sentence); err != nil {
		return err
	}
	arcs := sentence.Dependencies.Arcs
	if err := validateDependencyArcs(ctx, arcs); err != nil {
		return err
	}
	return validateDependencyPaths(ctx, arcs)
}

func validateDependencyTokens(ctx context.Context, mapped document.MappedText, sentence document.Sentence) error {
	last := -1
	for i, token := range sentence.Tokens {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := validateDependencyToken(mapped, token, last); err != nil {
			return fmt.Errorf("dependency token %d: %w", i, err)
		}
		last = token.End
	}
	first := sentence.Tokens[0].Start
	if sentence.Text != mapped.Text[first:last] || strings.ContainsRune(sentence.Text, 0) {
		return fmt.Errorf("dependency sentence differs from mapped prose or crosses a protected boundary")
	}
	spans := mapped.Spans(first, last)
	if !slices.Equal(sentence.Spans, spans) || sentence.Span != document.Bounds(spans) {
		return fmt.Errorf("dependency sentence has invalid source spans")
	}
	return nil
}

func validateDependencyToken(mapped document.MappedText, token document.Token, last int) error {
	if !(document.Span{Start: token.Start, End: token.End}).Valid(len(mapped.Text)) || token.Start < last {
		return fmt.Errorf("invalid or overlapping byte range")
	}
	if last >= 0 && dependencyGapHasProse(mapped.Text[last:token.Start]) {
		return fmt.Errorf("dependency tokens omit extracted prose")
	}
	if token.Protected || token.Text != mapped.Text[token.Start:token.End] || !utf8.ValidString(token.Text) {
		return fmt.Errorf("protected token or text differs from mapped bytes")
	}
	if strings.ContainsRune(token.Text, 0) || !slices.Equal(token.Spans, mapped.Spans(token.Start, token.End)) {
		return fmt.Errorf("protected boundary or invalid source spans")
	}
	return nil
}

func validateDependencyArcs(ctx context.Context, arcs []document.DependencyArc) error {
	roots := 0
	for i, arc := range arcs {
		if err := ctx.Err(); err != nil {
			return err
		}
		if arc.Head < -1 || arc.Head >= len(arcs) || arc.Head == i {
			return fmt.Errorf("dependency token %d has invalid head %d", i, arc.Head)
		}
		if !validDependencyRelation(arc.Relation) {
			return fmt.Errorf("dependency token %d has an invalid relation label", i)
		}
		if arc.Head == -1 {
			roots++
		}
	}
	if roots != 1 {
		return fmt.Errorf("dependency tree has %d roots; require exactly one", roots)
	}
	return nil
}

func validateDependencyPaths(ctx context.Context, arcs []document.DependencyArc) error {
	// Each token enters a path at most once, so even a long chain is linear.
	state := make([]uint8, len(arcs))
	for start := range arcs {
		i := start
		for i != -1 && state[i] == 0 {
			if err := ctx.Err(); err != nil {
				return err
			}
			state[i] = 1
			i = arcs[i].Head
		}
		if i != -1 && state[i] == 1 {
			return fmt.Errorf("dependency tree contains a cycle at token %d", i)
		}
		for i = start; i != -1 && state[i] == 1; i = arcs[i].Head {
			state[i] = 2
		}
	}
	return nil
}

func validDependencyRelation(relation string) bool {
	return relation != "" && len(relation) <= 128 && utf8.ValidString(relation) &&
		!strings.ContainsFunc(relation, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) })
}

func dependencyGapHasProse(text string) bool {
	return strings.ContainsFunc(text, func(r rune) bool { return r != 0 && !unicode.IsSpace(r) })
}
