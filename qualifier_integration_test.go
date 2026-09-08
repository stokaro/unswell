package unswell_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestQualifierActivationsDoNotRestoreExcludedProse(t *testing.T) {
	c := qt.New(t)
	options := qualifierOptions()
	options.AllowEmpty = true
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown,
		Bytes: []byte("`may possibly perhaps`\n\n```text\ngame-changing solution\n```\n")})
	c.Assert(err, qt.IsNil)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
	c.Assert(result.Findings, qt.HasLen, 0)
	c.Assert(result.Features.Sources, qt.HasLen, 1)
	c.Assert(result.Features.Sources[0].Units, qt.HasLen, 0)
}

func TestQualifierActivationsRetainFailuresAndConcurrentOwnership(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{NoGate: true, Features: []string{"activation/hype.vague-praise"}, Config: []byte(
		"version: 1\nextends: [builtin:custom]\nanalysis: {max_candidates: 1}\nrules:\n" +
			"  hype.vague-praise: {enabled: true, parameters: {phrases: [alpha beta, alpha gamma]}}\n")})
	c.Assert(err, qt.IsNil)
	result, err := engine.Analyze(t.Context(), document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte("Alpha")})
	c.Assert(err, qt.ErrorMatches, ".*editorial pattern checks exceed max_candidates.*")
	c.Assert(result.Manifest.Complete, qt.IsFalse)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	assertPhraseMeasurements(t, result, []string{"evaluation_failed"}, nil)
	engine, err = unswell.New(qualifierOptions())
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(
		"This is a game-changing solution.\n\nThe service guarantees complete safety.\n\nThe result may possibly perhaps change.")}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(want.Findings, qt.HasLen, 3)
	for range 4 {
		t.Run("shared engine", func(t *testing.T) {
			t.Parallel()
			c := qt.New(t)
			got, err := engine.Analyze(t.Context(), source)
			c.Assert(err, qt.IsNil)
			c.Assert(got, qt.DeepEquals, want)
		})
	}
}

func qualifierOptions() unswell.Options {
	options := unswell.Options{}
	config := "version: 1\nextends: [builtin:custom]\nrules:\n"
	for _, id := range qualifierActivationIDs() {
		options.Features = append(options.Features, "activation/"+id)
		config += "  " + id + ": {enabled: true}\n"
	}
	options.Config = []byte(config)
	return options
}
