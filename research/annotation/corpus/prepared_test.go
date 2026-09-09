package corpus_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/research/annotation/internal/testfixture"
)

func TestPreparedCorpusReusesExactMappedTargets(t *testing.T) {
	c := qt.New(t)
	input := testfixture.Load(t, "../training/testdata")
	candidates, _ := input.Compile(t)
	prepared, err := corpus.Prepare(t.Context(), candidates, input.Files)
	c.Assert(err, qt.IsNil)
	c.Assert(prepared.Verification.Status, qt.Equals, "source_and_candidates_reproduced")
	c.Assert(prepared.Targets, qt.HasLen, len(candidates.Units))
	for i, target := range prepared.Targets {
		candidate := candidates.Units[i]
		c.Assert(target.UnitID, qt.Equals, candidate.Unit.ID)
		c.Assert(target.Unit.Block().Text, qt.Equals, candidate.Unit.Text)
		c.Assert(target.Unit.Context(), qt.Equals, candidate.Unit.Context)
		c.Assert(target.Unit.Binding().Segments, qt.DeepEquals, candidate.Unit.Source.Segments)
		c.Assert(target.Unit.Binding().Kind, qt.Equals, candidate.Unit.Kind)
	}
	input.Files["d000001.txt"][0]++
	result, err := corpus.Prepare(t.Context(), candidates, input.Files)
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, corpus.Prepared{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = corpus.Prepare(ctx, candidates, input.Files)
	c.Assert(err, qt.ErrorIs, context.Canceled)
}
