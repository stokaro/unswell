package llmdetcommand_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/stokaro/unswell/research/annotation/internal/llmdetcommand"
	"github.com/stokaro/unswell/research/annotation/llmdet"
)

const proxyPack = `{"version":"unswell-llmdet-proxy-v1","vocab_size":16,"rows":[]}`

func TestRunNumericEnsemble(t *testing.T) {
	c := qt.New(t)
	data, err := json.Marshal(llmdet.EnsembleSpec{
		Version: llmdet.EnsembleVersion, FeatureCount: 1, Classes: []string{"first", "second"},
		Trees: []llmdet.TreeSpec{{Comparison: "numeric-le", Splits: []llmdet.Split{
			{Threshold: 1, Left: -1, Right: -2},
		}, Leaves: []float64{2, -2}}},
	})
	c.Assert(err, qt.IsNil)
	args := []string{"--kind", "ensemble", "--model", writePack(t, string(data))}
	var output bytes.Buffer
	err = llmdetcommand.Run(t.Context(), args, strings.NewReader("[[0],[1],[2]]"), &output)
	c.Assert(err, qt.IsNil)
	var result struct {
		Results []llmdet.EnsembleResult `json:"results"`
	}
	c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
	c.Assert(result.Results, qt.HasLen, 3)
	c.Assert(result.Results[0].Raw, qt.DeepEquals, []float64{2, 0})
	c.Assert(result.Results[1].Raw, qt.DeepEquals, []float64{2, 0})
	c.Assert(result.Results[2].Raw, qt.DeepEquals, []float64{-2, 0})
	output.Reset()
	err = llmdetcommand.Run(t.Context(), args, strings.NewReader("[[0],[null]]"), &output)
	c.Assert(err, qt.IsNotNil)
	c.Assert(output.Len(), qt.Equals, 0)
}

func writePack(t *testing.T, data string) string {
	t.Helper()
	c := qt.New(t)
	path := filepath.Join(t.TempDir(), "pack.json")
	c.Assert(os.WriteFile(path, []byte(data), 0o600), qt.IsNil)
	return path
}

func TestRunRetainsAbsentFeaturesAndPackIdentity(t *testing.T) {
	c := qt.New(t)
	path := writePack(t, proxyPack)
	var output bytes.Buffer
	err := llmdetcommand.Run(t.Context(), []string{"--kind", "proxy", "--model", path},
		strings.NewReader(`[[1,2,3,4],[]]`), &output)
	c.Assert(err, qt.IsNil)
	var result struct {
		Version string               `json:"version"`
		Kind    string               `json:"kind"`
		Hash    string               `json:"model_sha256"`
		Results []llmdet.ProxyResult `json:"results"`
	}
	c.Assert(json.Unmarshal(output.Bytes(), &result), qt.IsNil)
	c.Assert(result.Version, qt.Equals, "unswell-llmdet-probe-v1")
	c.Assert(result.Kind, qt.Equals, "proxy")
	hash := sha256.Sum256([]byte(proxyPack))
	c.Assert(result.Hash, qt.Equals, hex.EncodeToString(hash[:]))
	c.Assert(result.Results, qt.HasLen, 2)
	c.Assert(result.Results[0].Reason, qt.Equals, "no_matching_context")
	c.Assert(result.Results[1].Reason, qt.Equals, "insufficient_tokens")
	c.Assert(bytes.Count(output.Bytes(), []byte(`"value": null`)), qt.Equals, 2)
}

func TestRunRejectsInvalidBatchesWithoutPartialOutput(t *testing.T) {
	path := writePack(t, proxyPack)
	for _, input := range []string{
		"", "null", "[]", "[null]", "[[null]]", "[[1.5]]", "[[1,2,3,4],[16]]",
		"[[1]] {}", "[[1e999]]", "[[\"1\"]]", "[" + strings.Repeat("[],", 1000) + "[]]",
	} {
		t.Run(input[:min(len(input), 30)], func(t *testing.T) {
			c := qt.New(t)
			var output bytes.Buffer
			err := llmdetcommand.Run(t.Context(), []string{"--kind", "proxy", "--model", path}, strings.NewReader(input), &output)
			c.Assert(err, qt.IsNotNil)
			c.Assert(output.Len(), qt.Equals, 0)
		})
	}
}

func TestRunRejectsInvalidOptionsAndPacks(t *testing.T) {
	c := qt.New(t)
	path := writePack(t, "{}")
	for _, args := range [][]string{
		nil, {"--unknown"}, {"--kind", "bad", "--model", path}, {"--kind", "proxy"},
		{"--kind", "proxy", "--model", path, "extra"},
		{"--kind", "ensemble", "--model", filepath.Join(t.TempDir(), "absent")},
		{"--kind", "ensemble", "--model", path}, {"--kind", "proxy", "--model", path},
	} {
		var output bytes.Buffer
		err := llmdetcommand.Run(t.Context(), args, strings.NewReader("[[0]]"), &output)
		c.Assert(err, qt.IsNotNil)
		c.Assert(output.Len(), qt.Equals, 0)
	}
}

type brokenIO struct{ err error }

func (b brokenIO) Read([]byte) (int, error)  { return 0, b.err }
func (b brokenIO) Write([]byte) (int, error) { return 0, b.err }

func TestRunPropagatesIOFailures(t *testing.T) {
	c := qt.New(t)
	args := []string{"--kind", "proxy", "--model", writePack(t, proxyPack)}
	failure := errors.New("injected I/O failure")
	err := llmdetcommand.Run(t.Context(), args, brokenIO{failure}, io.Discard)
	c.Assert(err, qt.ErrorIs, failure)
	err = llmdetcommand.Run(t.Context(), args, strings.NewReader("[[]]"), brokenIO{failure})
	c.Assert(err, qt.ErrorIs, failure)
	err = llmdetcommand.Run(t.Context(), args, strings.NewReader("[[]]"), brokenIO{})
	c.Assert(err, qt.ErrorIs, io.ErrShortWrite)
}

type blockedReader struct{ started, release chan struct{} }

func (b blockedReader) Read([]byte) (int, error) {
	close(b.started)
	<-b.release
	return 0, io.EOF
}

func TestRunReturnsWhenInputIsCanceled(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	input := blockedReader{make(chan struct{}), make(chan struct{})}
	t.Cleanup(func() { close(input.release) })
	args := []string{"--kind", "proxy", "--model", writePack(t, proxyPack)}
	done := make(chan error, 1)
	go func() { done <- llmdetcommand.Run(ctx, args, input, io.Discard) }()
	select {
	case <-input.started:
		cancel()
	case <-ctx.Done():
		c.Fatal("input was never read")
	}
	select {
	case err := <-done:
		c.Assert(err, qt.ErrorIs, context.Canceled)
	case <-time.After(10 * time.Second):
		c.Fatal("probe ignored cancellation")
	}
}

func TestBuiltProbeExitCodes(t *testing.T) {
	c := qt.New(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	binary := filepath.Join(t.TempDir(), "llmdetprobe.exe")
	// #nosec G204 -- Build the repository's fixed probe into a test-owned temporary directory.
	build := exec.CommandContext(ctx, "go", "build", "-o", binary, "../../cmd/llmdetprobe")
	output, err := build.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("%s", output))
	pack := writePack(t, proxyPack)
	for _, test := range []struct {
		name, input string
		code        int
	}{
		{"numeric feature absent", "[[]]", 0},
		{"invalid token", "[[16]]", 2},
		{"null is not zero", "[[null]]", 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := qt.New(t)
			// #nosec G204 -- Execute only the just-built probe with test-authored arguments and pack.
			command := exec.CommandContext(ctx, binary, "--kind", "proxy", "--model", pack)
			command.Stdin = strings.NewReader(test.input)
			var stderr bytes.Buffer
			command.Stderr = &stderr
			stdout, err := command.Output()
			if test.code == 0 {
				c.Assert(err, qt.IsNil)
				c.Assert(json.Valid(stdout), qt.IsTrue)
				c.Assert(stderr.Len(), qt.Equals, 0)
				return
			}
			var exit *exec.ExitError
			c.Assert(err, qt.ErrorAs, &exit)
			c.Assert(exit.ExitCode(), qt.Equals, test.code)
			c.Assert(stdout, qt.HasLen, 0)
			c.Assert(stderr.Len() > 0, qt.IsTrue)
		})
	}
}
