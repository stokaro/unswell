package feature

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

type statistics struct {
	counts                            Counts
	nouns, verbs, adjectives, adverbs int
	lengths                           []int
	spans                             []document.Span
	frequencies                       map[string]int
	// Repetition counts stay inside one unit: adjacent pairs never cross a
	// sentence boundary, and a sequence identity is a digest, not stored prose.
	bigrams   map[string]int
	openers   map[string]int
	sequences map[uint64]int
	pairs     int
	sentences []sentenceIdentity
}

type sentenceIdentity struct {
	opener   string
	sequence uint64
}

// Measure computes the catalog on one enriched paragraph, comment, or string.
// Inputs must remain unchanged until it returns. The result retains no input
// buffers. Missing capabilities produce unavailable values, never guessed tags.
func Measure(ctx context.Context, block document.Block, identity Identity, limits Limits) (Measurements, error) {
	if err := validateInputs(ctx, block, identity, limits); err != nil {
		return Measurements{}, err
	}
	hash, err := unitHash(block, identity)
	if err != nil {
		return Measurements{}, err
	}
	if err := ctx.Err(); err != nil {
		return Measurements{}, err
	}
	if !SupportsBlock(block.Kind) {
		return missingMeasurements(hash, "unsupported_unit"), nil
	}
	if block.Excluded {
		return missingMeasurements(hash, "excluded_unit"), nil
	}
	if !slices.Contains(identity.Capabilities, nlp.Tokens) || !slices.Contains(identity.Capabilities, nlp.Sentences) {
		return missingMeasurements(hash, "capability_missing"), nil
	}
	return measureAvailable(ctx, block, identity, limits, hash)
}

func measureAvailable(ctx context.Context, block document.Block, identity Identity, limits Limits, hash string) (Measurements, error) {
	stats, err := count(ctx, block, limits)
	if err != nil {
		return Measurements{}, err
	}
	m := Measurements{counts: stats.counts, lengths: stats.lengths, spans: stats.spans, hash: hash}
	m.values = stats.values()
	if !slices.Contains(identity.Capabilities, nlp.POS) {
		for i := 9; i <= 12; i++ {
			m.values[i] = number{reason: "capability_missing"}
		}
	}
	if err := ctx.Err(); err != nil {
		return Measurements{}, err
	}
	return m, nil
}

func missingMeasurements(hash, reason string) Measurements {
	m := Measurements{hash: hash, values: make([]number, len(definitions()))}
	for i := range m.values {
		m.values[i].reason = reason
	}
	return m
}

func count(ctx context.Context, block document.Block, limits Limits) (statistics, error) {
	stats := statistics{frequencies: make(map[string]int), bigrams: make(map[string]int),
		openers: make(map[string]int), sequences: make(map[uint64]int)}
	for _, sentence := range block.Sentences {
		words, sequence, err := stats.countSentence(ctx, sentence, limits)
		if err != nil {
			return statistics{}, err
		}
		if words > 0 {
			stats.lengths = append(stats.lengths, words)
			stats.counts.Words += words
			stats.addSentence(sentence, sequence)
		}
	}
	stats.counts.Sentences = len(stats.lengths)
	stats.counts.Available = true
	return stats, nil
}

// countSentence accumulates one sentence's words, adjacent pairs, and sequence
// identity. Adjacent pairs never cross a sentence boundary.
func (s *statistics) countSentence(ctx context.Context, sentence document.Sentence, limits Limits) (int, uint64, error) {
	words, previous := 0, ""
	digest := fnv.New64a()
	for _, token := range sentence.Tokens {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}
		s.counts.TokenVisits++
		if !token.Word || token.Protected {
			continue
		}
		words++
		s.addWord(token)
		if previous != "" {
			s.bigrams[previous+"\x00"+token.Normal]++
			s.pairs++
		}
		previous = token.Normal
		_, _ = digest.Write([]byte(token.Normal + "\x00"))
		if len(s.frequencies) > limits.MaxUniqueWords || len(s.bigrams) > limits.MaxUniqueWords {
			return 0, 0, fmt.Errorf("features exceed max_unique_words")
		}
	}
	return words, digest.Sum64(), nil
}

// addSentence records one sentence's opener and sequence identity, so repeated
// openers and duplicated sentences are counted without keeping their prose.
func (s *statistics) addSentence(sentence document.Sentence, sequence uint64) {
	opener := ""
	for _, token := range sentence.Tokens {
		if token.Word && !token.Protected {
			opener = token.Normal
			break
		}
	}
	s.openers[opener]++
	s.sequences[sequence]++
	s.sentences = append(s.sentences, sentenceIdentity{opener: opener, sequence: sequence})
}

func (s *statistics) addWord(token document.Token) {
	s.frequencies[token.Normal]++
	s.spans = append(s.spans, token.Spans...)
	for _, r := range token.Text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			s.counts.Characters++
		}
	}
	switch {
	case strings.HasPrefix(token.Tag, "NN"):
		s.nouns++
	case strings.HasPrefix(token.Tag, "VB"):
		s.verbs++
	case strings.HasPrefix(token.Tag, "JJ"):
		s.adjectives++
	case strings.HasPrefix(token.Tag, "RB"):
		s.adverbs++
	}
}

func (s statistics) values() []number {
	mean, deviation, shortest, longest := lengthStatistics(s.lengths)
	values := []number{{value: float64(s.counts.Words)}, {value: float64(s.counts.Characters)},
		{value: float64(s.counts.Sentences)}, {value: mean}, {value: deviation},
		{value: float64(shortest)}, {value: float64(longest)}}
	hapax := 0
	for _, count := range s.frequencies {
		if count == 1 {
			hapax++
		}
	}
	for _, count := range []int{len(s.frequencies), hapax, s.nouns, s.verbs, s.adjectives, s.adverbs} {
		values = append(values, number{value: float64(count) / float64(max(1, s.counts.Words))})
	}
	ari := 4.71*float64(s.counts.Characters)/float64(max(1, s.counts.Words)) +
		0.5*float64(s.counts.Words)/float64(max(1, s.counts.Sentences)) - 21.43
	values = append(values, number{value: ari})
	values = append(values, s.repetitionValues()...)
	if s.counts.Words == 0 {
		for i := 3; i < len(values); i++ {
			values[i] = number{reason: "no_prose_words"}
		}
	}
	return values
}

// repetitionValues describe how much of a unit repeats itself. They are
// descriptive ratios: technical prose repeats terms for good reasons.
func (s statistics) repetitionValues() []number {
	peak := 0
	for _, count := range s.frequencies {
		peak = max(peak, count)
	}
	repeatedPairs := 0
	for _, count := range s.bigrams {
		if count > 1 {
			repeatedPairs += count
		}
	}
	repeatedOpeners, duplicated := 0, 0
	for _, identity := range s.sentences {
		if s.openers[identity.opener] > 1 {
			repeatedOpeners++
		}
		if s.sequences[identity.sequence] > 1 {
			duplicated++
		}
	}
	sentences := float64(max(1, len(s.sentences)))
	pairs := number{value: float64(repeatedPairs) / float64(max(1, s.pairs))}
	if s.pairs == 0 {
		pairs = number{reason: "no_eligible_pair"}
	}
	return []number{
		{value: float64(peak) / float64(max(1, s.counts.Words))},
		pairs,
		{value: float64(repeatedOpeners) / sentences},
		{value: float64(duplicated) / sentences},
	}
}

func lengthStatistics(lengths []int) (mean, deviation float64, shortest, longest int) {
	if len(lengths) == 0 {
		return 0, 0, 0, 0
	}
	shortest = lengths[0]
	m2 := 0.0
	for i, words := range lengths {
		delta := float64(words) - mean
		mean += delta / float64(i+1)
		m2 += delta * (float64(words) - mean)
		shortest, longest = min(shortest, words), max(longest, words)
	}
	return mean, math.Sqrt(m2 / float64(len(lengths))), shortest, longest
}

// SupportsBlock reports whether the current block contract measures this kind.
func SupportsBlock(kind string) bool {
	return kind == "paragraph" || kind == "comment" || kind == "string"
}
