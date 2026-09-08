package builtin

import (
	"context"
	"slices"

	"github.com/stokaro/unswell/rule"
)

type overlapEdge struct {
	left, right   int
	shared, union int
}

type overlapAnalysis struct {
	view      rule.View
	budget    repetitionBudget
	units     []lexicalUnit
	parents   []int
	edges     []overlapEdge
	summaries bool
}

func paragraphOverlap(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return analyzeOverlap(ctx, view, emit, false)
}

func summaryEcho(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return analyzeOverlap(ctx, view, emit, true)
}

func analyzeOverlap(ctx context.Context, view rule.View, emit rule.Emitter, summaries bool) error {
	a := overlapAnalysis{view: view, budget: repetitionBudget{ctx, view.MaxCandidates}, summaries: summaries}
	units, err := overlapUnits(view, summaries, &a.budget)
	if err != nil {
		return err
	}
	a.units, a.parents = units, make([]int, len(units))
	for i := range a.parents {
		a.parents[i] = i
	}
	if err := a.compare(); err != nil {
		return err
	}
	return a.emit(emit)
}

func (a *overlapAnalysis) compare() error {
	index := overlapIndex{postings: make(map[string][]int), budget: &a.budget}
	oldest := 0
	for i, unit := range a.units {
		first, err := a.expire(&index, oldest, i)
		if err != nil {
			return err
		}
		oldest = first
		candidates, err := index.candidates(unit.keys)
		if err != nil {
			return err
		}
		for _, candidate := range candidates {
			if err := a.pair(candidate, i); err != nil {
				return err
			}
		}
		if err := index.add(unit.keys, i); err != nil {
			return err
		}
	}
	return a.budget.ctx.Err()
}

func (a *overlapAnalysis) expire(index *overlapIndex, oldest, at int) (int, error) {
	for oldest < at && a.units[at].ordinal-a.units[oldest].ordinal >= a.view.Parameters.WindowBlocks {
		if err := index.expire(a.units[oldest].keys, oldest); err != nil {
			return oldest, err
		}
		oldest++
	}
	return oldest, nil
}

func (a *overlapAnalysis) pair(left, right int) error {
	u, v := a.units[left], a.units[right]
	if a.summaries && (!v.summary || u.summary) {
		return nil
	}
	if err := a.budget.spend(len(u.words) + len(v.words)); err != nil {
		return err
	}
	if u.block.Kind != v.block.Kind || u.signature != v.signature {
		return nil
	}
	shared, union := wordOverlap(u.words, v.words)
	if union > 0 && float64(shared)/float64(union) >= a.view.Parameters.Similarity {
		a.parents[leader(a.parents, right)] = leader(a.parents, left)
		a.edges = append(a.edges, overlapEdge{left, right, shared, union})
	}
	return nil
}

func (a *overlapAnalysis) emit(emit rule.Emitter) error {
	groups := make(map[int][]int)
	for i := range a.units {
		root := leader(a.parents, i)
		groups[root] = append(groups[root], i)
	}
	edges := make(map[int][]overlapEdge)
	for _, edge := range a.edges {
		root := leader(a.parents, edge.left)
		edges[root] = append(edges[root], edge)
	}
	roots := make([]int, 0, len(edges))
	for root := range edges {
		roots = append(roots, root)
	}
	slices.Sort(roots)
	for _, root := range roots {
		if err := a.budget.spend(len(groups[root])); err != nil {
			return err
		}
		if len(groups[root]) <= a.view.Parameters.AllowedOccurrences {
			continue
		}
		if err := a.emitCluster(groups[root], edges[root], emit); err != nil {
			return err
		}
	}
	return a.budget.ctx.Err()
}

func (a *overlapAnalysis) emitCluster(ids []int, edges []overlapEdge, emit rule.Emitter) error {
	if a.summaries {
		slices.SortStableFunc(ids, func(left, right int) int {
			if a.units[left].summary == a.units[right].summary {
				return left - right
			}
			if a.units[left].summary {
				return -1
			}
			return 1
		})
	}
	var occurrences []rule.Occurrence
	for _, id := range ids {
		occurrences = append(occurrences, unitOccurrences(a.units[id])...)
	}
	p := a.view.Parameters
	evidence := measured("heuristic", "overlapping-blocks", "blocks", len(ids), p.AllowedOccurrences, p.SaturationOccurrences, occurrences)
	minimum, shared, union := 1.0, 0, 0
	for _, edge := range edges {
		minimum = min(minimum, float64(edge.shared)/float64(edge.union))
		shared += edge.shared
		union += edge.union
	}
	evidence.Metrics = append(evidence.Metrics,
		rule.Metric{Name: "qualifying-pairs", Value: float64(len(edges)), Unit: "pairs"},
		rule.Metric{Name: "minimum-pair-jaccard", Value: minimum, Unit: "ratio", Onset: p.Similarity, Saturation: 1},
		rule.Metric{Name: "pair-word-intersections", Value: float64(shared), Unit: "summed-distinct-words"},
		rule.Metric{Name: "pair-word-unions", Value: float64(union), Unit: "summed-distinct-words"})
	return emit.Emit(evidence)
}
