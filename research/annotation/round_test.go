package annotation_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

type roundInput struct {
	Version       string                    `json:"version"`
	ID            string                    `json:"round_id"`
	Purpose       string                    `json:"purpose"`
	Rubric        string                    `json:"rubric"`
	Profile       annotation.Profile        `json:"profile"`
	Units         []annotation.Unit         `json:"units"`
	Actors        []annotation.Actor        `json:"actors"`
	Judgments     []annotation.Judgment     `json:"judgments"`
	Adjudications []annotation.Adjudication `json:"adjudications"`
}

func fixture(c *qt.C) roundInput {
	c.Helper()
	data, err := os.ReadFile("testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	_, err = annotation.Load(context.Background(), data)
	c.Assert(err, qt.IsNil)
	var input roundInput
	c.Assert(json.Unmarshal(data, &input), qt.IsNil)
	return input
}

func encode(c *qt.C, data roundInput) []byte {
	c.Helper()
	encoded, err := json.Marshal(data)
	c.Assert(err, qt.IsNil)
	return encoded
}

func TestPacketBlindingAndOwnership(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Units[0].Origin.Evidence = "HIDDEN_ORIGIN"
	data.Units[0].Source.Reference = "HIDDEN_SOURCE"
	data.Units[0].Rights.Evidence = "HIDDEN_RIGHTS"
	data.Judgments[0].Rationale = "HIDDEN_JUDGMENT"
	data.Adjudications[0].Rationale = "HIDDEN_DECISION"
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	packet, err := round.Packet(t.Context())
	c.Assert(err, qt.IsNil)
	output, err := json.Marshal(packet)
	c.Assert(err, qt.IsNil)
	c.Assert(bytes.Contains(output, []byte("HIDDEN_")), qt.IsFalse)
	c.Assert(packet.Units[0].Text, qt.Equals, data.Units[0].Text)
	c.Assert(packet.Units[0].Context, qt.Equals, data.Units[0].Context)
	packet.Units[0].Text = "changed"
	again, err := round.Packet(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(again.Units[0].Text, qt.Equals, data.Units[0].Text)
}

func TestAgreementIgnoresOriginAndAdjudication(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	before, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(before.Basis, qt.Equals, "simulation")
	c.Assert(before.PrimaryRaters, qt.Equals, 2)
	c.Assert(before.AuxiliaryJudgments, qt.Equals, 1)
	data.Adjudications = []annotation.Adjudication{}
	data.Judgments = data.Judgments[:len(data.Judgments)-1]
	data.Units[0].Origin = annotation.Origin{Label: "unknown", Scope: "repository", Evidence: "Unverified source-wide claim."}
	slices.Reverse(data.Units)
	slices.Reverse(data.Judgments)
	round, err = annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	after, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(after.Groups, qt.DeepEquals, before.Groups)
	c.Assert(after.AuxiliaryJudgments, qt.Equals, 0)
}

func TestRoundRejectsInvalidRecords(t *testing.T) {
	cases := []struct {
		name string
		edit func(*roundInput)
	}{
		{"version", func(r *roundInput) { r.Version = "unknown" }},
		{"empty", func(r *roundInput) { r.Units = nil }},
		{"duplicate unit", func(r *roundInput) { r.Units[1].ID = r.Units[0].ID }},
		{"profile hash", func(r *roundInput) { r.Profile.Instructions += "changed" }},
		{"source span", func(r *roundInput) { r.Units[0].Source.Segments[0].End++ }},
		{"source language", func(r *roundInput) { r.Units[0].Source.Language = "unknown" }},
		{"unknown role", func(r *roundInput) { r.Units[0].Role = "unknown" }},
		{"source origin", func(r *roundInput) { r.Units[0].Origin.Scope = "repository" }},
		{"generation record", func(r *roundInput) { r.Units[0].Origin.GenerationRecord = "" }},
		{"protected boundary", func(r *roundInput) { r.Units[0].Text += "\x00text" }},
		{"human tutorial", func(r *roundInput) { r.Actors[0].Kind = "human" }},
		{"simulated corpus", func(r *roundInput) { r.Purpose = "corpus" }},
		{"single primary", func(r *roundInput) { r.Actors[0].Kind = "assistant" }},
		{"duplicate actor", func(r *roundInput) { r.Actors[1].ID = r.Actors[0].ID }},
		{"unknown actor", func(r *roundInput) { r.Judgments[0].ActorID = "a999" }},
		{"unknown unit", func(r *roundInput) { r.Judgments[0].UnitID = "u999999" }},
		{"duplicate judgment", func(r *roundInput) { r.Judgments = append(r.Judgments, r.Judgments[0]) }},
		{"missing category", func(r *roundInput) { r.Judgments[0].Categories = nil }},
		{"incompatible category", func(r *roundInput) { r.Judgments[0].Label = "acceptable" }},
		{"unknown category", func(r *roundInput) { r.Judgments[0].Categories[0] = "origin" }},
		{"context", func(r *roundInput) { r.Judgments[0].Context = "insufficient" }},
		{"timestamp", func(r *roundInput) { r.Judgments[0].RecordedAt = "yesterday" }},
		{"early decision", func(r *roundInput) { r.Adjudications[0].RecordedAt = "2026-09-07T00:00:00Z" }},
		{"assistant decision", func(r *roundInput) { r.Adjudications[0].Reviewers = []string{"a003"} }},
		{"duplicate decision", func(r *roundInput) { r.Adjudications = append(r.Adjudications, r.Adjudications[0]) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			data := fixture(c)
			tc.edit(&data)
			_, err := annotation.Load(t.Context(), encode(c, data))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestLoadRejectsJSONAmbiguity(t *testing.T) {
	cases := []struct{ name, value string }{
		{"duplicate", `{"version":"unswell-annotation-v1","version":"unswell-annotation-v1"}`},
		{"trailing", `{} {}`},
		{"depth", strings.Repeat("[", 34) + strings.Repeat("]", 34)},
		{"utf8", "\xff"},
		{"unknown field", `{"anything":1}`},
		{"truncated", `{"version":`},
		{"surrogate", `{"text":"\ud800"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := annotation.Load(t.Context(), []byte(tc.value))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestMissingAnswersAndPermissions(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Judgments = []annotation.Judgment{}
	data.Adjudications = []annotation.Adjudication{}
	data.Units[0].Rights.AllowedUses = []string{}
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	_, err = round.Packet(t.Context())
	c.Assert(err, qt.ErrorMatches, ".*annotation permission.*")
	agreement, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(agreement.MissingAnswers, qt.Equals, len(data.Units)*2)
	c.Assert(agreement.Groups[0].Quality.Alpha.Value, qt.IsNil)
	c.Assert(agreement.Groups[0].Quality.Uncertain.Value, qt.IsNil)
}

func TestPacketBindingAndUnknownFields(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Units[0].Context += " New context."
	_, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.ErrorMatches, ".*frozen annotation packet.*")
	data = fixture(c)
	data.Adjudications[0].PacketSHA256 = strings.Repeat("0", 64)
	_, err = annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.ErrorMatches, ".*frozen annotation packet.*")
	encoded := encode(c, fixture(c))
	encoded = bytes.Replace(encoded, []byte(`"version":`), []byte(`"unknown":true,"version":`), 1)
	_, err = annotation.Load(t.Context(), encoded)
	c.Assert(err, qt.ErrorMatches, "(?s)annotation schema:.*additional properties.*")
}

func TestPlannedHumanRoundDoesNotInventRatings(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Purpose = "pilot"
	data.Judgments = []annotation.Judgment{}
	data.Adjudications = []annotation.Adjudication{}
	for i := range data.Actors {
		if data.Actors[i].Kind == "simulation" {
			data.Actors[i].Kind = "human"
		}
	}
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	result, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Basis, qt.Equals, "human")
	c.Assert(result.MissingAnswers, qt.Equals, 34)
	c.Assert(result.Groups[0].Quality.Ratings, qt.Equals, 0)
	c.Assert(result.Groups[0].Quality.RawAgreement.Value, qt.IsNil)
}

func TestUnloadedRound(t *testing.T) {
	cases := []struct {
		name  string
		round *annotation.Round
	}{{"nil", nil}, {"zero", &annotation.Round{}}}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, err := tc.round.Packet(t.Context())
			c.Assert(err, qt.IsNotNil)
			_, err = tc.round.Agreement(t.Context())
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestStringRolesAndFragments(t *testing.T) {
	cases := []struct{ role string }{{"error_message"}, {"log_message"}, {"ui_text"}, {"other_string"}, {"string"}}
	for _, tc := range cases {
		t.Run(tc.role, func(t *testing.T) {
			c := qt.New(t)
			data := fixture(c)
			data.Judgments = []annotation.Judgment{}
			data.Adjudications = []annotation.Adjudication{}
			data.Units[0].Kind = "fragment"
			data.Units[0].Role = tc.role
			round, err := annotation.Load(t.Context(), encode(c, data))
			c.Assert(err, qt.IsNil)
			packet, err := round.Packet(t.Context())
			c.Assert(err, qt.IsNil)
			c.Assert(packet.Units[0].Kind, qt.Equals, "fragment")
			c.Assert(packet.Units[0].Role, qt.Equals, tc.role)
			result, err := round.Agreement(t.Context())
			c.Assert(err, qt.IsNil)
			found := slices.ContainsFunc(result.Groups, func(group annotation.GroupStats) bool { return group.Group == "role:"+tc.role })
			c.Assert(found, qt.IsTrue)
		})
	}
}

func TestTutorialSourceReferences(t *testing.T) {
	c := qt.New(t)
	for _, unit := range fixture(c).Units {
		source, err := os.ReadFile(unit.Source.Reference)
		c.Assert(err, qt.IsNil)
		c.Assert(fmt.Sprintf("%x", sha256.Sum256(source)), qt.Equals, unit.Source.SHA256)
		c.Assert(source, qt.HasLen, unit.Source.Bytes)
		c.Assert(string(source), qt.Equals, unit.Text)
		c.Assert(unit.Source.Segments[0].Start, qt.Equals, 0)
		c.Assert(unit.Source.Segments[0].End, qt.Equals, len(source))
	}
}

func TestCancellationAndSize(t *testing.T) {
	c := qt.New(t)
	data := encode(c, fixture(c))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := annotation.Load(ctx, data)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	round, err := annotation.Load(t.Context(), data)
	c.Assert(err, qt.IsNil)
	_, err = round.Packet(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = round.Agreement(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = annotation.Load(t.Context(), bytes.Repeat([]byte(" "), annotation.MaxBytes+1))
	c.Assert(err, qt.IsNotNil)
}

func FuzzLoad(f *testing.F) {
	data, err := os.ReadFile("testdata/tutorial.json")
	if err != nil {
		f.Fatal(err)
	}
	f.Add(data)
	f.Add([]byte(`{"version":null}`))
	f.Fuzz(func(t *testing.T, input []byte) {
		round, err := annotation.Load(t.Context(), input)
		if err != nil {
			return
		}
		c := qt.New(t)
		result, err := round.Agreement(t.Context())
		c.Assert(err, qt.IsNil)
		_, err = json.Marshal(result)
		c.Assert(err, qt.IsNil)
	})
}
