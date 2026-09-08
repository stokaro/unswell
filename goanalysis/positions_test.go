package goanalysis_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"golang.org/x/tools/go/analysis"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/goanalysis"
)

func overlayPass(t *testing.T, source string) (*analysis.Pass, *[]analysis.Diagnostic) {
	t.Helper()
	c := qt.New(t)
	filename := filepath.Join(t.TempDir(), "overlay.go")
	fset := token.NewFileSet()
	syntax, err := parser.ParseFile(fset, filename, source, parser.ParseComments)
	c.Assert(err, qt.IsNil)
	var diagnostics []analysis.Diagnostic
	return &analysis.Pass{Fset: fset, Files: []*ast.File{syntax},
		ReadFile: func(name string) ([]byte, error) {
			if name != filename {
				return nil, fmt.Errorf("unexpected driver read: %s", name)
			}
			return []byte(source), nil
		},
		Report: func(diagnostic analysis.Diagnostic) { diagnostics = append(diagnostics, diagnostic) },
	}, &diagnostics
}

func TestOverlayCoordinatesRetainOriginalByteRanges(t *testing.T) {
	for _, row := range []struct{ name, source, matched string }{
		{"unicode", "package p\n// Café 🙂. It is important to note that the client retries.\n", "It is important to note that"},
		{"CRLF BOM line directive", "\ufeffpackage p\r\n//line virtual.go:90\r\n" +
			"// It is important to note that the client retries.\r\n", "It is important to note that"},
		{"escaped string", "package p\nconst Message = \"It is \\u0069mportant to note that the client retries.\"\n",
			"It is \\u0069mportant to note that"},
		{"raw string", "package p\nconst Message = `Café 🙂.\nIt is important to note that the client retries.`\n",
			"It is important to note that"},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			analyzer, err := goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, policy), Context: t.Context()})
			c.Assert(err, qt.IsNil)
			pass, diagnostics := overlayPass(t, row.source)
			_, err = analyzer.Run(pass)
			c.Assert(err, qt.IsNil)
			c.Assert(*diagnostics, qt.HasLen, 1)
			file := pass.Fset.File(pass.Files[0].FileStart)
			start := strings.Index(row.source, row.matched)
			c.Assert(file.Offset((*diagnostics)[0].Pos), qt.Equals, start)
			c.Assert(file.Offset((*diagnostics)[0].End), qt.Equals, start+len(row.matched))
		})
	}
}

func TestRelatedPositionsMatchEngineEvidence(t *testing.T) {
	c := qt.New(t)
	config := "version: 1\nextends: [builtin:custom]\nrules:\n  repetition.exact-sentence: {enabled: true, gate: forbid}\n"
	analyzer, err := goanalysis.New(goanalysis.Options{Engine: configuredEngine(t, config), Context: t.Context()})
	c.Assert(err, qt.IsNil)
	const prose = "The client opens a new connection after the server closes the previous connection."
	pass, diagnostics := overlayPass(t, "package p\n// "+prose+"\nfunc First() {}\n// "+prose+"\nfunc Second() {}\n")
	value, err := analyzer.Run(pass)
	c.Assert(err, qt.IsNil)
	result, ok := value.(unswell.RunResult)
	c.Assert(ok, qt.IsTrue)
	c.Assert(*diagnostics, qt.HasLen, 1)
	c.Assert(len((*diagnostics)[0].Related) > 0, qt.IsTrue)
	file := pass.Fset.File(pass.Files[0].FileStart)
	for i, related := range (*diagnostics)[0].Related {
		c.Assert(file.Offset(related.Pos), qt.Equals, result.Findings[0].Related[i].Span.Start)
		c.Assert(file.Offset(related.End), qt.Equals, result.Findings[0].Related[i].Span.End)
	}
}
