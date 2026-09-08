package feature

import (
	"context"
	"fmt"
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
	if !supported(block.Kind) {
		return missingMeasurements(hash, "unsupported_unit"), nil
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
	stats := statistics{frequencies: make(map[string]int)}
	for _, sentence := range block.Sentences {
		words := 0
		for _, token := range sentence.Tokens {
			if err := ctx.Err(); err != nil {
				return statistics{}, err
			}
			stats.counts.TokenVisits++
			if !token.Word || token.Protected {
				continue
			}
			words++
			stats.addWord(token)
			if len(stats.frequencies) > limits.MaxUniqueWords {
				return statistics{}, fmt.Errorf("features exceed max_unique_words")
			}
		}
		if words > 0 {
			stats.lengths = append(stats.lengths, words)
			stats.counts.Words += words
		}
	}
	stats.counts.Sentences = len(stats.lengths)
	stats.counts.Available = true
	return stats, nil
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
	if s.counts.Words == 0 {
		for i := 3; i < len(values); i++ {
			values[i] = number{reason: "no_prose_words"}
		}
	}
	return values
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

func supported(kind string) bool { return kind == "paragraph" || kind == "comment" || kind == "string" }
