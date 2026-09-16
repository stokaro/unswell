package builtin_test

import (
	"context"
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func TestFrameRulesPropagateObservationCancellationAndBudget(t *testing.T) {
	for _, implementation := range builtin.Rules() {
		id := implementation.Descriptor().ID
		if id != "syntax.repeated-reframing" && id != "filler.document-metadiscourse" {
			continue
		}
		t.Run(id, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: "paragraph"}}},
				Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer, MaxCandidates: 100}
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
			c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{{BlockID: 0, Status: "inapplicable", Reason: "no_prose_words"}})
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			c.Assert(implementation.Evaluate(ctx, view, observer), qt.ErrorIs, context.Canceled)
			view.Document.Blocks[0].Sentences = []document.Sentence{{ID: 0, BlockID: 0, Text: "Text."}}
			view.MaxCandidates = 0
			var abstention *rule.Abstention
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorAs, &abstention)
			c.Assert(abstention.Reason, qt.Equals, rule.ReasonBudgetExhausted)
		})
	}
}

func TestFrameCatalogDeclaresNonblockingRecognition(t *testing.T) {
	c := qt.New(t)
	count := 0
	for _, implementation := range builtin.Rules() {
		d := implementation.Descriptor()
		if d.ID != "syntax.repeated-reframing" && d.ID != "filler.document-metadiscourse" {
			continue
		}
		count++
		c.Assert(d.Defaults.Enabled, qt.IsTrue)
		c.Assert(d.Defaults.Score, qt.Equals, rule.Score{})
		c.Assert(d.Defaults.Gate, qt.Equals, "none")
		c.Assert(d.Defaults.Severity, qt.Equals, "note")
		c.Assert(d.Limitations, qt.Contains, "unqualified")
	}
	c.Assert(count, qt.Equals, 2)
}
