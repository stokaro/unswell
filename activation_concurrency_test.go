package unswell_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
)

func checkConcurrentActivations(t *testing.T, ids []string, text string, findings int) {
	t.Helper()
	c := qt.New(t)
	options := unswell.Options{}
	config := "version: 1\nextends: [builtin:custom]\nrules:\n"
	for _, id := range ids {
		options.Features = append(options.Features, "activation/"+id)
		config += "  " + id + ": {enabled: true}\n"
	}
	options.Config = []byte(config)
	engine, err := unswell.New(options)
	c.Assert(err, qt.IsNil)
	source := document.Source{Name: "guide.md", Format: document.Markdown, Bytes: []byte(text)}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = engine.Analyze(ctx, source)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	want, err := engine.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	c.Assert(want.Findings, qt.HasLen, findings)
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
