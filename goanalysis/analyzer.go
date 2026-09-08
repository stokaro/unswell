// Package goanalysis adapts the public Unswell engine to Go analysis drivers.
// Drivers supply source bytes; the adapter performs no direct filesystem access.
package goanalysis

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell"
)

// Options configures the adapter without discovering policy or reading files.
type Options struct {
	// Engine is reused across passes. Nil builds the default engine once with
	// AllowEmpty, since a Go file need not contain eligible prose. A supplied
	// engine retains its own empty-input, extraction, baseline, and gate policy.
	Engine *unswell.Engine
	// Root is an optional absolute project directory for logical source names.
	// Without Root, names are file basenames local to each analysis pass.
	Root string
	// Context supplies cancellation absent from analysis.Pass. Nil uses Background.
	// The adapter cannot interrupt a blocking driver ReadFile callback.
	Context context.Context
	// ReportAll reports every unsuppressed finding, including advisory findings
	// and accepted baseline debt. Go drivers can fail on any diagnostic. The
	// default reports only reasons for a failing completed engine gate.
	ReportAll bool
}

// New returns an independent analyzer with a shared immutable engine. It reports
// original source ranges and returns an unswell.RunResult to dependent analyzers.
// Callers must not change Analyzer fields or flags while passes are running.
func New(options Options) (*analysis.Analyzer, error) {
	if options.Root != "" {
		if !filepath.IsAbs(options.Root) {
			return nil, fmt.Errorf("analysis root must be absolute")
		}
		options.Root = filepath.Clean(options.Root)
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.Engine == nil {
		engine, err := unswell.New(unswell.Options{AllowEmpty: true})
		if err != nil {
			return nil, err
		}
		options.Engine = engine
	}
	return &analysis.Analyzer{
		Name: "unswell", Doc: "check Go prose against Unswell editorial policy",
		URL:        "https://github.com/stokaro/unswell/blob/main/docs/go-analysis.md",
		ResultType: reflect.TypeFor[unswell.RunResult](),
		Run:        options.run,
	}, nil
}

func (o Options) run(pass *analysis.Pass) (any, error) {
	if err := o.Context.Err(); err != nil {
		return nil, err
	}
	input, err := o.readSources(pass)
	if err != nil {
		return nil, err
	}
	result, err := o.Engine.AnalyzeAll(o.Context, input.sources)
	if err != nil {
		return result, err
	}
	diagnostics, err := resultDiagnostics(o.Context, result, input.files, o.ReportAll)
	if err != nil {
		return result, err
	}
	for _, diagnostic := range diagnostics {
		if err := o.Context.Err(); err != nil {
			return result, err
		}
		pass.Report(diagnostic)
	}
	return result, nil
}
