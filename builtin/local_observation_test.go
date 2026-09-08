package builtin_test

import (
	"errors"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestLocalObservationErrorsPropagateFromSkippedBlocks(t *testing.T) {
	ids := []string{"syntax.long-sentence", "hype.modifier-cluster", "density.connective-overuse",
		"syntax.nominalization-chain", "syntax.noun-stack", "syntax.parenthetical-load", "format.em-dash-density"}
	for _, implementation := range builtin.Rules() {
		if !slices.Contains(ids, implementation.Descriptor().ID) {
			continue
		}
		for _, kind := range []string{"paragraph", "heading"} {
			t.Run(implementation.Descriptor().ID+"/"+kind, func(t *testing.T) {
				c := qt.New(t)
				observer := &phraseObserver{err: errors.New("observation unavailable")}
				view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: kind}}},
					Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer}
				c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
				c.Assert(observer.values, qt.HasLen, 1)
				c.Assert(observer.values[0].Status, qt.Equals, "inapplicable")
			})
		}
	}
}
