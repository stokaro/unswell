// Package e2e_test verifies the built CLI against annotated source fixtures.
package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell"
)

type scenario struct {
	Files        []string `json:"files"`
	Resources    []string `json:"resources"`
	ExitCode     int      `json:"exit_code"`
	Args         []string `json:"args"`
	Stdin        bool     `json:"stdin"`
	Directory    bool     `json:"directory"`
	CRLF         bool     `json:"crlf"`
	BOM          bool     `json:"bom"`
	StartupError bool     `json:"startup_error"`
}

func TestCLI(t *testing.T) {
	c := qt.New(t)
	binary := buildCLI(t)
	entries, err := os.ReadDir("testdata")
	c.Assert(err, qt.IsNil)
	c.Assert(len(entries) > 0, qt.IsTrue)
	for _, entry := range entries {
		if entry.IsDir() {
			t.Run(entry.Name(), func(t *testing.T) {
				runScenario(t, binary, filepath.Join("testdata", entry.Name()))
			})
		}
	}
}

func buildCLI(t *testing.T) string {
	t.Helper()
	c := qt.New(t)
	name := "unswell"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	// #nosec G204 -- Build the fixed local CLI package into a test-owned temporary directory.
	command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./cmd/unswell")
	command.Dir = ".."
	command.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := command.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("CLI build: %s", output))
	return binary
}

func runScenario(t *testing.T, binary, fixture string) {
	t.Helper()
	c := qt.New(t)
	var spec scenario
	decodeFile(t, filepath.Join(fixture, "case.json"), &spec)
	c.Assert(len(spec.Files) > 0, qt.IsTrue)
	workspace := t.TempDir()
	sources, wants := prepareSources(t, fixture, workspace, spec)
	prepareResources(t, fixture, workspace, spec.Resources, sources)
	policy, err := fs.ReadFile(os.DirFS(fixture), "policy.yaml")
	if errors.Is(err, os.ErrNotExist) {
		policy, err = os.ReadFile("testdata/policy.yaml")
	}
	c.Assert(err, qt.IsNil)
	// #nosec G703 -- The output is a fixed filename in t.TempDir; fixture bytes never enter its path.
	c.Assert(os.WriteFile(filepath.Join(workspace, "policy.yaml"), policy, 0o600), qt.IsNil)
	sources["policy.yaml"] = policy
	args := []string{"check", "--config", "policy.yaml", "--report", "text:-", "--report", "json:result.json",
		"--report", "sarif:result.sarif"}
	args = append(args, spec.Args...)
	var input []byte
	switch {
	case spec.Stdin:
		c.Assert(spec.Files, qt.HasLen, 1)
		args = append(args, "--stdin", "--filename", spec.Files[0])
		input = sources[spec.Files[0]]
	case spec.Directory:
		args = append(args, ".")
	default:
		args = append(args, spec.Files...)
	}
	stdout, stderr, code := invoke(t, binary, workspace, args, input)
	c.Assert(code, qt.Equals, spec.ExitCode, qt.Commentf("%s\n%s", stdout, stderr))
	assertInputsUnchanged(t, workspace, sources)
	if spec.StartupError {
		_, err = os.Stat(filepath.Join(workspace, "result.json"))
		c.Assert(err, qt.ErrorIs, os.ErrNotExist)
	} else {
		var result unswell.RunResult
		decodeFile(t, filepath.Join(workspace, "result.json"), &result)
		c.Assert(matchWants(result.Findings, wants), qt.IsNil)
		verifyLocations(t, result, sources)
		verifySARIF(t, workspace, result)
		assertGoldenJSON(t, filepath.Join(fixture, "diagnostics.golden.json"), diagnostics(result))
	}
	assertGolden(t, filepath.Join(fixture, "stdout.golden"), []byte(normalizeOutput(stdout, workspace)))
	assertGolden(t, filepath.Join(fixture, "stderr.golden"), []byte(normalizeOutput(stderr, workspace)))
}

func invoke(t *testing.T, binary, workspace string, args []string, input []byte) (string, string, int) {
	t.Helper()
	c := qt.New(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	// #nosec G204 -- Run only the CLI compiled from this checkout with reviewed fixture arguments, without a shell.
	command := exec.CommandContext(ctx, binary, args...)
	command.Dir = workspace
	command.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	c.Assert(ctx.Err(), qt.IsNil)
	if err != nil {
		var exit *exec.ExitError
		c.Assert(err, qt.ErrorAs, &exit, qt.Commentf("process failure: %v", err))
	}
	return stdout.String(), stderr.String(), command.ProcessState.ExitCode()
}

func decodeFile(t *testing.T, path string, target any) {
	t.Helper()
	c := qt.New(t)
	data, err := fs.ReadFile(os.DirFS(filepath.Dir(path)), filepath.Base(path))
	c.Assert(err, qt.IsNil)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	c.Assert(decoder.Decode(target), qt.IsNil)
	var extra any
	c.Assert(decoder.Decode(&extra), qt.ErrorIs, io.EOF)
}

func normalizeOutput(value, workspace string) string {
	value = strings.ReplaceAll(value, workspace, "<workspace>")
	return strings.ReplaceAll(value, "\r\n", "\n")
}
