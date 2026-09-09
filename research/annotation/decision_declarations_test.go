package annotation_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

func TestParticipationDeclarationsNeverQualifyCorpus(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	// These scripted declarations test the boundary; no human responses were collected.
	data.Purpose = "pilot"
	for i := range data.Actors {
		if data.Actors[i].Kind == "simulation" {
			data.Actors[i].Kind = "human"
		}
	}
	unrated := data
	unrated.Judgments = []annotation.Judgment{}
	unrated.Adjudications = []annotation.Adjudication{}
	pending, err := annotation.Load(t.Context(), encode(c, unrated))
	c.Assert(err, qt.IsNil)
	packet, err := pending.Packet(t.Context())
	c.Assert(err, qt.IsNil)
	for i := range data.Judgments {
		data.Judgments[i].PacketSHA256 = packet.SHA256
	}
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	result, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Basis, qt.Equals, "declared_human")
	c.Assert(result.Purpose, qt.Equals, "pilot")
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(result.Units[0].Status, qt.Equals, "resolved")
}
