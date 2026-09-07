package config

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Bundle provides all configuration resources as bytes, without an implicit loader.
// Root and Files keys use project-relative slash paths. AllowOutsideRoot is an
// explicit trusted-client permission for parent-relative local dependencies.
type Bundle struct {
	Root             string
	Files            map[string][]byte
	AllowOutsideRoot bool
}

// Reference identifies a local configuration or vocabulary dependency.
type Reference struct {
	Path string
	Kind string
}

// SourceIdentity records a canonical YAML resource, excluding formatting/comments.
type SourceIdentity struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
	Hash string `json:"sha256"`
}

// References validates the YAML structure and lists direct local dependencies.
// Builtin profiles have no file dependency. Paths remain relative to the source.
func References(data []byte) ([]Reference, error) {
	raw, err := loadInput(data)
	if err != nil {
		return nil, err
	}
	return inputReferences(raw)
}

func inputReferences(raw input) ([]Reference, error) {
	result, err := parentReferences(raw.Extends)
	if err != nil {
		return nil, err
	}
	nodes := []yaml.Node{raw.Vocabulary}
	for _, override := range raw.Overrides {
		nodes = append(nodes, override.Vocabulary)
	}
	for _, node := range nodes {
		var vocabulary Vocabulary
		if node.Kind != 0 {
			if err := mergeNode(node, &vocabulary); err != nil {
				return nil, err
			}
		}
		for _, name := range vocabulary.Dictionaries {
			result = append(result, Reference{Path: name, Kind: "dictionary"})
		}
	}
	return result, nil
}

func parentReferences(parents []string) ([]Reference, error) {
	if len(parents) > 64 {
		return nil, fmt.Errorf("configuration exceeds 64 extends references")
	}
	var result []Reference
	seen := make(map[string]bool)
	for _, name := range parents {
		key, err := parentKey(name)
		if err != nil {
			return nil, err
		}
		if seen[key] {
			return nil, fmt.Errorf("duplicate extends reference %q", name)
		}
		seen[key] = true
		if !strings.HasPrefix(name, "builtin:") {
			result = append(result, Reference{Path: name, Kind: "config"})
		}
	}
	return result, nil
}

func parentKey(name string) (string, error) {
	if strings.HasPrefix(name, "builtin:") {
		profile, err := resolveProfile([]string{name})
		return "builtin:" + profile, err
	}
	return resourceName(name, true)
}

// ResolveReference resolves a portable local dependency relative to its source.
// Network addresses, absolute paths, and platform-specific separators are invalid.
func ResolveReference(source, reference string, allowOutside bool) (string, error) {
	if reference == "" || strings.ContainsAny(reference, "\\:\x00") || path.IsAbs(reference) {
		return "", fmt.Errorf("invalid local configuration reference %q", reference)
	}
	return resourceName(path.Join(path.Dir(source), reference), allowOutside)
}

func resourceName(name string, allowOutside bool) (string, error) {
	if name == "" || len(name) > 1000 || strings.ContainsAny(name, "\\:\x00") || path.IsAbs(name) {
		return "", fmt.Errorf("invalid configuration resource name %q", name)
	}
	name = path.Clean(name)
	check := name
	if allowOutside {
		for strings.HasPrefix(check, "../") {
			check = strings.TrimPrefix(check, "../")
		}
	}
	if check == "." || !fs.ValidPath(check) {
		return "", fmt.Errorf("configuration reference escapes project root: %q", name)
	}
	return name, nil
}

func normalizeBundle(bundle Bundle) (Bundle, error) {
	if len(bundle.Files) == 0 || len(bundle.Files) > 64 {
		return Bundle{}, fmt.Errorf("configuration bundle requires 1 to 64 resources")
	}
	root, err := resourceName(bundle.Root, bundle.AllowOutsideRoot)
	if err != nil {
		return Bundle{}, err
	}
	result := Bundle{Root: root, Files: make(map[string][]byte), AllowOutsideRoot: bundle.AllowOutsideRoot}
	keys := make([]string, 0, len(bundle.Files))
	for name := range bundle.Files {
		keys = append(keys, name)
	}
	slices.Sort(keys)
	total := 0
	for _, key := range keys {
		name, err := resourceName(key, bundle.AllowOutsideRoot)
		if err != nil {
			return Bundle{}, err
		}
		if _, exists := result.Files[name]; exists {
			return Bundle{}, fmt.Errorf("duplicate configuration resource %q", name)
		}
		data := bundle.Files[key]
		total += len(data)
		if len(data) > 1<<20 || total > 8<<20 {
			return Bundle{}, fmt.Errorf("configuration resources exceed 1 MiB each or 8 MiB total")
		}
		result.Files[name] = bytes.Clone(data)
	}
	return result, nil
}

func sourceIdentity(name, kind string, data []byte) (SourceIdentity, error) {
	var value map[string]any
	if len(bytes.TrimSpace(data)) != 0 {
		if err := decode(data, &value); err != nil {
			return SourceIdentity{}, err
		}
	}
	canonicalContextSets(value)
	canonical, err := json.Marshal(value)
	if err != nil {
		return SourceIdentity{}, err
	}
	return SourceIdentity{Path: name, Kind: kind, Hash: fmt.Sprintf("%x", sha256.Sum256(canonical))}, nil
}

func canonicalContextSets(value map[string]any) {
	extraction, _ := value["extraction"].(map[string]any)
	sortContextSet(extraction)
	languages, _ := extraction["languages"].(map[string]any)
	for _, language := range languages {
		settings, _ := language.(map[string]any)
		sortContextSet(settings)
	}
	overrides, _ := value["overrides"].([]any)
	for _, override := range overrides {
		settings, _ := override.(map[string]any)
		canonicalContextSets(settings)
	}
}

func sortContextSet(settings map[string]any) {
	values, ok := settings["contexts"].([]any)
	if !ok {
		return
	}
	contexts := make([]string, len(values))
	for i, value := range values {
		text, ok := value.(string)
		if !ok {
			return
		}
		contexts[i] = text
	}
	slices.Sort(contexts)
	settings["contexts"] = contexts
}
