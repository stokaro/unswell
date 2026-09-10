package config_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
)

func vocabularyRoot(extra string) []byte {
	return []byte("version: 1\nextends: [builtin:strict-v1]\n" + extra)
}

// Shared dictionaries and term exemptions decide which candidates a rule keeps,
// so a broken reference must stop compilation rather than silently shrink the
// vocabulary a policy claims to use.
func TestBrokenVocabularyReferencesFailCompilation(t *testing.T) {
	for _, row := range []struct {
		name  string
		files map[string][]byte
		want  string
	}{
		{"missing resource", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml]\n")},
			`missing dictionary resource "terms.yaml"`},
		{"unsupported version", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml]\n"),
			"terms.yaml":    []byte("version: 2\nterms: [control plane]\n")},
			"terms.yaml: dictionary requires version 1 and nonempty terms"},
		{"empty term list", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml]\n"),
			"terms.yaml":    []byte("version: 1\nterms: []\n")},
			"terms.yaml: dictionary requires version 1 and nonempty terms"},
		{"empty term", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml]\n"),
			"terms.yaml":    []byte("version: 1\nterms: ['']\n")},
			"terms.yaml: terms require 1 to 1000 valid UTF-8 bytes without protected markers"},
		{"duplicate key in a dictionary", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml]\n"),
			"terms.yaml":    []byte("version: 1\nversion: 1\nterms: [control plane]\n")},
			`terms.yaml: duplicate configuration key "version"`},
		{"same dictionary twice", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: [terms.yaml, ./terms.yaml]\n"),
			"terms.yaml":    []byte("version: 1\nterms: [control plane]\n")},
			`duplicate dictionary "terms.yaml"`},
		{"reference outside the project", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  dictionaries: ['../terms.yaml']\n")},
			`configuration reference escapes project root: "\.\./terms\.yaml"`},
		{"unknown exemption", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  term_exemptions: [no.such.rule]\n")},
			`unknown term exemption rule "no.such.rule"`},
		{"rule without term support", map[string][]byte{
			".unswell.yaml": vocabularyRoot("vocabulary:\n  term_exemptions: [syntax.long-sentence]\n")},
			"rule syntax.long-sentence does not support candidate-level term exemptions"},
		{"repeated exemption", map[string][]byte{
			".unswell.yaml": vocabularyRoot(
				"vocabulary:\n  term_exemptions: [policy.banned-phrases, policy.banned-phrases]\n")},
			`duplicate term exemption "policy.banned-phrases"`},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			_, _, err := config.CompileBundle(config.Bundle{Root: ".unswell.yaml", Files: row.files}, bundleCatalog())
			c.Assert(err, qt.ErrorMatches, row.want)
		})
	}
}

// A resolved term keeps the name of the file it came from so a reader can tell
// an inline term from a dictionary entry.
func TestResolvedTermsRecordTheirDictionary(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: ".unswell.yaml", Files: map[string][]byte{
		".unswell.yaml":       vocabularyRoot("vocabulary:\n  terms: [control plane]\n  dictionaries: [policies/terms.yaml]\n"),
		"policies/terms.yaml": []byte("version: 1\nterms: [robust estimator, retry budget]\n"),
	}}
	plan, _, err := config.CompileBundle(bundle, bundleCatalog())
	c.Assert(err, qt.IsNil)
	policy, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(policy.Vocabulary.ResolvedTerms, qt.DeepEquals, []string{"control plane", "robust estimator", "retry budget"})
	c.Assert(policy.Origins["/vocabulary/resolved_terms/0"], qt.Equals, ".unswell.yaml")
	c.Assert(policy.Origins["/vocabulary/resolved_terms/1"], qt.Equals, "policies/terms.yaml")
	c.Assert(policy.Origins["/vocabulary/resolved_terms/2"], qt.Equals, "policies/terms.yaml")
}
