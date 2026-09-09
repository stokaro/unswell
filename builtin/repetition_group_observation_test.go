package builtin_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func TestRepetitionGroupObservationsRetainEmptyBlocksAndFailures(t *testing.T) {
	ids := []string{"repetition.exact-sentence", "repetition.sentence-openers", "repetition.paragraph-openers"}
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(ids, d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: "paragraph"}, {ID: 1, Kind: "heading"}}},
				Parameters: d.Defaults.Parameters, Observer: observer}
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
			headingReason := "unsupported_unit"
			if d.ID == "repetition.exact-sentence" {
				headingReason = "no_sentences"
			}
			c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{
				{BlockID: 0, Status: "inapplicable", Reason: "no_sentences"},
				{BlockID: 1, Status: "inapplicable", Reason: headingReason},
			})
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			c.Assert(implementation.Evaluate(ctx, view, observer), qt.ErrorIs, context.Canceled)
		})
	}
}
