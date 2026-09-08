package goanalysis_test

import (
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/baseline"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/goanalysis"
)

const goProse = "package p\n// It is important to note that the client retries.\n"

func TestDiagnosticsRespectGateAndExplicitAllFindingsMode(t *testing.T) {
	for _, row := range []struct {
		name    string
		config  string
		noGate  bool
		all     bool
		count   int
		passed  bool
		derived bool
	}{
		{"forbidden", policy, false, false, 1, false, false},
		{"no gate", policy, true, false, 0, true, false},
		{"all with no gate", policy, true, true, 1, true, false},
		{"advisory", strings.ReplaceAll(policy, "gate: forbid", "gate: none"), false, false, 0, true, false},
		{"all advisory", strings.ReplaceAll(policy, "gate: forbid", "gate: none"), false, true, 1, true, false},
		{"threshold", strings.ReplaceAll(policy, "gate: forbid", "gate: none") +
			"gate:\n  sentence_score: {fail_at: 1, min_words: 1}\n  paragraph_score: {fail_at: 100, min_words: 100}\n",
			false, false, 1, false, true},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			engine, err := unswell.New(unswell.Options{Config: []byte(row.config), NoGate: row.noGate})
			c.Assert(err, qt.IsNil)
			analyzer, err := goanalysis.New(goanalysis.Options{Engine: engine, ReportAll: row.all, Context: t.Context()})
			c.Assert(err, qt.IsNil)
			pass, diagnostics := overlayPass(t, goProse)
			value, err := analyzer.Run(pass)
			c.Assert(err, qt.IsNil)
			result, ok := value.(unswell.RunResult)
			c.Assert(ok, qt.IsTrue)
			c.Assert(result.Gate.Passed, qt.Equals, row.passed)
			c.Assert(*diagnostics, qt.HasLen, row.count)
			if row.derived {
				c.Assert((*diagnostics)[0].Category, qt.Equals, "gate.sentence-score")
			}
		})
	}
}

func TestAcceptedBaselineRemainsInResultWithoutDefaultDiagnostics(t *testing.T) {
	c := qt.New(t)
	source := document.Source{Name: "overlay.go", Format: document.Go, Bytes: []byte(goProse)}
	collector, err := unswell.New(unswell.Options{Config: []byte(policy), CollectBaseline: true})
	c.Assert(err, qt.IsNil)
	collected, err := collector.Analyze(t.Context(), source)
	c.Assert(err, qt.IsNil)
	file, err := baseline.Create(t.Context(), *collected.BaselineSnapshot)
	c.Assert(err, qt.IsNil)
	data, err := baseline.Encode(t.Context(), file)
	c.Assert(err, qt.IsNil)
	engine, err := unswell.New(unswell.Options{Config: []byte(policy), Baseline: data, GateMode: "new"})
	c.Assert(err, qt.IsNil)
	analyzer, err := goanalysis.New(goanalysis.Options{Engine: engine})
	c.Assert(err, qt.IsNil)
	pass, diagnostics := overlayPass(t, goProse)
	value, err := analyzer.Run(pass)
	c.Assert(err, qt.IsNil)
	result, ok := value.(unswell.RunResult)
	c.Assert(ok, qt.IsTrue)
	c.Assert(result.Gate.Passed, qt.IsTrue)
	c.Assert(result.Gate.Accepted, qt.HasLen, 1)
	c.Assert(result.Findings, qt.HasLen, 1)
	c.Assert(result.Findings[0].BaselineState, qt.Equals, "existing")
	c.Assert(*diagnostics, qt.HasLen, 0)
}
