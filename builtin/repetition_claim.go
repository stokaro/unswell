package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type repeatedClaim struct {
	key        string
	occurrence rule.Occurrence
	complete   bool
	novel      bool
}

type claimGroup struct {
	first       int
	complete    bool
	novel       bool
	occurrences []rule.Occurrence
}

type claimCollector struct {
	view         rule.View
	scopes       map[int]int
	budget       *repetitionBudget
	groups       map[string]*claimGroup
	observations *candidateObservations
	previous     map[int]claimCandidate
	ordinal      int
}

func repeatedClaims(ctx context.Context, view rule.View, emit rule.Emitter) error {
	scopes, err := repetitionScopes(ctx, view)
	if err != nil {
		return err
	}
	c := &claimCollector{view: view, scopes: scopes, budget: &repetitionBudget{ctx, view.MaxCandidates},
		groups: make(map[string]*claimGroup), previous: make(map[int]claimCandidate),
		observations: newCandidateObservations(view, func(block document.Block) bool {
			return proseBlock(block) && block.List == nil
		})}
	for _, block := range view.Document.Blocks {
		if err := c.addBlock(block, emit); err != nil {
			return err
		}
	}
	keys := make([]string, 0, len(c.groups))
	for key := range c.groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if err := c.budget.spend(1); err != nil {
			return err
		}
		if err := emitClaimGroup(c.groups[key], emit); err != nil {
			return err
		}
	}
	if err := reformulatedClaims(view, c.budget, emit); err != nil {
		return err
	}
	return c.observations.finish(ctx, view)
}

func (c *claimCollector) addBlock(block document.Block, emit rule.Emitter) error {
	if err := c.budget.spend(1); err != nil {
		return err
	}
	if claimBlockReason(block, c.view.Parameters.MinWords) != "" {
		c.observations.advance(block.ID, candidateInsufficientWords)
		c.ordinal += len(block.Sentences)
		return nil
	}
	for _, sentence := range block.Sentences {
		if err := c.addSentence(block, sentence, emit); err != nil {
			return err
		}
		c.ordinal++
	}
	return nil
}

func (c *claimCollector) addSentence(block document.Block, sentence document.Sentence, emit rule.Emitter) error {
	claim, err := makeRepeatedClaim(c.view, block, sentence, c.budget)
	if err != nil {
		return err
	}
	c.observations.lexicalCandidate(block.ID, sentence.Words >= c.view.Parameters.MinWords, claim.key != "")
	if claim.key == "" {
		return nil
	}
	scope := c.scopes[block.ID]
	// Separate comments and string literals can document different declarations.
	if block.Kind == "comment" || block.Kind == "string" {
		scope = -block.ID - 1
	}
	window := c.view.Parameters.WindowSentences
	observeClaimPair(c.previous, c.observations, scope, block.ID, c.ordinal, window)
	return collectClaim(c.groups, fmt.Sprintf("%d/%s", scope, claim.key), claim, c.ordinal, window, emit)
}

type claimCandidate struct{ blockID, ordinal int }

func observeClaimPair(previous map[int]claimCandidate, observations *candidateObservations, scope, blockID, ordinal, window int) {
	if prior, ok := previous[scope]; ok && ordinal-prior.ordinal < window {
		observations.advance(blockID, candidateEvaluated)
		observations.advance(prior.blockID, candidateEvaluated)
	}
	previous[scope] = claimCandidate{blockID, ordinal}
}

func collectClaim(groups map[string]*claimGroup, key string, claim repeatedClaim, ordinal, window int, emit rule.Emitter) error {
	group := groups[key]
	if group != nil && ordinal-group.first >= window {
		if err := emitClaimGroup(group, emit); err != nil {
			return err
		}
		group = nil
	}
	if group == nil {
		group = &claimGroup{first: ordinal}
		groups[key] = group
	}
	group.complete = group.complete || claim.complete
	group.novel = group.novel || claim.novel
	group.occurrences = append(group.occurrences, claim.occurrence)
	return nil
}

func emitClaimGroup(group *claimGroup, emit rule.Emitter) error {
	if len(group.occurrences) < 2 || !group.complete || !group.novel {
		return nil
	}
	return emit.Emit(measured("exact", "repeated-assertions", "occurrences", len(group.occurrences), 1, 2, group.occurrences))
}

func claimBlockReason(block document.Block, minimum int) string {
	if block.List != nil {
		return "unsupported_unit"
	}
	return proseMinimumReason(block, minimum)
}

func makeRepeatedClaim(view rule.View, block document.Block, sentence document.Sentence, budget *repetitionBudget) (repeatedClaim, error) {
	if err := budget.spend(len(sentence.Tokens) + 1); err != nil {
		return repeatedClaim{}, err
	}
	tokens := sentence.Tokens
	if !eligibleClaimSentence(sentence) {
		return repeatedClaim{}, nil
	}
	end, complete := claimEnd(tokens)
	if !complete && qualifiedClaimTail(tokens[end:]) {
		return repeatedClaim{}, nil
	}
	tokens = tokens[:end]
	if !assertionShape(tokens, view.Parameters.MinWords) || view.Exempts(sentence, 0, end) {
		return repeatedClaim{}, nil
	}
	key, opaque, err := claimIdentity(view.Document.Source, block, tokens, budget)
	if err != nil || key == "" {
		return repeatedClaim{}, err
	}
	return repeatedClaim{key, tokenOccurrence(sentence, 0, end), complete, novelClaim(complete, opaque, sentence.Words)}, nil
}

func novelClaim(complete, opaque bool, words int) bool {
	return !complete || opaque || words < 12
}

func eligibleClaimSentence(sentence document.Sentence) bool {
	return len(sentence.Tokens) >= 2 && len(sentence.Tokens) <= 64 && !quotedClaim(sentence.Tokens) && !question(sentence)
}

func qualifiedClaimTail(tokens []document.Token) bool {
	return slices.ContainsFunc(tokens, func(t document.Token) bool {
		return t.Protected || t.Tag == "CD" || frameWord(t, "not", "never", "only", "if", "when", "unless", "except",
			"must", "may", "might", "should", "before", "after", "until")
	})
}

func claimEnd(tokens []document.Token) (int, bool) {
	for i := 1; i+1 < len(tokens); i++ {
		if tokens[i].Text == "," && frameWord(tokens[i+1], "which") {
			return i, false
		}
	}
	end := len(tokens)
	if slices.Contains([]string{".", "!"}, tokens[end-1].Text) {
		end--
	}
	return end, true
}

// Finite verbs are surface cues, not a parse. Imperatives and incomplete noun
// phrases do not qualify; identity includes every condition inside the claim.
func assertionShape(tokens []document.Token, minimum int) bool {
	words, finite := 0, false
	for i, token := range tokens {
		if token.Word && !token.Protected {
			words++
		}
		if i > 0 && !token.Protected && slices.Contains([]string{"VBZ", "VBP", "VBD", "MD"}, token.Tag) {
			finite = true
		}
	}
	return len(tokens) > 0 && tokens[0].Tag != "VB" && finite && words >= minimum
}

// Code operands remain opaque and retain their delimiters, case and punctuation.
// Spaces, commands and expressions do not become editorial vocabulary.
func opaqueClaimAtom(source []byte, token document.Token, budget *repetitionBudget) (string, error) {
	if len(token.Spans) != 1 || !token.Spans[0].Valid(len(source)) {
		return "", nil
	}
	span := token.Spans[0]
	if err := budget.spend(span.End - span.Start); err != nil {
		return "", err
	}
	if span.End-span.Start > 128 {
		return "", nil
	}
	raw := string(source[span.Start:span.End])
	if !validClaimAtom(raw) {
		return "", nil
	}
	return raw, nil
}

func validClaimAtom(raw string) bool {
	if len(raw) < 3 || raw[0] != '`' || raw[len(raw)-1] != '`' {
		return false
	}
	atom := strings.Trim(raw, "`")
	if atom == "" {
		return false
	}
	return !strings.ContainsFunc(atom, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("._-", r)
	})
}
