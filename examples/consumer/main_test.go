package main

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/rule"
)

func TestPublicConsumer(t *testing.T) {
	c := qt.New(t)
	result, err := analyze(t.Context())
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].RuleID, qt.Equals, "team.avoid-magic")
	c.Assert(result.Findings[0].Primary.Span.Start, qt.Equals, 17)
}

func TestPublicBaseline(t *testing.T) {
	c := qt.New(t)
	options := unswell.Options{Rules: []rule.Rule{teamRule{}}, CollectBaseline: true}
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.txt", Format: document.Plain, Bytes: []byte("The service uses magic.")}
	result, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	accepted, err := baseline.Create(t.Context(), *result.BaselineSnapshot)
	c.Assert(err, qt.IsNil)
	options.Baseline, err = baseline.Encode(t.Context(), accepted)
	c.Assert(err, qt.IsNil)
	options.GateMode = "new"
	engine, err = unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err = engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "existing")
	c.Assert(result.Findings[0].Suppressed, qt.IsFalse)
}
