package corpus_test

import (
	"encoding/json"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/builtin"
	"github.com/stokaro/unswell/research/annotation/corpus"
	"github.com/stokaro/unswell/rule"
)

func committedClasses(c *qt.C) ([]byte, corpus.RuleClasses) {
	c.Helper()
	data, err := os.ReadFile("../../methods/rule-classes-v1.json")
	c.Assert(err, qt.IsNil)
	classes, err := corpus.LoadRuleClasses(c.TB.(*testing.T).Context(), data)
	c.Assert(err, qt.IsNil)
	return data, classes
}

func builtinCatalog() []rule.Descriptor {
	rules := builtin.Rules()
	catalog := make([]rule.Descriptor, 0, len(rules))
	for _, item := range rules {
		catalog = append(catalog, item.Descriptor())
	}
	return catalog
}

func TestCommittedRuleClassesCoverTheBuiltinCatalog(t *testing.T) {
	c := qt.New(t)
	_, classes := committedClasses(c)
	c.Assert(classes.Format, qt.Equals, corpus.RuleClassesVersion)
	c.Assert(classes.Protocol, qt.Equals, "unswell-llm-patterns-v1")
	c.Assert(classes.Cover(builtinCatalog()), qt.IsNil)
	counts := map[string]int{}
	for _, entry := range classes.Rules {
		counts[entry.Class]++
		if entry.HypothesisSource != nil {
			c.Assert(*entry.HypothesisSource, qt.Matches, `[a-z]+-20[0-9]{2}`)
		}
	}
	c.Assert(counts["explicit_prohibition"], qt.Equals, 1)
	c.Assert(counts["general_style"] > 0, qt.IsTrue)
	c.Assert(counts["llm_associated_candidate"] > 0, qt.IsTrue)
}

func TestRuleClassesRejectGapsAndUnknowns(t *testing.T) {
	c := qt.New(t)
	data, classes := committedClasses(c)
	catalog := builtinCatalog()
	// A missing or invented rule breaks coverage in either direction.
	c.Assert(classes.Cover(catalog[1:]), qt.IsNotNil)
	extra := append([]rule.Descriptor{{ID: "zzz.invented"}}, catalog...)
	c.Assert(classes.Cover(extra), qt.IsNotNil)
	for _, row := range []struct {
		name string
		edit func(map[string]any)
	}{
		{"format", func(m map[string]any) { m["format"] = "future" }},
		{"revision", func(m map[string]any) { m["revision"] = 0 }},
		{"unknown field", func(m map[string]any) { m["extra"] = true }},
		{"unknown class name", func(m map[string]any) { m["classes"].(map[string]any)["other"] = "x" }},
		{"unknown class", func(m map[string]any) { m["rules"].([]any)[0].(map[string]any)["class"] = "other" }},
		{"unknown role", func(m map[string]any) { m["rules"].([]any)[0].(map[string]any)["roles"] = []any{"poem"} }},
		{"empty roles", func(m map[string]any) { m["rules"].([]any)[0].(map[string]any)["roles"] = []any{} }},
		{"empty reason", func(m map[string]any) { m["rules"].([]any)[0].(map[string]any)["reason"] = " " }},
		{"empty source", func(m map[string]any) { m["rules"].([]any)[0].(map[string]any)["hypothesis_source"] = "" }},
		{"duplicate", func(m map[string]any) { rules := m["rules"].([]any); m["rules"] = append(rules, rules[0]) }},
		{"order", func(m map[string]any) { rules := m["rules"].([]any); rules[0], rules[1] = rules[1], rules[0] }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			var m map[string]any
			c.Assert(json.Unmarshal(data, &m), qt.IsNil)
			row.edit(m)
			_, err := corpus.LoadRuleClasses(t.Context(), encoded(c, m))
			c.Assert(err, qt.IsNotNil)
		})
	}
}
