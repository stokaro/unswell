package main

// White-box tests: Check the grouping, the rendered bytes, and the two --check
// failures. This package is a command with nothing exported, so a blackbox test
// cannot reach build, render, or compare. Running the built binary would test
// the toolchain, not the rules under test here.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/rule"
)

func TestFamilyOfTakesTheIdentifierPrefix(t *testing.T) {
	c := qt.New(t)
	c.Assert(familyOf("filler.announced-importance"), qt.Equals, "filler")
	c.Assert(familyOf("repetition.near-sentence"), qt.Equals, "repetition")
	c.Assert(familyOf("solitary"), qt.Equals, "solitary")
	c.Assert(familyOf(".leading"), qt.Equals, ".leading")
}

func TestBuildOrdersFamiliesAndRules(t *testing.T) {
	c := qt.New(t)
	document := build([]rule.Descriptor{
		{ID: "syntax.long-sentence"},
		{ID: "filler.wordy-phrase"},
		{ID: "filler.announced-importance"},
	})
	c.Assert(document.Schema, qt.Equals, schemaID)
	c.Assert(document.RuleCount, qt.Equals, 3)
	c.Assert(document.Families, qt.HasLen, 2)
	c.Assert(document.Families[0].Name, qt.Equals, "filler")
	c.Assert(document.Families[1].Name, qt.Equals, "syntax")
	c.Assert(document.Families[0].Rules[0].ID, qt.Equals, "filler.announced-importance")
	c.Assert(document.Families[0].Rules[1].ID, qt.Equals, "filler.wordy-phrase")
}

func TestRenderCoversTheWholeBuiltinCatalog(t *testing.T) {
	c := qt.New(t)
	rendered, err := render()
	c.Assert(err, qt.IsNil)
	var document catalog
	c.Assert(json.Unmarshal(rendered, &document), qt.IsNil)
	c.Assert(document.RuleCount > 0, qt.IsTrue)
	counted := 0
	for _, group := range document.Families {
		counted += len(group.Rules)
		for _, item := range group.Rules {
			c.Assert(item.ID, qt.Not(qt.Equals), "")
			c.Assert(item.Version, qt.Not(qt.Equals), "")
			c.Assert(item.Summary, qt.Not(qt.Equals), "")
			c.Assert(strings.HasPrefix(item.ID, group.Name), qt.IsTrue)
		}
	}
	c.Assert(counted, qt.Equals, document.RuleCount)
	again, err := render()
	c.Assert(err, qt.IsNil)
	c.Assert(string(again), qt.Equals, string(rendered))
}

func TestCheckRejectsMissingAndStaleFiles(t *testing.T) {
	c := qt.New(t)
	directory := t.TempDir()
	path := filepath.Join(directory, "rules-catalog.json")

	err := run([]string{"--output", path, "--check"}, os.Stdout)
	c.Assert(err, qt.ErrorMatches, ".*is missing.*")

	c.Assert(run([]string{"--output", path}, os.Stdout), qt.IsNil)
	c.Assert(run([]string{"--output", path, "--check"}, os.Stdout), qt.IsNil)

	c.Assert(os.WriteFile(path, []byte("{}\n"), 0o600), qt.IsNil)
	err = run([]string{"--output", path, "--check"}, os.Stdout)
	c.Assert(err, qt.ErrorMatches, ".*is stale.*")
}

func TestCommittedCatalogMatchesTheEngine(t *testing.T) {
	c := qt.New(t)
	c.Assert(run([]string{"--output", "../../docs/site/src/data/rules-catalog.json", "--check"}, os.Stdout), qt.IsNil)
}
