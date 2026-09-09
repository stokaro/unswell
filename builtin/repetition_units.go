package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

type lexicalUnit struct {
	block     document.Block
	ordinal   int
	words     feature.WordSet
	keys      []string
	signature string
	summary   bool
}

func makeLexicalUnit(view rule.View, block document.Block, ordinal int, budget *repetitionBudget,
	observations *candidateObservations) (lexicalUnit, bool, error) {
	unit := lexicalUnit{block: block, ordinal: ordinal}
	var signatures []string
	var words []string
	for _, sentence := range block.Sentences {
		if err := budget.spend(len(sentence.Tokens)); err != nil {
			return unit, false, err
		}
		if protectedSentence(sentence) {
			observations.advance(block.ID, candidateNoTokens)
			return unit, false, nil
		}
		words = proseWords(view, sentence, words)
		if signature := repetitionSignature(view, sentence); signature != "" {
			signatures = append(signatures, signature)
		}
	}
	var err error
	unit.words, err = makeWordSet(budget.ctx, words, len(block.Text))
	if err != nil {
		return unit, false, err
	}
	unit.keys = unit.words.ContentKeys()
	unit.signature = strings.Join(signatures, "|")
	enoughWords := len(words) >= view.Parameters.MinWords
	ok := enoughWords && len(unit.keys) >= 2
	observations.lexicalCandidate(block.ID, enoughWords, ok)
	return unit, ok, nil
}

func overlapUnits(view rule.View, summaries bool, budget *repetitionBudget, observations *candidateObservations) ([]lexicalUnit, error) {
	var units []lexicalUnit
	var scope []string
	ordinal := 0
	for _, block := range view.Document.Blocks {
		if summaries && block.Kind == "heading" {
			scope = nextSummaryScope(block, scope, view.Parameters.Phrases)
		}
		if !proseBlock(block) {
			continue
		}
		if len(scope) > 0 && !withinSection(block.Context, scope) {
			scope = nil
		}
		unit, ok, err := makeLexicalUnit(view, block, ordinal, budget, observations)
		if err != nil {
			return nil, err
		}
		ordinal++
		if !ok {
			continue
		}
		unit.summary = len(scope) > 0 && withinSection(block.Context, scope)
		units = append(units, unit)
	}
	return units, nil
}

func nextSummaryScope(block document.Block, current, headings []string) []string {
	label := normalizedHeading(block.Text)
	for _, heading := range headings {
		if label == normalizedHeading(heading) {
			return slices.Clone(block.Context)
		}
	}
	if !withinSection(block.Context, current) {
		return nil
	}
	return current
}

func normalizedHeading(text string) string {
	return strings.Join(strings.Fields(document.Normalize(text)), " ")
}

func withinSection(context, parent []string) bool {
	return len(parent) > 0 && len(context) >= len(parent) && slices.Equal(context[:len(parent)], parent)
}

func unitOccurrences(unit lexicalUnit) []rule.Occurrence {
	return blockOccurrences(unit.block)
}
