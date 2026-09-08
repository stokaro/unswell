package builtin

import (
	"context"

	"github.com/stokaro/unswell/document"
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
		stats, err := measureProse(m, block)
		if err != nil {
			return err
		}
		if stats.words == 0 {
			continue
		}
		var evidence rule.Evidence
		if grade {
			evidence = gradeEvidence(view.Parameters, block, stats)
		} else {
			evidence = paragraphEvidence(view.Parameters, block, stats)
		}
		if len(evidence.Occurrences) == 0 {
			continue
		}
		evidence.Metrics = append(evidence.Metrics, stats.metrics()...)
		if err := emit.Emit(evidence); err != nil {
			return err
		}
	}
	return ctx.Err()
}

func gradeEvidence(p rule.Parameters, block document.Block, stats proseMeasurements) rule.Evidence {
	if stats.words < p.MinWords || len(stats.lengths) < p.MinSentences {
		return rule.Evidence{}
	}
	value := automatedReadability(stats.characters, stats.words, len(stats.lengths))
	if value <= float64(p.Onset) {
		return rule.Evidence{}
	}
	return rule.Evidence{Kind: "heuristic", Activation: metricActivation(value, p.Onset, p.Saturation),
		Occurrences: blockOccurrences(block), Metrics: []rule.Metric{
			{Name: "automated-readability-index", Value: value, Unit: "ARI-formula-units",
				Onset: float64(p.Onset), Saturation: float64(p.Saturation)},
		}}
}

func automatedReadability(characters, words, sentences int) float64 {
	return 4.71*float64(characters)/float64(words) + 0.5*float64(words)/float64(sentences) - 21.43
}

func paragraphEvidence(p rule.Parameters, block document.Block, stats proseMeasurements) rule.Evidence {
	long := 0
	for _, words := range stats.lengths {
		if words > p.SentenceWords {
			long++
		}
	}
	if stats.words <= p.Onset || long < p.MinLongSentences {
		return rule.Evidence{}
	}
	evidence := measured("heuristic", "paragraph-length", "prose-words", stats.words, p.Onset, p.Saturation, blockOccurrences(block))
	evidence.Metrics = append(evidence.Metrics,
		rule.Metric{Name: "long-sentences", Value: float64(long), Unit: "sentences",
			Onset: float64(p.MinLongSentences)},
		rule.Metric{Name: "long-sentence-boundary", Value: float64(p.SentenceWords), Unit: "words"})
	return evidence
}
