// Package llmdet compares bounded numerical components with published LLMDet code.
// It does not provide a qualified detector or tokenize source text.
package llmdet

// Proxy accumulation follows TrustedLLM/LLMDet llmdet/detector.py at
// 5d038354006ca0c8e6aa0dadb75e8840accb51a8, function perplexity.
// Copyright (c) 2023 Kangxi Wu, Liang Pang, TryMore Group.
// Licensed under the MIT License; see licenses/LLMDet_LICENSE.
// Unswell adds validation, bounded immutable rows, cancellation, and coverage.

import (
	"context"
	"fmt"
	"math"
)

const (
	// ProxyVersion identifies the explicit token and probability-row format.
	ProxyVersion = "unswell-llmdet-proxy-v1"
	// MaxTokens bounds one numerical proxy input.
	MaxTokens = 100_000
	maxRows   = 100_000
	maxTopK   = 4096
	maxValues = 1_000_000
)

// ProbabilityRow stores an ordered context and its retained continuations.
type ProbabilityRow struct {
	Context       []int     `json:"context"`
	Continuations []int     `json:"continuations"`
	Probabilities []float64 `json:"probabilities"`
}

// ProxySpec describes one model's numerical vocabulary and probability rows.
// VocabSize preserves the reference argument, not an inferred tokenizer size.
type ProxySpec struct {
	Version   string           `json:"version"`
	VocabSize int              `json:"vocab_size"`
	Rows      []ProbabilityRow `json:"rows"`
}

type probabilityRow struct {
	values   map[int]float64
	residual float64
}

// Proxy owns validated rows and can be shared between independent calls.
type Proxy struct {
	vocab int
	rows  map[[4]int]probabilityRow
}

// ProxyResult separates reference arithmetic from observed coverage.
// Value is absent when no position contributes a likelihood. A present value
// is a numerical feature, not a probability or a claim of model applicability.
type ProxyResult struct {
	ReferenceScore float64  `json:"reference_score"`
	Value          *float64 `json:"value"`
	Reason         string   `json:"reason,omitempty"`
	Possible       int      `json:"possible"`
	Matched        int      `json:"matched"`
	Evaluated      int      `json:"evaluated"`
	Orders         [3]int   `json:"orders"` // Matched bigram, trigram, and four-gram contexts.
	Residual       int      `json:"residual"`
	Skipped        int      `json:"skipped"`
	Coverage       float64  `json:"context_coverage"`
	ValueCoverage  float64  `json:"value_coverage"`
}

// NewProxy validates and copies a bounded set of probability rows.
func NewProxy(ctx context.Context, spec ProxySpec) (*Proxy, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !validProxyDimensions(spec) {
		return nil, fmt.Errorf("unsupported proxy version or dimensions")
	}
	p := &Proxy{vocab: spec.VocabSize, rows: make(map[[4]int]probabilityRow, len(spec.Rows))}
	count := 0
	for i, row := range spec.Rows {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		count += len(row.Continuations)
		if count > maxValues {
			return nil, fmt.Errorf("proxy exceeds %d retained values", maxValues)
		}
		key, value, err := validateRow(row, spec.VocabSize)
		if err != nil {
			return nil, fmt.Errorf("proxy row %d: %w", i, err)
		}
		if _, exists := p.rows[key]; exists {
			return nil, fmt.Errorf("duplicate proxy context at row %d", i)
		}
		p.rows[key] = value
	}
	return p, nil
}

func validProxyDimensions(spec ProxySpec) bool {
	return spec.Version == ProxyVersion && spec.VocabSize >= 2 && spec.VocabSize <= maxValues && len(spec.Rows) <= maxRows
}

func validateRow(row ProbabilityRow, vocab int) ([4]int, probabilityRow, error) {
	var key [4]int
	if len(row.Context) < 1 || len(row.Context) > 3 || len(row.Continuations) > min(vocab, maxTopK) ||
		len(row.Continuations) != len(row.Probabilities) {
		return key, probabilityRow{}, fmt.Errorf("invalid context or continuation dimensions")
	}
	if err := validTokens(row.Context, vocab); err != nil {
		return key, probabilityRow{}, err
	}
	if err := validTokens(row.Continuations, vocab); err != nil {
		return key, probabilityRow{}, err
	}
	key[0] = len(row.Context)
	copy(key[1:], row.Context)
	values, sum, err := retainedValues(row)
	if err != nil {
		return key, probabilityRow{}, err
	}
	value := probabilityRow{values: values}
	if sum < 1 && len(values) < vocab {
		value.residual = (1 - sum) / float64(vocab-len(values))
	}
	return key, value, nil
}

func retainedValues(row ProbabilityRow) (map[int]float64, float64, error) {
	values, sum := make(map[int]float64, len(row.Continuations)), 0.0
	for i, token := range row.Continuations {
		probability := row.Probabilities[i]
		if _, exists := values[token]; exists || !finite(probability) || probability < 0 || probability > 1 {
			return nil, 0, fmt.Errorf("duplicate continuation or invalid probability")
		}
		values[token] = probability
		sum += probability
	}
	if sum > 1+1e-12 {
		return nil, 0, fmt.Errorf("probability mass exceeds one")
	}
	return values, sum, nil
}

// Measure preserves the pinned reference's positions, backoff, and denominator.
// Unmatched positions do not contribute; matched zero probabilities are counted
// in the reference denominator but skipped in its log sum.
func (p *Proxy) Measure(ctx context.Context, tokens []int) (ProxyResult, error) {
	var result ProxyResult
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if p == nil || p.vocab < 2 || len(tokens) > MaxTokens {
		return result, fmt.Errorf("missing proxy or excessive token count")
	}
	if err := validTokens(tokens, p.vocab); err != nil {
		return result, err
	}
	result.Possible = max(0, len(tokens)-3)
	sum := 0.0
	for i := 2; i < len(tokens)-1; i++ {
		if err := ctx.Err(); err != nil {
			return ProxyResult{}, err
		}
		row, order, found := p.context(tokens[i-2 : i+1])
		if !found {
			continue
		}
		sum += observeRow(&result, row, order, tokens[i+1])
	}
	result.ReferenceScore = -sum / float64(result.Matched+1)
	finishProxyResult(&result)
	return result, nil
}

func observeRow(result *ProxyResult, row probabilityRow, order, token int) float64 {
	result.Matched++
	result.Orders[order-1]++
	probability, retained := row.values[token]
	if !retained {
		probability = row.residual
		result.Residual++
	}
	if probability > 0 {
		result.Evaluated++
		return math.Log2(probability)
	}
	result.Skipped++
	return 0
}

func (p *Proxy) context(tokens []int) (probabilityRow, int, bool) {
	for size := 3; size > 0; size-- {
		key := [4]int{size}
		copy(key[1:], tokens[3-size:])
		if row, exists := p.rows[key]; exists {
			return row, size, true
		}
	}
	return probabilityRow{}, 0, false
}

func finishProxyResult(result *ProxyResult) {
	if result.Possible > 0 {
		result.Coverage = float64(result.Matched) / float64(result.Possible)
		result.ValueCoverage = float64(result.Evaluated) / float64(result.Possible)
	}
	switch {
	case result.Possible == 0:
		result.Reason = "insufficient_tokens"
	case result.Matched == 0:
		result.Reason = "no_matching_context"
	case result.Evaluated == 0:
		result.Reason = "no_evaluated_probability"
	default:
		value := result.ReferenceScore
		result.Value = &value
	}
}

func validTokens(tokens []int, vocab int) error {
	for _, token := range tokens {
		if token < 0 || token >= vocab {
			return fmt.Errorf("token ID %d outside vocabulary", token)
		}
	}
	return nil
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
