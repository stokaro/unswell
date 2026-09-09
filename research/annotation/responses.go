package annotation

import "context"

type primaryResponses struct {
	raters    map[string]bool
	byUnit    map[string][]Judgment
	auxiliary map[string]int
	auxTotal  int
}

func (r roundData) responses(ctx context.Context) (primaryResponses, error) {
	result := primaryResponses{raters: make(map[string]bool), byUnit: make(map[string][]Judgment), auxiliary: make(map[string]int)}
	for _, actor := range r.Actors {
		if actor.Kind == r.primaryKind() && actor.Role == "rater" {
			result.raters[actor.ID] = true
		}
	}
	for _, judgment := range r.Judgments {
		if err := ctx.Err(); err != nil {
			return primaryResponses{}, err
		}
		if result.raters[judgment.ActorID] {
			result.byUnit[judgment.UnitID] = append(result.byUnit[judgment.UnitID], judgment)
		} else {
			result.auxiliary[judgment.UnitID]++
			result.auxTotal++
		}
	}
	return result, ctx.Err()
}
