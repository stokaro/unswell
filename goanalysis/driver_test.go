package goanalysis_test

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestGoVetDriverRejectsProseAndAcceptsRevision(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	driver := filepath.Join(root, "unswell-vet")
	if runtime.GOOS == "windows" {
		driver += ".exe"
	}
	output, err := goCommand(t, "", "build", "-o", driver, "./cmd/unswell-vet")
	c.Assert(err, qt.IsNil, qt.Commentf("%s", output))
	c.Assert(os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/prosefixture\n"), 0o600), qt.IsNil)
	for _, row := range []struct{ name, text, location string }{
		{"comment draft", "package prosefixture\n// Certainly! The client retries.\n", "sample.go:2:4:"},
		{"string draft", "package prosefixture\nconst Message = \"Certainly! The server waits.\"\n", "sample.go:2:18:"},
		{"revision", "package prosefixture\n// The client retries.\nconst Message = \"The server waits.\"\n", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			c := qt.New(t)
			c.Assert(os.WriteFile(filepath.Join(root, "sample.go"), []byte(row.text), 0o600), qt.IsNil)
			output, err := goCommand(t, root, "vet", "-vettool="+driver, ".")
			if row.location != "" {
				var failure *exec.ExitError
				c.Assert(err, qt.ErrorAs, &failure, qt.Commentf("%s", output))
				c.Assert(failure.ExitCode(), qt.Equals, 1)
				c.Assert(strings.Count(output, "scaffold.chat-preamble"), qt.Equals, 1, qt.Commentf("%s", output))
				c.Assert(output, qt.Contains, row.location)
			} else {
				c.Assert(err, qt.IsNil, qt.Commentf("%s", output))
			}
			data, err := fs.ReadFile(os.DirFS(root), "sample.go")
			c.Assert(err, qt.IsNil)
			c.Assert(string(data), qt.Equals, row.text)
		})
	}
}

func goCommand(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	// #nosec G204 -- Fixed Go build/vet arguments and test-owned paths use argv without a shell.
	command := exec.CommandContext(t.Context(), "go", args...)
	command.Dir = dir
	command.Env = append(os.Environ(), "GOWORK=off", "GOPROXY=off", "CGO_ENABLED=0")
	output, err := command.CombinedOutput()
	return string(output), err
}
