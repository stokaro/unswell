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

func TestWindowPhraseObservationsAccountForEmptyAndUnsupportedBlocks(t *testing.T) {
	ids := []string{"filler.section-announcement", "filler.empty-transition", "hype.metaphor-cluster"}
	for _, implementation := range builtin.Rules() {
		if !slices.Contains(ids, implementation.Descriptor().ID) {
			continue
		}
		t.Run(implementation.Descriptor().ID, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{
				{ID: 0, Kind: "paragraph"}, {ID: 1, Kind: "heading"},
			}}, Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer}
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
			c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{
				{BlockID: 0, Status: "inapplicable", Reason: "no_sentences"},
				{BlockID: 1, Status: "inapplicable", Reason: "unsupported_unit"},
			})
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
		})
	}
}
