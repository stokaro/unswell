package builtin

import (
	"sort"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type mappedSentenceRange struct{ start, end, id int }

type mappedRangeIndex struct {
	block     document.Block
	sentences []mappedSentenceRange
}

func newMappedRangeIndex(block document.Block) mappedRangeIndex {
	index := mappedRangeIndex{block: block}
	for _, sentence := range block.Sentences {
		if len(sentence.Tokens) > 0 && sentence.Words > 0 {
			index.sentences = append(index.sentences, mappedSentenceRange{
				sentence.Tokens[0].Start, sentence.Tokens[len(sentence.Tokens)-1].End, sentence.ID,
			})
		}
	}
	return index
}

func (index mappedRangeIndex) occurrences(start, end int) []rule.Occurrence {
	first := sort.Search(len(index.sentences), func(i int) bool { return index.sentences[i].end > start })
	var result []rule.Occurrence
	for i := first; i < len(index.sentences) && index.sentences[i].start < end; i++ {
		sentence := index.sentences[i]
		left, right := max(start, sentence.start), min(end, sentence.end)
		result = append(result, rule.Occurrence{BlockID: index.block.ID, SentenceID: sentence.id, Spans: index.block.Spans(left, right)})
	}
	return result
}
