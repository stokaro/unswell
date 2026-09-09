package builtin

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func repetitionRules() []rule.Rule {
	exact := descriptor(
		"repetition.exact-sentence",
		"This sentence is repeated; consider whether every occurrence is needed.",
		"repetition",
		"document",
		30,
	)
	exact.Defaults.Parameters = rule.Parameters{MinWords: 12, Window: "document"}
	exact.BlockObservations = true
	exact.Parameters = []string{"min_words", "window"}
	sample := "The client opens a connection to the server and sends the request with its credentials."
	exact.Examples = []rule.Example{{Text: sample + " " + sample, Match: true}, {Text: sample, Match: false}}
	near := descriptor(
		"repetition.near-sentence",
		"These sentences have high lexical overlap; check for repeated information.",
		"repetition",
		"document",
		22,
	)
	near.Requires = append(near.Requires, nlp.POS)
	near.Defaults.Parameters = rule.Parameters{
		MinWords:           12,
		Similarity:         0.85,
		Window:             "document",
		ProtectNegation:    true,
		ProtectNumbers:     true,
		ProtectIdentifiers: true,
	}
	near.Parameters = []string{"min_words", "similarity", "window", "protect_negation", "protect_numbers", "protect_identifiers"}
	near.Examples = []rule.Example{
		{Text: sample + " " + strings.Replace(sample, "opens", "creates", 1), Match: true},
		{Text: sample + " " + strings.Replace(sample, "opens", "never opens", 1), Match: false},
	}
	sentence := descriptor(
		"repetition.sentence-openers",
		"Several sentences repeat the same opening; consider removing repeated setup.",
		"repetition",
		"document",
		15,
	)
	sentence.Defaults.Parameters = rule.Parameters{MinWords: 8, OpenerWords: 3, AllowedOccurrences: 2, SaturationOccurrences: 5}
	sentence.BlockObservations = true
	sentence.Parameters = []string{"min_words", "opener_words", "allowed_occurrences", "saturation_occurrences"}
	sentence.Examples = []rule.Example{
		{
			Text: "The client opens a connection to the database. The client opens a file from the disk. " +
				"The client opens a channel to the service.",
			Match: true,
		},
		{Text: "Open the file. Open the connection. Open the channel.", Match: false},
	}
	paragraph := descriptor(
		"repetition.paragraph-openers",
		"Several paragraphs repeat the same opening formula.",
		"repetition",
		"document",
		18,
	)
	paragraph.Defaults.Parameters = sentence.Defaults.Parameters
	paragraph.BlockObservations = true
	paragraph.Parameters = slices.Clone(sentence.Parameters)
	paragraph.Examples = []rule.Example{
		{Text: strings.ReplaceAll(sentence.Examples[0].Text, ". ", ".\n\n"), Match: true},
		{Text: "The client opens connections.\n\nThe client opens files.", Match: false},
	}
	return []rule.Rule{
		check{exact, exactRepetition},
		check{near, nearRepetition},
		check{sentence, sentenceOpeners},
		check{paragraph, paragraphOpeners},
	}
}

func exactRepetition(ctx context.Context, view rule.View, emit rule.Emitter) error {
	groups := make(map[string][]rule.Occurrence)
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := addExactBlock(ctx, view, block, groups); err != nil {
			return err
		}
	}
	return emitGroups(groups, 1, 2, "exact", emit)
}

func emitGroups(groups map[string][]rule.Occurrence, allowed, saturation int, kind string, emit rule.Emitter) error {
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		occurrences := groups[key]
		if len(occurrences) <= allowed {
			continue
		}
		evidence := measured(kind, "repeated-occurrences", "occurrences", len(occurrences), allowed, saturation, occurrences)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return nil
}

func sentenceOpeners(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return openers(ctx, view, emit, false)
}
func paragraphOpeners(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return openers(ctx, view, emit, true)
}

func openers(ctx context.Context, view rule.View, emit rule.Emitter, paragraphs bool) error {
	groups := make(map[string][]rule.Occurrence)
	for _, block := range view.Document.Blocks {
		if err := ctx.Err(); err != nil {
			return err
		}
		if block.Kind != "paragraph" {
			if err := observeBlock(view, block, "unsupported_unit"); err != nil {
				return err
			}
			continue
		}
		if err := addOpenerBlock(ctx, view, block, groups, paragraphs); err != nil {
			return err
		}
	}
	return emitGroups(groups, view.Parameters.AllowedOccurrences, view.Parameters.SaturationOccurrences, "heuristic", emit)
}

func nearRepetition(ctx context.Context, view rule.View, emit rule.Emitter) error {
	sentences := allSentences(view.Document)
	parents, err := linkNearCandidates(ctx, view, sentences)
	if err != nil {
		return err
	}
	groups := make(map[string][]rule.Occurrence)
	for i, sentence := range sentences {
		key := fmt.Sprint(leader(parents, i))
		groups[key] = append(groups[key], sentenceOccurrence(sentence))
	}
	return emitGroups(groups, 1, 2, "heuristic", emit)
}

func linkNearCandidates(ctx context.Context, view rule.View, sentences []document.Sentence) ([]int, error) {
	units := make([]nearSentence, len(sentences))
	index := candidateIndex{postings: make(map[string][]int), remaining: view.MaxCandidates}
	parents := make([]int, len(sentences))
	for i := range parents {
		parents[i] = i
	}
	for i, sentence := range sentences {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if sentence.Words < view.Parameters.MinWords || sentenceKey(sentence) == "" {
			continue
		}
		words := normalizedWords(sentence)
		set, err := makeWordSet(ctx, words, len(sentence.Text))
		if err != nil {
			return nil, err
		}
		units[i] = nearSentence{key: sentenceKey(sentence), signature: protectedSignature(sentence, view.Parameters), words: set}
		shingles := bigrams(words)
		ordered, err := index.candidates(ctx, shingles)
		if err != nil {
			return nil, err
		}
		if err := joinNearPairs(ctx, units, i, ordered, parents, view.Parameters.Similarity); err != nil {
			return nil, err
		}
		for _, shingle := range shingles {
			index.postings[shingle] = append(index.postings[shingle], i)
		}
	}
	return parents, nil
}

type candidateIndex struct {
	postings  map[string][]int
	remaining int
}

func (index *candidateIndex) candidates(ctx context.Context, shingles []string) ([]int, error) {
	set := make(map[int]bool)
	for _, shingle := range shingles {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		for _, candidate := range index.postings[shingle] {
			index.remaining--
			if index.remaining < 0 {
				return nil, fmt.Errorf("near-sentence candidate index budget exceeded")
			}
			set[candidate] = true
		}
	}
	ordered := make([]int, 0, len(set))
	for candidate := range set {
		ordered = append(ordered, candidate)
	}
	slices.Sort(ordered)
	return ordered, nil
}

type nearSentence struct {
	key, signature string
	words          feature.WordSet
}

func joinNearPairs(ctx context.Context, units []nearSentence, index int, candidates, parents []int, similarity float64) error {
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		matches, err := nearPair(ctx, units[index], units[candidate], similarity)
		if err != nil {
			return err
		}
		if matches {
			parents[leader(parents, index)] = leader(parents, candidate)
		}
	}
	return nil
}

func leader(parents []int, index int) int {
	for parents[index] != index {
		parents[index] = parents[parents[index]]
		index = parents[index]
	}
	return index
}

func bigrams(words []string) []string {
	result := make([]string, 0)
	for i := 0; i+1 < len(words); i++ {
		result = append(result, words[i]+" "+words[i+1])
	}
	slices.Sort(result)
	return slices.Compact(result)
}

func nearPair(ctx context.Context, left, right nearSentence, similarity float64) (bool, error) {
	if left.key == right.key || left.signature != right.signature {
		return false, nil
	}
	overlap, err := feature.CompareWords(ctx, left.words, right.words)
	if err != nil {
		return false, err
	}
	return overlap.Union() > 0 && float64(overlap.Intersection())/float64(overlap.Union()) >= similarity, nil
}

func protectedSignature(sentence document.Sentence, parameters rule.Parameters) string {
	values := make([]string, 0)
	for i, token := range sentence.Tokens {
		negation := parameters.ProtectNegation &&
			slices.Contains([]string{"not", "no", "never", "without", "cannot", "n't"}, token.Normal)
		number := parameters.ProtectNumbers && strings.ContainsFunc(token.Text, unicode.IsDigit)
		identifier := parameters.ProtectIdentifiers && protectedIdentifier(token, i)
		if negation || number || identifier {
			values = append(values, token.Normal)
		}
	}
	return strings.Join(values, "|")
}

func protectedIdentifier(token document.Token, index int) bool {
	if strings.ContainsAny(token.Text, "_./") && token.Word {
		return true
	}
	if len(token.Text) > 1 && strings.ContainsFunc(token.Text[1:], unicode.IsUpper) {
		return true
	}
	return index > 0 && strings.HasPrefix(token.Tag, "NNP")
}
