package builtin

import (
	"context"
	"fmt"

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
	if reason := readabilityAbsence(m.view.Parameters, stats.Counts(), grade); reason != "" {
		return rule.Evidence{}, m.view.Observe(feature.BlockObservation{BlockID: block.ID, Status: "inapplicable", Reason: reason})
	}
	evidence, err := configuredReadability(m.view.Parameters, block, stats, grade)
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

func readabilityAbsence(p rule.Parameters, counts feature.Counts, grade bool) string {
	if counts.Words == 0 {
		return "no_prose_words"
	}
	if grade && counts.Words < p.MinWords {
		return "insufficient_words"
	}
	if grade && counts.Sentences < p.MinSentences {
		return "insufficient_sentences"
	}
	return ""
}

func configuredReadability(p rule.Parameters, block document.Block, stats feature.Measurements, grade bool) (rule.Evidence, error) {
	if grade {
		return gradeEvidence(p, block, stats)
	}
	return paragraphEvidence(p, block, stats), nil
}

func gradeEvidence(p rule.Parameters, block document.Block, stats feature.Measurements) (rule.Evidence, error) {
	metric, err := stats.Value("automated-readability-index")
	if err != nil {
		return rule.Evidence{}, err
	}
	if metric.Number == nil {
		return rule.Evidence{}, fmt.Errorf("required ARI feature is unavailable: %s", metric.Reason)
	}
	value := *metric.Number
	if value <= float64(p.Onset) {
		return rule.Evidence{}, nil
	}
	return rule.Evidence{Kind: "heuristic", Activation: metricActivation(value, p.Onset, p.Saturation),
		Occurrences: blockOccurrences(block), Metrics: []rule.Metric{
			{Name: "automated-readability-index", Value: value, Unit: "ARI-formula-units",
				Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
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
