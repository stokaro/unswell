package builtin

import (
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

// restriction keeps the arguments of a written exclusion. It is a bounded
// surface representation, not a dependency parse or a semantic embedding.
type restriction struct {
	actor    []document.Token
	verb     document.Token
	property []document.Token
	object   string
	tense    string
}

func reformulatedClaims(view rule.View, budget *repetitionBudget, emit rule.Emitter) error {
	for _, block := range view.Document.Blocks {
		if err := budget.spend(1); err != nil {
			return err
		}
		if claimBlockReason(block, view.Parameters.MinWords) != "" || view.Parameters.WindowSentences < 2 {
			continue
		}
		if err := reformulatedBlock(view, block, budget, emit); err != nil {
			return err
		}
	}
	return budget.ctx.Err()
}

func reformulatedBlock(view rule.View, block document.Block, budget *repetitionBudget, emit rule.Emitter) error {
	for i := 1; i < len(block.Sentences); i++ {
		first, second := block.Sentences[i-1], block.Sentences[i]
		if err := budget.spend(1); err != nil {
			return err
		}
		// The marker check reads at most four tokens. Unmarked pairs never
		// enter the argument comparison and must not pay for that walk.
		if reformulationStart(restrictionTokens(second)) == 0 {
			continue
		}
		if err := budget.spend(len(first.Tokens) + len(second.Tokens)); err != nil {
			return err
		}
		if err := emitRestriction(view, block, first, second, budget, emit); err != nil {
			return err
		}
	}
	return nil
}

func emitRestriction(view rule.View, block document.Block, first, second document.Sentence,
	budget *repetitionBudget, emit rule.Emitter) error {
	a, b, ok := matchedRestrictions(view, first, second)
	if !ok {
		return nil
	}
	for _, tokens := range [][]document.Token{a, b} {
		valid, err := restrictionMappingEligible(view.Document.Source, block, tokens, budget)
		if err != nil || !valid {
			return err
		}
	}
	parts := []rule.Occurrence{tokenOccurrence(first, 0, len(a)), tokenOccurrence(second, 0, len(b))}
	evidence := measured("heuristic", "restriction-restatements", "patterns", 1, 0, 1, parts)
	evidence.Suggestion = "State the restriction once. Keep the original actor, scope, modality, property and exceptions."
	return emit.Emit(evidence)
}

// Restrictions use the shared source-map walk without serializing an identity.
func restrictionMappingEligible(source []byte, block document.Block, tokens []document.Token,
	budget *repetitionBudget) (bool, error) {
	start, end := tokens[0].Start, tokens[len(tokens)-1].End
	if start < 0 || end > len(block.Map) || start >= end {
		return false, nil
	}
	walk := claimIdentityWalk{source: source, mapped: block.Map, budget: budget,
		position: start, previous: block.Map[start].Start, previousStart: block.Map[start].Start}
	return walk.advance(end, "")
}

func restrictionTokens(sentence document.Sentence) []document.Token {
	tokens := sentence.Tokens
	if len(tokens) > 0 && frameWord(tokens[len(tokens)-1], ".", "!") {
		tokens = tokens[:len(tokens)-1]
	}
	return tokens
}

func eligibleRestriction(sentence document.Sentence) bool {
	if !eligibleClaimSentence(sentence) || !proseTokens(sentence.Tokens) {
		return false
	}
	return !slices.ContainsFunc(sentence.Tokens, func(t document.Token) bool {
		return t.Tag == "CD" || frameWord(t, "if", "when", "unless", "except", "before", "after", "until", "while",
			"because", "without", "may", "might", "can", "could", "must", "should", "according", "says", "said", "claims")
	})
}

func reformulationStart(tokens []document.Token) int {
	if len(tokens) > 3 && matches(tokens[:3], []string{"that", "is", ","}) {
		return 3
	}
	if len(tokens) > 4 && matches(tokens[:4], []string{"in", "other", "words", ","}) {
		return 4
	}
	return 0
}

func sameRestriction(a, b restriction) bool {
	return a.tense == b.tense && a.object == b.object && sameFiniteVerb(a.verb, b.verb) &&
		sameRestrictionWords(a.actor, b.actor) && sameRestrictionWords(a.property, b.property)
}

func sameRestrictionWords(a, b []document.Token) bool {
	return slices.EqualFunc(a, b, func(x, y document.Token) bool { return x.Normal == y.Normal })
}

func sameFiniteVerb(a, b document.Token) bool {
	if a.Normal == b.Normal {
		return true
	}
	return inflectedVerb(a, b) || inflectedVerb(b, a)
}

func inflectedVerb(finite, base document.Token) bool {
	if finite.Tag != "VBZ" || !slices.Contains([]string{"VB", "VBP"}, base.Tag) {
		return false
	}
	return finite.Normal == base.Normal+"s" || finite.Normal == base.Normal+"es" ||
		strings.HasSuffix(base.Normal, "y") && finite.Normal == strings.TrimSuffix(base.Normal, "y")+"ies"
}

func restrictionActor(tokens []document.Token) bool {
	return len(tokens) > 0 && len(tokens) <= 8 && nominalSubject(tokens) &&
		strings.HasPrefix(tokens[len(tokens)-1].Tag, "NN") &&
		!slices.ContainsFunc(tokens, func(t document.Token) bool { return t.Tag == "CC" || t.Tag == "POS" })
}

func matchedRestrictions(view rule.View, first, second document.Sentence) ([]document.Token, []document.Token, bool) {
	if !restrictionSentencePair(view, first, second) {
		return nil, nil, false
	}
	a, b := restrictionTokens(first), restrictionTokens(second)
	start := reformulationStart(b)
	if start == 0 || view.Exempts(first, 0, len(a)) || view.Exempts(second, 0, len(b)) {
		return nil, nil, false
	}
	left, ok := positiveRestriction(a)
	if !ok {
		left, ok = dependencyRestriction(a)
	}
	right, found := negativeRestriction(b[start:])
	if !ok || !found || !sameRestriction(left, right) {
		return nil, nil, false
	}
	return a, b, true
}

func restrictionSentencePair(view rule.View, first, second document.Sentence) bool {
	return first.Words >= view.Parameters.MinWords && second.Words >= view.Parameters.MinWords &&
		eligibleRestriction(first) && eligibleRestriction(second)
}
