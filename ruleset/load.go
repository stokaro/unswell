// Package ruleset compiles bounded, declarative editorial rules from YAML bytes.
// Compiled rules use the normal engine contract and never load files or execute
// code. A Set is immutable and can be shared by concurrent engines.
package ruleset

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/rule"
)

// Set contains validated rules and the identity of their complete definition.
type Set struct {
	origin rule.Origin
	rules  []rule.Rule
}

// Origin returns the ruleset's content identity and declared provenance.
func (s *Set) Origin() rule.Origin { return s.origin }

// Rules returns an independent registry slice of immutable rule implementations.
func (s *Set) Rules() []rule.Rule { return slices.Clone(s.rules) }

type definition struct {
	Version    int        `yaml:"version"`
	Namespace  string     `yaml:"namespace"`
	Release    string     `yaml:"release"`
	License    string     `yaml:"license"`
	Provenance string     `yaml:"provenance"`
	Rules      []ruleSpec `yaml:"rules"`
}

type ruleSpec struct {
	Version  int              `yaml:"version"`
	ID       string           `yaml:"id"`
	Summary  string           `yaml:"summary"`
	Message  string           `yaml:"message"`
	Severity string           `yaml:"severity"`
	Gate     string           `yaml:"gate"`
	Scope    string           `yaml:"scope"`
	Group    string           `yaml:"group"`
	Score    rule.Score       `yaml:"score"`
	Requires []nlp.Capability `yaml:"requires"`
	Contexts []string         `yaml:"contexts"`
	Match    yaml.Node        `yaml:"match"`
	Except   []yaml.Node      `yaml:"except"`
	Examples struct {
		Fail []exampleSpec `yaml:"fail"`
		Pass []exampleSpec `yaml:"pass"`
	} `yaml:"examples"`
}

type exampleSpec struct {
	Text   string          `yaml:"text"`
	Format document.Format `yaml:"format"`
}

// UnmarshalYAML accepts either plain prose or a typed source example.
func (e *exampleSpec) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.Tag == "!!str" {
		e.Text, e.Format = node.Value, document.Plain
		return nil
	}
	type fields exampleSpec
	return decodeNode(*node, (*fields)(e))
}

// Load validates one version-1 ruleset, including every matcher and example.
// Examples are parsed here and executed by the CLI's rules test command.
// Definitions are limited to 1 MiB, 100 rules, and 512 matcher nodes per rule.
func Load(data []byte) (*Set, error) {
	if len(data) == 0 || len(data) > 1<<20 {
		return nil, fmt.Errorf("ruleset must contain 1 byte to 1 MiB")
	}
	var tree yaml.Node
	if err := decode(data, &tree); err != nil {
		return nil, err
	}
	if err := validateTree(&tree, 0); err != nil {
		return nil, err
	}
	var input definition
	if err := decode(data, &input); err != nil {
		return nil, err
	}
	if err := validateDefinition(input); err != nil {
		return nil, err
	}
	canonical, err := yaml.Marshal(tree.Content[0])
	if err != nil {
		return nil, err
	}
	set := &Set{origin: rule.Origin{Namespace: input.Namespace, Version: input.Release,
		License: input.License, Provenance: input.Provenance, Hash: fmt.Sprintf("%x", sha256.Sum256(canonical))}}
	if err := set.compileRules(input.Rules); err != nil {
		return nil, err
	}
	return set, nil
}

func (s *Set) compileRules(specs []ruleSpec) error {
	ids := make(map[string]bool)
	for _, spec := range specs {
		if ids[spec.ID] {
			return fmt.Errorf("duplicate rule ID %q", spec.ID)
		}
		ids[spec.ID] = true
		compiled, err := compileRule(spec, s.origin)
		if err != nil {
			return fmt.Errorf("%s: %w", spec.ID, err)
		}
		s.rules = append(s.rules, compiled)
	}
	return nil
}

func validateDefinition(input definition) error {
	if input.Version != 1 || !validName(input.Namespace) || strings.Contains(input.Namespace, ".") {
		return fmt.Errorf("ruleset requires version 1 and a namespace such as company")
	}
	for _, value := range []string{input.Release, input.License, input.Provenance} {
		if strings.TrimSpace(value) == "" || len(value) > 1000 {
			return fmt.Errorf("ruleset release, license, and provenance require 1 to 1000 bytes")
		}
	}
	if len(input.Rules) == 0 || len(input.Rules) > 100 {
		return fmt.Errorf("ruleset requires 1 to 100 rules")
	}
	return nil
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z][a-z0-9-]*)*$`)

func validName(name string) bool { return len(name) <= 100 && namePattern.MatchString(name) }

func decode(data []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("ruleset YAML: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return fmt.Errorf("ruleset must contain exactly one YAML document")
	}
	return nil
}

func validateTree(node *yaml.Node, depth int) error {
	if depth > 48 || node.Kind == yaml.AliasNode || node.Tag == "!!null" || node.Tag == "!!merge" {
		return fmt.Errorf("ruleset YAML rejects aliases, nulls, merges, and nesting beyond 48")
	}
	for _, child := range node.Content {
		if err := validateTree(child, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func decodeNode(node yaml.Node, target any) error {
	data, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	return decode(data, target)
}
