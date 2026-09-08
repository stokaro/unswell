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
	if stats.Counts().Words == 0 {
		return rule.Evidence{}, nil
	}
	evidence, err := configuredReadability(m.view.Parameters, block, stats, grade)
	if err != nil || len(evidence.Occurrences) == 0 {
		return evidence, err
	}
	metrics, err := proseMetrics(stats)
	if err != nil {
		return rule.Evidence{}, err
	}
	evidence.Metrics = append(evidence.Metrics, metrics...)
	return evidence, nil
}

func configuredReadability(p rule.Parameters, block document.Block, stats feature.Measurements, grade bool) (rule.Evidence, error) {
	if grade {
		return gradeEvidence(p, block, stats)
	}
	return paragraphEvidence(p, block, stats), nil
}

func gradeEvidence(p rule.Parameters, block document.Block, stats feature.Measurements) (rule.Evidence, error) {
	if stats.Counts().Words < p.MinWords || stats.Counts().Sentences < p.MinSentences {
		return rule.Evidence{}, nil
	}
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
