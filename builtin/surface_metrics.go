package builtin

import (
	"errors"
	"fmt"
	"math"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

func measureProse(m *editorialMatcher, block document.Block) (feature.Measurements, error) {
	var stats feature.Measurements
	var err error
	if m.view.Features != nil {
		stats, err = m.view.Features.Block(block.ID)
	} else {
		// Direct Rule.Evaluate callers supply already enriched, unversioned input.
		// Engine calls use a shared set with the actual provider identity.
		capabilities := []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.POS}
		stats, err = feature.Measure(m.ctx, block, feature.Identity{
			NLP:          nlp.Identity{Name: "caller-supplied", Version: "unspecified", Capabilities: capabilities},
			Capabilities: capabilities, Source: "unversioned-rule-view", Policy: "unversioned-rule-view",
			Vocabulary: "no-term-exemptions", Preprocessing: "provided-tokens-v1",
		}, feature.Limits{MaxTokens: m.view.MaxCandidates, MaxUniqueWords: m.view.MaxCandidates,
			MaxBytes: max(1, len(block.Text)), MaxBlocks: 1})
	}
	if err != nil {
		if errors.Is(err, feature.ErrTokenLimit) {
			return feature.Measurements{}, rule.Abstain(rule.ReasonBudgetExhausted,
				fmt.Errorf("editorial pattern checks exceed max_candidates: %w", err))
		}
		return feature.Measurements{}, err
	}
	if !stats.Counts().Available {
		return feature.Measurements{}, fmt.Errorf("required prose measurements are unavailable")
	}
	m.checks += stats.Counts().TokenVisits
	if m.checks > m.view.MaxCandidates {
		return feature.Measurements{}, rule.Abstain(rule.ReasonBudgetExhausted,
			fmt.Errorf("editorial pattern checks exceed max_candidates"))
	}
	return stats, m.ctx.Err()
}

func proseMetrics(stats feature.Measurements) ([]rule.Metric, error) {
	var result []rule.Metric
	for _, value := range stats.Values() {
		if value.ID == "automated-readability-index" {
			continue
		}
		if value.Number == nil {
			return nil, fmt.Errorf("required readability feature %s is unavailable: %s", value.ID, value.Reason)
		}
		result = append(result, rule.Metric{Name: value.ID, Value: *value.Number, Unit: value.Unit})
	}
	return result, nil
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
