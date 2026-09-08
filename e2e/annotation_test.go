package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestAnnotationProtocol(t *testing.T) {
	c := qt.New(t)
	binary := buildAnnotationTool(t)
	data, err := os.ReadFile("../research/annotation/testdata/tutorial.json")
	c.Assert(err, qt.IsNil)
	t.Run("blinded packet", func(t *testing.T) {
		c := qt.New(t)
		output := annotationCommand(t, binary, "packet", data, 0)
		expected, err := os.ReadFile("annotationdata/packet.golden.json")
		c.Assert(err, qt.IsNil)
		c.Assert(output, qt.DeepEquals, expected)
	})
	t.Run("agreement", func(t *testing.T) {
		c := qt.New(t)
		output := annotationCommand(t, binary, "agreement", data, 0)
		var result map[string]json.RawMessage
		c.Assert(json.Unmarshal(output, &result), qt.IsNil)
		var groups []struct {
			Group   string          `json:"group"`
			Quality json.RawMessage `json:"quality"`
		}
		c.Assert(json.Unmarshal(result["groups"], &groups), qt.IsNil)
		c.Assert(groups[0].Group, qt.Equals, "all")
		projection := map[string]json.RawMessage{"basis": result["basis"], "purpose": result["purpose"],
			"packet_sha256":  result["packet_sha256"],
			"primary_raters": result["primary_raters"], "missing_answers": result["missing_answers"],
			"auxiliary_judgments": result["auxiliary_judgments"], "quality": groups[0].Quality}
		actual, err := json.MarshalIndent(projection, "", "  ")
		c.Assert(err, qt.IsNil)
		expected, err := os.ReadFile("annotationdata/agreement.golden.json")
		c.Assert(err, qt.IsNil)
		c.Assert(append(actual, '\n'), qt.DeepEquals, expected)
	})
	t.Run("duplicate key rejection", func(t *testing.T) {
		invalid := bytes.Replace(data, []byte(`"purpose": "tutorial"`), []byte(`"purpose": "tutorial", "purpose": "tutorial"`), 1)
		c := qt.New(t)
		c.Assert(bytes.Equal(invalid, data), qt.IsFalse)
		annotationCommand(t, binary, "validate", invalid, 2)
	})
	t.Run("simulated labels cannot count as corpus", func(t *testing.T) {
		invalid := bytes.Replace(data, []byte(`"purpose": "tutorial"`), []byte(`"purpose": "corpus"`), 1)
		annotationCommand(t, binary, "validate", invalid, 2)
	})
}

func buildAnnotationTool(t *testing.T) string {
	t.Helper()
	c := qt.New(t)
	name := "annotate"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	// #nosec G204 -- Compile the fixed local research command into a test-owned directory.
	command := exec.CommandContext(ctx, "go", "build", "-trimpath", "-o", binary, "./cmd/annotate")
	command.Dir = "../research/annotation"
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOWORK=off")
	output, err := command.CombinedOutput()
	c.Assert(err, qt.IsNil, qt.Commentf("annotation tool build: %s", output))
	return binary
}

func annotationCommand(t *testing.T, binary, name string, data []byte, exit int) []byte {
	t.Helper()
	c := qt.New(t)
	ctx, cancel := context.WithTimeout(t.Context(), time.Minute)
	defer cancel()
	// #nosec G204 -- The test selects the locally built command and a fixed operation.
	command := exec.CommandContext(ctx, binary, name)
	command.Stdin = bytes.NewReader(data)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if exit == 0 {
		c.Assert(err, qt.IsNil, qt.Commentf("stderr: %s", stderr.Bytes()))
	} else {
		c.Assert(err, qt.IsNotNil)
		c.Assert(stdout.Len(), qt.Equals, 0)
		c.Assert(stderr.Len() > 0, qt.IsTrue)
	}
	c.Assert(command.ProcessState, qt.IsNotNil)
	c.Assert(command.ProcessState.ExitCode(), qt.Equals, exit)
	return stdout.Bytes()
}
