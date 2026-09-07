// Package config compiles strict, offline YAML policy against an explicit catalog.
package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/rule"
)

// Threshold gates eligible units at or above FailAt index points.
type Threshold struct {
	FailAt   int `json:"fail_at"   yaml:"fail_at"`
	MinWords int `json:"min_words" yaml:"min_words"`
}

// Gate is independent of reporter filtering and severity.
type Gate struct {
	Sentence         Threshold `json:"sentence_score"     yaml:"sentence_score"`
	Paragraph        Threshold `json:"paragraph_score"    yaml:"paragraph_score"`
	FailOnEmpty      bool      `json:"fail_on_empty"      yaml:"fail_on_empty"`
	FailOnIncomplete bool      `json:"fail_on_incomplete" yaml:"fail_on_incomplete"`
}

// Analysis bounds the amount of work performed on untrusted documents.
type Analysis struct {
	NLP             string `json:"nlp"              yaml:"nlp"`
	RequireComplete bool   `json:"require_complete" yaml:"require_complete"`
	MaxFileBytes    int    `json:"max_file_bytes"   yaml:"max_file_bytes"`
	MaxTotalBytes   int    `json:"max_total_bytes"  yaml:"max_total_bytes"`
	MaxBlocks       int    `json:"max_blocks"       yaml:"max_blocks"`
	MaxTokens       int    `json:"max_tokens"       yaml:"max_tokens"`
	MaxFindings     int    `json:"max_findings"     yaml:"max_findings"`
	MaxCandidates   int    `json:"max_candidates"   yaml:"max_candidates"`
	IncludeQuotes   bool   `json:"include_quotes"   yaml:"include_quotes"`
}

// Files controls recursive discovery in the CLI. Library calls use explicit sources.
type Files struct {
	Include []string `json:"include" yaml:"include"`
	Exclude []string `json:"exclude" yaml:"exclude"`
}

// Policy is an effective policy with provenance and a canonical content hash.
type Policy struct {
	Version   int                      `json:"version"`
	Profile   string                   `json:"profile"`
	Language  string                   `json:"language"`
	Rules     map[string]rule.Settings `json:"rules"`
	Gate      Gate                     `json:"gate"`
	Analysis  Analysis                 `json:"analysis"`
	Files     Files                    `json:"files"`
	GroupCaps map[string]int           `json:"group_caps"`
	Origins   map[string]string        `json:"origins"`
	Hash      string                   `json:"hash"`
}

type input struct {
	Version     int                  `yaml:"version"`
	Extends     []string             `yaml:"extends"`
	Language    string               `yaml:"language"`
	Rules       map[string]yaml.Node `yaml:"rules"`
	Gate        yaml.Node            `yaml:"gate"`
	Analysis    yaml.Node            `yaml:"analysis"`
	Files       yaml.Node            `yaml:"files"`
	Calibration struct {
		Model          string `yaml:"model"`
		OnIncompatible string `yaml:"on_incompatible"`
	} `yaml:"calibration"`
}

// Load returns a fully validated policy. Alpha extends accepts versioned builtin
// profiles only; local inheritance and file overrides are reserved for stage 2.
func Load(data []byte, catalog []rule.Descriptor) (Policy, error) {
	raw, err := loadInput(data)
	if err != nil {
		return Policy{}, err
	}
	profile, err := resolveProfile(raw.Extends)
	if err != nil {
		return Policy{}, err
	}
	policy, err := defaults(profile, catalog)
	if err != nil {
		return Policy{}, err
	}
	if err := applyNodes(raw, &policy, catalog); err != nil {
		return Policy{}, err
	}
	if err := validate(policy, catalog); err != nil {
		return Policy{}, err
	}
	canonical, err := json.Marshal(policy)
	if err != nil {
		return Policy{}, err
	}
	policy.Hash = fmt.Sprintf("%x", sha256.Sum256(canonical))
	return policy, nil
}

func loadInput(data []byte) (input, error) {
	var raw input
	if len(data) > 1<<20 {
		return raw, fmt.Errorf("configuration exceeds 1 MiB")
	}
	if len(bytes.TrimSpace(data)) == 0 {
		raw.Version = 1
	} else if err := decode(data, &raw); err != nil {
		return raw, err
	}
	return raw, validateInput(raw)
}

func validateInput(raw input) error {
	if raw.Version != 1 {
		return fmt.Errorf("unsupported config version %d", raw.Version)
	}
	if raw.Language != "" && raw.Language != "en" {
		return fmt.Errorf("unsupported language %q", raw.Language)
	}
	if raw.Calibration.Model != "" && raw.Calibration.Model != "none" {
		return fmt.Errorf("calibration model %q is unavailable in this alpha", raw.Calibration.Model)
	}
	if raw.Calibration.OnIncompatible != "" && raw.Calibration.OnIncompatible != "unavailable" {
		return fmt.Errorf("invalid calibration unavailable policy")
	}
	return nil
}

func decode(data []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("configuration: %w", err)
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err != nil {
			return fmt.Errorf("configuration: %w", err)
		}
		return fmt.Errorf("configuration must contain one YAML document")
	}
	return nil
}

func resolveProfile(extends []string) (string, error) {
	profile := "technical"
	if len(extends) > 1 {
		return "", fmt.Errorf("alpha configuration accepts one builtin profile")
	}
	if len(extends) == 1 {
		if !strings.HasPrefix(extends[0], "builtin:") {
			return "", fmt.Errorf("alpha extends requires a builtin profile")
		}
		profile = strings.TrimPrefix(extends[0], "builtin:")
		profile = strings.TrimSuffix(profile, "-v1")
	}
	if !slices.Contains([]string{"technical", "strict", "minimal", "business", "reference", "custom"}, profile) {
		return "", fmt.Errorf("unknown profile %q", profile)
	}
	return profile, nil
}

func defaults(profile string, catalog []rule.Descriptor) (Policy, error) {
	policy := Policy{
		Version:  1,
		Profile:  profile + "-v1",
		Language: "en",
		Rules:    make(map[string]rule.Settings),
		Origins:  make(map[string]string),
		Gate: Gate{
			Sentence:         Threshold{FailAt: 80, MinWords: 12},
			Paragraph:        Threshold{FailAt: 65, MinWords: 30},
			FailOnEmpty:      true,
			FailOnIncomplete: true,
		},
		Analysis: Analysis{
			NLP:             "builtin-en",
			RequireComplete: true,
			MaxFileBytes:    2 << 20,
			MaxTotalBytes:   100 << 20,
			MaxBlocks:       10000,
			MaxTokens:       200000,
			MaxFindings:     10000,
			MaxCandidates:   100000,
		},
		Files: Files{
			Include: []string{"**/*.md", "**/*.txt", "**/*.go"},
			Exclude: []string{".git/**", "vendor/**", "testdata/**", "artifacts/**", "dist/**", "rules/**", "**/*.generated.go"},
		},
		GroupCaps: map[string]int{
			"scaffolding":         40,
			"inflation":           35,
			"rhetorical-patterns": 35,
			"syntax-load":         40,
			"repetition":          45,
			"readability":         25,
		},
	}
	for _, descriptor := range catalog {
		if _, exists := policy.Rules[descriptor.ID]; exists {
			return Policy{}, fmt.Errorf("duplicate rule ID %q", descriptor.ID)
		}
		settings := descriptor.Defaults
		applyProfile(profile, descriptor.ID, &settings)
		policy.Rules[descriptor.ID] = settings
		policy.Origins["rules."+descriptor.ID] = "builtin:" + profile + "-v1"
	}
	if profile == "strict" {
		policy.Gate.Paragraph.FailAt = 50
		policy.Gate.Sentence.FailAt = 65
	}
	return policy, nil
}

func applyProfile(profile, id string, settings *rule.Settings) {
	switch profile {
	case "strict":
		if strictForbid(id) {
			settings.Gate = "forbid"
			settings.Severity = "error"
		}
	case "minimal":
		settings.Enabled = settings.Gate == "forbid" || id == "repetition.exact-sentence"
	case "business":
		if slices.Contains([]string{"scaffold.follow-up-offer", "scaffold.chat-preamble"}, id) {
			settings.Enabled = false
		}
	case "reference":
		if referenceDisabled(id) {
			settings.Enabled = false
		}
	case "custom":
		settings.Enabled = false
	}
}

func strictForbid(id string) bool {
	return strings.HasPrefix(id, "scaffold.") ||
		slices.Contains([]string{"filler.announced-importance", "filler.modern-world-opening"}, id)
}

func referenceDisabled(id string) bool {
	return strings.HasPrefix(id, "repetition.") && id != "repetition.exact-sentence"
}

func applyNodes(raw input, policy *Policy, catalog []rule.Descriptor) error {
	ids := make([]string, 0, len(raw.Rules))
	for id := range raw.Rules {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		node := raw.Rules[id]
		settings, ok := policy.Rules[id]
		if !ok {
			return fmt.Errorf("unknown rule ID %q", id)
		}
		if err := validateRuleNode(id, node, catalog); err != nil {
			return err
		}
		if err := mergeNode(node, &settings); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
		policy.Rules[id] = settings
		policy.Origins["rules."+id] = "project configuration"
	}
	return applyPolicyNodes(raw, policy)
}

func validateRuleNode(id string, node yaml.Node, catalog []rule.Descriptor) error {
	for _, descriptor := range catalog {
		if descriptor.ID != id {
			continue
		}
		if err := validateParameterNames(node, descriptor.Parameters); err != nil {
			return fmt.Errorf("%s: %w", id, err)
		}
	}
	return nil
}

func applyPolicyNodes(raw input, policy *Policy) error {
	for _, item := range []struct {
		node   yaml.Node
		target any
	}{{raw.Gate, &policy.Gate}, {raw.Analysis, &policy.Analysis}, {raw.Files, &policy.Files}} {
		if item.node.Kind != 0 {
			if err := mergeNode(item.node, item.target); err != nil {
				return err
			}
		}
	}
	return nil
}

func mergeNode(node yaml.Node, target any) error {
	if err := rejectAliasesAndNull(&node, 0); err != nil {
		return err
	}
	data, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	return decode(data, target)
}

func rejectAliasesAndNull(node *yaml.Node, depth int) error {
	if depth > 32 {
		return fmt.Errorf("YAML nesting exceeds 32")
	}
	if node.Kind == yaml.AliasNode || node.Tag == "!!null" {
		return fmt.Errorf("YAML aliases and null overrides are not supported")
	}
	for _, child := range node.Content {
		if err := rejectAliasesAndNull(child, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func validateParameterNames(node yaml.Node, allowed []string) error {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value != "parameters" {
			continue
		}
		params := node.Content[i+1]
		for j := 0; j+1 < len(params.Content); j += 2 {
			if !slices.Contains(allowed, params.Content[j].Value) {
				return fmt.Errorf("unknown parameter %q", params.Content[j].Value)
			}
		}
	}
	return nil
}
