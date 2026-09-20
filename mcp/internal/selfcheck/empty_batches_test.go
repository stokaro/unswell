package selfcheck_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/config"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/selfcheck"
	"github.com/stokaro/unswell/mcp/internal/server"
	"github.com/stokaro/unswell/report"
)

func TestRunVerifiesCodeOnlyBatchUnderDiscoveredPolicy(t *testing.T) {
	binary := wireServer(t)
	for _, failOnEmpty := range []bool{true, false} {
		t.Run(fmt.Sprintf("fail_on_empty=%t", failOnEmpty), func(t *testing.T) {
			c := qt.New(t)
			policyBytes := []byte(fmt.Sprintf("version: 1\ngate:\n  fail_on_empty: %t\n", failOnEmpty))
			bundle := config.Bundle{Root: "policy.yaml", Files: map[string][]byte{"policy.yaml": policyBytes}}
			engine, err := unswell.New(unswell.Options{ConfigBundle: &bundle, IncludeSource: true})
			c.Assert(err, qt.IsNil)
			sources := make([]document.Source, server.MaxSources+1)
			for i := range sources {
				sources[i] = document.Source{Name: fmt.Sprintf("draft-%03d.go", i), Format: document.Go,
					Bytes: []byte("package sample\n")}
			}
			sources[len(sources)-1].Bytes = []byte("package sample\n// The client opens connections.\n")
			expected, err := engine.AnalyzeAll(t.Context(), sources)
			c.Assert(err, qt.IsNil)
			c.Assert(expected.Gate.Passed, qt.IsTrue)
			var serialized bytes.Buffer
			c.Assert(report.Write(&serialized, "json", expected, report.Options{}), qt.IsNil)
			directory := t.TempDir()
			input, output := filepath.Join(directory, "expected.json"), filepath.Join(directory, "checked.json")
			policy := filepath.Join(directory, "policy.yaml")
			c.Assert(os.WriteFile(input, serialized.Bytes(), 0o600), qt.IsNil)
			c.Assert(os.WriteFile(policy, policyBytes, 0o600), qt.IsNil)
			var diagnostics bytes.Buffer
			count, err := selfcheck.Run(t.Context(), input, output, []string{binary, "--project-root", directory, "--config", policy}, &diagnostics)
			c.Assert(err, qt.IsNil, qt.Commentf("%s", diagnostics.String()))
			c.Assert(count, qt.Equals, len(sources))
			assertCodeOnlyBatch(t, output, expected, failOnEmpty)
		})
	}
}

func assertCodeOnlyBatch(t *testing.T, path string, expected unswell.RunResult, failOnEmpty bool) {
	t.Helper()
	c := qt.New(t)
	// #nosec G304 -- Read the evidence written into this test's temporary directory.
	data, err := os.ReadFile(path)
	c.Assert(err, qt.IsNil)
	var evidence struct {
		Batches []server.CheckOutput `json:"repository_batches"`
	}
	c.Assert(json.Unmarshal(data, &evidence), qt.IsNil)
	c.Assert(len(evidence.Batches) > 1, qt.IsTrue)
	var names, want []string
	for _, batch := range evidence.Batches {
		for _, doc := range batch.Result.Documents {
			names = append(names, doc.Name)
		}
	}
	for _, doc := range expected.Documents {
		want = append(want, doc.Name)
	}
	c.Assert(names, qt.DeepEquals, want)
	last := evidence.Batches[0]
	for _, doc := range last.Result.Documents {
		c.Assert(doc.ProseWords, qt.Equals, 0)
	}
	c.Assert(last.Result.Gate.Passed, qt.Equals, !failOnEmpty)
	c.Assert(last.Result.Manifest.Complete, qt.Equals, !failOnEmpty)
	if failOnEmpty {
		c.Assert(last.Outcome, qt.Equals, "error")
		c.Assert(last.Result.Status, qt.Equals, "incomplete")
		c.Assert(last.Result.Errors, qt.DeepEquals, []unswell.RunError{{Message: "scan contains no applicable English prose"}})
	} else {
		c.Assert(last.Outcome, qt.Equals, "pass")
		c.Assert(last.Result.Status, qt.Equals, "complete")
		c.Assert(last.Result.Errors, qt.HasLen, 0)
	}
}
