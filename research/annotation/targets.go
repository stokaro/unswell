package annotation

import (
	"context"
	"fmt"
	"reflect"
	"slices"
)

// MatchTargets checks that every round unit occurs in an expected candidate set.
// IDs, prose, context, source metadata, extraction, and rights must agree. Origin
// is independently curated and is not part of an editorial target binding.
// Callers must reproduce expected candidates from sources before trusting them.
// This check neither establishes permissions nor qualifies human participation.
func (r *Round) MatchTargets(ctx context.Context, expected []Unit) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r == nil || r.inputSHA256 == "" || r.data.packetSHA256 == "" {
		return fmt.Errorf("load a validated annotation round before matching targets")
	}
	byID, err := indexTargets(ctx, expected)
	if err != nil {
		return err
	}
	for _, unit := range r.data.Units {
		if err := ctx.Err(); err != nil {
			return err
		}
		index, exists := byID[unit.ID]
		if !exists || !sameTarget(unit, expected[index]) {
			return fmt.Errorf("round unit %s does not match the expected target and provenance", unit.ID)
		}
	}
	return ctx.Err()
}

func indexTargets(ctx context.Context, expected []Unit) (map[string]int, error) {
	if len(expected) == 0 || len(expected) > 10000 {
		return nil, fmt.Errorf("expected targets must contain 1 to 10000 units")
	}
	byID := make(map[string]int, len(expected))
	for i, unit := range expected {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if _, exists := byID[unit.ID]; unit.ID == "" || exists {
			return nil, fmt.Errorf("expected target IDs must be nonempty and unique")
		}
		byID[unit.ID] = i
	}
	return byID, nil
}

func sameTarget(a, b Unit) bool {
	return a.Kind == b.Kind && a.Role == b.Role && a.Text == b.Text && a.Context == b.Context &&
		reflect.DeepEqual(a.Source, b.Source) && a.Extraction == b.Extraction &&
		a.Rights.License == b.Rights.License && a.Rights.Evidence == b.Rights.Evidence &&
		slices.Equal(sortedUses(a.Rights.AllowedUses), sortedUses(b.Rights.AllowedUses))
}

func sortedUses(values []string) []string {
	result := slices.Clone(values)
	slices.Sort(result)
	return result
}
