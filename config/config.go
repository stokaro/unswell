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

	"github.com/stokaro/unswell/extract"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
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
	Version    int                      `json:"version"`
	Profile    string                   `json:"profile"`
	Language   string                   `json:"language"`
	Rules      map[string]rule.Settings `json:"rules"`
	Gate       Gate                     `json:"gate"`
	Analysis   Analysis                 `json:"analysis"`
	Files      Files                    `json:"files"`
	Extraction extract.Policy           `json:"extraction"`
	GroupCaps  map[string]int           `json:"group_caps"`
	Origins    map[string]string        `json:"origins"`
	Hash       string                   `json:"hash"`
	RuleSets   []rule.Origin            `json:"rule_sets,omitempty"`
}

type input struct {
	Version     int                  `yaml:"version"`
	Extends     []string             `yaml:"extends"`
	Language    string               `yaml:"language"`
	Rules       map[string]yaml.Node `yaml:"rules"`
	Gate        yaml.Node            `yaml:"gate"`
	Analysis    yaml.Node            `yaml:"analysis"`
	Files       yaml.Node            `yaml:"files"`
	Extraction  yaml.Node            `yaml:"extraction"`
	RuleSets    []yaml.Node          `yaml:"rule_sets"`
	Calibration struct {
		Model          string `yaml:"model"`
		OnIncompatible string `yaml:"on_incompatible"`
	} `yaml:"calibration"`
}

// Load returns a fully validated policy. Alpha extends accepts versioned builtin
// profiles only; local inheritance and file overrides are reserved for stage 2.
// Use Compile when the caller also needs implementations from inline rule_sets.
func Load(data []byte, catalog []rule.Descriptor) (Policy, error) {
	policy, _, err := Compile(data, catalog)
	return policy, err
}

// Compile returns the effective policy and additional rules declared in rule_sets.
// The caller supplies the existing catalog; duplicate IDs always fail. All input
// and ruleset definitions are bytes in memory, with no implicit file loading.
func Compile(data []byte, catalog []rule.Descriptor) (Policy, []rule.Rule, error) {
	raw, err := loadInput(data)
	if err != nil {
		return Policy{}, nil, err
	}
	additional, err := compileRuleSets(raw.RuleSets)
	if err != nil {
		return Policy{}, nil, err
	}
	catalog = slices.Clone(catalog)
	for _, implementation := range additional {
		catalog = append(catalog, implementation.Descriptor())
	}
	if len(catalog) > 1000 {
		return Policy{}, nil, fmt.Errorf("catalog exceeds 1000 rules")
	}
	policy, err := compilePolicy(raw, catalog)
	return policy, additional, err
}

func compilePolicy(raw input, catalog []rule.Descriptor) (Policy, error) {
	profile, err := resolveProfile(raw.Extends)
	if err != nil {
		return Policy{}, err
	}
	policy, err := defaults(profile, catalog)
	if err != nil {
		return Policy{}, err
	}
	policy.RuleSets = catalogOrigins(catalog)
	if err := applyNodes(raw, &policy, catalog); err != nil {
		return Policy{}, err
	}
	if err := validate(policy, catalog); err != nil {
		return Policy{}, err
	}
	slices.Sort(policy.Extraction.Contexts)
	for format, override := range policy.Extraction.Languages {
		slices.Sort(override.Contexts)
		policy.Extraction.Languages[format] = override
	}
	canonical, err := json.Marshal(policy)
	if err != nil {
		return Policy{}, err
	}
	policy.Hash = fmt.Sprintf("%x", sha256.Sum256(canonical))
	return policy, nil
}

func compileRuleSets(nodes []yaml.Node) ([]rule.Rule, error) {
	if len(nodes) > 10 {
		return nil, fmt.Errorf("configuration exceeds 10 rule_sets")
	}
	var implementations []rule.Rule
	for _, node := range nodes {
		data, err := yaml.Marshal(node)
		if err != nil {
			return nil, err
		}
		set, err := ruleset.Load(data)
		if err != nil {
			return nil, err
		}
		implementations = append(implementations, set.Rules()...)
	}
	return implementations, nil
}

func catalogOrigins(catalog []rule.Descriptor) []rule.Origin {
	var origins []rule.Origin
	for _, descriptor := range catalog {
		if descriptor.Origin != nil && !slices.Contains(origins, *descriptor.Origin) {
			origins = append(origins, *descriptor.Origin)
		}
	}
	slices.SortFunc(origins, func(a, b rule.Origin) int {
		return strings.Compare(a.Namespace+"/"+a.Version+"/"+a.Hash, b.Namespace+"/"+b.Version+"/"+b.Hash)
	})
	return origins
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
		Version:    1,
		Profile:    profile + "-v1",
		Language:   "en",
		Extraction: extract.Policy{Contexts: extract.DefaultContexts()},
		Rules:      make(map[string]rule.Settings),
		Origins:    make(map[string]string),
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
			Include: defaultIncludes(),
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
		if descriptor.Origin != nil {
			policy.Origins["rules."+descriptor.ID] = "ruleset:" + descriptor.Origin.Namespace + "@" + descriptor.Origin.Version
		}
	}
	if profile == "strict" {
		policy.Gate.Paragraph.FailAt = 50
		policy.Gate.Sentence.FailAt = 65
	}
	return policy, nil
}

func defaultIncludes() []string {
	return []string{
		"**/*.md", "**/*.markdown", "**/*.txt", "**/*.go", "**/*.js", "**/*.jsx", "**/*.mjs", "**/*.cjs",
		"**/*.ts", "**/*.mts", "**/*.cts", "**/*.tsx", "**/*.py", "**/*.pyi", "**/*.rs", "**/*.java",
		"**/*.c", "**/*.h", "**/*.C", "**/*.cc", "**/*.cpp", "**/*.cxx", "**/*.hpp", "**/*.hh", "**/*.hxx",
		"**/*.cs", "**/*.csx", "**/*.yaml", "**/*.yml",
		"**/*.sh", "**/*.bash", "**/*.zsh", "**/*.fish", "**/*.ps1", "**/*.psm1", "**/*.psd1",
		"**/.bashrc", "**/.bash_profile", "**/.bash_login", "**/.bash_logout", "**/.profile",
		"**/.zshrc", "**/.zprofile", "**/.zshenv", "**/.zlogin", "**/.zlogout",
	}
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
	}{{raw.Gate, &policy.Gate}, {raw.Analysis, &policy.Analysis}, {raw.Files, &policy.Files}, {raw.Extraction, &policy.Extraction}} {
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
