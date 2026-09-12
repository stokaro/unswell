package unswell

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/rule"
)

func (e *Engine) runRule(
	ctx context.Context, implementation rule.Rule, view rule.View, terms *rule.TermMatches, result *RunResult,
) error {
	descriptor := implementation.Descriptor()
	settings := e.policy.Rules[descriptor.ID]
	if !settings.Enabled {
		return nil
	}
	emitter := &collector{ctx: ctx, doc: view.Document, descriptor: descriptor, settings: settings,
		limit: e.policy.Analysis.MaxFindings - len(result.Findings), includeSource: e.includeSource}
	if _, requested := e.activationIndices[descriptor.ID]; requested {
		builder, err := feature.NewActivations(len(view.Document.Blocks), e.policy.Analysis.MaxBlocks)
		if err != nil {
			return err
		}
		emitter.activations, view.Observer = builder, emitter
	}
	view.Parameters, view.MaxCandidates = cloneParameters(settings.Parameters), e.policy.Analysis.MaxCandidates
	if slices.Contains(e.policy.Vocabulary.TermExemptions, descriptor.ID) {
		view.TermExemptions = terms
	}
	err := implementation.Evaluate(ctx, view, emitter)
	if err == nil {
		err = ctx.Err()
	}
	if abstention, declared := declaredAbstention(ctx, emitter, err); declared {
		e.abstainRule(result, emitter, view.Document.Name, abstention)
		return nil
	}
	result.Findings = append(result.Findings, emitter.findings...)
	return e.finishRuleFeatures(result, emitter, evaluationError(emitter, err))
}

// declaredAbstention recognizes a rule that declined the document for a valid
// reason. Cancellation, emitter failures, and undeclared reasons keep the
// fatal path, so a rule cannot hide an operational failure behind an abstention.
func declaredAbstention(ctx context.Context, emitter *collector, err error) (rule.Abstention, bool) {
	var abstention *rule.Abstention
	if ctx.Err() != nil || emitter.err != nil || !errors.As(err, &abstention) {
		return rule.Abstention{}, false
	}
	if !feature.ValidApplicabilityReason(abstention.Reason) {
		return rule.Abstention{}, false
	}
	return *abstention, true
}

// abstainRule records the abstention and drops the rule's partial evidence.
// A truncated rule's findings depend on block order and on the budget, so
// they must not reach the gate, the baseline, or the activation values.
func (e *Engine) abstainRule(result *RunResult, emitter *collector, path string, abstention rule.Abstention) {
	result.Abstentions = append(result.Abstentions, RuleAbstention{Path: path, RuleID: emitter.descriptor.ID,
		RuleVersion: emitter.descriptor.Version, Reason: abstention.Reason, Detail: abstention.Detail()})
	if emitter.activations != nil {
		e.absentRuleFeatures(result, emitter, "inapplicable/"+abstention.Reason)
	}
}

func evaluationError(emitter *collector, err error) error {
	if emitter.err != nil {
		return fmt.Errorf("%s emitted invalid evidence: %w", emitter.descriptor.ID, emitter.err)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", emitter.descriptor.ID, err)
	}
	return nil
}

func (e *Engine) finishRuleFeatures(result *RunResult, emitter *collector, evaluationErr error) error {
	if emitter.activations == nil {
		return evaluationErr
	}
	if evaluationErr != nil {
		e.absentRuleFeatures(result, emitter, "evaluation_failed")
		return evaluationErr
	}
	values, err := emitter.activations.Values(emitter.descriptor.ID, emitter.descriptor.BlockObservations)
	if err != nil {
		e.absentRuleFeatures(result, emitter, "evaluation_failed")
		return fmt.Errorf("%s activation collection: %w", emitter.descriptor.ID, err)
	}
	column := e.activationIndices[emitter.descriptor.ID]
	units := result.Features.Sources[0].Units
	for i := range units {
		units[i].Values[column] = values[i]
	}
	return nil
}

// absentRuleFeatures replaces the rule's whole activation column with one
// absence reason, discarding any value collected before the rule stopped.
func (e *Engine) absentRuleFeatures(result *RunResult, emitter *collector, reason string) {
	column := e.activationIndices[emitter.descriptor.ID]
	units := result.Features.Sources[0].Units
	for i := range units {
		units[i].Values[column].Number, units[i].Values[column].Reason = nil, reason
	}
}

// Observe shares the emitter's latched failure state for optional observations.
func (c *collector) Observe(observation feature.BlockObservation) error {
	if c.err != nil {
		return c.err
	}
	if c.activations == nil {
		return nil
	}
	if c.err = c.ctx.Err(); c.err != nil {
		return c.err
	}
	c.err = c.activations.Observe(observation)
	return c.err
}

func (c *collector) observeOccurrences(evidence rule.Evidence) error {
	if c.activations == nil {
		return nil
	}
	for _, occurrence := range evidence.Occurrences {
		if err := c.activations.Record(occurrence.BlockID, evidence.Activation); err != nil {
			return err
		}
	}
	return nil
}
