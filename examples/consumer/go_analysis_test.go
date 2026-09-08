package main_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	qt "github.com/frankban/quicktest"
	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/goanalysis"
)

func TestPublicGoAnalyzerReportsTheSharedGate(t *testing.T) {
	c := qt.New(t)
	analyzer, err := goanalysis.New(goanalysis.Options{Context: t.Context()})
	c.Assert(err, qt.IsNil)
	c.Assert(analysis.Validate([]*analysis.Analyzer{analyzer}), qt.IsNil)
	const source = "package example\n// Certainly! The client retries.\n"
	fset := token.NewFileSet()
	syntax, err := parser.ParseFile(fset, "example.go", source, parser.ParseComments)
	c.Assert(err, qt.IsNil)
	var diagnostics []analysis.Diagnostic
	value, err := analyzer.Run(&analysis.Pass{Fset: fset, Files: []*ast.File{syntax},
		ReadFile: func(string) ([]byte, error) { return []byte(source), nil },
		Report:   func(diagnostic analysis.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
	})
	c.Assert(err, qt.IsNil)
	result, ok := value.(unswell.RunResult)
	c.Assert(ok, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsFalse)
	c.Assert(diagnostics, qt.HasLen, 1)
	c.Assert(diagnostics[0].Category, qt.Equals, "scaffold.chat-preamble")
}
