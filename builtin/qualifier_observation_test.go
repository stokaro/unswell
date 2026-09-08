package builtin_test

import (
	"errors"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func TestQualifierObservationsRetainEmptyBlocksAndObserverErrors(t *testing.T) {
	ids := []string{"hype.vague-praise", "hype.absolute-claim", "filler.weak-intensifiers", "filler.stacked-hedging"}
	for _, implementation := range builtin.Rules() {
		if !slices.Contains(ids, implementation.Descriptor().ID) {
			continue
		}
		t.Run(implementation.Descriptor().ID, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: "comment"}}},
				Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer}
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
			reason := "no_sentences"
			if implementation.Descriptor().ID == "filler.weak-intensifiers" {
				reason = "no_prose_words"
			}
			c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{{BlockID: 0, Status: "inapplicable", Reason: reason}})
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
		})
	}
}
