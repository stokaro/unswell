package goanalysis_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/goanalysis"
)

const policy = "version: 1\nextends: [builtin:custom]\nrules:\n" +
	"  filler.announced-importance: {enabled: true, gate: forbid}\n"

func configuredEngine(t *testing.T, config string) *unswell.Engine {
	t.Helper()
	engine, err := unswell.New(unswell.Options{Config: []byte(config), AllowEmpty: true})
	qt.New(t).Assert(err, qt.IsNil)
	return engine
}

func runFixture(t *testing.T, analyzer *analysis.Analyzer, pattern string) *analysistest.Result {
	t.Helper()
	results := analysistest.Run(t, analysistest.TestData(), analyzer, pattern)
	if len(results) != 1 {
		// Print the count without formatting a potentially cyclic analysis graph.
		t.Fatalf("expected one analysis result; got %d", len(results))
	}
	return results[0]
}

func TestAnalyzerFixturesUseTheSharedEngine(t *testing.T) {
	for _, context := range []struct{ name, values string }{{"comments", "comment"}, {"literals", "string"}, {"both", "comment, string"}} {
		t.Run(context.name, func(t *testing.T) {
			c := qt.New(t)
			engine := configuredEngine(t, policy+"extraction:\n  contexts: ["+context.values+"]\n")
			analyzer, err := goanalysis.New(goanalysis.Options{Engine: engine, Context: t.Context()})
			c.Assert(err, qt.IsNil)
			action := runFixture(t, analyzer, context.name).Action
			result, ok := action.Result.(unswell.RunResult)
			c.Assert(ok, qt.IsTrue)
			c.Assert(result.Manifest.Complete, qt.IsTrue)
			sources := make([]document.Source, 0, len(action.Package.Syntax))
			for _, syntax := range action.Package.Syntax {
				file := action.Package.Fset.File(syntax.FileStart)
				data, err := fs.ReadFile(os.DirFS(filepath.Dir(file.Name())), filepath.Base(file.Name()))
				c.Assert(err, qt.IsNil)
				sources = append(sources, document.Source{Name: filepath.Base(file.Name()), Format: document.Go, Bytes: data})
			}
			// The direct-engine comparison uses the same immutable fixture bytes
			// without invoking callbacks on a completed analysis pass.
			direct, err := engine.AnalyzeAll(t.Context(), sources)
			c.Assert(err, qt.IsNil)
			c.Assert(result, qt.DeepEquals, direct)
			for _, diagnostic := range action.Diagnostics {
				c.Assert(diagnostic.SuggestedFixes, qt.HasLen, 0)
				c.Assert(diagnostic.Category, qt.Equals, "filler.announced-importance")
				c.Assert(diagnostic.End > diagnostic.Pos, qt.IsTrue)
			}
		})
	}
}

func TestAnalyzerUsesPerLanguageAndFileOverrides(t *testing.T) {
	c := qt.New(t)
	root := filepath.Join(analysistest.TestData(), "src")
	config := policy + "extraction:\n  contexts: [string]\n  languages:\n    go:\n      contexts: [comment]\n" +
		"overrides:\n  - files: [overrides/skip.go]\n    rules:\n      filler.announced-importance: {enabled: false}\n"
	analyzer, err := goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, config), Root: root, Context: t.Context()})
	c.Assert(err, qt.IsNil)
	action := runFixture(t, analyzer, "overrides").Action
	c.Assert(action.Diagnostics, qt.HasLen, 1)
	result, ok := action.Result.(unswell.RunResult)
	c.Assert(ok, qt.IsTrue)
	c.Assert(result.Documents, qt.HasLen, 2)
	c.Assert(result.Findings[0].Primary.Path, qt.Equals, "overrides/check.go")
}

func TestDefaultAnalyzerAcceptsGoFilesWithoutProse(t *testing.T) {
	c := qt.New(t)
	analyzer, err := goanalysis.New(goanalysis.Options{})
	c.Assert(err, qt.IsNil)
	action := runFixture(t, analyzer, "clean").Action
	c.Assert(action.Diagnostics, qt.HasLen, 0)
	result, ok := action.Result.(unswell.RunResult)
	c.Assert(ok, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Manifest.Complete, qt.IsTrue)
}

func ExampleNew() {
	engine, err := unswell.New(unswell.Options{AllowEmpty: true, Config: []byte("version: 1\nextraction:\n  contexts: [comment]\n")})
	if err != nil {
		panic(err)
	}
	analyzer, err := goanalysis.New(goanalysis.Options{Engine: engine})
	if err != nil {
		panic(err)
	}
	fmt.Println(analyzer.Name)
	// Output: unswell
}
