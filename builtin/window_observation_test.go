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

func TestWindowObservationsRetainEmptyBlocksAndObserverErrors(t *testing.T) {
	ids := []string{"syntax.not-only-density", "syntax.paired-contrast-density", "syntax.triad-density",
		"syntax.whether-preface-density", "syntax.rhetorical-question-density", "syntax.passive-candidate-density"}
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
			c.Assert(observer.values, qt.DeepEquals, []feature.BlockObservation{{BlockID: 0, Status: "inapplicable", Reason: "no_sentences"}})
			observer.err = errors.New("observation unavailable")
			c.Assert(implementation.Evaluate(t.Context(), view, observer), qt.ErrorIs, observer.err)
		})
	}
}

func TestWindowObservationsPreserveMatcherBudget(t *testing.T) {
	ids := []string{"syntax.not-only-density", "syntax.paired-contrast-density", "syntax.triad-density",
		"syntax.whether-preface-density", "syntax.rhetorical-question-density", "syntax.passive-candidate-density"}
	sentences := []document.Sentence{
		{ID: 0, BlockID: 0, Text: "not only powerful", Words: 3, Tokens: []document.Token{
			{Normal: "not", Word: true}, {Normal: "only", Word: true}, {Normal: "powerful", Word: true, Tag: "JJ"}}},
		{ID: 1, BlockID: 0, Text: "The result?", Words: 2, Tokens: []document.Token{
			{Normal: "the", Word: true}, {Normal: "result", Word: true}, {Normal: "?"}}},
		{ID: 2, BlockID: 0, Text: "An answer.", Words: 2, Tokens: []document.Token{
			{Normal: "an", Word: true}, {Normal: "answer", Word: true}, {Normal: "."}}},
	}
	for _, implementation := range builtin.Rules() {
		if !slices.Contains(ids, implementation.Descriptor().ID) {
			continue
		}
		t.Run(implementation.Descriptor().ID, func(t *testing.T) {
			c := qt.New(t)
			observer := &phraseObserver{}
			view := rule.View{Document: &document.Document{Blocks: []document.Block{{ID: 0, Kind: "comment", Sentences: sentences}}},
				Parameters: implementation.Descriptor().Defaults.Parameters, Observer: observer}
			view.Parameters.MinWords = 0
			err := implementation.Evaluate(t.Context(), view, observer)
			c.Assert(err, qt.ErrorMatches, ".*editorial pattern checks exceed max_candidates.*")
			var abstention *rule.Abstention
			c.Assert(err, qt.ErrorAs, &abstention)
			c.Assert(abstention.Reason, qt.Equals, rule.ReasonBudgetExhausted)
			c.Assert(observer.values, qt.HasLen, 0)
			view.Observer = nil
			err = implementation.Evaluate(t.Context(), view, observer)
			c.Assert(err, qt.ErrorMatches, ".*editorial pattern checks exceed max_candidates.*")
			c.Assert(err, qt.ErrorAs, &abstention)
		})
	}
}
