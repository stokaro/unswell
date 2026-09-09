package builtin

import (
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func newListObservations(view rule.View) *candidateObservations {
	o := newCandidateObservations(view, func(block document.Block) bool { return block.Kind == "list-item" })
	if o == nil {
		return nil
	}
	o.unvisited, o.missingCandidate, o.missingPair = "no_eligible_list", "no_eligible_list", "no_eligible_window"
	return o
}

func (o *candidateObservations) listGroup(group shortList, stage candidateStage) {
	if o == nil {
		return
	}
	for _, block := range group.blocks {
		o.advance(block.ID, stage)
	}
}
