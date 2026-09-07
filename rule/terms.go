package rule

import (
	"cmp"
	"fmt"
	"slices"
	"sort"

	"github.com/stokaro/unswell/document"
)

// TermMatches is an immutable index of approved token ranges. A nil index has no
// matches. Construct it with NewTermMatches and consult it through View.Exempts.
type TermMatches struct {
	sentences map[[2]int][]TokenRange
}

// NewTermMatches owns and indexes up to 100000 ranges for logarithmic lookups.
// Each candidate must fit within one term; overlapping terms do not combine into
// a larger exemption. Callers must supply ranges from the same analyzed document.
func NewTermMatches(ranges []TokenRange) (*TermMatches, error) {
	if len(ranges) > 100000 {
		return nil, fmt.Errorf("term matches exceed 100000 ranges")
	}
	result := &TermMatches{sentences: make(map[[2]int][]TokenRange)}
	for _, term := range ranges {
		if term.BlockID < 0 || term.SentenceID < 0 || term.Start < 0 || term.End <= term.Start {
			return nil, fmt.Errorf("invalid term token range")
		}
		key := [2]int{term.BlockID, term.SentenceID}
		result.sentences[key] = append(result.sentences[key], term)
	}
	for _, terms := range result.sentences {
		slices.SortFunc(terms, func(a, b TokenRange) int { return cmp.Compare(a.Start, b.Start) })
		for i := 1; i < len(terms); i++ {
			terms[i].End = max(terms[i].End, terms[i-1].End)
		}
	}
	return result, nil
}

func (m *TermMatches) contains(sentence document.Sentence, start, end int) bool {
	if m == nil {
		return false
	}
	terms := m.sentences[[2]int{sentence.BlockID, sentence.ID}]
	i := sort.Search(len(terms), func(i int) bool { return terms[i].Start > start })
	return i > 0 && terms[i-1].End >= end
}
