package config

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode"

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/internal/pathglob"
	"github.com/stokaro/unswell/rule"
)

func validate(policy Policy, catalog []rule.Descriptor) error {
	if err := validateFiles(policy.Files); err != nil {
		return err
	}
	if err := extract.ValidatePolicy(policy.Extraction); err != nil {
		return err
	}
	if err := validateAnalysis(policy.Analysis); err != nil {
		return err
	}
	if err := validateGate(policy.Gate); err != nil {
		return err
	}
	if err := validateProbabilityGate(policy); err != nil {
		return err
	}
	for _, descriptor := range catalog {
		if err := validateSettings(policy.Rules[descriptor.ID], descriptor.Parameters); err != nil {
			return fmt.Errorf("%s: %w", descriptor.ID, err)
		}
	}
	return nil
}

func validateFiles(files Files) error {
	for _, patterns := range [][]string{files.Include, files.Exclude} {
		if len(patterns) > 1000 {
			return fmt.Errorf("file selection exceeds 1000 patterns")
		}
		for _, pattern := range patterns {
			if err := relativeGlob(pattern); err != nil {
				return err
			}
		}
		if _, err := pathglob.Compile(patterns); err != nil {
			return err
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
	if gate.Mode != "all" && gate.Mode != "new" {
		return fmt.Errorf("gate.mode must be all or new")
	}
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

// validateProbabilityGate keeps a probability gate tied to an accepted model.
// An experimental pack may report estimates; it cannot decide a build.
func validateProbabilityGate(policy Policy) error {
	gate := policy.Gate.Probability
	if gate == nil {
		return nil
	}
	if gate.FailAt <= 0 || gate.FailAt > 1 {
		return fmt.Errorf("gate.probability.fail_at must be greater than 0 and at most 1")
	}
	if policy.Calibration == nil || policy.Calibration.Model != "pack" {
		return fmt.Errorf("gate.probability requires calibration.model: pack")
	}
	if policy.Calibration.AcceptExperimental {
		return fmt.Errorf("gate.probability requires an accepted pack; calibration.accept_experimental cannot gate a build")
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
	if err := validateWordLists(p.Verbs, p.Nouns); err != nil {
		return err
	}
	return validatePhrases(p.Phrases)
}

func parameterValidity(p rule.Parameters) map[string]bool {
	return map[string]bool{
		"min_words": inRange(p.MinWords, 0, 10000), "onset": p.Onset >= 0, "saturation": p.Saturation > p.Onset,
		"allowed_occurrences": p.AllowedOccurrences >= 0, "saturation_occurrences": p.SaturationOccurrences > p.AllowedOccurrences,
		"window_sentences": inRange(p.WindowSentences, 1, 1000), "opener_words": inRange(p.OpenerWords, 2, 5),
		"max_answer_words": inRange(p.MaxAnswerWords, 1, 100),
		"min_ngram_words":  inRange(p.MinNgramWords, 3, 8) && p.MinNgramWords <= p.MaxNgramWords,
		"max_ngram_words":  inRange(p.MaxNgramWords, 3, 8) && p.MaxNgramWords >= p.MinNgramWords,
		"window_blocks":    inRange(p.WindowBlocks, 1, 128),
		"min_sentences":    inRange(p.MinSentences, 1, 1000), "sentence_words": inRange(p.SentenceWords, 1, 10000),
		"min_long_sentences": inRange(p.MinLongSentences, 1, 1000),
		"allowed_depth":      inRange(p.AllowedDepth, 0, 16), "saturation_depth": inRange(p.SaturationDepth, p.AllowedDepth+1, 32),
		"max_item_words": inRange(p.MaxItemWords, 1, 1000), "max_list_items": inRange(p.MaxListItems, 1, 100),
		"similarity": !math.IsNaN(p.Similarity) && p.Similarity > 0 && p.Similarity <= 1, "window": p.Window == "document",
	}
}

func validateWordLists(lists ...[]string) error {
	for _, words := range lists {
		if err := validatePhrases(words); err != nil {
			return err
		}
		for _, word := range words {
			if strings.ContainsFunc(word, func(r rune) bool { return !unicode.IsLetter(r) }) {
				return fmt.Errorf("verb and noun dictionaries require single words containing only letters")
			}
		}
	}
	return nil
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
