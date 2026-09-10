package annotation_test

import (
	"context"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

func TestOriginsRemainIndependentOfEditorialResponses(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	data.Judgments, data.Adjudications = []annotation.Judgment{}, []annotation.Adjudication{}
	data.Units[0].Origin = annotation.Origin{Label: "human", Scope: "unit", Evidence: "Simulated provenance for this tutorial test."}
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	origins, err := round.Origins(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(origins.Version, qt.Equals, "unswell-origin-claims-v1")
	c.Assert(origins.Basis, qt.Equals, "simulation")
	c.Assert(origins.Units, qt.HasLen, len(data.Units))
	c.Assert(origins.Units[0].Origin, qt.Equals, data.Units[0].Origin)
	c.Assert(origins.Units[0].Target.SourceSHA256, qt.Equals, data.Units[0].Source.SHA256)
	c.Assert(origins.Units[0].Target.Segments, qt.DeepEquals, data.Units[0].Source.Segments)
	encoded, err := json.Marshal(origins)
	c.Assert(err, qt.IsNil)
	c.Assert(string(encoded), qt.Not(qt.Contains), data.Units[0].Text)
	decisions, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(decisions.Units[0].Reason, qt.Equals, "missing_judgments")
	c.Assert(decisions.Units[0].Label, qt.IsNil)
	c.Assert(origins.RoundSHA256, qt.Equals, decisions.RoundSHA256)
	origins.Units[0].Target.Segments[0].End++
	origins.Units[0].Target.AllowedUses[0] = "changed"
	again, err := round.Origins(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(again.Units[0].Target.Segments, qt.DeepEquals, data.Units[0].Source.Segments)
	c.Assert(again.SHA256, qt.Equals, origins.SHA256)
	c.Assert(again.Units[0].Target.AllowedUses, qt.Not(qt.Contains), "changed")
	data.Units[0].Origin.Evidence += " Revised curator assertion."
	changed, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	different, err := changed.Origins(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(different.SHA256, qt.Not(qt.Equals), again.SHA256)
}

func TestOriginsRejectMissingRoundAndCancellation(t *testing.T) {
	c := qt.New(t)
	var missing *annotation.Round
	result, err := missing.Origins(t.Context())
	c.Assert(err, qt.IsNotNil)
	c.Assert(result, qt.DeepEquals, annotation.OriginSet{})
	round, err := annotation.Load(t.Context(), encode(c, decisionFixture(c)))
	c.Assert(err, qt.IsNil)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	result, err = round.Origins(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	c.Assert(result, qt.DeepEquals, annotation.OriginSet{})
}
