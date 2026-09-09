package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type candidateStage uint8

const (
	candidateNoContext candidateStage = iota
	candidateInsufficientWords
	candidateNoTokens
	candidatePrepared
	candidateEvaluated
)

type candidateObservations struct {
	stages                                   map[int]candidateStage
	supported                                func(document.Block) bool
	unvisited, missingCandidate, missingPair string
}

func newCandidateObservations(view rule.View, supported func(document.Block) bool) *candidateObservations {
	if view.Observer == nil {
		return nil
	}
	return &candidateObservations{stages: make(map[int]candidateStage), supported: supported, unvisited: "no_eligible_pair",
		missingCandidate: "no_eligible_tokens", missingPair: "no_eligible_pair"}
}

func (o *candidateObservations) advance(blockID int, stage candidateStage) {
	if o != nil {
		o.stages[blockID] = max(o.stages[blockID], stage)
	}
}

func (o *candidateObservations) lexicalCandidate(blockID int, enoughWords, usable bool) {
	if o == nil {
		return
	}
	stage := candidateInsufficientWords
	if enoughWords {
		stage = candidateNoTokens
	}
	if usable {
		stage = candidatePrepared
	}
	o.advance(blockID, stage)
}

func (o *candidateObservations) finish(ctx context.Context, view rule.View) error {
	if o == nil {
		return ctx.Err()
	}
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := observeBlock(view, block, o.reason(block)); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func (o *candidateObservations) reason(block document.Block) string {
	if o.supported != nil && !o.supported(block) {
		return "unsupported_unit"
	}
	if len(block.Sentences) == 0 {
		return "no_sentences"
	}
	switch o.stages[block.ID] {
	case candidateNoContext:
		return o.unvisited
	case candidateInsufficientWords:
		return "insufficient_words"
	case candidateNoTokens:
		return o.missingCandidate
	case candidatePrepared:
		return o.missingPair
	case candidateEvaluated:
		return ""
	}
	return o.missingCandidate
}
