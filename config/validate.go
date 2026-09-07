package config

import (
	"fmt"
	"math"
	"slices"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/rule"
)

func validate(policy Policy, catalog []rule.Descriptor) error {
	if err := extract.ValidatePolicy(policy.Extraction); err != nil {
		return err
	}
	if err := validateAnalysis(policy.Analysis); err != nil {
		return err
	}
	if err := validateGate(policy.Gate); err != nil {
		return err
	}
	for _, descriptor := range catalog {
		if err := validateSettings(policy.Rules[descriptor.ID], descriptor.Parameters); err != nil {
			return fmt.Errorf("%s: %w", descriptor.ID, err)
		}
	}
	return nil
}

func validateAnalysis(analysis Analysis) error {
	if !analysis.RequireComplete {
		return fmt.Errorf("alpha requires complete analysis")
	}
	if analysis.NLP != "builtin-en" {
		return fmt.Errorf("unknown NLP backend %q", analysis.NLP)
	}
	limits := []int{
		analysis.MaxFileBytes,
		analysis.MaxTotalBytes,
		analysis.MaxBlocks,
		analysis.MaxTokens,
		analysis.MaxFindings,
		analysis.MaxCandidates,
	}
	for _, limit := range limits {
		if limit <= 0 || limit > 1<<30 {
			return fmt.Errorf("analysis limits must be positive and at most 1 GiB/count")
		}
	}
	return nil
}

func validateGate(gate Gate) error {
	if !gate.FailOnIncomplete {
		return fmt.Errorf("alpha requires complete analysis")
	}
	for _, threshold := range []Threshold{gate.Sentence, gate.Paragraph} {
		if threshold.FailAt < 1 || threshold.FailAt > 100 || threshold.MinWords < 0 {
			return fmt.Errorf("invalid score threshold")
		}
	}
	return nil
}

func validateSettings(settings rule.Settings, accepted []string) error {
	if !slices.Contains([]string{"note", "warning", "error"}, settings.Severity) {
		return fmt.Errorf("invalid severity %q", settings.Severity)
	}
	if settings.Gate != "none" && settings.Gate != "forbid" {
		return fmt.Errorf("invalid gate %q", settings.Gate)
	}
	if !inRange(settings.Score.Weight, 0, 100) || !inRange(settings.Score.Cap, 0, 100) {
		return fmt.Errorf("score weight and cap must be in [0,100]")
	}
	return validateParameters(settings.Parameters, accepted)
}

func inRange(value, low, high int) bool { return value >= low && value <= high }

func validateParameters(p rule.Parameters, accepted []string) error {
	for _, position := range p.Positions {
		if !slices.Contains([]string{"any", "sentence-start", "document-start", "document-end"}, position) {
			return fmt.Errorf("invalid phrase position %q", position)
		}
	}
	valid := parameterValidity(p)
	for _, name := range accepted {
		if ok, exists := valid[name]; exists && !ok {
			return fmt.Errorf("invalid %s parameter", name)
		}
	}
	return validatePhrases(p.Phrases)
}

func parameterValidity(p rule.Parameters) map[string]bool {
	return map[string]bool{
		"min_words": inRange(p.MinWords, 0, 10000), "onset": p.Onset >= 0, "saturation": p.Saturation > p.Onset,
		"allowed_occurrences": p.AllowedOccurrences >= 0, "saturation_occurrences": p.SaturationOccurrences > p.AllowedOccurrences,
		"window_sentences": inRange(p.WindowSentences, 1, 1000), "opener_words": inRange(p.OpenerWords, 2, 5),
		"similarity": !math.IsNaN(p.Similarity) && p.Similarity > 0 && p.Similarity <= 1, "window": p.Window == "document",
	}
}

func validatePhrases(phrases []string) error {
	if len(phrases) > 1000 {
		return fmt.Errorf("too many phrases")
	}
	for _, phrase := range phrases {
		if len(phrase) == 0 || len(phrase) > 1000 {
			return fmt.Errorf("phrases must contain 1 to 1000 bytes")
		}
	}
	return nil
}
