package report

import (
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/feature"
	"github.com/stokaro/unswell/nlp"
)

func activationDefinitions(result unswell.RunResult) []feature.Descriptor {
	definitions := make([]feature.Descriptor, 0, len(result.Manifest.Rules))
	for _, d := range result.Manifest.Rules {
		definition := feature.ActivationDescriptor(d.ID)
		definition.Requires = slices.Clone(d.Requires)
		definitions = append(definitions, definition)
	}
	return definitions
}

func validateActivationContract(result unswell.RunResult, definitions []feature.Descriptor) error {
	requested := slices.ContainsFunc(definitions, func(d feature.Descriptor) bool { return d.Family == "rule-activation" })
	expected := ""
	if requested {
		expected = feature.ActivationContract
	}
	if result.Features.ActivationContract != expected {
		return fmt.Errorf("incompatible activation collection contract")
	}
	seen := make(map[string]bool)
	for _, d := range result.Manifest.Rules {
		if d.ID == "" || d.Version == "" || seen[d.ID] {
			return fmt.Errorf("invalid feature rule identity")
		}
		seen[d.ID] = true
	}
	return nil
}

func validateActivationSource(source unswell.FeatureSource, result unswell.RunResult) error {
	if result.Features.ActivationContract == "" {
		if source.RulesetHash != "" {
			return fmt.Errorf("unexpected activation source identity")
		}
		return nil
	}
	if !hashString(source.RulesetHash) || source.RulesetHash != result.Manifest.RulesetHash {
		return fmt.Errorf("inconsistent activation source identity")
	}
	return nil
}

func validateActivationValue(value feature.Value, d feature.Descriptor, capabilities []nlp.Capability, complete bool) error {
	if err := feature.ValidateActivation(value); err != nil {
		return err
	}
	if complete && (value.Reason == "not_evaluated" || value.Reason == "evaluation_failed") {
		return fmt.Errorf("complete run contains an unfinished activation")
	}
	computed := value.Number != nil || value.Reason == "applicability_unknown" || strings.HasPrefix(value.Reason, "inapplicable/")
	if computed {
		for _, required := range d.Requires {
			if !slices.Contains(capabilities, required) {
				return fmt.Errorf("activation requires missing capability %s", required)
			}
		}
	}
	return nil
}
