// Package unswell analyzes English prose offline with an extensible rule engine.
// It never discovers files, reads environment variables, prints, exits, or uses
// the network. Style findings are results; operational failures are Go errors.
package unswell

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
)

// Options injects all inputs. A nil Rules slice selects the builtin catalog;
// a nonnil slice replaces that catalog. RuleSets adds declarative YAML packs to
// the registry. Config and RuleSets contain bytes, never filenames.
type Options struct {
	// Baseline contains an explicitly selected artifact; nil disables comparison.
	Baseline []byte
	// CollectBaseline requests a validated snapshot for explicit create or update.
	CollectBaseline bool
	// GateMode overrides configuration with all or new; empty follows the policy.
	GateMode      string
	Config        []byte
	ConfigBundle  *config.Bundle
	Rules         []rule.Rule
	RuleSets      [][]byte
	NLP           nlp.Provider
	Jobs          int
	IncludeSource bool
	AllowEmpty    bool
	NoGate        bool
}

// Engine is immutable after construction and supports concurrent calls. Custom
// rule and NLP implementations must uphold their documented concurrency contract.
type Engine struct {
	trustedSources        map[string]sourceIdentities
	baselineFile          *baseline.File
	collectBaseline       bool
	baselineCompatibility baseline.Compatibility
	gateMode              string
	policy                config.Policy
	plan                  *config.Plan
	rules                 []rule.Rule
	descriptors           []rule.Descriptor
	nlp                   nlp.Provider
	capabilities          []nlp.Capability
	jobs                  int
	includeSource         bool
	allowEmpty            bool
	noGate                bool
	rulesetHash           string
}

// New validates the entire policy and required capabilities before reading prose.
func New(options Options) (*Engine, error) {
	registry := options.Rules
	if registry == nil {
		registry = builtin.Rules()
	}
	registry, err := extendRegistry(registry, options.RuleSets)
	if err != nil {
		return nil, err
	}
	e := &Engine{
		rules:         slices.Clone(registry),
		jobs:          options.Jobs,
		includeSource: options.IncludeSource,
		allowEmpty:    options.AllowEmpty,
		noGate:        options.NoGate,
	}
	jobs, err := workerCount(options.Jobs)
	if err != nil {
		return nil, err
	}
	e.jobs = jobs
	if err := e.snapshotDescriptors(); err != nil {
		return nil, err
	}
	if err := e.configure(options); err != nil {
		return nil, err
	}
	e.nlp = options.NLP
	if e.nlp == nil {
		e.nlp, err = english.New()
		if err != nil {
			return nil, err
		}
	}
	if err := e.planCapabilities(); err != nil {
		return nil, err
	}
	if err := e.configureBaseline(options); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *Engine) configure(options Options) error {
	if options.ConfigBundle != nil && len(options.Config) != 0 {
		return fmt.Errorf("config bytes and ConfigBundle cannot be combined")
	}
	bundle := config.Bundle{Root: ".unswell.yaml", Files: map[string][]byte{".unswell.yaml": options.Config}}
	if options.ConfigBundle != nil {
		bundle = *options.ConfigBundle
	}
	plan, additional, err := config.CompileBundle(bundle, e.descriptors)
	if err != nil {
		return err
	}
	e.plan = plan
	e.policy, err = plan.Policy()
	if err != nil {
		return err
	}
	if len(additional) == 0 {
		return nil
	}
	e.rules = append(e.rules, additional...)
	e.descriptors = nil
	return e.snapshotDescriptors()
}

// PolicyForFile returns an owned effective policy for a project-relative logical
// filename. An empty filename returns the base discovery and run-wide policy.
func (e *Engine) PolicyForFile(name string) (config.Policy, error) {
	return e.plan.ForFile(name)
}

// Catalog returns owned descriptors for the complete registered rule catalog.
func (e *Engine) Catalog() []rule.Descriptor {
	return cloneDescriptors(e.descriptors)
}

func extendRegistry(registry []rule.Rule, definitions [][]byte) ([]rule.Rule, error) {
	if len(definitions) > 10 {
		return nil, fmt.Errorf("options exceed 10 rulesets")
	}
	registry = slices.Clone(registry)
	for _, data := range definitions {
		set, err := ruleset.Load(data)
		if err != nil {
			return nil, err
		}
		registry = append(registry, set.Rules()...)
	}
	return registry, nil
}

func workerCount(jobs int) (int, error) {
	if jobs == 0 {
		jobs = 1
	}
	if jobs < 1 || jobs > 64 {
		return 0, fmt.Errorf("jobs must be in [1,64]")
	}
	return jobs, nil
}

func (e *Engine) snapshotDescriptors() error {
	if len(e.rules) > 1000 {
		return fmt.Errorf("registry exceeds 1000 rules")
	}
	for _, implementation := range e.rules {
		if implementation == nil {
			return fmt.Errorf("nil rule implementation")
		}
		descriptor := implementation.Descriptor()
		if !strings.Contains(descriptor.ID, ".") || descriptor.Version == "" || descriptor.Group == "" {
			return fmt.Errorf("incomplete rule descriptor %q", descriptor.ID)
		}
		e.descriptors = append(e.descriptors, descriptor)
	}
	data, err := json.Marshal(e.descriptors)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &e.descriptors); err != nil {
		return err
	}
	slices.SortFunc(e.descriptors, func(a, b rule.Descriptor) int { return strings.Compare(a.ID, b.ID) })
	data, err = json.Marshal(e.descriptors)
	if err != nil {
		return err
	}
	e.rulesetHash = fmt.Sprintf("%x", sha256.Sum256(data))
	return nil
}

func (e *Engine) planCapabilities() error {
	e.capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences}
	available := e.nlp.Identity().Capabilities
	enabled := e.plan.EnabledRuleIDs()
	for _, descriptor := range e.descriptors {
		if !slices.Contains(enabled, descriptor.ID) {
			continue
		}
		for _, capability := range descriptor.Requires {
			if !slices.Contains(available, capability) {
				return fmt.Errorf("rule %s requires unavailable capability %s", descriptor.ID, capability)
			}
			if !slices.Contains(e.capabilities, capability) {
				e.capabilities = append(e.capabilities, capability)
			}
		}
	}
	slices.Sort(e.capabilities)
	return nil
}

// Analyze runs the same pipeline used by AnalyzeAll on one in-memory source.
func (e *Engine) Analyze(ctx context.Context, src document.Source) (Result, error) {
	return e.AnalyzeAll(ctx, []document.Source{src})
}

// AnalyzeAll returns a deterministic, path-sorted result regardless of worker
// count or source order. It returns partial evidence and an error on incompleteness.
func (e *Engine) AnalyzeAll(ctx context.Context, sources []document.Source) (RunResult, error) {
	result, _, err := e.analyzeAll(ctx, sources, false)
	return result, err
}

func (e *Engine) analyzeAll(ctx context.Context, sources []document.Source, identify bool) (RunResult, []sourceIdentities, error) {
	result := e.emptyResult()
	if err := ctx.Err(); err != nil {
		return incompleteBatch(result, err)
	}
	sources = slices.Clone(sources)
	slices.SortFunc(sources, func(a, b document.Source) int { return strings.Compare(a.Name, b.Name) })
	engines, err := e.prepareSources(ctx, sources)
	if err != nil {
		return incompleteBatch(result, err)
	}
	if err := ctx.Err(); err != nil {
		return incompleteBatch(result, err)
	}
	partials := make([]RunResult, len(sources))
	identities := newBatchIdentities(identify, len(sources))
	errorsBySource := make([]error, len(sources))
	var workers sync.WaitGroup
	for worker := 0; worker < min(e.jobs, len(sources)); worker++ {
		workers.Go(func() {
			for index := worker; index < len(sources); index += e.jobs {
				partials[index], errorsBySource[index] = engines[index].analyzeSource(ctx, sources[index], batchIdentity(identities, index))
			}
		})
	}
	workers.Wait()
	for i, part := range partials {
		result.Documents = append(result.Documents, part.Documents...)
		result.Findings = append(result.Findings, part.Findings...)
		result.Assessments = append(result.Assessments, part.Assessments...)
		result.Suppressions = append(result.Suppressions, part.Suppressions...)
		if result.BaselineSnapshot != nil && part.BaselineSnapshot != nil {
			result.BaselineSnapshot.Documents = append(result.BaselineSnapshot.Documents, part.BaselineSnapshot.Documents...)
			result.BaselineSnapshot.Candidates = append(result.BaselineSnapshot.Candidates, part.BaselineSnapshot.Candidates...)
		}
		result.Gate.Reasons = append(result.Gate.Reasons, part.Gate.Reasons...)
		if errorsBySource[i] != nil {
			result.Errors = append(result.Errors, RunError{Path: sources[i].Name, Message: errorsBySource[i].Error()})
		}
	}
	result, err = e.finishBaseline(ctx, result, errors.Join(errorsBySource...))
	return result, identities, err
}

func newBatchIdentities(identify bool, count int) []sourceIdentities {
	if identify {
		return make([]sourceIdentities, count)
	}
	return nil
}

func incompleteBatch(result RunResult, err error) (RunResult, []sourceIdentities, error) {
	result, err = incomplete(result, err)
	return result, nil, err
}

func batchIdentity(identities []sourceIdentities, index int) *sourceIdentities {
	if identities == nil {
		return nil
	}
	return &identities[index]
}

func (e *Engine) emptyResult() RunResult {
	result := RunResult{
		SchemaVersion: SchemaVersion,
		Status:        "complete",
		Documents:     []DocumentResult{},
		Findings:      []Finding{},
		Assessments:   []Assessment{},
		Errors:        []RunError{},
		Gate:          GateDecision{Passed: true, Reasons: []GateReason{}},
		Manifest: Manifest{
			GateMode:        e.selectedGateMode(),
			ToolVersion:     Version,
			ToolCommit:      BuildCommit,
			ConfigHash:      e.policy.Hash,
			ConfigIdentity:  e.policy.Identity,
			ConfigSources:   slices.Clone(e.policy.Sources),
			ConfigOverrides: cloneOverrides(e.policy.Overrides),
			RulesetHash:     e.rulesetHash,
			ScoringProfile:  e.policy.Profile,
			FeatureContract: "unswell-features-v1",
			NLP:             e.nlp.Identity(),
			Rules:           cloneDescriptors(e.descriptors),
			SelectionMode:   "explicit",
			Complete:        true,
			NoGate:          e.noGate,
			IncludeSource:   e.includeSource,
			SkippedRules:    []string{},
		},
	}
	if e.collectBaseline {
		result.BaselineSnapshot = &baseline.Snapshot{Complete: true, Compatibility: e.baselineCompatibility,
			Documents: []baseline.Document{}, Candidates: []baseline.Identity{}}
	}
	return result
}

func (e *Engine) prepareSources(ctx context.Context, sources []document.Source) ([]*Engine, error) {
	engines := make([]*Engine, len(sources))
	resolved := make(map[string]*Engine)
	total := 0
	for i, source := range sources {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if i > 0 && sources[i-1].Name == source.Name {
			return nil, fmt.Errorf("duplicate source name %q", source.Name)
		}
		policy, err := e.PolicyForFile(source.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source.Name, err)
		}
		engine, err := e.sourceEngine(policy, resolved)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", source.Name, err)
		}
		engines[i] = engine
		if len(source.Bytes) > policy.Analysis.MaxFileBytes {
			return nil, fmt.Errorf("%s exceeds max_file_bytes", source.Name)
		}
		total += len(source.Bytes)
		if total > e.policy.Analysis.MaxTotalBytes {
			return nil, fmt.Errorf("sources exceed max_total_bytes")
		}
	}
	return engines, nil
}

func (e *Engine) sourceEngine(policy config.Policy, resolved map[string]*Engine) (*Engine, error) {
	if engine := resolved[policy.Hash]; engine != nil {
		return engine, nil
	}
	local := *e
	local.policy = policy
	if local.selectedGateMode() == "new" && local.baselineFile == nil {
		return nil, fmt.Errorf("gate.mode new requires a baseline")
	}
	resolved[policy.Hash] = &local
	return &local, nil
}

func incomplete(result RunResult, err error) (RunResult, error) {
	result.Status = "incomplete"
	result.Manifest.Complete = false
	if result.BaselineSnapshot != nil {
		result.BaselineSnapshot.Complete = false
	}
	result.Gate.Passed = false
	result.Errors = append(result.Errors, RunError{Message: err.Error()})
	return result, err
}

func (e *Engine) finish(result RunResult, err error) (RunResult, error) {
	words := 0
	for _, doc := range result.Documents {
		words += doc.ProseWords
	}
	if words == 0 && !e.allowEmpty && e.policy.Gate.FailOnEmpty {
		err = errors.Join(err, fmt.Errorf("scan contains no applicable English prose"))
	}
	if err != nil {
		return incomplete(result, err)
	}
	result.Gate.Passed = len(result.Gate.Reasons) == 0 || e.noGate
	return result, nil
}
