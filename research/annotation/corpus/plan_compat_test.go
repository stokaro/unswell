package corpus_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

// A manifest frozen before a policy field existed omits that field, and
// canonicalization must leave the omission alone. The plan digest covers the
// canonical manifest. Writing a default there renumbers every plan on disk and
// every artifact that cites one. The extractor already reads an empty value as
// that same default.
func TestPlanIdentitySurvivesAnAddedPolicyField(t *testing.T) {
	c := qt.New(t)
	manifest, _ := sample()
	c.Assert(manifest.Policy.GitHubActions, qt.Equals, "")

	plan, err := corpus.MakePlan(t.Context(), manifest)
	c.Assert(err, qt.IsNil)
	c.Assert(plan.Manifest.Policy.GitHubActions, qt.Equals, "",
		qt.Commentf("canonicalization must not write a default into the manifest it hashes"))
	c.Assert(corpus.ValidatePlan(t.Context(), plan), qt.IsNil)

	// Stating the default explicitly names the same plan, which is what the
	// field's own test asks for, and it carries the identity the omitted form
	// already had rather than a new one.
	stated := manifest
	stated.Policy.GitHubActions = "shell"
	other, err := corpus.MakePlan(t.Context(), stated)
	c.Assert(err, qt.IsNil)
	c.Assert(corpus.ValidatePlan(t.Context(), other), qt.IsNil)
	c.Assert(other.ManifestSHA256, qt.Equals, plan.ManifestSHA256)

	// A mode that is not the default keeps an identity of its own.
	strings := manifest
	strings.Policy.GitHubActions = "strings"
	third, err := corpus.MakePlan(t.Context(), strings)
	c.Assert(err, qt.IsNil)
	c.Assert(third.ManifestSHA256, qt.Not(qt.Equals), plan.ManifestSHA256)
}
