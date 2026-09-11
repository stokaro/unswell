package config_test

import (
	"reflect"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
)

// A plan resolves one override combination once. Resolution shares that
// result between files of the same combination, ForFile hands out owned copies
// of it, and a file outside the override carries the base resolution.
func TestResolutionIsSharedAndForFileIsOwned(t *testing.T) {
	c := qt.New(t)
	bundle := config.Bundle{Root: ".unswell.yaml", Files: map[string][]byte{
		".unswell.yaml": []byte(`version: 1
extends: [builtin:strict-v1]
overrides:
  - files: ["docs/**"]
    rules:
      policy.banned-phrases:
        parameters: {phrases: [robust]}
`),
	}}
	plan, _, err := config.CompileBundle(bundle, bundleCatalog())
	c.Assert(err, qt.IsNil)

	first, err := plan.Resolution("docs/a.md")
	c.Assert(err, qt.IsNil)
	second, err := plan.Resolution("docs/b.md")
	c.Assert(err, qt.IsNil)
	c.Assert(first.AppliedOverrides, qt.HasLen, 1)
	c.Assert(second.Hash, qt.Equals, first.Hash)
	c.Assert(reflect.ValueOf(second.Rules).Pointer(), qt.Equals, reflect.ValueOf(first.Rules).Pointer())

	owned, err := plan.ForFile("docs/c.md")
	c.Assert(err, qt.IsNil)
	c.Assert(owned.Hash, qt.Equals, first.Hash)
	c.Assert(reflect.ValueOf(owned.Rules).Pointer(), qt.Not(qt.Equals), reflect.ValueOf(first.Rules).Pointer())
	owned.Rules["policy.banned-phrases"].Parameters.Phrases[0] = "corrupted"
	again, err := plan.Resolution("docs/d.md")
	c.Assert(err, qt.IsNil)
	c.Assert(again.Rules["policy.banned-phrases"].Parameters.Phrases[0], qt.Equals, "robust")

	base, err := plan.Resolution("src/a.go")
	c.Assert(err, qt.IsNil)
	c.Assert(base.AppliedOverrides, qt.HasLen, 0)
	c.Assert(base.Hash, qt.Not(qt.Equals), first.Hash)
	snapshot, err := plan.Policy()
	c.Assert(err, qt.IsNil)
	c.Assert(snapshot.Hash, qt.Equals, base.Hash)
}
