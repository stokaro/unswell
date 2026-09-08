package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type lexicalUnit struct {
	block     document.Block
	ordinal   int
	words     map[string]bool
	keys      []string
	signature string
	summary   bool
}

func makeLexicalUnit(view rule.View, block document.Block, ordinal int, budget *repetitionBudget) (lexicalUnit, bool, error) {
	unit := lexicalUnit{block: block, ordinal: ordinal, words: make(map[string]bool)}
	var signatures []string
	words := 0
	for _, sentence := range block.Sentences {
		if err := budget.spend(len(sentence.Tokens)); err != nil {
			return unit, false, err
		}
		if protectedSentence(sentence) {
			return unit, false, nil
		}
		words += proseWordSet(view, sentence, unit.words)
		if signature := repetitionSignature(view, sentence); signature != "" {
			signatures = append(signatures, signature)
		}
	}
	unit.keys = contentKeys(unit.words)
	unit.signature = strings.Join(signatures, "|")
	return unit, words >= view.Parameters.MinWords && len(unit.keys) >= 2, nil
}

func overlapUnits(view rule.View, summaries bool, budget *repetitionBudget) ([]lexicalUnit, error) {
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
		unit, ok, err := makeLexicalUnit(view, block, ordinal, budget)
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
	var occurrences []rule.Occurrence
	for _, sentence := range unit.block.Sentences {
		if sentence.Words > 0 {
			occurrences = append(occurrences, sentenceOccurrence(sentence))
		}
	}
	return occurrences
}
