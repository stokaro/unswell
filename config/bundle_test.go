package config_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/rule"
)

func bundleCatalog() []rule.Descriptor {
	var result []rule.Descriptor
	for _, implementation := range builtin.Rules() {
		result = append(result, implementation.Descriptor())
	}
	return result
}

func TestOrderedInheritanceAndFilePolicy(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: ".unswell.yaml", Files: map[string][]byte{
		".unswell.yaml": []byte(`version: 1
extends: [policies/base.yaml, policies/second.yaml]
rules:
  policy.banned-phrases:
    parameters: {phrases: [robust]}
gate: {paragraph_score: {fail_at: 60}}
overrides:
  - files: [docs/**]
    gate: {paragraph_score: {fail_at: 45}}
    analysis: {include_quotes: false}
  - files: [docs/reference/**]
    gate: {paragraph_score: {fail_at: 20}}
    rules:
      policy.banned-phrases: {parameters: {phrases: []}}
`),
		"policies/base.yaml": []byte(`version: 1
extends: [builtin:strict-v1]
rules:
  policy.banned-phrases:
    parameters: {phrases: [old phrase]}
    score: {weight: 25, cap: 40}
gate: {paragraph_score: {fail_at: 55, min_words: 70}}
analysis: {include_quotes: true}
vocabulary:
  dictionaries: [terms.yaml]
  term_exemptions: [policy.banned-phrases]
`),
		"policies/second.yaml": []byte("version: 1\ngate: {paragraph_score: {min_words: 31}}\n"),
		"policies/terms.yaml":  []byte("version: 1\nterms: [robust estimator, control plane]\n"),
	}}
	plan, _, err := config.CompileBundle(bundle, bundleCatalog())
	c.Assert(err, qt.IsNil)
	base, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(base.Gate.Paragraph, qt.Equals, config.Threshold{FailAt: 60, MinWords: 31})
	c.Assert(base.Rules["policy.banned-phrases"].Score, qt.Equals, rule.Score{Weight: 25, Cap: 40})
	c.Assert(base.Rules["policy.banned-phrases"].Parameters.Phrases, qt.DeepEquals, []string{"robust"})
	c.Assert(base.Vocabulary.ResolvedTerms, qt.DeepEquals, []string{"robust estimator", "control plane"})
	c.Assert(base.Origins["/gate/paragraph_score/min_words"], qt.Equals, "policies/second.yaml")
	c.Assert(base.Origins["/rules/policy.banned-phrases/score/weight"], qt.Equals, "policies/base.yaml")
	c.Assert(base.Origins["/vocabulary/resolved_terms/0"], qt.Equals, "policies/terms.yaml")
	reference, err := plan.ForFile("docs/reference/api.md")
	c.Assert(err, qt.IsNil)
	c.Assert(reference.Gate.Paragraph, qt.Equals, config.Threshold{FailAt: 20, MinWords: 31})
	c.Assert(reference.Analysis.IncludeQuotes, qt.IsFalse)
	c.Assert(reference.Rules["policy.banned-phrases"].Parameters.Phrases, qt.HasLen, 0)
	c.Assert(reference.AppliedOverrides, qt.HasLen, 2)
	c.Assert(reference.Origins["/gate/paragraph_score/fail_at"], qt.Equals, ".unswell.yaml#overrides[1]")
	c.Assert(reference.Hash, qt.Not(qt.Equals), base.Hash)
	unchanged, err := plan.ForFile("src/api.go")
	c.Assert(err, qt.IsNil)
	c.Assert(unchanged, qt.DeepEquals, base)
}

func TestConfigGraphValidation(t *testing.T) {
	cases := []struct{ name, root string }{
		{"cycle", "extends: [.unswell.yaml]"},
		{"normalized duplicate", "extends: [base.yaml, ./base.yaml]"},
		{"builtin duplicate", "extends: [builtin:technical, builtin:technical-v1]"},
		{"missing", "extends: [missing.yaml]"},
		{"parent escape", "extends: [../base.yaml]"},
		{"URL", "extends: ['https://example.com/config.yaml']"},
		{"Windows path", "extends: ['C:\\base.yaml']"},
		{"absolute", "extends: [/base.yaml]"},
		{"null list", "extends: null"},
		{"alias", "extends: &parent [builtin:technical]\nvocabulary: {terms: *parent}"},
		{"duplicate nested", "gate: {paragraph_score: {fail_at: 40, fail_at: 50}}"},
		{"unknown override", "overrides: [{files: ['*.md'], rules: {missing.rule: {enabled: false}}}]"},
		{"unknown override field", "overrides: [{files: ['*.md'], analysis: {max_words: 10}}]"},
		{"global override", "overrides: [{files: ['*.md'], analysis: {max_total_bytes: 100}}]"},
		{"discovery override", "overrides: [{files: ['*.md'], extends: [base.yaml]}]"},
		{"invalid glob", "overrides: [{files: ['../*.md']}]"},
		{"unsupported exemption", "vocabulary: {term_exemptions: [syntax.long-sentence]}"},
		{"duplicate exemption", "vocabulary: {term_exemptions: [policy.banned-phrases, policy.banned-phrases]}"},
		{"duplicate terms", "vocabulary: {terms: [Control Plane, control plane]}"},
		{"output-only terms", "vocabulary: {resolved_terms: [control plane]}"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			_, _, err := config.CompileBundle(config.Bundle{Root: ".unswell.yaml", Files: map[string][]byte{
				".unswell.yaml": []byte("version: 1\n" + tc.root + "\n"), "base.yaml": []byte("version: 1\n"),
			}}, bundleCatalog())
			c.Assert(err, qt.IsNotNil)
		})
	}
}

func TestDictionaryIdentityAndOwnership(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
		"policy.yaml": []byte("version: 1\nvocabulary: {dictionaries: [terms.yaml]}\n"),
		"terms.yaml":  []byte("version: 1\nterms: [control plane]\n"),
	}}
	catalog := bundleCatalog()
	plan, _, err := config.CompileBundle(bundle, catalog)
	c.Assert(err, qt.IsNil)
	base, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	bundle.Files["terms.yaml"] = []byte("# same meaning\nterms: ['control plane']\nversion: 1\n")
	same, _, err := config.CompileBundle(bundle, catalog)
	c.Assert(err, qt.IsNil)
	samePolicy, err := same.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(samePolicy.Hash, qt.Equals, base.Hash)
	bundle.Files["terms.yaml"] = []byte("version: 1\nterms: [data plane]\n")
	changed, _, err := config.CompileBundle(bundle, catalog)
	c.Assert(err, qt.IsNil)
	changedPolicy, err := changed.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(changedPolicy.Hash, qt.Not(qt.Equals), base.Hash)
	base.Vocabulary.ResolvedTerms[0] = "mutated"
	base.Origins["/profile"] = "mutated"
	catalog[0].Parameters[0] = "mutated"
	again, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(again.Vocabulary.ResolvedTerms, qt.DeepEquals, []string{"control plane"})
	c.Assert(again.Origins["/profile"], qt.Equals, "builtin:technical-v1")
}

func TestSelectedOverridesRejectIncompatibleCombination(t *testing.T) {
	c := qt.New(t)
	data := []byte(`version: 1
overrides:
  - files: [docs/**]
    rules: {syntax.long-sentence: {parameters: {onset: 50}}}
  - files: ['**/*.md']
    rules: {syntax.long-sentence: {parameters: {saturation: 40}}}
`)
	plan, _, err := config.CompileBundle(config.Bundle{Root: "policy.yaml", Files: map[string][]byte{"policy.yaml": data}}, bundleCatalog())
	c.Assert(err, qt.IsNil)
	_, err = plan.ForFile("docs/api.md")
	c.Assert(err, qt.ErrorMatches, ".*invalid saturation parameter")
	_, err = plan.ForFile("api.md")
	c.Assert(err, qt.IsNil)
}

func TestBundleResourceLimits(t *testing.T) {
	c := qt.New(t)
	_, _, err := config.CompileBundle(config.Bundle{Root: "a.yaml", Files: map[string][]byte{
		"a.yaml": []byte("version: 1\n"), "./a.yaml": []byte("version: 1\n"),
	}}, nil)
	c.Assert(err, qt.ErrorMatches, ".*duplicate configuration resource.*")
	_, _, err = config.CompileBundle(config.Bundle{Root: "a.yaml", Files: map[string][]byte{
		"a.yaml": []byte(strings.Repeat(" ", (1<<20)+1)),
	}}, nil)
	c.Assert(err, qt.ErrorMatches, ".*exceed 1 MiB.*")
	outside := config.Bundle{Root: "a.yaml", AllowOutsideRoot: true, Files: map[string][]byte{
		"a.yaml": []byte("version: 1\nextends: [../shared.yaml]\n"), "../shared.yaml": []byte("version: 1\n"),
	}}
	_, _, err = config.CompileBundle(outside, nil)
	c.Assert(err, qt.IsNil)
	outside.AllowOutsideRoot = false
	_, _, err = config.CompileBundle(outside, nil)
	c.Assert(err, qt.IsNotNil)
}

func TestReplacedOverridesRemainStrictAndListsReplace(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
		"policy.yaml": []byte("version: 1\nextends: [base.yaml]\noverrides: []\n"),
		"base.yaml": []byte("version: 1\noverrides:\n  - files: ['*.md']\n" +
			"    rules: {policy.banned-phrases: {enabled: false}}\n"),
	}}
	plan, _, err := config.CompileBundle(bundle, bundleCatalog())
	c.Assert(err, qt.IsNil)
	policy, err := plan.ForFile("draft.md")
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Rules["policy.banned-phrases"].Enabled, qt.IsTrue)
	c.Assert(policy.AppliedOverrides, qt.HasLen, 0)
	bundle.Files["base.yaml"] = []byte("version: 1\noverrides:\n  - files: ['*.md']\n" +
		"    rules: {missing.rule: {enabled: false}}\n")
	_, _, err = config.CompileBundle(bundle, bundleCatalog())
	c.Assert(err, qt.ErrorMatches, ".*unknown rule ID.*")
}
