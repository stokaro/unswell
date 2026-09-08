package main_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func TestCompareCompleteSourcesThroughThePublicEngine(t *testing.T) {
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	sources := []document.Source{{Name: "guide.txt", Format: document.Plain, Bytes: []byte("Certainly! The client retries.")}}
	result, err := engine.AnalyzeChanged(t.Context(), sources, sources)
	c.Assert(err, qt.IsNil)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Changes.Complete, qt.IsTrue)
	c.Assert(result.Changes.SelectedUnits, qt.Equals, 0)
	c.Assert(len(result.Findings) > 0, qt.IsTrue)
	c.Assert(len(result.Gate.Unchanged) > 0, qt.IsTrue)
	c.Assert(result.Manifest.Git, qt.IsNil)
	c.Assert(result.Baseline, qt.IsNil)
}
