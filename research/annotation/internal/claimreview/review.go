package claimreview

import (
	"context"
	"fmt"
	"slices"
	"unicode/utf8"
)

// Review requires an explicit disposition for every original claim in a stage.
// Approvals are separately supplied by the trusted calling application, not read
// from Stage. Structural validity never establishes that a diagnosis is correct.
func (in Inventory) Review(ctx context.Context, stage Stage, approvals []Approval) (Account, error) {
	if err := ctx.Err(); err != nil {
		return Account{}, err
	}
	if err := in.stageLimits(stage, approvals); err != nil {
		return Account{}, err
	}
	decisions, err := in.decisions(ctx, stage)
	if err != nil {
		return Account{}, err
	}
	if err := in.validateEdits(stage, decisions); err != nil {
		return Account{}, err
	}
	used, err := validateApprovals(stage, approvals, decisions)
	if err != nil {
		return Account{}, err
	}
	displayed, err := in.displayed(stage, approvals, decisions, used)
	if err != nil {
		return Account{}, err
	}
	if len(used) != len(approvals) {
		return Account{}, fmt.Errorf("unused or repeated review approval")
	}
	out := Account{Version: Version, SourceHash: in.document.Hash, StageID: stage.ID,
		Claims: in.Claims(), Decisions: slices.Clone(stage.Decisions), Edits: slices.Clone(stage.Edits),
		DisplayedClaims: displayed, Duplicates: cloneDuplicates(stage.Duplicates), Approvals: cloneApprovals(approvals), Complete: true}
	out.Complete = !slices.ContainsFunc(stage.Decisions, func(decision Decision) bool { return decision.Status == "uncertain" })
	return out, ctx.Err()
}

func (in Inventory) stageLimits(stage Stage, approvals []Approval) error {
	if in.document.Hash == "" || !validText(stage.ID) || len(stage.Edits) > maximumClaims ||
		len(stage.Duplicates) > maximumClaims || len(approvals) > maximumClaims {
		return fmt.Errorf("unbound inventory or invalid stage limits")
	}
	return nil
}

func (in Inventory) decisions(ctx context.Context, stage Stage) (map[string]Decision, error) {
	if len(stage.Decisions) != len(in.claims) {
		return nil, fmt.Errorf("every original claim requires exactly one decision")
	}
	known := make(map[string]bool, len(in.claims))
	for _, claim := range in.claims {
		known[claim.ID] = true
	}
	decisions := make(map[string]Decision, len(in.claims))
	for _, decision := range stage.Decisions {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if !known[decision.ClaimID] || decisions[decision.ClaimID].ClaimID != "" || !validText(decision.Reason) {
			return nil, fmt.Errorf("unknown, repeated, or unexplained claim decision")
		}
		switch decision.Status {
		case "retained", "resolved", "rejected", "uncertain":
		default:
			return nil, fmt.Errorf("unsupported claim disposition %q", decision.Status)
		}
		if (decision.Status == "resolved") != (decision.EditID != "") {
			return nil, fmt.Errorf("only a resolved claim must cite a concrete edit")
		}
		decisions[decision.ClaimID] = decision
	}
	return decisions, nil
}

func (in Inventory) validateEdits(stage Stage, decisions map[string]Decision) error {
	edits := make(map[string]Edit, len(stage.Edits))
	for _, edit := range stage.Edits {
		if err := in.validEdit(edit, edits); err != nil {
			return err
		}
		edits[edit.ID] = edit
	}
	used := make(map[string]bool)
	for _, claim := range in.claims {
		decision := decisions[claim.ID]
		if decision.Status != "resolved" {
			continue
		}
		edit, ok := edits[decision.EditID]
		if !ok || !slices.ContainsFunc(claim.Targets, func(ref Reference) bool {
			return ref.Block == edit.Target.Block && contains(ref.Span, edit.Target.Span)
		}) {
			return fmt.Errorf("resolution edit is not within this claim's original targets")
		}
		used[edit.ID] = true
	}
	if len(used) != len(edits) {
		return fmt.Errorf("an edit has no resolved original claim")
	}
	return disjointEdits(stage.Edits)
}

func (in Inventory) validEdit(edit Edit, known map[string]Edit) error {
	if !validText(edit.ID) || known[edit.ID].ID != "" || len(edit.Replacement) > 16000 ||
		edit.Replacement == edit.Target.Quote || !utf8.ValidString(edit.Replacement) {
		return fmt.Errorf("invalid, duplicate, unchanged, or excessive edit")
	}
	return validateReference(in.document, edit.Target)
}

func disjointEdits(edits []Edit) error {
	ordered := slices.Clone(edits)
	slices.SortFunc(ordered, func(a, b Edit) int { return a.Target.Span.Start - b.Target.Span.Start })
	for i := 1; i < len(ordered); i++ {
		if overlaps(ordered[i-1].Target.Span, ordered[i].Target.Span) {
			return fmt.Errorf("simultaneous original-source edits overlap")
		}
	}
	return nil
}

func cloneApprovals(approvals []Approval) []Approval {
	result := slices.Clone(approvals)
	for i := range result {
		result[i].Claims = slices.Clone(result[i].Claims)
	}
	return result
}

func cloneDuplicates(duplicates []Duplicate) []Duplicate {
	result := slices.Clone(duplicates)
	for i := range result {
		result[i].Claims = slices.Clone(result[i].Claims)
	}
	return result
}
