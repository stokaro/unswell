package builtin_test

import (
	"errors"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type phraseObserver struct {
	values []feature.BlockObservation
	err    error
}

func (o *phraseObserver) Observe(value feature.BlockObservation) error {
	o.values = append(o.values, value)
	return o.err
}

func (*phraseObserver) Emit(rule.Evidence) error {
	return errors.New("empty block must not emit evidence")
}

func TestPhraseObservationAccountsForBlocksWithoutSentences(t *testing.T) {
	c := qt.New(t)
	var implementation rule.Rule
	for _, candidate := range builtin.Rules() {
		if candidate.Descriptor().ID == "filler.wordy-phrase" {
			implementation = candidate
		}
	}
	c.Assert(implementation, qt.IsNotNil)
	c.Assert(implementation.Descriptor().BlockObservations, qt.IsTrue)
	observer := &phraseObserver{}
	view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: "comment"}}},
		Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer}
	c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.IsNil)
	c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{
		{BlockID: 0, Status: "inapplicable", Reason: "no_sentences"},
	})
	observer.err = errors.New("observer unavailable")
	c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
}
