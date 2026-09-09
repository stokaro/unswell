package annotation

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestDecisionsBindInputWithoutExposingProse(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	data.Units[0].Source.Reference = "HIDDEN_SOURCE"
	data.Units[0].Source.AuthorGroup = "HIDDEN_AUTHOR"
	data.Units[0].Origin.Evidence = "HIDDEN_ORIGIN"
	data.Judgments[0].Rationale = "HIDDEN_RATIONALE"
	input := encode(c, data)
	round, err := Load(t.Context(), input)
	c.Assert(err, qt.IsNil)
	result, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.RoundSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256(input)))
	c.Assert(result.PacketSHA256, qt.Equals, data.Judgments[0].PacketSHA256)
	c.Assert(result.Units[0].Target.SourceSHA256, qt.Equals, data.Units[0].Source.SHA256)
	c.Assert(result.Units[0].Target.Segments, qt.DeepEquals, data.Units[0].Source.Segments)
	c.Assert(result.Units[0].Target.TextSHA256, qt.Equals, fmt.Sprintf("%x", sha256.Sum256([]byte(data.Units[0].Text))))
	output, err := json.Marshal(result)
	c.Assert(err, qt.IsNil)
	c.Assert(bytes.Contains(output, []byte("HIDDEN_")), qt.IsFalse)
	c.Assert(bytes.Contains(output, []byte(data.Units[0].Text)), qt.IsFalse)
	digest := result.SHA256
	result.SHA256 = ""
	unsigned, err := json.Marshal(result)
	c.Assert(err, qt.IsNil)
	c.Assert(digest, qt.Equals, fmt.Sprintf("%x", sha256.Sum256(unsigned)))
}

func TestDecisionOrderingAndRightsAreExplicit(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	data.Units[0].Rights.AllowedUses = []string{"annotation"}
	for i := range data.Judgments {
		data.Judgments[i].Categories = []string{"wordiness", "empty_framing"}
	}
	before, err := Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	want, err := before.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(want.Units[0].Target.AllowedUses, qt.DeepEquals, []string{"annotation"})
	slices.Reverse(data.Units)
	slices.Reverse(data.Actors)
	slices.Reverse(data.Judgments)
	slices.Reverse(data.Judgments[0].Categories)
	after, err := Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	got, err := after.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(got.RoundSHA256, qt.Not(qt.Equals), want.RoundSHA256)
	c.Assert(got.Units, qt.DeepEquals, want.Units)
	c.Assert(got.PrimaryRaters, qt.DeepEquals, want.PrimaryRaters)
	c.Assert(got.PacketSHA256, qt.Equals, want.PacketSHA256)
}

func TestDecisionOwnershipAndCancellation(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	round, err := Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	want, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(want.Units[0].Label, qt.IsNotNil)
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			got, err := round.Decisions(t.Context())
			if !c.Check(err, qt.IsNil) {
				return
			}
			c.Check(got, qt.DeepEquals, want)
			got.PrimaryRaters[0] = "changed"
			*got.Units[0].Label = "changed"
			got.Units[0].Categories[0] = "changed"
			got.Units[0].Target.AllowedUses[0] = "changed"
			got.Units[0].Target.Segments[0].End++
		})
	}
	wg.Wait()
	again, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(again, qt.DeepEquals, want)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	empty, err := round.Decisions(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(empty, qt.DeepEquals, DecisionSet{})
	var invalid *Round
	_, err = invalid.Decisions(t.Context())
	c.Assert(err, qt.ErrorMatches, "load a validated.*")
	_, err = (&Round{}).Decisions(t.Context())
	c.Assert(err, qt.ErrorMatches, "load a validated.*")
}
