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

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
)

// Options injects all inputs. A nil Rules slice selects the builtin catalog;
// a nonnil slice is the complete registry. Config is YAML bytes, never a path.
type Options struct {
	Config        []byte
	Rules         []rule.Rule
	NLP           nlp.Provider
	Jobs          int
	IncludeSource bool
	AllowEmpty    bool
	NoGate        bool
}

// Engine is immutable after construction and supports concurrent calls. Custom
// rule and NLP implementations must uphold their documented concurrency contract.
type Engine struct {
	policy        config.Policy
	rules         []rule.Rule
	descriptors   []rule.Descriptor
	nlp           nlp.Provider
	capabilities  []nlp.Capability
	jobs          int
	includeSource bool
	allowEmpty    bool
	noGate        bool
	rulesetHash   string
}

// New validates the entire policy and required capabilities before reading prose.
func New(options Options) (*Engine, error) {
	registry := options.Rules
	if registry == nil {
		registry = builtin.Rules()
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
	policy, err := config.Load(options.Config, e.descriptors)
	if err != nil {
		return nil, err
	}
	e.policy = policy
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
	return e, nil
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
	for _, descriptor := range e.descriptors {
		if !e.policy.Rules[descriptor.ID].Enabled {
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
	result := e.emptyResult()
	sources = slices.Clone(sources)
	slices.SortFunc(sources, func(a, b document.Source) int { return strings.Compare(a.Name, b.Name) })
	if err := e.validateSources(sources); err != nil {
		return incomplete(result, "", err)
	}
	if err := ctx.Err(); err != nil {
		return incomplete(result, "", err)
	}
	partials := make([]RunResult, len(sources))
	errorsBySource := make([]error, len(sources))
	var workers sync.WaitGroup
	for worker := 0; worker < min(e.jobs, len(sources)); worker++ {
		workers.Go(func() {
			for index := worker; index < len(sources); index += e.jobs {
				partials[index], errorsBySource[index] = e.analyzeSource(ctx, sources[index])
			}
		})
	}
	workers.Wait()
	for i, part := range partials {
		result.Documents = append(result.Documents, part.Documents...)
		result.Findings = append(result.Findings, part.Findings...)
		result.Assessments = append(result.Assessments, part.Assessments...)
		result.Gate.Reasons = append(result.Gate.Reasons, part.Gate.Reasons...)
		if errorsBySource[i] != nil {
			result.Errors = append(result.Errors, RunError{Path: sources[i].Name, Message: errorsBySource[i].Error()})
		}
	}
	return e.finish(result, errors.Join(errorsBySource...))
}

func (e *Engine) emptyResult() RunResult {
	return RunResult{
		SchemaVersion: SchemaVersion,
		Status:        "complete",
		Documents:     []DocumentResult{},
		Findings:      []Finding{},
		Assessments:   []Assessment{},
		Errors:        []RunError{},
		Gate:          GateDecision{Passed: true, Reasons: []GateReason{}},
		Manifest: Manifest{
			ToolVersion:     Version,
			ToolCommit:      BuildCommit,
			ConfigHash:      e.policy.Hash,
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
}

func (e *Engine) validateSources(sources []document.Source) error {
	total := 0
	for i, source := range sources {
		if i > 0 && sources[i-1].Name == source.Name {
			return fmt.Errorf("duplicate source name %q", source.Name)
		}
		if len(source.Bytes) > e.policy.Analysis.MaxFileBytes {
			return fmt.Errorf("%s exceeds max_file_bytes", source.Name)
		}
		total += len(source.Bytes)
		if total > e.policy.Analysis.MaxTotalBytes {
			return fmt.Errorf("sources exceed max_total_bytes")
		}
	}
	return nil
}

func incomplete(result RunResult, path string, err error) (RunResult, error) {
	result.Status = "incomplete"
	result.Manifest.Complete = false
	result.Gate.Passed = false
	result.Errors = append(result.Errors, RunError{Path: path, Message: err.Error()})
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
		return incomplete(result, "", err)
	}
	result.Gate.Passed = len(result.Gate.Reasons) == 0 || e.noGate
	return result, nil
}
