// Package builtin constructs the versioned alpha rule catalog without init hooks.
package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

type check struct {
	descriptor rule.Descriptor
	evaluate   func(context.Context, rule.View, rule.Emitter) error
}

// Descriptor returns this explicitly registered rule's contract.
func (c check) Descriptor() rule.Descriptor { return c.descriptor }

// Evaluate executes the configured signal against a read-only document view.
func (c check) Evaluate(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return c.evaluate(ctx, view, emit)
}

// Rules returns a fresh catalog. Heuristics remain experimental pending corpus validation.
func Rules() []rule.Rule {
	result := phraseRules()
	result = append(result, contextRules()...)
	result = append(result, editorialRules()...)
	result = append(result, extendedRepetitionRules()...)
	result = append(result, surfaceRules()...)
	slices.SortFunc(result, func(a, b rule.Rule) int {
		if a.Descriptor().ID < b.Descriptor().ID {
			return -1
		}
		if a.Descriptor().ID > b.Descriptor().ID {
			return 1
		}
		return 0
	})
	return result
}

func descriptor(id, summary, group, scope string, weight int) rule.Descriptor {
	return rule.Descriptor{
		ID: id, Version: "1", Summary: summary, Description: summary, Group: group, Scope: scope, Status: "experimental",
		Contexts: []string{"paragraph", "comment", "string", "heading", "list-item", "table-cell"},
		Requires: []nlp.Capability{nlp.Tokens, nlp.Sentences},
		Defaults: rule.Settings{
			Enabled:  true,
			Severity: "warning",
			Gate:     "none",
			Score:    rule.Score{Weight: weight, Cap: weight * 2},
		},
		Limitations: "Alpha defaults are editorial policy proposals; corpus precision has not been established.",
		Parameters:  []string{}, Examples: []rule.Example{},
	}
}
