package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"sync"

	"go.yaml.in/yaml/v3"

	"github.com/stokaro/unswell/internal/pathglob"
	"github.com/stokaro/unswell/rule"
)

// OverrideIdentity identifies an ordered per-file policy layer.
type OverrideIdentity struct {
	ID    string   `json:"id"`
	Files []string `json:"files"`
	Hash  string   `json:"sha256"`
}

type overrideInput struct {
	Files        []string             `yaml:"files"`
	Rules        map[string]yaml.Node `yaml:"rules"`
	Gate         yaml.Node            `yaml:"gate"`
	Analysis     yaml.Node            `yaml:"analysis"`
	Extraction   yaml.Node            `yaml:"extraction"`
	Vocabulary   yaml.Node            `yaml:"vocabulary"`
	Suppressions yaml.Node            `yaml:"suppressions"`
}

type policyLayer struct {
	name, profile string
	raw           input
}
type fileLayer struct {
	identity OverrideIdentity
	raw      input
	patterns []*regexp.Regexp
}

// Plan is an immutable compiled configuration graph. Policy returns an owned
// snapshot that callers may modify. ForFile returns the shared resolution of
// one override combination, which callers read and never modify in place.
type Plan struct {
	base         Policy
	catalog      []rule.Descriptor
	overrides    []fileLayer
	dictionaries map[string][]string
	enabled      map[string]bool
	resolved     sync.Map
}

type bundleCompiler struct {
	bundle       Bundle
	layers       []policyLayer
	active       map[string]bool
	sources      map[string]SourceIdentity
	dictionaries map[string][]string
}

// CompileBundle compiles ordered inheritance and file overrides from explicit
// in-memory resources. It returns inline rule implementations for registration.
func CompileBundle(bundle Bundle, catalog []rule.Descriptor) (*Plan, []rule.Rule, error) {
	bundle, err := normalizeBundle(bundle)
	if err != nil {
		return nil, nil, err
	}
	c := bundleCompiler{bundle: bundle, active: make(map[string]bool), sources: make(map[string]SourceIdentity),
		dictionaries: make(map[string][]string)}
	if err := c.walk(bundle.Root, 0); err != nil {
		return nil, nil, err
	}
	additional, err := c.ruleSets()
	if err != nil {
		return nil, nil, err
	}
	catalog = slices.Clone(catalog)
	for _, implementation := range additional {
		catalog = append(catalog, implementation.Descriptor())
	}
	if len(catalog) > 1000 {
		return nil, nil, fmt.Errorf("catalog exceeds 1000 rules")
	}
	catalog, err = copyCatalog(catalog)
	if err != nil {
		return nil, nil, err
	}
	plan, err := c.compile(catalog)
	return plan, additional, err
}

func (c *bundleCompiler) ruleSets() ([]rule.Rule, error) {
	var selected []rule.Rule
	for _, layer := range c.layers {
		if layer.raw.RuleSets == nil {
			continue
		}
		compiled, err := compileRuleSets(layer.raw.RuleSets)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", layer.name, err)
		}
		selected = compiled
	}
	return selected, nil
}

func copyCatalog(catalog []rule.Descriptor) ([]rule.Descriptor, error) {
	data, err := json.Marshal(catalog)
	if err != nil {
		return nil, err
	}
	var result []rule.Descriptor
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *bundleCompiler) walk(name string, depth int) error {
	if depth > 16 || len(c.layers) >= 256 {
		return fmt.Errorf("configuration inheritance exceeds depth 16 or 256 layers")
	}
	if c.active[name] {
		return fmt.Errorf("configuration inheritance cycle at %q", name)
	}
	raw, refs, err := c.readLayer(name)
	if err != nil {
		return err
	}
	c.active[name] = true
	defer delete(c.active, name)
	if err := c.inherit(name, raw.Extends, depth); err != nil {
		return err
	}
	if err := c.resources(name, &raw, refs); err != nil {
		return err
	}
	return c.appendLayer(policyLayer{name: name, raw: raw})
}

func (c *bundleCompiler) readLayer(name string) (input, []Reference, error) {
	data, exists := c.bundle.Files[name]
	if !exists {
		return input{}, nil, fmt.Errorf("missing configuration resource %q", name)
	}
	raw, err := loadInput(data)
	if err != nil {
		return input{}, nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := c.remember(name, "config", data); err != nil {
		return input{}, nil, err
	}
	refs, err := inputReferences(raw)
	if err != nil {
		return input{}, nil, fmt.Errorf("%s: %w", name, err)
	}
	return raw, refs, nil
}

func (c *bundleCompiler) inherit(name string, parents []string, depth int) error {
	for _, parent := range parents {
		if strings.HasPrefix(parent, "builtin:") {
			profile, err := resolveProfile([]string{parent})
			if err != nil {
				return err
			}
			if err := c.appendLayer(policyLayer{name: "builtin:" + profile + "-v1", profile: profile}); err != nil {
				return err
			}
			continue
		}
		resolved, err := ResolveReference(name, parent, c.bundle.AllowOutsideRoot)
		if err != nil {
			return err
		}
		if err := c.walk(resolved, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (c *bundleCompiler) resources(name string, raw *input, refs []Reference) error {
	for _, ref := range refs {
		if ref.Kind == "dictionary" {
			if err := c.dictionary(name, ref.Path); err != nil {
				return err
			}
		}
	}
	if err := resolveVocabularyPaths(&raw.Vocabulary, name, c.bundle.AllowOutsideRoot); err != nil {
		return err
	}
	for i := range raw.Overrides {
		if err := resolveVocabularyPaths(&raw.Overrides[i].Vocabulary, name, c.bundle.AllowOutsideRoot); err != nil {
			return err
		}
	}
	return nil
}

func (c *bundleCompiler) appendLayer(layer policyLayer) error {
	if len(c.layers) >= 256 {
		return fmt.Errorf("configuration inheritance exceeds 256 layers")
	}
	c.layers = append(c.layers, layer)
	return nil
}

func (c *bundleCompiler) remember(name, kind string, data []byte) error {
	identity, err := sourceIdentity(name, kind, data)
	if err != nil {
		return err
	}
	if old, exists := c.sources[name]; exists && old.Kind != kind {
		return fmt.Errorf("resource %q has conflicting kinds", name)
	}
	c.sources[name] = identity
	return nil
}

func (c *bundleCompiler) compile(catalog []rule.Descriptor) (*Plan, error) {
	policy, err := defaults("technical", catalog)
	if err != nil {
		return nil, err
	}
	plan := &Plan{catalog: catalog, dictionaries: c.dictionaries, enabled: make(map[string]bool)}
	if err := defaultOrigins(&policy, catalog); err != nil {
		return nil, err
	}
	overrides, declared, err := c.applyLayers(&policy, catalog)
	if err != nil {
		return nil, err
	}
	policy.RuleSets = catalogOrigins(catalog)
	for _, identity := range c.sources {
		policy.Sources = append(policy.Sources, identity)
	}
	slices.SortFunc(policy.Sources, func(a, b SourceIdentity) int { return strings.Compare(a.Path, b.Path) })
	plan.overrides = overrides
	for _, override := range overrides {
		policy.Overrides = append(policy.Overrides, override.identity)
	}
	if err := plan.finish(&policy); err != nil {
		return nil, err
	}
	plan.base = policy
	plan.recordEnabled(policy)
	if err := plan.validateOverrides(declared); err != nil {
		return nil, err
	}
	return plan, nil
}

func (c *bundleCompiler) applyLayers(policy *Policy, catalog []rule.Descriptor) ([]fileLayer, []fileLayer, error) {
	var selected, declared []fileLayer
	for _, layer := range c.layers {
		if layer.profile != "" {
			if err := applyBuiltin(policy, layer.profile, catalog); err != nil {
				return nil, nil, err
			}
			continue
		}
		if err := applyLayer(layer.raw, policy, catalog, layer.name); err != nil {
			return nil, nil, err
		}
		if layer.raw.Overrides == nil {
			continue
		}
		compiled, err := compileOverrides(layer.raw.Overrides, layer.name)
		if err != nil {
			return nil, nil, err
		}
		selected = compiled
		declared = append(declared, compiled...)
		if len(declared) > 256 {
			return nil, nil, fmt.Errorf("configuration graph exceeds 256 declared overrides")
		}
	}
	return selected, declared, nil
}

func (p *Plan) validateOverrides(overrides []fileLayer) error {
	for _, override := range overrides {
		candidate, err := p.Policy()
		if err != nil {
			return err
		}
		if err := applyLayer(override.raw, &candidate, p.catalog, override.identity.ID); err != nil {
			return err
		}
		if err := p.finish(&candidate); err != nil {
			return fmt.Errorf("%s: %w", override.identity.ID, err)
		}
		if slices.ContainsFunc(p.overrides, func(active fileLayer) bool { return active.identity.ID == override.identity.ID }) {
			p.recordEnabled(candidate)
		}
	}
	return nil
}

func (p *Plan) recordEnabled(policy Policy) {
	for id, settings := range policy.Rules {
		if settings.Enabled {
			p.enabled[id] = true
		}
	}
}

// EnabledRuleIDs lists rules enabled by the base or any individual file override.
// This conservative union lets the engine validate capabilities before scanning.
func (p *Plan) EnabledRuleIDs() []string {
	result := make([]string, 0, len(p.enabled))
	for id := range p.enabled {
		result = append(result, id)
	}
	slices.Sort(result)
	return result
}

// Policy returns an owned base policy, before selecting any file overrides.
func (p *Plan) Policy() (Policy, error) {
	data, err := json.Marshal(p.base)
	if err != nil {
		return Policy{}, err
	}
	var result Policy
	err = json.Unmarshal(data, &result)
	return result, err
}

// ForFile resolves ordered overrides for a project-relative logical source name.
// It validates the complete selected combination before prose analysis begins.
// The returned policy is an owned snapshot that the caller may modify.
func (p *Plan) ForFile(name string) (Policy, error) {
	resolved, err := p.resolution(name)
	if err != nil {
		return Policy{}, err
	}
	var owned Policy
	err = json.Unmarshal(resolved.data, &owned)
	return owned, err
}

// Resolution returns the shared policy of a source name. Its override
// combination is resolved once per plan, so a scan of many files pays one map
// lookup per file instead of a copy of the whole policy. Callers read the
// result and never modify its maps or slices; ForFile returns an owned copy
// of the same resolution.
func (p *Plan) Resolution(name string) (Policy, error) {
	resolved, err := p.resolution(name)
	if err != nil {
		return Policy{}, err
	}
	return resolved.policy, nil
}

// resolution holds one resolved override combination and its serialized form,
// from which owned copies are decoded.
type resolution struct {
	policy Policy
	data   []byte
}

func (p *Plan) resolution(name string) (resolution, error) {
	var selected []fileLayer
	if len(p.overrides) > 0 && name != "" {
		resource, err := resourceName(name, false)
		if err != nil {
			return resolution{}, fmt.Errorf("override source name: %w", err)
		}
		for _, override := range p.overrides {
			if pathglob.Matches(resource, override.patterns) {
				selected = append(selected, override)
			}
		}
	}
	key := make([]string, 0, len(selected))
	for _, override := range selected {
		key = append(key, override.identity.ID)
	}
	return p.resolve(strings.Join(key, "\x00"), selected)
}

// resolve returns one override combination, computing it on the first request.
// A failed resolution is not remembered, so the error repeats.
func (p *Plan) resolve(key string, selected []fileLayer) (resolution, error) {
	if cached, ok := p.resolved.Load(key); ok {
		return cached.(resolution), nil
	}
	policy, err := p.Policy()
	if err != nil {
		return resolution{}, err
	}
	for _, override := range selected {
		if err := applyLayer(override.raw, &policy, p.catalog, override.identity.ID); err != nil {
			return resolution{}, err
		}
		policy.AppliedOverrides = append(policy.AppliedOverrides, override.identity.ID)
	}
	if len(policy.AppliedOverrides) > 0 {
		if err := p.finish(&policy); err != nil {
			return resolution{}, err
		}
	}
	data, err := json.Marshal(policy)
	if err != nil {
		return resolution{}, err
	}
	stored, _ := p.resolved.LoadOrStore(key, resolution{policy: policy, data: data})
	return stored.(resolution), nil
}

func (p *Plan) finish(policy *Policy) error {
	if err := p.vocabulary(policy); err != nil {
		return err
	}
	if err := validate(*policy, p.catalog); err != nil {
		return err
	}
	slices.Sort(policy.Extraction.Contexts)
	for format, override := range policy.Extraction.Languages {
		slices.Sort(override.Contexts)
		policy.Extraction.Languages[format] = override
	}
	policy.Hash = ""
	data, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	policy.Hash = fmt.Sprintf("%x", sha256.Sum256(data))
	return nil
}
