package claimreview_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/research/annotation/internal/claimreview"
)

func reviewFixture(t *testing.T) (claimreview.Inventory, claimreview.Stage, []claimreview.Approval) {
	c := qt.New(t)
	t.Helper()
	doc := parsed(t, "Four settings decides where requests go.", document.Plain)
	target := reference(t, doc, string(doc.Source), 1)
	inventory, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{
		specification(target, "grammar"), specification(target, "rhetoric"),
	})
	c.Assert(err, qt.IsNil)
	claims := inventory.Claims()
	edit := claimreview.Edit{ID: "edit-grammar", Target: reference(t, doc, "decides", 1), Replacement: "decide"}
	stage := claimreview.Stage{ID: "rewrite", Edits: []claimreview.Edit{edit}, Decisions: []claimreview.Decision{
		{ClaimID: claims[0].ID, Status: "resolved", Reason: "Verb agreement is corrected.", EditID: edit.ID},
		{ClaimID: claims[1].ID, Status: "retained", Reason: "The numerical framing is unchanged."},
	}}
	approvals := []claimreview.Approval{{StageID: stage.ID, Kind: "resolution", Claims: []string{claims[0].ID},
		EditDigest: claimreview.EditDigest(edit), Reviewer: "assistant-constructed-control",
		Reason: "Only grammar is changed; the independently declared rhetorical claim remains.", MeaningPreserved: true}}
	return inventory, stage, approvals
}

func TestGrammarEditCannotEraseRhetoricalClaim(t *testing.T) {
	c := qt.New(t)
	inventory, stage, approvals := reviewFixture(t)
	out, err := inventory.Review(context.Background(), stage, approvals)
	c.Assert(err, qt.IsNil)
	c.Assert(out.Complete, qt.IsTrue)
	c.Assert(out.EditorialQualified, qt.IsFalse)
	c.Assert(out.Claims, qt.DeepEquals, inventory.Claims())
	c.Assert(out.DisplayedClaims, qt.DeepEquals, []string{inventory.Claims()[1].ID})
	stage.Decisions[1].Status, stage.Decisions[1].EditID = "resolved", stage.Edits[0].ID
	_, err = inventory.Review(context.Background(), stage, approvals)
	c.Assert(err, qt.ErrorMatches, "resolution lacks separate claim-specific review of the exact edit")
}

func TestStageMustAccountForEveryOriginalClaim(t *testing.T) {
	tests := []struct {
		name string
		edit func(*claimreview.Stage, *[]claimreview.Approval)
	}{
		{"missing claim", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Decisions = s.Decisions[:1] }},
		{"invented claim", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Decisions[1].ClaimID = "invented" }},
		{"repeated claim", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Decisions[1] = s.Decisions[0] }},
		{"unsupported resolution", func(_ *claimreview.Stage, a *[]claimreview.Approval) { *a = nil }},
		{"changed edit", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Edits[0].Replacement = "never decides" }},
		{"unsafe edit verdict", func(_ *claimreview.Stage, a *[]claimreview.Approval) { (*a)[0].MeaningPreserved = false }},
		{"unexplained rejection", func(s *claimreview.Stage, _ *[]claimreview.Approval) {
			s.Decisions[1].Status, s.Decisions[1].Reason = "rejected", ""
		}},
		{"invented status", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Decisions[1].Status = "acceptable" }},
		{"hidden edit", func(s *claimreview.Stage, _ *[]claimreview.Approval) { s.Decisions[1].EditID = s.Edits[0].ID }},
		{"approval wrong stage", func(_ *claimreview.Stage, a *[]claimreview.Approval) { (*a)[0].StageID = "other-stage" }},
		{"unused approval", func(_ *claimreview.Stage, a *[]claimreview.Approval) { *a = append(*a, (*a)[0]) }},
		{"silent overlap merge", func(s *claimreview.Stage, _ *[]claimreview.Approval) {
			s.Decisions[0].Status, s.Decisions[0].EditID, s.Edits = "retained", "", nil
			s.Duplicates = []claimreview.Duplicate{{Claims: []string{s.Decisions[0].ClaimID, s.Decisions[1].ClaimID},
				Representative: s.Decisions[0].ClaimID, Reason: "The quoted paragraph overlaps."}}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			inventory, stage, approvals := reviewFixture(t)
			test.edit(&stage, &approvals)
			_, err := inventory.Review(context.Background(), stage, approvals)
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestSupportOnlyContextCannotBecomeEditTarget(t *testing.T) {
	c := qt.New(t)
	doc := parsed(t, "The API retries on failure.\n\nThe API retries on failure.", document.Plain)
	spec := specification(reference(t, doc, "The API retries on failure.", 2), "repetition")
	spec.Support = []claimreview.Reference{reference(t, doc, "The API retries on failure.", 1)}
	inventory, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{spec})
	c.Assert(err, qt.IsNil)
	edit := claimreview.Edit{ID: "edit-context", Target: spec.Support[0], Replacement: ""}
	stage := claimreview.Stage{ID: "rewrite", Edits: []claimreview.Edit{edit}, Decisions: []claimreview.Decision{
		{ClaimID: inventory.Claims()[0].ID, Status: "resolved", Reason: "Removed the retained antecedent.", EditID: edit.ID},
	}}
	_, err = inventory.Review(context.Background(), stage, nil)
	c.Assert(err, qt.ErrorMatches, "resolution edit is not within this claim's original targets")
}

func TestUncertainAccountingDoesNotBecomeComplete(t *testing.T) {
	c := qt.New(t)
	inventory, stage, approvals := reviewFixture(t)
	stage.Decisions[1].Status = "uncertain"
	out, err := inventory.Review(context.Background(), stage, approvals)
	c.Assert(err, qt.IsNil)
	c.Assert(out.Complete, qt.IsFalse)
	c.Assert(out.EditorialQualified, qt.IsFalse)
	c.Assert(out.Claims, qt.HasLen, 2)
	c.Assert(out.Decisions, qt.HasLen, 2)
}

func TestReviewedDuplicatesKeepEveryOriginalClaim(t *testing.T) {
	c := qt.New(t)
	doc := parsed(t, "The introductory sentence repeats the section heading.", document.Plain)
	target := reference(t, doc, string(doc.Source), 1)
	first, second := specification(target, "heading-echo"), specification(target, "heading-echo")
	second.Origin.Candidate = "candidate-2"
	inventory, err := claimreview.Bind(context.Background(), doc, []claimreview.Specification{first, second})
	c.Assert(err, qt.IsNil)
	stage := claimreview.Stage{ID: "selection", Decisions: []claimreview.Decision{
		{ClaimID: inventory.Claims()[0].ID, Status: "retained", Reason: "Retain the reviewed criticism."},
		{ClaimID: inventory.Claims()[1].ID, Status: "retained", Reason: "Retain its independent provenance."},
	}}
	ids := []string{stage.Decisions[0].ClaimID, stage.Decisions[1].ClaimID}
	stage.Duplicates = []claimreview.Duplicate{{Claims: ids, Representative: ids[1], Reason: "Declared same-defect control."}}
	approvals := []claimreview.Approval{{StageID: stage.ID, Kind: "same_defect", Claims: ids,
		Reviewer: "constructed-control", Reason: "This test supplies an explicit relationship, not a semantic verdict."}}
	out, err := inventory.Review(context.Background(), stage, approvals)
	c.Assert(err, qt.IsNil)
	c.Assert(out.DisplayedClaims, qt.DeepEquals, []string{ids[1]})
	c.Assert(out.Claims, qt.DeepEquals, inventory.Claims())
	expected := slices.Clone(ids)
	stage.Duplicates[0].Claims[0] = "changed"
	approvals[0].Claims[1] = "changed"
	c.Assert(out.Duplicates[0].Claims, qt.DeepEquals, expected)
}
