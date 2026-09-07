package ruleset

import (
	"fmt"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

type compiler struct {
	nodes    int
	requires []nlp.Capability
}

type matcherBuilder func(*compiler, yaml.Node, int) (matcher, error)

func (c *compiler) compile(node yaml.Node, depth int) (matcher, error) {
	c.nodes++
	if depth > 16 || c.nodes > 512 {
		return nil, fmt.Errorf("matcher exceeds depth 16 or 512 nodes")
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("matcher must be a mapping")
	}
	kind := ""
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == "type" {
			kind = node.Content[i+1].Value
		}
	}
	builders := map[string]matcherBuilder{
		"phrase": compilePhrase, "token-set": compilePhrase, "regex": compileRegex,
		"sequence": compileSequence, "all": compileBoolean, "any": compileBoolean,
		"not": compileNot, "count": compileCounter, "density": compileCounter, "feature": compileFeature,
	}
	build, ok := builders[kind]
	if !ok {
		return nil, fmt.Errorf("unknown matcher type %q", kind)
	}
	return build(c, node, depth)
}

func (c *compiler) require(capability nlp.Capability) {
	if !slices.Contains(c.requires, capability) {
		c.requires = append(c.requires, capability)
	}
}

func compileRule(spec ruleSpec, origin rule.Origin) (*compiledRule, error) {
	if err := validateRuleSpec(&spec, origin.Namespace); err != nil {
		return nil, err
	}
	c := &compiler{requires: []nlp.Capability{nlp.Tokens, nlp.Sentences}}
	compiled, err := c.compile(spec.Match, 0)
	if err != nil {
		return nil, err
	}
	r := &compiledRule{match: compiled, message: spec.Message}
	for _, node := range spec.Except {
		exception, err := c.compile(node, 0)
		if err != nil {
			return nil, fmt.Errorf("exception: %w", err)
		}
		r.except = append(r.except, exception)
	}
	if err := validateCapabilities(spec.Requires, c.requires); err != nil {
		return nil, err
	}
	r.descriptor = rule.Descriptor{
		ID: spec.ID, Version: fmt.Sprint(spec.Version), Summary: spec.Summary, Description: spec.Message,
		Scope: spec.Scope, Group: spec.Group, Status: "custom", Requires: c.requires, Contexts: spec.Contexts,
		Origin: &origin, Limitations: "User-defined editorial policy; no measured precision or authorship claim.",
		Defaults: rule.Settings{Enabled: true, Severity: spec.Severity, Gate: spec.Gate, Score: spec.Score},
	}
	for _, example := range spec.Examples.Fail {
		r.descriptor.Examples = append(r.descriptor.Examples, rule.Example{Text: example.Text, Format: example.Format, Match: true})
	}
	for _, example := range spec.Examples.Pass {
		r.descriptor.Examples = append(r.descriptor.Examples, rule.Example{Text: example.Text, Format: example.Format})
	}
	slices.Sort(r.descriptor.Requires)
	r.heuristic = slices.Contains(c.requires, nlp.POS) || slices.Contains(c.requires, nlp.Chunks)
	return r, nil
}

func validateRuleSpec(spec *ruleSpec, namespace string) error {
	if !validName(spec.ID) || !strings.HasPrefix(spec.ID, namespace+".") {
		return fmt.Errorf("rule ID must belong to namespace %s", namespace)
	}
	if err := validateRuleVersion(spec); err != nil {
		return err
	}
	if err := validateRulePolicy(spec); err != nil {
		return err
	}
	if err := validateContexts(spec); err != nil {
		return err
	}
	if len(spec.Except) > 32 || len(spec.Examples.Fail) == 0 || len(spec.Examples.Pass) == 0 {
		return fmt.Errorf("rule needs pass and fail examples and at most 32 exceptions")
	}
	return validateExamples(spec)
}

func validateRuleVersion(spec *ruleSpec) error {
	if spec.Version == 0 {
		spec.Version = 1
	}
	if spec.Version < 1 || spec.Version > 1000000 {
		return fmt.Errorf("rule version must be in [1,1000000]")
	}
	return nil
}

func validateContexts(spec *ruleSpec) error {
	if len(spec.Contexts) == 0 {
		spec.Contexts = extract.DefaultContexts()
	}
	for _, kind := range spec.Contexts {
		if !slices.Contains(extract.DefaultContexts(), kind) {
			return fmt.Errorf("unknown rule context %q", kind)
		}
	}
	slices.Sort(spec.Contexts)
	spec.Contexts = slices.Compact(spec.Contexts)
	return nil
}

func validateRulePolicy(spec *ruleSpec) error {
	if spec.Severity == "" {
		spec.Severity = "warning"
	}
	if spec.Gate == "" {
		spec.Gate = "none"
	}
	if spec.Group == "" {
		spec.Group = "custom"
	}
	return validatePolicyValues(spec)
}

func validatePolicyValues(spec *ruleSpec) error {
	if !slices.Contains([]string{"sentence", "paragraph", "document"}, spec.Scope) {
		return fmt.Errorf("scope must be sentence, paragraph, or document")
	}
	if !slices.Contains([]string{"note", "warning", "error"}, spec.Severity) ||
		!slices.Contains([]string{"none", "forbid"}, spec.Gate) {
		return fmt.Errorf("invalid rule severity or gate")
	}
	if spec.Score.Weight < 0 || spec.Score.Weight > 100 || spec.Score.Cap < 0 || spec.Score.Cap > 100 {
		return fmt.Errorf("score weight and cap must be in [0,100]")
	}
	return validateDescriptions(spec)
}

func validateDescriptions(spec *ruleSpec) error {
	for _, text := range []string{spec.Summary, spec.Message, spec.Group} {
		if strings.TrimSpace(text) == "" || len(text) > 2000 {
			return fmt.Errorf("summary, message, and group require 1 to 2000 bytes")
		}
	}
	return nil
}

func validateExamples(spec *ruleSpec) error {
	if len(spec.Examples.Fail)+len(spec.Examples.Pass) > 100 {
		return fmt.Errorf("rule exceeds 100 examples")
	}
	for _, examples := range [][]exampleSpec{spec.Examples.Fail, spec.Examples.Pass} {
		for _, example := range examples {
			if strings.TrimSpace(example.Text) == "" || len(example.Text) > 10000 {
				return fmt.Errorf("examples require 1 to 10000 bytes")
			}
			if example.Format != "" && !slices.Contains(document.Formats(), example.Format) {
				return fmt.Errorf("unknown example format %q", example.Format)
			}
		}
	}
	return nil
}

func validateCapabilities(declared, inferred []nlp.Capability) error {
	if len(declared) == 0 {
		return nil
	}
	for _, capability := range declared {
		if !slices.Contains(inferred, capability) {
			return fmt.Errorf("unused or unsupported declared capability %q", capability)
		}
	}
	for _, capability := range inferred {
		if !slices.Contains(declared, capability) {
			return fmt.Errorf("matcher requires undeclared capability %q", capability)
		}
	}
	return nil
}
