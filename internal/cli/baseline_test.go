package cli_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/internal/cli"
)

func TestBaselineCommandsRejectUnsafeWritesAndIncompleteChecks(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	source := []byte("Certainly! The client retries.")
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), source, 0o600), qt.IsNil)
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	c.Assert(cli.Run(t.Context(), []string{"baseline", "create", "draft.md", "--output", "debt.json"}, environment), qt.Equals, 0)
	original, err := fs.ReadFile(os.DirFS(root), "debt.json")
	c.Assert(err, qt.IsNil)
	for _, args := range [][]string{
		{"baseline", "create", "draft.md", "--output", "debt.json"},
		{"baseline", "create", "draft.md", "--output", "draft.md"},
		{"baseline", "create", "--stdin", "--filename", "draft.md", "--output", "other.json"},
		{"baseline", "update", "--baseline", "debt.json", "--output", "other.json"},
		{"baseline", "check", "draft.md"},
		{"check", "draft.md", "--gate-mode", "new"},
		{"check", "draft.md", "--baseline", "missing.json", "--no-gate"},
		{"check", "draft.md", "--baseline", "debt.json", "--report", "json:debt.json"},
		{"baseline", "update", "--baseline", "debt.json", "missing.md"},
	} {
		c.Assert(cli.Run(t.Context(), args, environment), qt.Equals, 2, qt.Commentf("%v: %s", args, stderr.String()))
		actual, err := fs.ReadFile(os.DirFS(root), "debt.json")
		c.Assert(err, qt.IsNil)
		c.Assert(actual, qt.DeepEquals, original)
		actual, err = fs.ReadFile(os.DirFS(root), "draft.md")
		c.Assert(err, qt.IsNil)
		c.Assert(actual, qt.DeepEquals, source)
	}
	c.Assert(os.WriteFile(filepath.Join(root, "draft.md"), nil, 0o600), qt.IsNil)
	c.Assert(cli.Run(t.Context(), []string{"check", "draft.md", "--baseline", "debt.json", "--no-gate"}, environment), qt.Equals, 2)
	c.Assert(cli.Run(t.Context(), []string{"baseline", "update", "draft.md", "--baseline", "debt.json"}, environment), qt.Equals, 2)
	actual, err := fs.ReadFile(os.DirFS(root), "debt.json")
	c.Assert(err, qt.IsNil)
	c.Assert(actual, qt.DeepEquals, original)
}

func TestBaselineUpdateRequiresCoveredCompatibilityAndPreservesUnobservedDebt(t *testing.T) {
	c := qt.New(t)
	root := t.TempDir()
	for _, name := range []string{"first.md", "second.md"} {
		c.Assert(os.WriteFile(filepath.Join(root, name), []byte("Certainly! The client retries."), 0o600), qt.IsNil)
	}
	var out, stderr bytes.Buffer
	environment := cli.Environment{Dir: root, In: bytes.NewReader(nil), Out: &out, Err: &stderr}
	c.Assert(cli.Run(t.Context(), []string{"baseline", "create", "first.md", "second.md", "--output", "debt.json"}, environment), qt.Equals, 0)
	c.Assert(os.WriteFile(filepath.Join(root, "first.md"), []byte("The client retries."), 0o600), qt.IsNil)
	c.Assert(cli.Run(t.Context(), []string{"baseline", "update", "first.md", "--baseline", "debt.json"}, environment), qt.Equals, 0)
	c.Assert(cli.Run(t.Context(), []string{"baseline", "check", "second.md", "--baseline", "debt.json"}, environment), qt.Equals, 0)
	policy := []byte("version: 1\nrules:\n  scaffold.chat-preamble: {severity: warning, score: {weight: 70, cap: 70}}\n")
	c.Assert(os.WriteFile(filepath.Join(root, "policy.yaml"), policy, 0o600), qt.IsNil)
	args := []string{"baseline", "update", "first.md", "--baseline", "debt.json", "--config", "policy.yaml"}
	c.Assert(cli.Run(t.Context(), args, environment), qt.Equals, 2)
	c.Assert(stderr.String(), qt.Contains, "requires observing second.md")
	args = append(args, "second.md")
	c.Assert(cli.Run(t.Context(), args, environment), qt.Equals, 0, qt.Commentf("%s", stderr.String()))
}
