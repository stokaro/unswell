package ruleset

import (
	"fmt"
	"math"
	"slices"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

type booleanSpec struct {
	Type  string      `yaml:"type"`
	Items []yaml.Node `yaml:"items"`
}

type booleanMatcher struct {
	all   bool
	items []matcher
}

func compileBoolean(c *compiler, node yaml.Node, depth int) (matcher, error) {
	var spec booleanSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	if len(spec.Items) < 2 || len(spec.Items) > 32 {
		return nil, fmt.Errorf("all and any require 2 to 32 items")
	}
	m := &booleanMatcher{all: spec.Type == "all"}
	for _, child := range spec.Items {
		item, err := c.compile(child, depth+1)
		if err != nil {
			return nil, err
		}
		m.items = append(m.items, item)
	}
	return m, nil
}

func (m *booleanMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	result := matchResult{ok: m.all}
	for _, item := range m.items {
		if err := e.spend(1); err != nil {
			return matchResult{}, err
		}
		next, err := item.evaluate(e, current)
		if err != nil {
			return matchResult{}, err
		}
		if next.ok != m.all {
			return next, nil
		}
		if err := mergeResult(&result, next); err != nil {
			return matchResult{}, err
		}
	}
	result.occurrences = uniqueOccurrences(result.occurrences)
	return result, nil
}

func mergeResult(result *matchResult, next matchResult) error {
	for _, occurrence := range next.occurrences {
		if err := addOccurrence(result, occurrence); err != nil {
			return err
		}
	}
	result.metrics = append(result.metrics, next.metrics...)
	return nil
}

type unarySpec struct {
	Type  string    `yaml:"type"`
	Match yaml.Node `yaml:"match"`
}

type notMatcher struct{ child matcher }

func compileNot(c *compiler, node yaml.Node, depth int) (matcher, error) {
	var spec unarySpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	child, err := c.compile(spec.Match, depth+1)
	return &notMatcher{child: child}, err
}

func (m *notMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	if err := e.spend(1); err != nil {
		return matchResult{}, err
	}
	result, err := m.child.evaluate(e, current)
	return matchResult{ok: !result.ok}, err
}

type bounds struct {
	Min *float64 `yaml:"min"`
	Max *float64 `yaml:"max"`
}

func (b bounds) validate(integer bool) error {
	if b.Min == nil && b.Max == nil {
		return fmt.Errorf("comparison needs min or max")
	}
	for _, value := range []*float64{b.Min, b.Max} {
		if value == nil {
			continue
		}
		if !validBound(*value, integer) {
			return fmt.Errorf("comparison bounds must be finite nonnegative values up to 1e9; counts require integers")
		}
	}
	if b.Min != nil && b.Max != nil && *b.Min > *b.Max {
		return fmt.Errorf("min must not exceed max")
	}
	return nil
}

func validBound(value float64, integer bool) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0) && value >= 0 && value <= 1e9 && (!integer || value == math.Trunc(value))
}

func (b bounds) accepts(value float64) bool {
	return (b.Min == nil || value >= *b.Min) && (b.Max == nil || value <= *b.Max)
}

func (b bounds) metric(name, unit string, value float64) rule.Metric {
	metric := rule.Metric{Name: name, Unit: unit, Value: value}
	if b.Min != nil {
		metric.Onset = *b.Min
	}
	if b.Max != nil {
		metric.Saturation = *b.Max
	} else {
		metric.Saturation = metric.Onset
	}
	return metric
}

type counterSpec struct {
	Type   string    `yaml:"type"`
	Match  yaml.Node `yaml:"match"`
	Per    string    `yaml:"per"`
	Bounds bounds    `yaml:",inline"`
}

type counterMatcher struct {
	child   matcher
	per     string
	bounds  bounds
	density bool
}

func compileCounter(c *compiler, node yaml.Node, depth int) (matcher, error) {
	var spec counterSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	density := spec.Type == "density"
	if err := spec.Bounds.validate(!density); err != nil {
		return nil, err
	}
	if density && !slices.Contains([]string{"words", "sentences"}, spec.Per) || !density && spec.Per != "" {
		return nil, fmt.Errorf("density requires per: words or sentences; count does not accept per")
	}
	child, err := c.compile(spec.Match, depth+1)
	if err != nil {
		return nil, err
	}
	if !locatesMatches(child) {
		return nil, fmt.Errorf("count and density need a matcher that locates occurrences on every successful branch")
	}
	return &counterMatcher{child: child, per: spec.Per, bounds: spec.Bounds, density: density}, nil
}

func locatesMatches(m matcher) bool {
	switch m := m.(type) {
	case *phraseMatcher, *regexMatcher, *sequenceMatcher:
		return true
	case *booleanMatcher:
		for _, item := range m.items {
			if locatesMatches(item) == m.all {
				return m.all
			}
		}
		return !m.all
	default:
		return false
	}
}

func (m *counterMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	if err := e.spend(1); err != nil {
		return matchResult{}, err
	}
	result, err := m.child.evaluate(e, current)
	if err != nil {
		return matchResult{}, err
	}
	result.occurrences = uniqueOccurrences(result.occurrences)
	value := float64(len(result.occurrences))
	name, units := "match-count", "matches"
	if m.density {
		denominator := words(current)
		if m.per == "sentences" {
			denominator = len(current)
		}
		if denominator == 0 {
			return matchResult{}, nil
		}
		value = 100 * value / float64(denominator)
		name, units = "match-density", "matches/100-"+m.per
	}
	result.ok = m.bounds.accepts(value)
	result.metrics = append(result.metrics, m.bounds.metric(name, units, value))
	if !result.ok {
		result.occurrences = nil
	}
	return result, nil
}

type featureSpec struct {
	Type   string `yaml:"type"`
	Name   string `yaml:"name"`
	Bounds bounds `yaml:",inline"`
}

type featureMatcher struct {
	name   string
	bounds bounds
}

func compileFeature(c *compiler, node yaml.Node, _ int) (matcher, error) {
	var spec featureSpec
	if err := decodeNode(node, &spec); err != nil {
		return nil, err
	}
	if !slices.Contains([]string{"prose.words", "prose.tokens", "prose.sentences", "chunks.np", "chunks.vp", "chunks.pp"}, spec.Name) {
		return nil, fmt.Errorf("unknown shared feature %q", spec.Name)
	}
	if err := spec.Bounds.validate(true); err != nil {
		return nil, err
	}
	if slices.Contains([]string{"chunks.np", "chunks.vp", "chunks.pp"}, spec.Name) {
		c.require(nlp.Chunks)
	}
	return &featureMatcher{name: spec.Name, bounds: spec.Bounds}, nil
}

func (m *featureMatcher) evaluate(e *evaluation, current unit) (matchResult, error) {
	if err := e.spend(1); err != nil {
		return matchResult{}, err
	}
	value, err := sharedFeature(e, current, m.name)
	if err != nil {
		return matchResult{}, err
	}
	return matchResult{ok: m.bounds.accepts(float64(value)),
		metrics: []rule.Metric{m.bounds.metric(m.name, "count", float64(value))}}, nil
}

func sharedFeature(e *evaluation, current unit, name string) (int, error) {
	if name == "prose.words" {
		return words(current), nil
	}
	if name == "prose.sentences" {
		return len(current), nil
	}
	count := 0
	for _, view := range current {
		if err := e.spend(1 + len(view.sentence.Tokens)); err != nil {
			return 0, err
		}
		if name == "prose.tokens" {
			count += visibleTokens(view)
		} else {
			count += matchingChunks(view, name)
		}
	}
	return count, nil
}

func visibleTokens(view sentenceView) int {
	count := 0
	for _, token := range view.sentence.Tokens {
		if !token.Protected {
			count++
		}
	}
	return count
}

func matchingChunks(view sentenceView, name string) int {
	kind := map[string]string{"chunks.np": "NP", "chunks.vp": "VP", "chunks.pp": "PP"}[name]
	count := 0
	for _, chunk := range view.sentence.Chunks {
		if chunk.Kind == kind {
			count++
		}
	}
	return count
}
