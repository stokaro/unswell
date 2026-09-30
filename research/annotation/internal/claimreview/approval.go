package claimreview

import (
	"fmt"
	"slices"
)

func validateApprovals(stage Stage, approvals []Approval, decisions map[string]Decision) (map[int]bool, error) {
	for _, approval := range approvals {
		if err := validApproval(stage.ID, approval, decisions); err != nil {
			return nil, err
		}
	}
	used := make(map[int]bool)
	for _, decision := range stage.Decisions {
		if decision.Status != "resolved" {
			continue
		}
		edit := stage.Edits[slices.IndexFunc(stage.Edits, func(edit Edit) bool { return edit.ID == decision.EditID })]
		index := slices.IndexFunc(approvals, func(approval Approval) bool {
			return approval.Kind == "resolution" && slices.Equal(approval.Claims, []string{decision.ClaimID}) &&
				approval.EditDigest == EditDigest(edit) && approval.MeaningPreserved
		})
		if index < 0 {
			return nil, fmt.Errorf("resolution lacks separate claim-specific review of the exact edit")
		}
		used[index] = true
	}
	return used, nil
}

func validApproval(stageID string, approval Approval, decisions map[string]Decision) error {
	if approval.StageID != stageID || !validText(approval.Reviewer) || !validText(approval.Reason) ||
		len(approval.Claims) == 0 || len(approval.Claims) > 64 {
		return fmt.Errorf("invalid review identity, claim set, or rationale")
	}
	seen := make(map[string]bool)
	for _, id := range approval.Claims {
		if decisions[id].ClaimID == "" || seen[id] {
			return fmt.Errorf("review names an unknown or repeated claim")
		}
		seen[id] = true
	}
	return approvalKind(approval)
}

func approvalKind(approval Approval) error {
	switch approval.Kind {
	case "resolution":
		if len(approval.Claims) != 1 || !approval.MeaningPreserved || approval.EditDigest == "" {
			return fmt.Errorf("resolution review needs one claim, preserved meaning, and an edit digest")
		}
	case "same_defect":
		if len(approval.Claims) < 2 || approval.EditDigest != "" || approval.MeaningPreserved {
			return fmt.Errorf("same-defect review must name at least two claims without an edit verdict")
		}
	default:
		return fmt.Errorf("unsupported review approval kind")
	}
	return nil
}

func (in Inventory) displayed(stage Stage, approvals []Approval, decisions map[string]Decision, used map[int]bool) ([]string, error) {
	hidden := make(map[string]bool)
	membership := make(map[string]bool)
	for _, group := range stage.Duplicates {
		if err := validateDuplicate(group, decisions, membership); err != nil {
			return nil, err
		}
		index := slices.IndexFunc(approvals, func(approval Approval) bool {
			return approval.Kind == "same_defect" && sameMembers(approval.Claims, group.Claims)
		})
		if index < 0 || used[index] {
			return nil, fmt.Errorf("deduplication needs a distinct explicit same-defect review")
		}
		used[index] = true
		for _, id := range group.Claims {
			hidden[id] = id != group.Representative
		}
	}
	displayed := make([]string, 0)
	for _, claim := range in.claims {
		if decisions[claim.ID].Status == "retained" && !hidden[claim.ID] {
			displayed = append(displayed, claim.ID)
		}
	}
	return displayed, nil
}

func validateDuplicate(group Duplicate, decisions map[string]Decision, membership map[string]bool) error {
	if len(group.Claims) < 2 || len(group.Claims) > 64 || !slices.Contains(group.Claims, group.Representative) ||
		!validText(group.Reason) {
		return fmt.Errorf("invalid same-defect group or representative")
	}
	for _, id := range group.Claims {
		if decisions[id].Status != "retained" || membership[id] {
			return fmt.Errorf("only distinct retained claims may share one duplicate group")
		}
		membership[id] = true
	}
	return nil
}

func sameMembers(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}
