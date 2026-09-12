package builtin

import (
	"context"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func readabilityMetric(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return measureReadability(ctx, view, emit, true)
}

func longParagraph(ctx context.Context, view rule.View, emit rule.Emitter) error {
	return measureReadability(ctx, view, emit, false)
}

func measureReadability(ctx context.Context, view rule.View, emit rule.Emitter, grade bool) error {
	m := newEditorialMatcher(ctx, view)
	for _, block := range view.Document.Blocks {
		if !proseBlock(block) {
			if err := view.Observe(feature.BlockObservation{BlockID: block.ID, Status: "inapplicable", Reason: "unsupported_unit"}); err != nil {
				return err
			}
			continue
		}
		evidence, err := readabilityEvidence(m, block, grade)
		if err != nil {
			return err
		}
		if len(evidence.Occurrences) == 0 {
			continue
		}
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func readabilityEvidence(m *editorialMatcher, block document.Block, grade bool) (rule.Evidence, error) {
	stats, err := measureProse(m, block)
	if err != nil {
		return rule.Evidence{}, err
	}
	counts := ariCounts{words: stats.Counts().Words, sentences: stats.Counts().Sentences}
	if grade {
		if counts, err = countProse(m, block); err != nil {
			return rule.Evidence{}, err
		}
	}
	if reason := readabilityAbsence(m.view.Parameters, counts, grade); reason != "" {
		return rule.Evidence{}, m.view.Observe(feature.BlockObservation{BlockID: block.ID, Status: "inapplicable", Reason: reason})
	}
	evidence, err := configuredReadability(m.view.Parameters, block, stats, counts, grade)
	if err != nil {
		return evidence, err
	}
	if err := m.view.Observe(feature.BlockObservation{BlockID: block.ID, Status: "evaluated"}); err != nil {
		return rule.Evidence{}, err
	}
	if len(evidence.Occurrences) == 0 {
		return evidence, nil
	}
	metrics, err := proseMetrics(stats)
	if err != nil {
		return rule.Evidence{}, err
	}
	evidence.Metrics = append(evidence.Metrics, metrics...)
	return evidence, nil
}

func readabilityAbsence(p rule.Parameters, counts ariCounts, grade bool) string {
	if counts.words == 0 {
		return "no_prose_words"
	}
	if grade && counts.words < p.MinWords {
		return "insufficient_words"
	}
	if grade && counts.sentences < p.MinSentences {
		return "insufficient_sentences"
	}
	return ""
}

func configuredReadability(
	p rule.Parameters, block document.Block, stats feature.Measurements, counts ariCounts, grade bool,
) (rule.Evidence, error) {
	if grade {
		return gradeEvidence(p, block, stats, counts)
	}
	return paragraphEvidence(p, block, stats), nil
}

// ariCounts are the grade metric's own denominators: prose words without
// identifier-shaped tokens, with each part of a hyphenated compound counted as
// a word, and sentences containing at least one such word. The shared feature
// set keeps its documented counts; the rule reports both values.
type ariCounts struct {
	words, characters, sentences int
}

func (c ariCounts) index() float64 {
	return 4.71*float64(c.characters)/float64(max(1, c.words)) + 0.5*float64(c.words)/float64(max(1, c.sentences)) - 21.43
}

func countProse(m *editorialMatcher, block document.Block) (ariCounts, error) {
	var counts ariCounts
	for _, sentence := range block.Sentences {
		words := 0
		for _, token := range sentence.Tokens {
			if err := m.spend(); err != nil {
				return ariCounts{}, err
			}
			if !token.Word || token.Protected || identifierShape(token.Text) {
				continue
			}
			parts, characters := compoundParts(token.Text)
			words += parts
			counts.characters += characters
		}
		if words > 0 {
			counts.words += words
			counts.sentences++
		}
	}
	return counts, nil
}

// identifierShape recognizes tokens that name code rather than prose: an inner
// uppercase letter, a digit, an underscore, a period, or a slash. Protected
// code spans never reach this test.
func identifierShape(text string) bool {
	if strings.ContainsAny(text, "_./") || strings.ContainsFunc(text, unicode.IsDigit) {
		return true
	}
	_, size := utf8.DecodeRuneInString(text)
	return strings.ContainsFunc(text[size:], unicode.IsUpper)
}

// compoundParts counts each hyphen-separated part with a letter or digit as
// its own word and returns the letters and digits of the whole token.
func compoundParts(text string) (words, characters int) {
	for part := range strings.SplitSeq(text, "-") {
		counted := 0
		for _, r := range part {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				counted++
			}
		}
		if counted > 0 {
			words++
			characters += counted
		}
	}
	return words, characters
}

func gradeEvidence(p rule.Parameters, block document.Block, stats feature.Measurements, counts ariCounts) (rule.Evidence, error) {
	shared, err := stats.Value("automated-readability-index")
	if err != nil {
		return rule.Evidence{}, err
	}
	if shared.Number == nil {
		return rule.Evidence{}, fmt.Errorf("required ARI feature is unavailable: %s", shared.Reason)
	}
	value := counts.index()
	if value <= float64(p.Onset) {
		return rule.Evidence{}, nil
	}
	return rule.Evidence{Kind: "heuristic", Activation: metricActivation(value, p.Onset, p.Saturation),
		Occurrences: blockOccurrences(block), Metrics: []rule.Metric{
			{Name: "automated-readability-index-prose", Value: value, Unit: "ARI-formula-units",
				Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
			{Name: "ari-words", Value: float64(counts.words), Unit: "words"},
			{Name: "ari-characters", Value: float64(counts.characters), Unit: "letters-and-digits"},
			{Name: "ari-sentences", Value: float64(counts.sentences), Unit: "sentences"},
			{Name: "automated-readability-index", Value: *shared.Number, Unit: "ARI-formula-units"},
		}}, nil
}

func paragraphEvidence(p rule.Parameters, block document.Block, stats feature.Measurements) rule.Evidence {
	long := 0
	for _, words := range stats.SentenceLengths() {
		if words > p.SentenceWords {
			long++
		}
	}
	if stats.Counts().Words <= p.Onset || long < p.MinLongSentences {
		return rule.Evidence{}
	}
	evidence := measured("heuristic", "paragraph-length", "prose-words", stats.Counts().Words,
		p.Onset, p.Saturation, blockOccurrences(block))
	evidence.Metrics = append(evidence.Metrics,
		rule.Metric{Name: "long-sentences", Value: float64(long), Unit: "sentences",
			Onset: float64(p.MinLongSentences)},
		rule.Metric{Name: "long-sentence-boundary", Value: float64(p.SentenceWords), Unit: "words"})
	return evidence
}
