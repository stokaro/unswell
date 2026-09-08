package builtin

import (
	"math"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

type proseMeasurements struct {
	words, characters                 int
	nouns, verbs, adjectives, adverbs int
	lengths                           []int
	frequencies                       map[string]int
}

func measureProse(m *editorialMatcher, block document.Block) (proseMeasurements, error) {
	stats := proseMeasurements{frequencies: make(map[string]int)}
	for _, sentence := range block.Sentences {
		words := 0
		for _, token := range sentence.Tokens {
			if err := m.spend(); err != nil {
				return stats, err
			}
			if !token.Word || token.Protected {
				continue
			}
			words++
			stats.addWord(token)
		}
		if words > 0 {
			stats.lengths = append(stats.lengths, words)
			stats.words += words
		}
	}
	return stats, nil
}

func (s *proseMeasurements) addWord(token document.Token) {
	s.frequencies[token.Normal]++
	for _, r := range token.Text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			s.characters++
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

func (s proseMeasurements) metrics() []rule.Metric {
	if s.words == 0 {
		return nil
	}
	mean, deviation, shortest, longest := lengthStatistics(s.lengths)
	hapax := 0
	for _, count := range s.frequencies {
		if count == 1 {
			hapax++
		}
	}
	metrics := []rule.Metric{
		{Name: "prose-words", Value: float64(s.words), Unit: "words"},
		{Name: "counted-characters", Value: float64(s.characters), Unit: "letters-and-digits"},
		{Name: "prose-sentences", Value: float64(len(s.lengths)), Unit: "sentences"},
		{Name: "mean-sentence-words", Value: mean, Unit: "words/sentence"},
		{Name: "sentence-word-stddev", Value: deviation, Unit: "words"},
		{Name: "shortest-sentence", Value: float64(shortest), Unit: "words"},
		{Name: "longest-sentence", Value: float64(longest), Unit: "words"},
	}
	for _, fraction := range []struct {
		name  string
		count int
	}{
		{"type-token-ratio", len(s.frequencies)}, {"hapax-token-ratio", hapax},
		{"noun-token-ratio", s.nouns}, {"verb-token-ratio", s.verbs},
		{"adjective-token-ratio", s.adjectives}, {"adverb-token-ratio", s.adverbs},
	} {
		metrics = append(metrics, rule.Metric{Name: fraction.name, Value: float64(fraction.count) / float64(s.words), Unit: "ratio"})
	}
	return metrics
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

func metricActivation(value float64, onset, saturation int) int {
	return int(math.Min(1000, math.Max(0, (value-float64(onset))*1000/float64(saturation-onset))))
}

func blockOccurrences(block document.Block) []rule.Occurrence {
	var occurrences []rule.Occurrence
	for _, sentence := range block.Sentences {
		if sentence.Words > 0 {
			occurrences = append(occurrences, sentenceOccurrence(sentence))
		}
	}
	return occurrences
}
