package goanalysis_test

import (
	"context"
	"errors"
	"go/token"
	"strings"
	"sync"
	"testing"

	qt "github.com/frankban/quicktest"
	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/goanalysis"
)

func TestAdapterRejectsIncompleteInputs(t *testing.T) {
	for _, row := range []struct {
		name   string
		change func(*analysis.Pass)
	}{
		{"missing read callback", func(pass *analysis.Pass) { pass.ReadFile = nil }},
		{"missing report callback", func(pass *analysis.Pass) { pass.Report = nil }},
		{"missing files", func(pass *analysis.Pass) { pass.Files = nil }},
		{"missing syntax", func(pass *analysis.Pass) { pass.Files[0] = nil }},
		{"wrong FileSet", func(pass *analysis.Pass) { pass.Fset = token.NewFileSet() }},
		{"read error", func(pass *analysis.Pass) {
			pass.ReadFile = func(string) ([]byte, error) { return nil, errors.New("driver read failed") }
		}},
		{"wrong byte length", func(pass *analysis.Pass) {
			pass.ReadFile = func(string) ([]byte, error) { return nil, nil }
		}},
		{"duplicate file", func(pass *analysis.Pass) { pass.Files = append(pass.Files, pass.Files[0]) }},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{NoGate: true})
			c.Assert(err, qt.IsNil)
			analyzer, err := goanalysis.New(goanalysis.Options{Engine: engine, Context: t.Context()})
			c.Assert(err, qt.IsNil)
			pass, diagnostics := overlayPass(t, "package p\n")
			row.change(pass)
			_, err = analyzer.Run(pass)
			c.Assert(err, qt.IsNotNil)
			c.Assert(*diagnostics, qt.HasLen, 0)
		})
	}
}

func TestLimitsCancellationAndInvalidPolicy(t *testing.T) {
	c := qt.New(t)
	_, err := goanalysis.New(goanalysis.Options{Root: "relative"})
	c.Assert(err, qt.IsNotNil)
	analyzer, err := goanalysis.New(goanalysis.Options{Root: t.TempDir()})
	c.Assert(err, qt.IsNil)
	pass, _ := overlayPass(t, "package p\n")
	_, err = analyzer.Run(pass)
	c.Assert(err, qt.IsNotNil)
	ctx, cancel := context.WithCancel(t.Context())
	analyzer, err = goanalysis.New(goanalysis.Options{Context: ctx})
	c.Assert(err, qt.IsNil)
	cancel()
	_, err = analyzer.Run(pass)
	c.Assert(err, qt.ErrorIs, context.Canceled)
	analyzer, err = goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, policy+"analysis: {max_file_bytes: 64}\n")})
	c.Assert(err, qt.IsNil)
	pass, _ = overlayPass(t, "package p\n// "+strings.Repeat("word ", 100)+"\n")
	pass.ReadFile = func(string) ([]byte, error) {
		t.Fatal("oversized source must be rejected before reading")
		return nil, nil
	}
	_, err = analyzer.Run(pass)
	c.Assert(err, qt.IsNotNil)
	analyzer, err = goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, policy)})
	c.Assert(err, qt.IsNil)
	pass, _ = overlayPass(t, "package p\n// unswell-disable-next-block unknown.rule -- External wording.\n// Text.\n")
	_, err = analyzer.Run(pass)
	c.Assert(err, qt.IsNotNil)
}

func TestOneAnalyzerSupportsConcurrentPasses(t *testing.T) {
	c := qt.New(t)
	analyzer, err := goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, policy), Context: t.Context()})
	c.Assert(err, qt.IsNil)
	const count = 4
	passes := make([]*analysis.Pass, count)
	for i := range passes {
		passes[i], _ = overlayPass(t, "package p\n// It is important to note that the client retries.\n")
	}
	results, failures := make([]any, count), make([]error, count)
	var workers sync.WaitGroup
	for i := range passes {
		workers.Go(func() { results[i], failures[i] = analyzer.Run(passes[i]) })
	}
	workers.Wait()
	for i := range results {
		c.Assert(failures[i], qt.IsNil)
		c.Assert(results[i], qt.DeepEquals, results[0])
	}
}
