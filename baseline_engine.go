package unswell

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/config"
)

func (e *Engine) configureBaseline(options Options) error {
	if options.GateMode != "" && options.GateMode != "all" && options.GateMode != "new" {
		return fmt.Errorf("gate mode must be all or new")
	}
	e.gateMode = options.GateMode
	e.collectBaseline = options.CollectBaseline || options.Baseline != nil
	if options.Baseline != nil {
		file, err := baseline.Load(context.Background(), options.Baseline)
		if err != nil {
			return err
		}
		e.baselineFile = &file
	}
	if e.selectedGateMode() == "new" && e.baselineFile == nil {
		return fmt.Errorf("gate.mode new requires a baseline")
	}
	if !e.collectBaseline {
		return nil
	}
	hasher := &debtHasher{}
	e.baselineCompatibility = baseline.Compatibility{
		PolicyHash: hasher.hash(baselinePolicy(e.policy)), RulesHash: e.baselineRulesHash(hasher),
		NLPHash: hasher.hash(e.nlp.Identity()), FeatureContract: "unswell-features-v1",
		ScoringContract: "unswell-local-fixed-point-v1", ModelHash: e.probabilityModelHash(hasher),
	}
	return hasher.err
}

func (e *Engine) selectedGateMode() string {
	if e.gateMode != "" {
		return e.gateMode
	}
	return e.policy.Gate.Mode
}

// Baseline compatibility follows effective behavior, not YAML spelling or the
// origin of an override. The complete provenance remains in the run manifest.
func baselinePolicy(policy config.Policy) config.Policy {
	policy.Gate.Mode = ""
	policy.Hash, policy.Identity = "", ""
	policy.Sources, policy.Overrides, policy.AppliedOverrides = nil, nil, nil
	policy.Origins, policy.RuleSets = nil, nil
	policy.Rules = maps.Clone(policy.Rules)
	for id, settings := range policy.Rules {
		settings.Severity = ""
		policy.Rules[id] = settings
	}
	return policy
}

func (e *Engine) baselineRulesHash(hasher *debtHasher) string {
	descriptors := cloneDescriptors(e.descriptors)
	for i := range descriptors {
		descriptor := &descriptors[i]
		descriptor.Summary, descriptor.Description, descriptor.Limitations, descriptor.Status = "", "", "", ""
		descriptor.Examples = nil
		descriptor.Defaults.Severity = ""
	}
	return hasher.hash(descriptors)
}

type debtHasher struct{ err error }

func (h *debtHasher) hash(value any) string {
	if h.err != nil {
		return ""
	}
	data, err := json.Marshal(value)
	if err != nil {
		h.err = err
		return ""
	}
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func (e *Engine) finishBaseline(ctx context.Context, result RunResult, analysisErr error) (RunResult, error) {
	result, err := e.finish(result, analysisErr)
	if err != nil || result.BaselineSnapshot == nil {
		return result, err
	}
	if e.baselineFile == nil {
		_, err = baseline.Create(ctx, *result.BaselineSnapshot)
	} else {
		var comparison baseline.Comparison
		comparison, err = baseline.Compare(ctx, *e.baselineFile, *result.BaselineSnapshot)
		if err == nil {
			result.Baseline = &comparison
			applyBaseline(&result, comparison)
		}
	}
	if err != nil {
		return incomplete(result, err)
	}
	result.Gate.Passed = len(result.Gate.Reasons) == 0 || e.noGate
	return result, nil
}

func applyBaseline(result *RunResult, comparison baseline.Comparison) {
	states := make(map[string]string, len(comparison.Matches))
	for _, match := range comparison.Matches {
		states[match.Fingerprint] = match.State
	}
	newOnly := make(map[string]bool, len(result.Documents))
	for _, doc := range result.Documents {
		newOnly[doc.Name] = doc.GateMode == "new"
	}
	accepted := make(map[string]bool)
	for i := range result.Findings {
		finding := &result.Findings[i]
		if state := states[finding.BaselineFingerprint]; state != "" && !finding.Suppressed {
			finding.BaselineState = state
			accepted[finding.ID] = state == "existing" && newOnly[finding.Primary.Path]
		}
	}
	for i := range result.Assessments {
		assessment := &result.Assessments[i]
		if state := states[assessment.BaselineFingerprint]; state != "" {
			assessment.BaselineState = state
		}
	}
	result.Gate.Reasons = slices.DeleteFunc(result.Gate.Reasons, func(reason GateReason) bool {
		if accepted[reason.FindingID] {
			result.Gate.Accepted = append(result.Gate.Accepted, reason)
			return true
		}
		return false
	})
}
