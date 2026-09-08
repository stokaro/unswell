package builtin

import (
	"cmp"
	"slices"

	"github.com/stokaro/unswell/document"
)

type proseGapIndex struct {
	spans []document.Span
	next  int
}

func newProseGapIndex(exclusions []document.Exclusion) proseGapIndex {
	spans := make([]document.Span, 0, len(exclusions))
	for _, excluded := range exclusions {
		spans = append(spans, excluded.Span)
	}
	slices.SortFunc(spans, func(a, b document.Span) int { return cmp.Compare(a.Start, b.Start) })
	merged := spans[:0]
	for _, span := range spans {
		if len(merged) > 0 && merged[len(merged)-1].End >= span.Start {
			merged[len(merged)-1].End = max(merged[len(merged)-1].End, span.End)
			continue
		}
		merged = append(merged, span)
	}
	return proseGapIndex{spans: merged}
}

func (g *proseGapIndex) between(before, after *document.Block) bool {
	if before == nil {
		return false
	}
	for g.next < len(g.spans) && g.spans[g.next].End <= before.Span.End {
		g.next++
	}
	return g.next < len(g.spans) && g.spans[g.next].Start < after.Span.Start
}
