package annotation_test

import (
	"slices"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation"
)

func TestDecisionSelection(t *testing.T) {
	for _, row := range []struct {
		name, status, reason, basis, label string
		edit                               func(*roundInput)
	}{
		{"unanimous", "resolved", "", "unanimous", "needs_revision", func(*roundInput) {}},
		{"missing", "unresolved", "missing_judgments", "none", "", func(r *roundInput) { r.Judgments = r.Judgments[:1] }},
		{"disagreement", "unresolved", "judgment_disagreement", "none", "", func(r *roundInput) {
			r.Judgments[1].Label, r.Judgments[1].Categories = "acceptable", []string{}
		}},
		{"category disagreement", "unresolved", "judgment_disagreement", "none", "", func(r *roundInput) {
			r.Judgments[1].Categories = []string{"wordiness"}
		}},
		{"uncertain", "unresolved", "uncertain", "unanimous", "uncertain", func(r *roundInput) {
			for i := range r.Judgments {
				r.Judgments[i].Label, r.Judgments[i].Context = "uncertain", "insufficient"
			}
		}},
		{"adjudicated", "resolved", "", "adjudication", "acceptable", func(r *roundInput) {
			r.Adjudications = []annotation.Adjudication{decisionAdjudication(r, "acceptable")}
		}},
		{"uncertain adjudication", "unresolved", "uncertain", "adjudication", "uncertain", func(r *roundInput) {
			r.Adjudications = []annotation.Adjudication{decisionAdjudication(r, "uncertain")}
		}},
		{"unfinished adjudicated assignment", "unresolved", "missing_judgments", "none", "", func(r *roundInput) {
			r.Actors = append(r.Actors, annotation.Actor{ID: "a005", Kind: "simulation", Role: "rater", Declaration: "A missing test response."})
			r.Adjudications = []annotation.Adjudication{decisionAdjudication(r, "acceptable")}
		}},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			data := decisionFixture(c)
			row.edit(&data)
			round, err := annotation.Load(t.Context(), encode(c, data))
			c.Assert(err, qt.IsNil)
			result, err := round.Decisions(t.Context())
			c.Assert(err, qt.IsNil)
			decision := result.Units[0]
			c.Assert(decision.Status, qt.Equals, row.status)
			c.Assert(decision.Reason, qt.Equals, row.reason)
			c.Assert(decision.Basis, qt.Equals, row.basis)
			if row.label == "" {
				c.Assert(decision.Label, qt.IsNil)
				c.Assert(decision.Categories, qt.HasLen, 0)
			} else {
				c.Assert(decision.Label, qt.IsNotNil)
				c.Assert(*decision.Label, qt.Equals, row.label)
			}
			c.Assert(result.Basis, qt.Equals, "simulation")
			c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
		})
	}
}

func decisionFixture(c *qt.C) roundInput {
	c.Helper()
	data := fixture(c)
	data.Adjudications = []annotation.Adjudication{}
	data.Judgments = slices.DeleteFunc(data.Judgments, func(j annotation.Judgment) bool {
		return j.UnitID != "u000001" || j.ActorID == "a003"
	})
	slices.SortFunc(data.Judgments, func(a, b annotation.Judgment) int {
		return strings.Compare(a.ActorID, b.ActorID)
	})
	for i := range data.Judgments {
		data.Judgments[i].Label = "needs_revision"
		data.Judgments[i].Categories = []string{"empty_framing"}
		data.Judgments[i].Context = "adequate"
	}
	return data
}

func decisionAdjudication(r *roundInput, label string) annotation.Adjudication {
	return annotation.Adjudication{PacketSHA256: r.Judgments[0].PacketSHA256, UnitID: "u000001", Reviewers: []string{"a001", "a002"},
		Label: label, Categories: []string{}, Rationale: "Scripted final selection for this test.", RecordedAt: "2026-09-09T12:00:00Z"}
}

func TestAuxiliaryResponsesDoNotSupplyPrimaryDecisions(t *testing.T) {
	c := qt.New(t)
	data := decisionFixture(c)
	data.Judgments = data.Judgments[:1]
	auxiliary := data.Judgments[0]
	auxiliary.ActorID = "a003"
	data.Judgments = append(data.Judgments, auxiliary)
	round, err := annotation.Load(t.Context(), encode(c, data))
	c.Assert(err, qt.IsNil)
	result, err := round.Decisions(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Units[0].PrimaryJudgments, qt.Equals, 1)
	c.Assert(result.Units[0].AuxiliaryJudgments, qt.Equals, 1)
	c.Assert(result.Units[0].MissingRaters, qt.DeepEquals, []string{"a002"})
	c.Assert(result.Units[0].Label, qt.IsNil)
	c.Assert(result.Units[0].Reason, qt.Equals, "missing_judgments")
}
