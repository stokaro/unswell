package builtin

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func duplicateListRule() rule.Rule {
	d := descriptor("repetition.duplicate-list-item",
		"This complete list item is repeated in the same list; check whether both entries describe the same change.",
		"repetition", "document", 15)
	d.RequiresStructure, d.BlockObservations = true, true
	d.Contexts = []string{"list-item"}
	d.Description = "Compares complete simple unordered items within one grammar list and section," +
		" preserving case, punctuation and references."
	d.Limitations += " Only a single identifier may compare across plain text and inline-code formatting."
	d.Limitations += " Other protected content, ordered/task lists and complex items are inapplicable."
	d.Limitations += " Identity is not a semantic paraphrase test."
	d.Defaults.Parameters = rule.Parameters{MinWords: 4}
	d.Parameters = []string{"min_words"}
	d.Examples = []rule.Example{
		{Text: "- Add documentation for `Literal` type, #651.\n- Add documentation for Literal type, #651.\n",
			Format: document.Markdown, Match: true},
		{Text: "- Add documentation for `Literal` type, #651.\n- Add documentation for Literal type, #652.\n", Format: document.Markdown},
	}
	return check{d, duplicateListItems}
}

func duplicateListItems(ctx context.Context, view rule.View, emit rule.Emitter) error {
	scopes, err := repetitionScopes(ctx, view)
	if err != nil {
		return err
	}
	budget := &repetitionBudget{ctx, view.MaxCandidates}
	groups := make(map[string][]rule.Occurrence)
	for _, block := range view.Document.Blocks {
		if err := budget.spend(1); err != nil {
			return err
		}
		if reason := duplicateListReason(block, view.Parameters.MinWords); reason != "" {
			if err := observeBlock(view, block, reason); err != nil {
				return err
			}
			continue
		}
		key, err := listIdentity(block, view.Document.Source, budget)
		if err != nil {
			return err
		}
		reason := "no_eligible_tokens"
		if key != "" {
			reason = ""
			key = fmt.Sprintf("%d/list:%d/%q", scopes[block.ID], block.List.ID, key)
			groups[key] = append(groups[key], sentenceOccurrence(block.Sentences[0]))
		}
		if err := observeBlock(view, block, reason); err != nil {
			return err
		}
	}
	return emitGroups(groups, 1, 2, "exact", emit)
}

func duplicateListReason(block document.Block, minimum int) string {
	if block.Excluded || !grammarListBlock(&block) || block.List.Ordered || block.List.Task || block.List.Complex {
		return "unsupported_unit"
	}
	if len(block.Sentences) != 1 {
		return "unsupported_unit"
	}
	if block.Words < minimum {
		return "insufficient_words"
	}
	return ""
}

// This identity operation never supplies restored code to an editorial matcher.
// Whitespace folds only in prose. A code atom retains its case and punctuation;
// arbitrary commands, links and expressions remain outside this comparison.
func listIdentity(block document.Block, source []byte, budget *repetitionBudget) (string, error) {
	if err := budget.spend(2*len(block.Text) + block.Span.End - block.Span.Start); err != nil {
		return "", err
	}
	eligible, err := listMappingEligible(block, source)
	if err != nil || !eligible {
		return "", err
	}
	var key strings.Builder
	for i := range len(block.Text) {
		span := block.Map[i]
		if block.Text[i] != 0 {
			key.WriteByte(block.Text[i])
			continue
		}
		atom := inlineIdentifier(string(source[span.Start:span.End]))
		if atom == "" {
			return "", nil
		}
		key.WriteString(atom)
	}
	return strings.Join(strings.Fields(key.String()), " "), nil
}

func listMappingEligible(block document.Block, source []byte) (bool, error) {
	if !block.Span.Valid(len(source)) || len(block.Text) != len(block.Map) {
		return false, fmt.Errorf("list identity requires a complete source map")
	}
	end := block.Span.Start
	for _, span := range block.Map {
		if !span.Valid(len(source)) || span.Start < block.Span.Start || span.End > block.Span.End {
			return false, fmt.Errorf("list identity has an invalid source span")
		}
		if span.Start > end && !emphasisGap(string(source[end:span.Start])) {
			return false, nil
		}
		end = max(end, span.End)
	}
	return emphasisGap(string(source[end:block.Span.End])), nil
}

// Link targets and reference labels can disappear from extracted prose. Only
// emphasis delimiters may disappear from a complete-item identity comparison.
func emphasisGap(gap string) bool {
	for _, r := range gap {
		if r != '*' && r != '_' && !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

func inlineIdentifier(source string) string {
	width := len(source) - len(strings.TrimLeft(source, "`"))
	if width == 0 || len(source) <= 2*width || !strings.HasSuffix(source, strings.Repeat("`", width)) {
		return ""
	}
	atom := source[width : len(source)-width]
	for i, r := range atom {
		if !unicode.IsLetter(r) && r != '_' && (i == 0 || (!unicode.IsDigit(r) && r != '.')) {
			return ""
		}
	}
	return atom
}
