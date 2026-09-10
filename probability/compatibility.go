package probability

import (
	"fmt"
	"slices"

	"github.com/stokaro/unswell/nlp"
)

// Compatible reports whether a run computes this pack's columns identically.
// It compares the inputs that change measured values, not gate thresholds,
// severities, or report selection. An incompatible pack is never estimated.
func (p *Pack) Compatible(run Run) error {
	contract := p.file.Contract
	if run.FeatureContract != contract.FeatureContract || run.UnitContract != contract.UnitContract {
		return fmt.Errorf("probability pack %s expects feature contract %s and unit contract %s",
			p.file.ID, contract.FeatureContract, contract.UnitContract)
	}
	if run.PreparationHash != contract.PreparationHash {
		return fmt.Errorf("probability pack %s expects a different extraction and preparation policy", p.file.ID)
	}
	if run.IncludeQuotes != contract.IncludeQuotes || run.IncludeStructure != contract.IncludeStructure {
		return fmt.Errorf("probability pack %s expects include_quotes=%t and include_structure=%t",
			p.file.ID, contract.IncludeQuotes, contract.IncludeStructure)
	}
	if err := compatibleProvider(p.file.ID, contract.NLP, run.NLP); err != nil {
		return err
	}
	return compatibleCapabilities(p.file.ID, contract.Capabilities, run)
}

// compatibleProvider compares advertised provider identity without its
// capability list, which the requested set covers separately.
func compatibleProvider(id string, expected, actual nlp.Identity) error {
	expected.Capabilities, actual.Capabilities = nil, nil
	if !equalJSON(expected, actual) {
		return fmt.Errorf("probability pack %s expects NLP provider %s %s", id, expected.Name, expected.Version)
	}
	return nil
}

// compatibleCapabilities requires the exact requested set. A run with fewer or
// more representations measures different values, even with the same provider.
func compatibleCapabilities(id string, expected []nlp.Capability, run Run) error {
	actual := slices.Clone(run.Capabilities)
	required := slices.Clone(expected)
	slices.Sort(actual)
	slices.Sort(required)
	if !slices.Equal(actual, required) {
		return fmt.Errorf("probability pack %s expects the %v capabilities", id, expected)
	}
	for _, capability := range required {
		if !capabilityCovered(run.NLP.Capabilities, capability) {
			return fmt.Errorf("probability pack %s requires the %s capability from the provider", id, capability)
		}
	}
	return nil
}
