package command_test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/stokaro/unswell/mcp/internal/server"
)

func TestStdioProcess(t *testing.T) {
	c := qt.New(t)
	binary := buildServer(c, t)
	policy := filepath.Join(t.TempDir(), "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:strict-v1]\n"), 0o600), qt.IsNil)
	var diagnostics bytes.Buffer
	// #nosec G204 -- Execute only the server built from this checkout in the test's temporary directory.
	process := exec.CommandContext(t.Context(), binary, "--config", policy, "--feature", "prose-words",
		"--prepared-feature", "prose-words", "--prepared-kind", "sentence", "--prepared-kind", "paragraph")
	process.Stderr = &diagnostics
	client := mcp.NewClient(&mcp.Implementation{Name: "stdio-test", Version: "1"}, nil)
	session, err := client.Connect(t.Context(), &mcp.CommandTransport{Command: process}, nil)
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(session.Close(), qt.IsNil) })
	tools, err := session.ListTools(t.Context(), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(tools.Tools, qt.HasLen, 2)
	for _, tc := range []struct{ name, text, outcome string }{
		{"clean.md", "The client opens connections.", "pass"},
		{"bad.md", "Certainly! The client opens connections.", "policy_failure"},
		{"rewritten.md", "The client opens connections.", "pass"},
		{"broken.cs", "class Sample { string value = \"unfinished", "error"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := qt.New(t)
			result, err := session.CallTool(t.Context(), &mcp.CallToolParams{Name: "unswell_check", Arguments: server.CheckInput{
				Sources: []server.Source{{Name: tc.name, Text: tc.text}},
			}})
			c.Assert(err, qt.IsNil)
			data, err := json.Marshal(result.StructuredContent)
			c.Assert(err, qt.IsNil)
			var checked server.CheckOutput
			c.Assert(json.Unmarshal(data, &checked), qt.IsNil)
			c.Assert(checked.Outcome, qt.Equals, tc.outcome)
			c.Assert(result.IsError, qt.Equals, tc.outcome == "error")
			c.Assert(checked.Result.PreparedFeatures, qt.IsNotNil)
			c.Assert(checked.Result.PreparedFeatures.Requested, qt.DeepEquals, []string{"prose-words"})
			c.Assert(checked.Result.Features, qt.IsNotNil)
			c.Assert(checked.Result.Features.Requested, qt.DeepEquals, []string{"prose-words"})
		})
	}
	c.Assert(session.Close(), qt.IsNil)
	verifyStartupFailures(c, t, binary)
}

func buildServer(c *qt.C, t *testing.T) string {
	c.Helper()
	name := "unswell-mcp"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	// #nosec G204 -- Compile the fixed local server package to a test-owned temporary path.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", binary, "./cmd/unswell-mcp")
	build.Dir = "../.."
	output, err := build.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("%s", output))
	return binary
}

func verifyStartupFailures(c *qt.C, t *testing.T, binary string) {
	c.Helper()
	for _, args := range [][]string{{"--version"}, {"--config", filepath.Join(t.TempDir(), "absent.yaml")}, {"--feature", "unknown"},
		{"--prepared-feature", "prose-words"}, {"--prepared-feature", "unknown", "--prepared-kind", "sentence"}} {
		// #nosec G204 -- Test-owned executable and fixed startup cases; no source text enters arguments.
		process := exec.CommandContext(t.Context(), binary, args...)
		var stdout, stderr bytes.Buffer
		process.Stdout, process.Stderr = &stdout, &stderr
		err := process.Run()
		c.Assert(stdout.Len(), qt.Equals, 0)
		if args[0] == "--version" {
			c.Assert(err, qt.IsNil)
		} else {
			c.Assert(err, qt.IsNotNil)
			c.Assert(process.ProcessState.ExitCode(), qt.Equals, 2)
		}
		c.Assert(stderr.Len() > 0, qt.IsTrue)
	}
}
