package main

// White-box tests: Inject output failures and cancellation and probe bounded source readers;
// package main has no importable command API or process flag for these I/O faults.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/corpus"
)

func manifest(c *qt.C) []byte {
	c.Helper()
	data, err := os.ReadFile("../../corpus/testdata/ptah-manifest.json")
	c.Assert(err, qt.IsNil)
	return data
}

func TestCommandPlansExtractsAndVerifies(t *testing.T) {
	c := qt.New(t)
	var plan, artifact, verification bytes.Buffer
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(manifest(c)), &plan), qt.IsNil)
	root := "../../corpus/testdata/ptah"
	c.Assert(run(t.Context(), []string{"extract", "--root", root}, &plan, &artifact), qt.IsNil)
	candidates := bytes.Clone(artifact.Bytes())
	c.Assert(run(t.Context(), []string{"verify", "--root", root}, &artifact, &verification), qt.IsNil)
	var result corpus.Verification
	c.Assert(json.Unmarshal(verification.Bytes(), &result), qt.IsNil)
	c.Assert(result.Units, qt.Equals, 378)
	c.Assert(result.HumanCorpus, qt.Equals, "not_qualified")
	policy := filepath.Join(t.TempDir(), "policy.yaml")
	c.Assert(os.WriteFile(policy, []byte("version: 1\nextends: [builtin:custom]\nrules:\n"+
		"  policy.banned-phrases: {enabled: true, parameters: {phrases: [schema]}}\n"), 0o600), qt.IsNil)
	var measured bytes.Buffer
	c.Assert(run(t.Context(), []string{"measure", "--root", root, "--policy", policy},
		bytes.NewReader(candidates), &measured), qt.IsNil)
	var findings corpus.FindingsArtifact
	c.Assert(json.Unmarshal(measured.Bytes(), &findings), qt.IsNil)
	c.Assert(findings.Version, qt.Equals, corpus.FindingsVersion)
	c.Assert(findings.HumanCorpus, qt.Equals, "not_qualified")
	c.Assert(findings.Documents, qt.HasLen, 8)
	c.Assert(findings.Units, qt.HasLen, 378)
}

func TestCommandRejectsBadInputs(t *testing.T) {
	for _, args := range [][]string{nil, {"bad"}, {"plan", "extra"}, {"extract"}, {"plan", "--root", "."}, {"plan", "--bad"},
		{"join"}, {"join", "--root", "."}, {"join", "--root", ".", "--round", "round.json"},
		{"join", "--root", ".", "--feature", "prose-words"}, {"plan", "--round", "round.json"},
		{"verify", "--root", ".", "--feature", "prose-words"}, {"measure"}, {"measure", "--root", "."},
		{"verify", "--root", ".", "--policy", "policy.yaml"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			c.Assert(run(t.Context(), args, bytes.NewReader(manifest(c)), &output), qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

var errIO = errors.New("injected I/O failure")

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, errIO }

type badWriter struct{}

func (badWriter) Write([]byte) (int, error) { return 0, errIO }

type shortWriter struct{}

func (shortWriter) Write([]byte) (int, error) { return 0, nil }

func TestCommandIOAndCancellation(t *testing.T) {
	c := qt.New(t)
	data := manifest(c)
	c.Assert(run(t.Context(), []string{"plan"}, badReader{}, io.Discard), qt.ErrorIs, errIO)
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(data), badWriter{}), qt.ErrorIs, errIO)
	c.Assert(run(t.Context(), []string{"plan"}, bytes.NewReader(data), shortWriter{}), qt.ErrorIs, io.ErrShortWrite)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	c.Assert(run(ctx, []string{"plan"}, bytes.NewReader(data), io.Discard), qt.ErrorIs, context.Canceled)
}

func TestSourceRootRequiresRegularPinnedFiles(t *testing.T) {
	c := qt.New(t)
	m, err := corpus.LoadManifest(t.Context(), manifest(c))
	c.Assert(err, qt.IsNil)
	p, err := corpus.MakePlan(t.Context(), m)
	c.Assert(err, qt.IsNil)
	_, err = loadFiles(t.Context(), t.TempDir(), p)
	c.Assert(err, qt.IsNotNil)
	root, err := os.OpenRoot(t.TempDir())
	c.Assert(err, qt.IsNil)
	t.Cleanup(func() { c.Assert(root.Close(), qt.IsNil) })
	c.Assert(root.Mkdir("directory", 0o700), qt.IsNil)
	_, err = readFile(root, corpus.Notice{Path: "directory", Bytes: 1})
	c.Assert(err, qt.IsNotNil)
	c.Assert(root.WriteFile("source", []byte("short"), 0o600), qt.IsNil)
	_, err = readFile(root, corpus.Notice{Path: "source", Bytes: 100})
	c.Assert(err, qt.IsNotNil)
	_, err = readFile(root, corpus.Notice{Path: filepath.Join("absent", "source"), Bytes: 1})
	c.Assert(err, qt.IsNotNil)
}

type endlessInput struct{ bytes int }

func (r *endlessInput) Read(data []byte) (int, error) {
	for i := range data {
		data[i] = ' '
	}
	r.bytes += len(data)
	return len(data), nil
}

func TestManifestCommandsBoundReads(t *testing.T) {
	for _, args := range [][]string{{"plan"}, {"extract", "--root", "."}} {
		t.Run(args[0], func(t *testing.T) {
			c := qt.New(t)
			input := &endlessInput{}
			c.Assert(run(t.Context(), args, input, io.Discard), qt.IsNotNil)
			c.Assert(input.bytes, qt.Equals, corpus.MaxManifestBytes+1)
		})
	}
}
