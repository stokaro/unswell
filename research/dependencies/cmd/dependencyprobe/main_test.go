package main_test

import (
	"bytes"
	"debug/buildinfo"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestProbeReportsItsCompiledDependency(t *testing.T) {
	c := qt.New(t)
	name := "dependencyprobe"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	// #nosec G204 -- Build this package into a test-owned temporary directory, with fixed arguments and no shell.
	build := exec.CommandContext(t.Context(), "go", "build", "-trimpath", "-o", binary, ".")
	build.Env = append(os.Environ(), "CGO_ENABLED=0", "GOWORK=off")
	output, err := build.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("go build: %s", output))
	info, err := buildinfo.ReadFile(binary)
	c.Assert(err, qt.IsNil)
	dep := parserModule(t, info)
	c.Assert(dep.Version, qt.Not(qt.Equals), "")
	c.Assert(dep.Sum, qt.Not(qt.Equals), "")
	c.Assert(dep.Replace, qt.IsNil)

	// #nosec G204 -- Execute only the binary built by this test, with a fixed argument and no shell.
	command := exec.CommandContext(t.Context(), binary, "--build-info")
	command.Dir = t.TempDir() // No module manifests or model files at runtime.
	output, err = command.Output()
	c.Assert(err, qt.IsNil)
	var result struct {
		Schema   string                     `json:"schema"`
		Provider map[string]json.RawMessage `json:"provider"`
	}
	c.Assert(json.Unmarshal(output, &result), qt.IsNil)
	c.Assert(result.Schema, qt.Equals, "unswell-dependency-probe-build-v1")
	for field, want := range map[string]string{"path": dep.Path, "version": dep.Version, "sum": dep.Sum} {
		var value string
		c.Assert(json.Unmarshal(result.Provider[field], &value), qt.IsNil)
		c.Assert(value, qt.Equals, want)
	}
	c.Assert(string(result.Provider["revision"]), qt.Equals, "null")

	for _, args := range [][]string{{"--build-info", "--input", "missing"}, {"--build-info", "extra"}} {
		// #nosec G204 -- Execute the test-owned binary with the fixed cases above, without a shell.
		command := exec.CommandContext(t.Context(), binary, args...)
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		var exit *exec.ExitError
		c.Assert(err, qt.ErrorAs, &exit)
		c.Assert(exit.ExitCode(), qt.Equals, 2)
		c.Assert(output, qt.HasLen, 0)
		c.Assert(stderr.String(), qt.Contains, "cannot be combined")
	}
}

func parserModule(t *testing.T, info *debug.BuildInfo) *debug.Module {
	t.Helper()
	for _, dep := range info.Deps {
		if dep.Path == "github.com/bioshock/gospacy/v3" {
			return dep
		}
	}
	t.Fatal("compiled parser dependency missing")
	return nil
}
