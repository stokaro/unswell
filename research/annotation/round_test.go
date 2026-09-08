package annotation

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
)

func fixture(c *qt.C) roundData {
	c.Helper()
	data, err := os.ReadFile("testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	round, err := Load(context.Background(), data)
	c.Assert(err, qt.IsNil)
	return round.data
}

func encode(c *qt.C, data roundData) []byte {
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
	round, err := Load(t.Context(), encode(c, data))
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
	round, err := Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	before, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(before.Basis, qt.Equals, "simulation")
	c.Assert(before.PrimaryRaters, qt.Equals, 2)
	c.Assert(before.AuxiliaryJudgments, qt.Equals, 1)
	data.Adjudications = []Adjudication{}
	data.Judgments = data.Judgments[:len(data.Judgments)-1]
	data.Units[0].Origin = Origin{Label: "unknown", Scope: "repository", Evidence: "Unverified source-wide claim."}
	slices.Reverse(data.Units)
	slices.Reverse(data.Judgments)
	round, err = Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	after, err := round.Agreement(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(after.Groups, qt.DeepEquals, before.Groups)
	c.Assert(after.AuxiliaryJudgments, qt.Equals, 0)
}

func TestRoundRejectsInvalidRecords(t *testing.T) {
	cases := []struct {
		name string
		edit func(*roundData)
	}{
		{"version", func(r *roundData) { r.Version = "unknown" }},
		{"empty", func(r *roundData) { r.Units = nil }},
		{"duplicate unit", func(r *roundData) { r.Units[1].ID = r.Units[0].ID }},
		{"profile hash", func(r *roundData) { r.Profile.Instructions += "changed" }},
		{"source span", func(r *roundData) { r.Units[0].Source.Segments[0].End++ }},
		{"source language", func(r *roundData) { r.Units[0].Source.Language = "unknown" }},
		{"unknown role", func(r *roundData) { r.Units[0].Role = "unknown" }},
		{"source origin", func(r *roundData) { r.Units[0].Origin.Scope = "repository" }},
		{"generation record", func(r *roundData) { r.Units[0].Origin.GenerationRecord = "" }},
		{"protected boundary", func(r *roundData) { r.Units[0].Text += "\x00text" }},
		{"human tutorial", func(r *roundData) { r.Actors[0].Kind = "human" }},
		{"simulated corpus", func(r *roundData) { r.Purpose = "corpus" }},
		{"single primary", func(r *roundData) { r.Actors[0].Kind = "assistant" }},
		{"duplicate actor", func(r *roundData) { r.Actors[1].ID = r.Actors[0].ID }},
		{"unknown actor", func(r *roundData) { r.Judgments[0].ActorID = "a999" }},
		{"unknown unit", func(r *roundData) { r.Judgments[0].UnitID = "u999999" }},
		{"duplicate judgment", func(r *roundData) { r.Judgments = append(r.Judgments, r.Judgments[0]) }},
		{"missing category", func(r *roundData) { r.Judgments[0].Categories = nil }},
		{"incompatible category", func(r *roundData) { r.Judgments[0].Label = "acceptable" }},
		{"unknown category", func(r *roundData) { r.Judgments[0].Categories[0] = "origin" }},
		{"context", func(r *roundData) { r.Judgments[0].Context = "insufficient" }},
		{"timestamp", func(r *roundData) { r.Judgments[0].RecordedAt = "yesterday" }},
		{"early decision", func(r *roundData) { r.Adjudications[0].RecordedAt = "2026-09-07T00:00:00Z" }},
		{"assistant decision", func(r *roundData) { r.Adjudications[0].Reviewers = []string{"a003"} }},
		{"duplicate decision", func(r *roundData) { r.Adjudications = append(r.Adjudications, r.Adjudications[0]) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			data := fixture(c)
			tc.edit(&data)
			_, err := Load(t.Context(), encode(c, data))
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
			_, err := Load(t.Context(), []byte(tc.value))
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestUnicodeEscapes(t *testing.T) {
	cases := []struct {
		name, value string
		valid       bool
	}{
		{"pair", `"\ud83d\ude00"`, true},
		{"literal slash", `"\\ud800"`, true},
		{"ordinary", `"\u0061"`, true},
		{"low alone", `"\udc00"`, false},
		{"high then ordinary", `"\ud800\u0061"`, false},
		{"high at end", `"\ud800"`, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(validSurrogates([]byte(tc.value)) == nil, qt.Equals, tc.valid)
		})
	}
}

func TestMissingAnswersAndPermissions(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Judgments = []Judgment{}
	data.Adjudications = []Adjudication{}
	data.Units[0].Rights.AllowedUses = []string{}
	round, err := Load(t.Context(), encode(c, data))
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
	_, err := Load(t.Context(), encode(c, data))
	c.Assert(err, qt.ErrorMatches, ".*frozen annotation packet.*")
	data = fixture(c)
	data.Adjudications[0].PacketSHA256 = strings.Repeat("0", 64)
	_, err = Load(t.Context(), encode(c, data))
	c.Assert(err, qt.ErrorMatches, ".*frozen annotation packet.*")
	encoded := encode(c, fixture(c))
	encoded = bytes.Replace(encoded, []byte(`"version":`), []byte(`"unknown":true,"version":`), 1)
	_, err = Load(t.Context(), encoded)
	c.Assert(err, qt.ErrorMatches, "(?s)annotation schema:.*additional properties.*")
}

func TestPlannedHumanRoundDoesNotInventRatings(t *testing.T) {
	c := qt.New(t)
	data := fixture(c)
	data.Purpose = "pilot"
	data.Judgments = []Judgment{}
	data.Adjudications = []Adjudication{}
	for i := range data.Actors {
		if data.Actors[i].Kind == "simulation" {
			data.Actors[i].Kind = "human"
		}
	}
	round, err := Load(t.Context(), encode(c, data))
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
		round *Round
	}{{"nil", nil}, {"zero", &Round{}}}
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
			data.Judgments = []Judgment{}
			data.Adjudications = []Adjudication{}
			data.Units[0].Kind = "fragment"
			data.Units[0].Role = tc.role
			round, err := Load(t.Context(), encode(c, data))
			c.Assert(err, qt.IsNil)
			packet, err := round.Packet(t.Context())
			c.Assert(err, qt.IsNil)
			c.Assert(packet.Units[0].Kind, qt.Equals, "fragment")
			c.Assert(packet.Units[0].Role, qt.Equals, tc.role)
			result, err := round.Agreement(t.Context())
			c.Assert(err, qt.IsNil)
			found := slices.ContainsFunc(result.Groups, func(group GroupStats) bool { return group.Group == "role:"+tc.role })
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
	_, err := Load(ctx, data)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	round, err := Load(t.Context(), data)
	c.Assert(err, qt.IsNil)
	_, err = round.Packet(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = round.Agreement(ctx)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	_, err = Load(t.Context(), bytes.Repeat([]byte(" "), MaxBytes+1))
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
		round, err := Load(t.Context(), input)
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
