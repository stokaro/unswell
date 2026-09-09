package corpus

import (
	"context"
	"slices"
	"strings"
)

func (b *compressionBankBuilder) reserve(ctx context.Context, artifact Artifact) error {
	for _, group := range artifact.Plan.Groups {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !b.reserved[group.ID] {
			continue
		}
		if err := b.charge(group); err != nil {
			return err
		}
		b.result.ReservedGroups = append(b.result.ReservedGroups, group)
	}
	for _, target := range artifact.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := b.reserveTarget(target); err != nil {
			return err
		}
	}
	slices.SortFunc(b.result.ReservedTargets, func(a, b CompressionReservation) int { return strings.Compare(a.UnitID, b.UnitID) })
	return nil
}

func (b *compressionBankBuilder) reserveTarget(target Candidate) error {
	if b.reserved[target.GroupID] {
		reservation := CompressionReservation{UnitID: target.Unit.ID, SourceID: target.SourceID, GroupID: target.GroupID}
		if err := b.charge(reservation); err != nil {
			return err
		}
		b.result.ReservedTargets = append(b.result.ReservedTargets, reservation)
	} else if target.Partition == "training" && target.Unit.Kind == b.result.Options.Kind {
		b.result.RemainingTrainingUnits++
	}
	return nil
}
