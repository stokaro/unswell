package unswell_test

import (
	"context"
	"slices"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/nlp"
	"github.com/stokaro/unswell/nlp/english"
	"github.com/stokaro/unswell/rule"
)

type dependencyContractRule struct {
	scheme   string
	calls    *int
	requires []nlp.Capability
}

func (r dependencyContractRule) Descriptor() rule.Descriptor {
	required := r.requires
	if required == nil {
		required = []nlp.Capability{nlp.Dependencies}
	}
	return rule.Descriptor{ID: "test.dependency-contract", Version: "1", Group: "test", Scope: "sentence",
		Requires: required, DependencyScheme: r.scheme,
		Defaults: rule.Settings{Enabled: true, Severity: "warning", Gate: "none"}}
}

func (r dependencyContractRule) Evaluate(_ context.Context, _ rule.View, _ rule.Emitter) error {
	*r.calls++
	return nil
}

type dependencyContractNLP struct {
	nlp.Provider
	identity nlp.Identity
	empty    bool
}

func (p dependencyContractNLP) Identity() nlp.Identity { return p.identity }

func (p dependencyContractNLP) Analyze(ctx context.Context, mapped document.MappedText, _ []nlp.Capability) ([]document.Sentence, error) {
	// Deliberately violates its advertised capability: returns the real surface
	// backend's output without a dependency parse. The engine must reject it.
	if p.empty {
		return []document.Sentence{}, nil
	}
	return p.Provider.Analyze(ctx, mapped, []nlp.Capability{nlp.Tokens, nlp.Sentences})
}

func TestDependencyContractRejectsMissingAnalysis(t *testing.T) {
	for _, row := range []struct {
		name          string
		noGate, empty bool
		err           string
	}{
		{"missing tree", false, false, "(?s).*requested dependency tree is missing.*"},
		{"missing tree without gate", true, false, "(?s).*requested dependency tree is missing.*"},
		{"no sentences", false, true, "(?s).*dependency sentences omit extracted prose.*"},
		{"no sentences without gate", true, true, "(?s).*dependency sentences omit extracted prose.*"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			base, err := english.New()
			c.Assert(err, qt.IsNil)
			identity := base.Identity()
			identity.Capabilities = append(identity.Capabilities, nlp.Dependencies)
			identity.DependencyScheme = "test-v1"
			calls := 0
			engine, err := unswell.New(unswell.Options{Rules: []rule.Rule{dependencyContractRule{scheme: "test-v1", calls: &calls}},
				NLP: dependencyContractNLP{Provider: base, identity: identity, empty: row.empty}, NoGate: row.noGate})
			c.Assert(err, qt.IsNil)
			result, err := engine.Analyze(t.Context(), document.Source{Name: "a.txt", Format: document.Plain,
				Bytes: []byte("The request was rejected.")})
			c.Assert(err, qt.ErrorMatches, row.err)
			c.Assert(result.Gate.Passed, qt.IsFalse)
			c.Assert(result.Manifest.Complete, qt.IsFalse)
			c.Assert(calls, qt.Equals, 0)
		})
	}
}

func TestDependencyCapabilityPlanning(t *testing.T) {
	for _, row := range []struct {
		name       string
		capability nlp.Capability
		scheme     string
		ruleScheme string
		err        string
	}{
		{"missing dependencies", nlp.Dependencies, "", "test-v1", ".*unavailable capability dependencies"},
		{"missing tokens", nlp.Tokens, "test-v1", "test-v1", ".*unavailable capability tokens"},
		{"missing sentences", nlp.Sentences, "test-v1", "test-v1", ".*unavailable capability sentences"},
		{"missing scheme", "", "", "", ".*requires a dependency scheme.*"},
		{"different scheme", "", "ud-v2", "spacy-en-3.8", ".*requires dependency scheme.*"},
		{"compatible", "", "ud-v2", "ud-v2", ""},
		{"structure only", "", "ud-v2", "", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			base, err := english.New()
			c.Assert(err, qt.IsNil)
			identity := base.Identity()
			identity.DependencyScheme = row.scheme
			identity.Capabilities = []nlp.Capability{nlp.Tokens, nlp.Sentences, nlp.Dependencies}
			identity.Capabilities = slices.DeleteFunc(identity.Capabilities, func(value nlp.Capability) bool { return value == row.capability })
			calls := 0
			_, err = unswell.New(unswell.Options{Rules: []rule.Rule{dependencyContractRule{scheme: row.ruleScheme, calls: &calls}},
				NLP: dependencyContractNLP{Provider: base, identity: identity}})
			if row.err == "" {
				c.Assert(err, qt.IsNil)
			} else {
				c.Assert(err, qt.ErrorMatches, row.err)
			}
		})
	}
}

func TestDependencySchemeRequiresDeclaration(t *testing.T) {
	c := qt.New(t)
	calls := 0
	_, err := unswell.New(unswell.Options{Rules: []rule.Rule{dependencyContractRule{
		scheme: "ud-v2", calls: &calls, requires: []nlp.Capability{},
	}}})
	c.Assert(err, qt.ErrorMatches, ".*declares a dependency scheme without requiring dependencies")
}
