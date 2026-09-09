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

func TestCandidateObservationsRetainEmptyBlocksAndFailures(t *testing.T) {
	ids := []string{"format.list-fragmentation", "repetition.heading-echo", "repetition.near-sentence", "repetition.ngram-density",
		"repetition.paragraph-overlap", "repetition.summary-echo", "repetition.syntax-template"}
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if !slices.Contains(ids, d.ID) {
			continue
		}
		t.Run(d.ID, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{
				{ID: 0, Kind: "paragraph"}, {ID: 1, Kind: "heading"}, {ID: 2, Kind: "list-item"}}},
				Parameters: d.Defaults.Parameters, Observer: observer, MaxCandidates: 100}
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
			reasons := []string{"no_sentences", "unsupported_unit", "unsupported_unit"}
			switch d.ID {
			case "repetition.near-sentence":
				reasons = []string{"no_sentences", "no_sentences", "no_sentences"}
			case "repetition.heading-echo":
				reasons[1] = "no_sentences"
			case "format.list-fragmentation":
				reasons = []string{"unsupported_unit", "unsupported_unit", "no_sentences"}
			}
			var want []feature.BlockObservation
			for i, reason := range reasons {
				want = append(want, feature.BlockObservation{BlockID: i, Status: "inapplicable", Reason: reason})
			}
			c.Assert(observer.values, qt.DeepEquals, want)
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			c.Assert(implementation.Evaluate(ctx, view, observer), qt.ErrorIs, context.Canceled)
			view.Observer = nil
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
		})
	}
}
