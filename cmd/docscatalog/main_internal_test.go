package main

// White-box tests: Check the grouping, the profile comparison, the rendered
// bytes, and the two --check failures. This package is a command with nothing
// exported, so a blackbox test cannot reach build, compareProfiles, render, or
// compare. Running the built binary would test the toolchain, not the rules
// under test here.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/config"
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

func TestCompareProfilesListsWhereStrictGoesFurther(t *testing.T) {
	c := qt.New(t)
	on := func(gate string) rule.Settings { return rule.Settings{Enabled: true, Gate: gate} }
	technical := config.Policy{
		Gate: config.Gate{Paragraph: config.Threshold{FailAt: 65}, Sentence: config.Threshold{FailAt: 80}},
		Rules: map[string]rule.Settings{
			"a.same": on("none"), "a.forbid": on("forbid"), "b.raised": on("none"),
			"c.added": {Gate: "none"}, "c.added-forbid": {Gate: "forbid"}, "d.off": {Gate: "none"},
		},
	}
	strict := config.Policy{
		Gate: config.Gate{Paragraph: config.Threshold{FailAt: 50}, Sentence: config.Threshold{FailAt: 65}},
		Rules: map[string]rule.Settings{
			"a.same": on("none"), "a.forbid": on("forbid"), "b.raised": on("forbid"),
			"c.added": on("none"), "c.added-forbid": on("forbid"), "d.off": {Gate: "none"},
		},
	}
	result := compareProfiles(&technical, &strict)
	c.Assert(result.Technical, qt.Equals, profileTotals{Rules: 3, Forbid: 1, ParagraphFailAt: 65, SentenceFailAt: 80})
	c.Assert(result.Strict, qt.Equals, profileTotals{Rules: 5, Forbid: 3, ParagraphFailAt: 50, SentenceFailAt: 65})
	// A rule that only strict runs is an added rule, also when it fails the gate there.
	c.Assert(result.Added, qt.DeepEquals, []string{"c.added", "c.added-forbid"})
	c.Assert(result.Stricter, qt.DeepEquals, []string{"b.raised"})

	// Two equal profiles differ nowhere, and the lists are empty arrays in the file.
	same := compareProfiles(&technical, &technical)
	c.Assert(same.Added, qt.DeepEquals, []string{})
	c.Assert(same.Stricter, qt.DeepEquals, []string{})
}

func TestRenderSaysWhatTheStrictProfileChanges(t *testing.T) {
	c := qt.New(t)
	rendered, err := render()
	c.Assert(err, qt.IsNil)
	var document catalog
	c.Assert(json.Unmarshal(rendered, &document), qt.IsNil)
	shown := document.Profiles
	// The strict profile runs every rule of the default one and more, fails
	// the gate on more of them, and fails a paragraph and a sentence earlier.
	c.Assert(shown.Strict.Rules, qt.Equals, shown.Technical.Rules+len(shown.Added))
	c.Assert(shown.Strict.Forbid >= shown.Technical.Forbid+len(shown.Stricter), qt.IsTrue)
	c.Assert(len(shown.Added) > 0, qt.IsTrue)
	c.Assert(len(shown.Stricter) > 0, qt.IsTrue)
	c.Assert(shown.Strict.ParagraphFailAt < shown.Technical.ParagraphFailAt, qt.IsTrue)
	c.Assert(shown.Strict.SentenceFailAt < shown.Technical.SentenceFailAt, qt.IsTrue)
	// Every listed rule is a rule of the catalog.
	known := map[string]bool{}
	for _, group := range document.Families {
		for _, item := range group.Rules {
			known[item.ID] = true
		}
	}
	for _, id := range append(append([]string{}, shown.Added...), shown.Stricter...) {
		c.Assert(known[id], qt.IsTrue, qt.Commentf("%s is not in the catalog", id))
	}
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
