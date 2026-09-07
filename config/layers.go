package config

import (
	"fmt"
	"path"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/internal/pathglob"
	"github.com/stokaro/unswell/rule"
)

func compileOverrides(raw []overrideInput, source string) ([]fileLayer, error) {
	if len(raw) > 64 {
		return nil, fmt.Errorf("configuration exceeds 64 file overrides")
	}
	var result []fileLayer
	for i, override := range raw {
		if err := validateOverride(override); err != nil {
			return nil, err
		}
		patterns, err := pathglob.Compile(override.Files)
		if err != nil {
			return nil, err
		}
		data, err := yaml.Marshal(override)
		if err != nil {
			return nil, err
		}
		id := source + "#overrides[" + strconv.Itoa(i) + "]"
		identity, err := sourceIdentity(id, "override", data)
		if err != nil {
			return nil, err
		}
		result = append(result, fileLayer{patterns: patterns,
			identity: OverrideIdentity{ID: id, Files: slices.Clone(override.Files), Hash: identity.Hash},
			raw: input{Version: 1, Rules: override.Rules, Gate: override.Gate, Analysis: override.Analysis,
				Extraction: override.Extraction, Vocabulary: override.Vocabulary, Suppressions: override.Suppressions}})
	}
	return result, nil
}

func validateOverride(override overrideInput) error {
	if len(override.Files) == 0 || len(override.Files) > 100 {
		return fmt.Errorf("file override requires 1 to 100 patterns")
	}
	seen := make(map[string]bool)
	for _, pattern := range override.Files {
		if err := relativeGlob(pattern); err != nil {
			return err
		}
		if seen[pattern] {
			return fmt.Errorf("duplicate override glob %q", pattern)
		}
		seen[pattern] = true
	}
	for _, field := range []struct {
		node yaml.Node
		key  string
	}{
		{override.Analysis, "max_total_bytes"}, {override.Gate, "fail_on_empty"}, {override.Gate, "fail_on_incomplete"},
	} {
		if nodeField(field.node, field.key) != nil {
			return fmt.Errorf("file overrides cannot change run-wide %s", field.key)
		}
	}
	return nil
}

func relativeGlob(pattern string) error {
	if pattern == "" || path.IsAbs(pattern) || strings.ContainsAny(pattern, "\\:\x00") {
		return fmt.Errorf("invalid project glob %q", pattern)
	}
	for part := range strings.SplitSeq(pattern, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("glob must be project relative: %q", pattern)
		}
	}
	return nil
}

func nodeField(node yaml.Node, key string) *yaml.Node {
	for i := 0; i+1 < len(node.Content); i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}
	return nil
}

func applyLayer(raw input, policy *Policy, catalog []rule.Descriptor, origin string) error {
	if err := applyNodes(raw, policy, catalog); err != nil {
		return fmt.Errorf("%s: %w", origin, err)
	}
	for id, node := range raw.Rules {
		policy.Origins["rules."+id] = origin
		recordNodeOrigins(policy.Origins, "/rules/"+pointerKey(id), node, origin)
	}
	for _, item := range []struct {
		name string
		node yaml.Node
	}{
		{"gate", raw.Gate}, {"analysis", raw.Analysis}, {"files", raw.Files}, {"extraction", raw.Extraction}, {"vocabulary", raw.Vocabulary},
		{"suppressions", raw.Suppressions},
	} {
		if item.node.Kind != 0 {
			recordNodeOrigins(policy.Origins, "/"+item.name, item.node, origin)
		}
	}
	policy.Origins["/version"] = origin
	if raw.Language != "" {
		policy.Origins["/language"] = origin
	}
	return nil
}

func applyBuiltin(policy *Policy, profile string, catalog []rule.Descriptor) error {
	base, err := defaults(profile, catalog)
	if err != nil {
		return err
	}
	policy.Profile, policy.Rules, policy.Gate = base.Profile, base.Rules, base.Gate
	if err := recordValueOrigins(policy.Origins, "/gate", policy.Gate, "builtin:"+policy.Profile); err != nil {
		return err
	}
	policy.Origins["/profile"] = "builtin:" + policy.Profile
	return ruleOrigins(policy, catalog)
}

func defaultOrigins(policy *Policy, catalog []rule.Descriptor) error {
	origin := "builtin:" + policy.Profile
	for _, item := range []struct {
		name  string
		value any
	}{
		{"version", policy.Version}, {"profile", policy.Profile}, {"language", policy.Language}, {"gate", policy.Gate},
		{"analysis", policy.Analysis}, {"files", policy.Files}, {"extraction", policy.Extraction},
		{"group_caps", policy.GroupCaps}, {"vocabulary", policy.Vocabulary},
		{"suppressions", policy.Suppressions},
	} {
		if err := recordValueOrigins(policy.Origins, "/"+item.name, item.value, origin); err != nil {
			return err
		}
	}
	return ruleOrigins(policy, catalog)
}

func ruleOrigins(policy *Policy, catalog []rule.Descriptor) error {
	for _, descriptor := range catalog {
		origin := "builtin:" + policy.Profile
		if descriptor.Origin != nil {
			origin = "ruleset:" + descriptor.Origin.Namespace + "@" + descriptor.Origin.Version
		}
		policy.Origins["rules."+descriptor.ID] = origin
		if err := recordValueOrigins(policy.Origins, "/rules/"+pointerKey(descriptor.ID), policy.Rules[descriptor.ID], origin); err != nil {
			return err
		}
	}
	return nil
}

func recordValueOrigins(origins map[string]string, prefix string, value any, origin string) error {
	var node yaml.Node
	if err := node.Encode(value); err != nil {
		return err
	}
	recordNodeOrigins(origins, prefix, node, origin)
	return nil
}

func recordNodeOrigins(origins map[string]string, prefix string, node yaml.Node, origin string) {
	origins[prefix] = origin
	if node.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(node.Content); i += 2 {
			recordNodeOrigins(origins, prefix+"/"+pointerKey(node.Content[i].Value), *node.Content[i+1], origin)
		}
		return
	}
	for key := range origins {
		if strings.HasPrefix(key, prefix+"/") {
			delete(origins, key)
		}
	}
	if node.Kind == yaml.SequenceNode {
		for i, child := range node.Content {
			recordNodeOrigins(origins, prefix+"/"+strconv.Itoa(i), *child, origin)
		}
	}
}

func pointerKey(key string) string {
	return strings.ReplaceAll(strings.ReplaceAll(key, "~", "~0"), "/", "~1")
}
