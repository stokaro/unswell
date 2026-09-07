package main

import (
	"bytes"
	"encoding/json"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/report"
	"github.com/stokaro/unswell/rule"
	"github.com/stokaro/unswell/ruleset"
)

func TestDeclarativePublicConsumer(t *testing.T) {
	c := qt.New(t)
	set, err := ruleset.Load([]byte(`version: 1
namespace: team
release: "1"
license: MIT
provenance: Public API consumer test.
rules:
  - id: team.no-introduction
    summary: Start with the subject.
    message: Remove the conversational introduction.
    scope: sentence
    gate: forbid
    match: {type: phrase, at: start, values: ["Let's dive into"]}
    examples:
      fail: ["Let's dive into the settings."]
      pass: ["The settings define connection limits."]
`))
	c.Assert(err, qt.IsNil)
	registry := append([]rule.Rule{teamRule{}}, set.Rules()...)
	engine, err := unswell.New(unswell.Options{Rules: registry})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "memory.md", Format: document.Markdown,
		Bytes: []byte("Let's **dive** into the settings.")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleID, qt.Equals, "team.no-introduction")
	var output bytes.Buffer
	c.Assert(report.Write(&output, "json", result, report.Options{}), qt.IsNil)
	var saved unswell.RunResult
	c.Assert(json.Unmarshal(output.Bytes(), &saved), qt.IsNil)
	c.Assert(saved.Manifest.RulesetHash, qt.Equals, result.Manifest.RulesetHash)
	for _, descriptor := range saved.Manifest.Rules {
		if descriptor.ID == "team.no-introduction" {
			c.Assert(*descriptor.Origin, qt.Equals, set.Origin())
		}
	}
}
