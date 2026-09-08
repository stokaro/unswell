package unswell

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
)

func (e *Engine) planCapabilities() error {
	e.capabilities = nil
	identity := e.nlp.Identity()
	if err := e.requireCapabilities([]nlp.Capability{nlp.Tokens, nlp.Sentences}, identity.Capabilities, "NLP provider"); err != nil {
		return err
	}
	enabled := e.plan.EnabledRuleIDs()
	for _, descriptor := range e.descriptors {
		if !slices.Contains(enabled, descriptor.ID) {
			continue
		}
		if err := e.requireCapabilities(descriptor.Requires, identity.Capabilities, "rule "+descriptor.ID); err != nil {
			return err
		}
		if descriptor.DependencyScheme != "" && descriptor.DependencyScheme != identity.DependencyScheme {
			return fmt.Errorf("rule %s requires dependency scheme %q; provider declares %q",
				descriptor.ID, descriptor.DependencyScheme, identity.DependencyScheme)
		}
	}
	if slices.Contains(e.capabilities, nlp.Dependencies) && strings.TrimSpace(identity.DependencyScheme) == "" {
		return fmt.Errorf("NLP provider requires a dependency scheme for requested dependencies")
	}
	slices.Sort(e.capabilities)
	return nil
}

func (e *Engine) requireCapabilities(required, available []nlp.Capability, owner string) error {
	for _, capability := range required {
		if !slices.Contains(available, capability) {
			return fmt.Errorf("%s requires unavailable capability %s", owner, capability)
		}
		if !slices.Contains(e.capabilities, capability) {
			e.capabilities = append(e.capabilities, capability)
		}
	}
	return nil
}

func (e *Engine) validateDependencies(ctx context.Context, mapped document.MappedText, sentences []document.Sentence) error {
	if !slices.Contains(e.capabilities, nlp.Dependencies) {
		return nil
	}
	return nlp.ValidateDependencies(ctx, mapped, sentences)
}
