package config

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/internal/terms"
	"github.com/stokaro/unswell/rule"
)

// Vocabulary combines explicit inline and shared terms. ResolvedTerms is output
// only. TermExemptions selects rules that declare candidate-level term support.
type Vocabulary struct {
	Terms          []string `json:"terms" yaml:"terms"`
	Dictionaries   []string `json:"dictionaries" yaml:"dictionaries"`
	TermExemptions []string `json:"term_exemptions" yaml:"term_exemptions"`
	CaseSensitive  bool     `json:"case_sensitive" yaml:"case_sensitive"`
	ResolvedTerms  []string `json:"resolved_terms" yaml:"-"`
}

func (c *bundleCompiler) dictionary(source, reference string) error {
	name, err := ResolveReference(source, reference, c.bundle.AllowOutsideRoot)
	if err != nil {
		return err
	}
	if _, exists := c.dictionaries[name]; exists {
		return nil
	}
	data, exists := c.bundle.Files[name]
	if !exists {
		return fmt.Errorf("missing dictionary resource %q", name)
	}
	words, err := dictionaryWords(data)
	if err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	if err := c.remember(name, "dictionary", data); err != nil {
		return err
	}
	c.dictionaries[name] = words
	return nil
}

func dictionaryWords(data []byte) ([]string, error) {
	if err := validateYAML(data); err != nil {
		return nil, err
	}
	var dictionary struct {
		Version int      `yaml:"version"`
		Terms   []string `yaml:"terms"`
	}
	if err := decode(data, &dictionary); err != nil {
		return nil, err
	}
	if dictionary.Version != 1 || len(dictionary.Terms) == 0 {
		return nil, fmt.Errorf("dictionary requires version 1 and nonempty terms")
	}
	if _, err := terms.Compile(dictionary.Terms, true); err != nil {
		return nil, err
	}
	return dictionary.Terms, nil
}

func resolveVocabularyPaths(node *yaml.Node, source string, allowOutside bool) error {
	list := nodeField(*node, "dictionaries")
	if list == nil {
		return nil
	}
	var paths []string
	if err := list.Decode(&paths); err != nil {
		return err
	}
	seen := make(map[string]bool)
	for i, reference := range paths {
		name, err := ResolveReference(source, reference, allowOutside)
		if err != nil {
			return err
		}
		if seen[name] {
			return fmt.Errorf("duplicate dictionary %q", name)
		}
		seen[name] = true
		paths[i] = name
	}
	return list.Encode(paths)
}

func (p *Plan) vocabulary(policy *Policy) error {
	v := &policy.Vocabulary
	v.ResolvedTerms = slices.Clone(v.Terms)
	origins := make([]string, len(v.Terms))
	for i := range origins {
		origins[i] = policy.Origins["/vocabulary/terms"]
	}
	for _, name := range v.Dictionaries {
		words, exists := p.dictionaries[name]
		if !exists {
			return fmt.Errorf("missing compiled dictionary %q", name)
		}
		v.ResolvedTerms = append(v.ResolvedTerms, words...)
		for range words {
			origins = append(origins, name)
		}
	}
	if _, err := terms.Compile(v.ResolvedTerms, v.CaseSensitive); err != nil {
		return err
	}
	for key := range policy.Origins {
		if strings.HasPrefix(key, "/vocabulary/resolved_terms/") {
			delete(policy.Origins, key)
		}
	}
	for i, origin := range origins {
		policy.Origins["/vocabulary/resolved_terms/"+strconv.Itoa(i)] = origin
	}
	return p.validateExemptions(v.TermExemptions)
}

func (p *Plan) validateExemptions(ids []string) error {
	seen := make(map[string]bool)
	for _, id := range ids {
		if seen[id] {
			return fmt.Errorf("duplicate term exemption %q", id)
		}
		seen[id] = true
		index := slices.IndexFunc(p.catalog, func(d rule.Descriptor) bool { return d.ID == id })
		if index < 0 {
			return fmt.Errorf("unknown term exemption rule %q", id)
		}
		if !p.catalog[index].TermExemptions {
			return fmt.Errorf("rule %s does not support candidate-level term exemptions", id)
		}
	}
	return nil
}
