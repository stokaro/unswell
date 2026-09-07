package config_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
)

func TestSuppressionPolicyInheritanceAndFileOverrides(t *testing.T) {
	c := qt.New(t)
	plan, _, err := config.CompileBundle(config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
		"policy.yaml": []byte("version: 1\nextends: [base.yaml, builtin:technical]\nsuppressions: {reject_unused: false}\n" +
			"overrides:\n  - files: [contracts/**]\n    suppressions: {require_reason: false, allow_file_wide: false}\n"),
		"base.yaml": []byte("version: 1\nsuppressions: {allow_file_wide: true}\n"),
	}}, bundleCatalog())
	c.Assert(err, qt.IsNil)
	base, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(base.Suppressions, qt.Equals, config.Suppressions{RequireReason: true, AllowFileWide: true, RejectUnused: false})
	c.Assert(base.Origins["/suppressions/allow_file_wide"], qt.Equals, "base.yaml")
	selected, err := plan.ForFile("contracts/api.md")
	c.Assert(err, qt.IsNil)
	c.Assert(selected.Suppressions, qt.Equals, config.Suppressions{})
	c.Assert(selected.Origins["/suppressions/require_reason"], qt.Equals, "policy.yaml#overrides[0]")
	c.Assert(selected.Hash, qt.Not(qt.Equals), base.Hash)
	again, err := plan.ForFile("guide.md")
	c.Assert(err, qt.IsNil)
	c.Assert(again.Suppressions, qt.Equals, base.Suppressions)
}

func TestSuppressionSettingsAreStrict(t *testing.T) {
	for _, input := range []string{
		"suppressions: {require_reasons: false}",
		"suppressions: {reject_unused: maybe}",
		"suppressions: {allow_file_wide: null}",
		"suppressions: {require_reason: true, require_reason: false}",
		"overrides: [{files: ['*.md'], suppressions: {allow_all: true}}]",
	} {
		c := qt.New(t)
		_, _, err := config.CompileBundle(config.Bundle{Root: "policy.yaml", Files: map[string][]byte{
			"policy.yaml": []byte("version: 1\n" + input + "\n"),
		}}, bundleCatalog())
		c.Assert(err, qt.IsNotNil, qt.Commentf("%s", input))
	}
}
