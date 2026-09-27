package selfcheck_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell"
	"github.com/stokaro/unswell/document"
	"github.com/stokaro/unswell/mcp/internal/server"
	"github.com/stokaro/unswell/report"
)

func TestSelfcheckProcessRetainsFailedBatch(t *testing.T) {
	c := qt.New(t)
	directory := t.TempDir()
	binary := filepath.Join(directory, "selfcheck.exe")
	// #nosec G204 -- Compile the fixed checker package into this test's temporary directory.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/mcp-selfcheck")
	build.Dir = "../.."
	log, err := build.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("%s", log))
	helper, err := os.Executable()
	c.Assert(err, qt.IsNil)
	input, expected := failureInput(t, directory)
	checkSelfcheckDeadlines(t, binary, input, directory)
	for _, mode := range []string{"structured", "disconnect"} {
		t.Run(mode, func(t *testing.T) {
			c := qt.New(t)
			output := filepath.Join(directory, mode+".json")
			// #nosec G204 -- Run only binaries built from this checkout and this test executable.
			command := exec.CommandContext(t.Context(), binary, "--expected", input, "--output", output,
				"--", helper, "-test.run=^TestFailureMCPServer$", "--", "failure-server", input, mode)
			log, err := command.CombinedOutput()
			t.Logf("self-check: %s", log)
			var exit *exec.ExitError
			c.Assert(err, qt.ErrorAs, &exit, qt.Commentf("%s", log))
			c.Assert(exit.ExitCode(), qt.Equals, 1)
			checkFailedBatch(t, output, expected, mode)
		})
	}
}

func failureInput(t *testing.T, directory string) (string, unswell.RunResult) {
	t.Helper()
	c := qt.New(t)
	engine, err := unswell.New(unswell.Options{IncludeSource: true})
	c.Assert(err, qt.IsNil)
	sources := make([]document.Source, server.MaxSources+1)
	for i := range sources {
		sources[i] = document.Source{Name: fmt.Sprintf("draft-%03d.md", i), Format: document.Markdown,
			Bytes: []byte("The private sentinel phrase stays in memory.")}
	}
	expected, err := engine.AnalyzeAll(t.Context(), sources)
	c.Assert(err, qt.IsNil)
	c.Assert(expected.Gate.Passed, qt.IsTrue)
	var encoded bytes.Buffer
	c.Assert(report.Write(&encoded, "json", expected, report.Options{}), qt.IsNil)
	path := filepath.Join(directory, "expected.json")
	c.Assert(os.WriteFile(path, encoded.Bytes(), 0o600), qt.IsNil)
	return path, expected
}

func checkFailedBatch(t *testing.T, path string, expected unswell.RunResult, mode string) {
	t.Helper()
	c := qt.New(t)
	// #nosec G304 -- Read this test's explicit temporary output.
	data, err := os.ReadFile(path)
	c.Assert(err, qt.IsNil)
	c.Assert(bytes.Contains(data, []byte("The private sentinel phrase")), qt.IsFalse)
	var record struct {
		Complete bool                 `json:"complete"`
		Batches  []server.CheckOutput `json:"repository_batches"`
		Failure  struct {
			Stage   string                   `json:"stage"`
			Batch   int                      `json:"batch"`
			Sources []unswell.DocumentResult `json:"sources"`
			Actual  *server.CheckOutput      `json:"actual"`
		} `json:"failure"`
	}
	c.Assert(json.Unmarshal(data, &record), qt.IsNil)
	c.Assert(record.Complete, qt.IsFalse)
	c.Assert(record.Batches, qt.HasLen, 1)
	completed := len(record.Batches[0].Result.Documents)
	c.Assert(completed > 0 && completed < len(expected.Documents), qt.IsTrue)
	c.Assert(record.Failure.Stage, qt.Equals, "repository")
	c.Assert(record.Failure.Batch, qt.Equals, 2)
	c.Assert(len(record.Failure.Sources) > 0, qt.IsTrue)
	last := expected.Documents[completed]
	c.Assert(record.Failure.Sources[0].Name, qt.Equals, last.Name)
	c.Assert(record.Failure.Sources[0].SourceHash, qt.Equals, last.SourceHash)
	if mode == "disconnect" {
		c.Assert(record.Failure.Actual, qt.IsNil)
		return
	}
	c.Assert(record.Failure.Actual, qt.IsNotNil)
	c.Assert(record.Failure.Actual.Outcome, qt.Equals, "error")
	c.Assert(record.Failure.Actual.Result.Errors, qt.DeepEquals,
		[]unswell.RunError{{Path: last.Name, Message: "fixture analysis budget exhausted"}})
	c.Assert(record.Failure.Actual.Result.Documents[0].Source, qt.Equals, "")
}

// TestFailureMCPServer is a subprocess fixture serving real MCP transport frames.
func TestFailureMCPServer(t *testing.T) {
	if len(os.Args) < 4 || os.Args[len(os.Args)-3] != "failure-server" {
		return
	}
	c := qt.New(t)
	mode := os.Args[len(os.Args)-1]
	engine, err := unswell.New(unswell.Options{})
	c.Assert(err, qt.IsNil)
	instance, err := server.New(server.Options{})
	c.Assert(err, qt.IsNil)
	calls := 0
	mcp.AddTool(instance, &mcp.Tool{Name: "unswell_check"},
		func(ctx context.Context, _ *mcp.CallToolRequest, input server.CheckInput) (*mcp.CallToolResult, server.CheckOutput, error) {
			calls++
			if calls == 2 && mode == "disconnect" {
				os.Exit(7)
			}
			sources := make([]document.Source, len(input.Sources))
			for i, source := range input.Sources {
				sources[i] = document.Source{Name: source.Name, Format: source.Format, Bytes: []byte(source.Text)}
			}
			result, err := engine.AnalyzeAll(ctx, sources)
			if err != nil {
				return nil, server.CheckOutput{}, err
			}
			if calls == 2 {
				result.Status, result.Manifest.Complete, result.Gate.Passed = "incomplete", false, false
				result.Documents[0].Source = input.Sources[0].Text
				result.Errors = []unswell.RunError{{Path: input.Sources[0].Name, Message: "fixture analysis budget exhausted"}}
				return &mcp.CallToolResult{IsError: true}, server.CheckOutput{Outcome: "error", Result: result}, nil
			}
			return nil, server.CheckOutput{Outcome: "pass", Result: result}, nil
		})
	c.Assert(instance.Run(t.Context(), &mcp.StdioTransport{}), qt.IsNil)
	os.Exit(0)
}

func checkSelfcheckDeadlines(t *testing.T, binary, input, directory string) {
	t.Helper()
	for _, timeout := range []string{"0s", "-1s", "31m", "1ns"} {
		t.Run(timeout, func(t *testing.T) {
			c := qt.New(t)
			output := filepath.Join(directory, "deadline-"+timeout+".json")
			// #nosec G204 -- Invoke the locally built checker with fixed deadline probes.
			command := exec.CommandContext(t.Context(), binary, "--timeout", timeout, "--expected", input,
				"--output", output, "--", binary, "--help")
			log, err := command.CombinedOutput()
			var exit *exec.ExitError
			c.Assert(err, qt.ErrorAs, &exit, qt.Commentf("%s", log))
			c.Assert(exit.ExitCode(), qt.Equals, 1)
			if timeout != "1ns" {
				c.Assert(string(log), qt.Contains, "timeout must be greater than zero and at most 30m")
				return
			}
			// #nosec G304 -- Read this test's explicit timeout evidence.
			data, err := os.ReadFile(output)
			c.Assert(err, qt.IsNil)
			var record struct {
				Complete bool `json:"complete"`
				Failure  struct {
					Stage string `json:"stage"`
				} `json:"failure"`
			}
			c.Assert(json.Unmarshal(data, &record), qt.IsNil)
			c.Assert(record.Complete, qt.IsFalse)
			c.Assert(record.Failure.Stage, qt.Equals, "connect")
		})
	}
}
