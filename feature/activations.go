package feature

import (
	"fmt"
	"math"
	"strings"
)

// ActivationContract identifies observed block activations and applicability.
const ActivationContract = "unswell-rule-activations-v1"

// BlockObservation reports applicability discovered during a rule's computation.
// Status is evaluated or inapplicable. Only inapplicable requires a reason.
type BlockObservation struct {
	BlockID int
	Status  string
	Reason  string
}

// ActivationBuilder collects one rule evaluation. Calls must be sequential.
// All errors latch; Values cannot turn a failed collection into valid data.
type ActivationBuilder struct {
	blocks []activationBlock
	err    error
}

type activationBlock struct {
	observation BlockObservation
	observed    bool
	matched     bool
	maximum     int
}

// NewActivations bounds allocation using the caller's extracted block limit.
func NewActivations(blocks, maximum int) (*ActivationBuilder, error) {
	if blocks < 0 || maximum < 1 || blocks > maximum {
		return nil, fmt.Errorf("invalid activation block count or limit")
	}
	return &ActivationBuilder{blocks: make([]activationBlock, blocks)}, nil
}

// ActivationDescriptor defines the value for a registered rule ID. The caller
// supplies and validates the rule catalog, effective policy, and rule version.
func ActivationDescriptor(ruleID string) Descriptor {
	return Descriptor{ID: "activation/" + ruleID, Version: "1", Family: "rule-activation", Type: "number", Unit: "ratio", Scope: "block",
		Formula:       "Maximum raw evidence activation among this rule's occurrences in the block, divided by 1000.",
		Normalization: "Uses the existing fixed-point evidence; no weights, caps, suppressions, or derived findings.",
		Missing:       "Disabled, incomplete, inapplicable, unobserved, or unreached evaluation never establishes zero.",
		Limitations:   "A policy-dependent signal, not a probability or a quality label. Rule capabilities and versions remain required."}
}

// Observe records a block once. It rejects conflicting evidence and invalid states.
func (b *ActivationBuilder) Observe(observation BlockObservation) error {
	if b.err != nil {
		return b.err
	}
	b.err = b.observe(observation)
	return b.err
}

func (b *ActivationBuilder) observe(observation BlockObservation) error {
	if observation.BlockID < 0 || observation.BlockID >= len(b.blocks) {
		return fmt.Errorf("invalid observed block ID")
	}
	if !validObservation(observation) {
		return fmt.Errorf("invalid block observation state or reason")
	}
	block := &b.blocks[observation.BlockID]
	if block.observed || (block.matched && observation.Status == "inapplicable") {
		return fmt.Errorf("duplicate or contradictory block observation")
	}
	block.observation, block.observed = observation, true
	return nil
}

func validObservation(observation BlockObservation) bool {
	if observation.Status == "evaluated" {
		return observation.Reason == ""
	}
	return observation.Status == "inapplicable" && ValidApplicabilityReason(observation.Reason)
}

// ValidApplicabilityReason accepts bounded lowercase machine identifiers.
func ValidApplicabilityReason(reason string) bool {
	if len(reason) == 0 || len(reason) > 64 || reason[0] < 'a' || reason[0] > 'z' {
		return false
	}
	return strings.IndexFunc(reason, func(r rune) bool { return (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' }) < 0
}

// Record consumes already validated fixed-point evidence for one affected block.
// Repeated occurrences and duplicate evidence cannot increase a maximum.
func (b *ActivationBuilder) Record(blockID, activation int) error {
	if b.err != nil {
		return b.err
	}
	b.err = b.record(blockID, activation)
	return b.err
}

func (b *ActivationBuilder) record(blockID, activation int) error {
	if blockID < 0 || blockID >= len(b.blocks) || activation < 0 || activation > 1000 {
		return fmt.Errorf("invalid activation block ID or value")
	}
	block := &b.blocks[blockID]
	if block.observed && block.observation.Status == "inapplicable" {
		return fmt.Errorf("evidence contradicts an inapplicable block")
	}
	block.matched = true
	block.maximum = max(block.maximum, activation)
	return nil
}

// Values returns owned values after a successful rule evaluation. When full is
// true, every block must have an explicit observation, including matched blocks.
// The caller must discard these values if the rule evaluation itself failed.
func (b *ActivationBuilder) Values(ruleID string, full bool) ([]Value, error) {
	if b.err != nil {
		return nil, b.err
	}
	d := ActivationDescriptor(ruleID)
	values := make([]Value, len(b.blocks))
	for i, block := range b.blocks {
		if full && !block.observed {
			b.err = fmt.Errorf("missing declared block observation for %d", i)
			return nil, b.err
		}
		values[i] = activationValue(d, block)
	}
	return values, nil
}

func activationValue(d Descriptor, block activationBlock) Value {
	v := Value{ID: d.ID, Version: d.Version, Unit: d.Unit}
	switch {
	case block.observed && block.observation.Status == "inapplicable":
		v.Reason = "inapplicable/" + block.observation.Reason
	case block.observed || block.matched:
		value := float64(block.maximum) / 1000
		v.Number = &value
	default:
		v.Reason = "applicability_unknown"
	}
	return v
}

// ValidateActivation checks the formula version, range, and numeric/absence pair.
// Source, rule, capability, and completed-evaluation identities remain required.
func ValidateActivation(value Value) error {
	id, selected := strings.CutPrefix(value.ID, "activation/")
	if !selected || id == "" || value.Version != "1" || value.Unit != "ratio" {
		return fmt.Errorf("incompatible activation definition")
	}
	if value.Number == nil {
		if !validActivationAbsence(value.Reason) {
			return fmt.Errorf("invalid activation absence reason")
		}
		return nil
	}
	return validateActivationNumber(*value.Number, value.Reason)
}

func validateActivationNumber(number float64, reason string) error {
	if reason != "" || math.IsNaN(number) || math.IsInf(number, 0) || number < 0 || number > 1 {
		return fmt.Errorf("invalid activation value")
	}
	return nil
}

func validActivationAbsence(reason string) bool {
	switch reason {
	case "disabled", "not_evaluated", "evaluation_failed", "applicability_unknown":
		return true
	default:
		value, inapplicable := strings.CutPrefix(reason, "inapplicable/")
		return inapplicable && ValidApplicabilityReason(value)
	}
}
