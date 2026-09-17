package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// Heading/container labels describe sections, while heading IDs distinguish successive
// sections with identical titles. Cells are independent comparison units: equal
// values in different rows or columns do not establish redundant information.
func repetitionScopes(ctx context.Context, view rule.View) (map[int]int, error) {
	scopes := make(map[int]int, len(view.Document.Blocks))
	identities := make(map[string]int)
	budget := repetitionBudget{ctx, view.MaxCandidates}
	headings := make(map[string]int)
	for _, block := range view.Document.Blocks {
		if err := budget.spend(1); err != nil {
			return nil, err
		}
		owner, err := repetitionSection(block.Context, &budget)
		if err != nil {
			return nil, err
		}
		if block.Kind == "heading" {
			headings[owner] = block.ID + 1
		}
		key := fmt.Sprintf("%s/%d", owner, headings[owner])
		if block.Kind == "heading" || block.Kind == "table-cell" {
			key += fmt.Sprintf("/block:%d", block.ID)
		}
		if len(block.Sentences) > 0 {
			key += leadingCondition(block)
		}
		if _, exists := identities[key]; !exists {
			identities[key] = len(identities) + 1
		}
		scopes[block.ID] = identities[key]
	}
	return scopes, nil
}

func repetitionSection(labels []string, budget *repetitionBudget) (string, error) {
	var section []string
	for _, label := range labels {
		if strings.HasPrefix(label, "heading-") || strings.HasPrefix(label, "container:") {
			// A heading can recur in many blocks. Charge its bytes before copying.
			if err := budget.spend(len(label)); err != nil {
				return "", err
			}
			section = append(section, label)
		}
	}
	return fmt.Sprintf("%q", section), nil
}

// Compare explicit task antecedents, not just "If you're a". An unresolved
// antecedent stays local to its block rather than joining unrelated variants.
func leadingCondition(block document.Block) string {
	tokens := block.Sentences[0].Tokens
	if len(tokens) == 0 || !slices.Contains([]string{"if", "when", "unless"}, tokens[0].Normal) {
		return ""
	}
	var condition []string
	for i, token := range tokens {
		if i >= 64 || token.Protected {
			return fmt.Sprintf("/condition-block:%d", block.ID)
		}
		if token.Text == "," || token.Text == ":" || token.Normal == "then" {
			return fmt.Sprintf("/condition:%q", condition)
		}
		condition = append(condition, token.Text)
	}
	return fmt.Sprintf("/condition-block:%d", block.ID)
}

func scopeShingles(shingles []string, scope int) {
	for j := range shingles {
		shingles[j] = fmt.Sprintf("%d/%s", scope, shingles[j])
	}
}
